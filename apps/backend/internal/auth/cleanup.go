package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TokenCleanupService handles background cleanup of expired tokens and audit logs
type TokenCleanupService struct {
	db          *sql.DB
	jwtManager  *JWTManager
	auditLogger *AuditLogger
	stopChan    chan struct{}
	config      *CleanupConfig
}

// CleanupConfig holds configuration for the cleanup service
type CleanupConfig struct {
	// How often to run cleanup (default: 1 hour)
	CleanupInterval time.Duration
	
	// How old audit logs should be before cleanup (default: 30 days)
	AuditLogRetentionPeriod time.Duration
	
	// How old login attempt logs should be before cleanup (default: 7 days)
	LoginAttemptRetentionPeriod time.Duration
	
	// Maximum number of records to delete in each cleanup batch (default: 1000)
	BatchSize int
}

// DefaultCleanupConfig returns default cleanup configuration
func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		CleanupInterval:             time.Hour,
		AuditLogRetentionPeriod:     30 * 24 * time.Hour, // 30 days
		LoginAttemptRetentionPeriod: 7 * 24 * time.Hour,  // 7 days
		BatchSize:                   1000,
	}
}

// NewTokenCleanupService creates a new token cleanup service
func NewTokenCleanupService(db *sql.DB, jwtManager *JWTManager, auditLogger *AuditLogger, config *CleanupConfig) *TokenCleanupService {
	if config == nil {
		config = DefaultCleanupConfig()
	}

	return &TokenCleanupService{
		db:          db,
		jwtManager:  jwtManager,
		auditLogger: auditLogger,
		stopChan:    make(chan struct{}),
		config:      config,
	}
}

// Start begins the background cleanup process
func (c *TokenCleanupService) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	// Run initial cleanup
	c.runCleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Token cleanup service shutting down...")
			return ctx.Err()
		case <-c.stopChan:
			fmt.Println("Token cleanup service stopped")
			return nil
		case <-ticker.C:
			c.runCleanup(ctx)
		}
	}
}

// Stop stops the background cleanup process
func (c *TokenCleanupService) Stop() {
	close(c.stopChan)
}

// runCleanup performs all cleanup operations
func (c *TokenCleanupService) runCleanup(ctx context.Context) {
	fmt.Printf("Starting token cleanup at %s\n", time.Now().Format(time.RFC3339))

	// Clean up expired JWT tokens from blacklist
	c.jwtManager.CleanupExpiredTokens()

	// Clean up old audit logs
	if err := c.cleanupAuditLogs(ctx); err != nil {
		fmt.Printf("Error cleaning up audit logs: %v\n", err)
	}

	// Clean up old login attempt logs
	if err := c.cleanupLoginAttempts(ctx); err != nil {
		fmt.Printf("Error cleaning up login attempts: %v\n", err)
	}

	// Log cleanup completion
	c.auditLogger.LogEvent(&AuditEvent{
		OrganizationID: "00000000-0000-0000-0000-000000000000", // system
		Action:         "system.cleanup.completed",
		ResourceType:   "system",
		ResourceID:     "token_cleanup",
		ActorID:        "system",
		Success:        true,
		Metadata: map[string]interface{}{
			"cleanup_time": time.Now().Unix(),
		},
	})

	fmt.Printf("Completed token cleanup at %s\n", time.Now().Format(time.RFC3339))
}

// cleanupAuditLogs removes old audit logs based on retention policy
func (c *TokenCleanupService) cleanupAuditLogs(ctx context.Context) error {
	cutoffTime := time.Now().Add(-c.config.AuditLogRetentionPeriod)
	
	// Count records to be deleted
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE created_at < $1`
	var count int
	err := c.db.QueryRowContext(ctx, countQuery, cutoffTime).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count old audit logs: %w", err)
	}

	if count == 0 {
		return nil
	}

	fmt.Printf("Cleaning up %d old audit logs older than %s\n", count, cutoffTime.Format(time.RFC3339))

	// Delete in batches to avoid long-running transactions
	deleteQuery := `
		DELETE FROM audit_logs 
		WHERE id IN (
			SELECT id FROM audit_logs 
			WHERE created_at < $1 
			ORDER BY created_at 
			LIMIT $2
		)
	`

	totalDeleted := 0
	for totalDeleted < count {
		result, err := c.db.ExecContext(ctx, deleteQuery, cutoffTime, c.config.BatchSize)
		if err != nil {
			return fmt.Errorf("failed to delete audit logs batch: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += int(rowsAffected)

		if rowsAffected == 0 {
			break // No more rows to delete
		}

		// Small delay between batches to reduce DB load
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Successfully deleted %d old audit log records\n", totalDeleted)
	return nil
}

// cleanupLoginAttempts removes old login attempt logs
func (c *TokenCleanupService) cleanupLoginAttempts(ctx context.Context) error {
	cutoffTime := time.Now().Add(-c.config.LoginAttemptRetentionPeriod)
	
	// Clean up old login attempt records from audit logs
	countQuery := `
		SELECT COUNT(*) FROM audit_logs 
		WHERE action IN ($1, $2) AND created_at < $3
	`
	
	var count int
	err := c.db.QueryRowContext(
		ctx, 
		countQuery, 
		ActionUserLogin, 
		ActionUserLoginFailed, 
		cutoffTime,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count old login attempts: %w", err)
	}

	if count == 0 {
		return nil
	}

	fmt.Printf("Cleaning up %d old login attempt logs older than %s\n", count, cutoffTime.Format(time.RFC3339))

	// Delete in batches
	deleteQuery := `
		DELETE FROM audit_logs 
		WHERE id IN (
			SELECT id FROM audit_logs 
			WHERE action IN ($1, $2) AND created_at < $3 
			ORDER BY created_at 
			LIMIT $4
		)
	`

	totalDeleted := 0
	for totalDeleted < count {
		result, err := c.db.ExecContext(
			ctx, 
			deleteQuery, 
			ActionUserLogin, 
			ActionUserLoginFailed, 
			cutoffTime, 
			c.config.BatchSize,
		)
		if err != nil {
			return fmt.Errorf("failed to delete login attempts batch: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += int(rowsAffected)

		if rowsAffected == 0 {
			break // No more rows to delete
		}

		// Small delay between batches
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Successfully deleted %d old login attempt records\n", totalDeleted)
	return nil
}

// GetCleanupStats returns statistics about cleanup operations
func (c *TokenCleanupService) GetCleanupStats(ctx context.Context) (*CleanupStats, error) {
	stats := &CleanupStats{}

	// Count total audit logs
	err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&stats.TotalAuditLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Count old audit logs
	cutoffTime := time.Now().Add(-c.config.AuditLogRetentionPeriod)
	err = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_logs WHERE created_at < $1", cutoffTime).Scan(&stats.OldAuditLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to count old audit logs: %w", err)
	}

	// Count login attempts
	err = c.db.QueryRowContext(
		ctx, 
		"SELECT COUNT(*) FROM audit_logs WHERE action IN ($1, $2)", 
		ActionUserLogin, 
		ActionUserLoginFailed,
	).Scan(&stats.TotalLoginAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed to count login attempts: %w", err)
	}

	// Count old login attempts
	loginCutoffTime := time.Now().Add(-c.config.LoginAttemptRetentionPeriod)
	err = c.db.QueryRowContext(
		ctx, 
		"SELECT COUNT(*) FROM audit_logs WHERE action IN ($1, $2) AND created_at < $3", 
		ActionUserLogin, 
		ActionUserLoginFailed,
		loginCutoffTime,
	).Scan(&stats.OldLoginAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed to count old login attempts: %w", err)
	}

	return stats, nil
}

// CleanupStats holds statistics about cleanup operations
type CleanupStats struct {
	TotalAuditLogs     int `json:"total_audit_logs"`
	OldAuditLogs       int `json:"old_audit_logs"`
	TotalLoginAttempts int `json:"total_login_attempts"`
	OldLoginAttempts   int `json:"old_login_attempts"`
}

// ForceCleanup performs an immediate cleanup operation
func (c *TokenCleanupService) ForceCleanup(ctx context.Context) error {
	fmt.Println("Performing forced cleanup...")
	c.runCleanup(ctx)
	return nil
}