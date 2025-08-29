package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// AuditLogger handles authentication audit logging
type AuditLogger struct {
	db *sql.DB
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{
		db: db,
	}
}

// AuditEvent represents an authentication audit event
type AuditEvent struct {
	OldValues      map[string]interface{}
	NewValues      map[string]interface{}
	Metadata       map[string]interface{}
	OrganizationID string
	Action         string
	ResourceType   string
	ResourceID     string
	ActorID        string
	ErrorMessage   string
	ActorIP        net.IP
	Success        bool
}

// Authentication audit actions
const (
	ActionUserLogin          = "user.login"
	ActionUserLoginFailed    = "user.login.failed"
	ActionUserLogout         = "user.logout"
	ActionTokenRefresh       = "user.token.refresh"
	ActionTokenInvalidate    = "user.token.invalidate"
	ActionUserCreated        = "user.created"
	ActionUserUpdated        = "user.updated"
	ActionUserDeleted        = "user.deleted"
	ActionAPIKeyCreated      = "api_key.created"
	ActionAPIKeyRevoked      = "api_key.revoked"
	ActionPasswordChanged    = "user.password.changed"
	ActionAccountLocked      = "user.account.locked"
	ActionAccountUnlocked    = "user.account.unlocked"
	ActionSuspiciousActivity = "security.suspicious_activity"
)

// LogEvent logs an audit event to the database
func (a *AuditLogger) LogEvent(event *AuditEvent) error {
	query := `
		INSERT INTO audit_logs (
			organization_id, action, resource_type, resource_id,
			actor_id, actor_ip, old_values, new_values, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`

	var oldValuesJSON, newValuesJSON, metadataJSON []byte
	var err error

	if event.OldValues != nil {
		// Clean old values
		cleanedOldValues := make(map[string]interface{})
		for k, v := range event.OldValues {
			if str, ok := v.(string); ok {
				cleanedOldValues[k] = strings.ReplaceAll(str, "\x00", "")
			} else {
				cleanedOldValues[k] = v
			}
		}
		oldValuesJSON, err = json.Marshal(cleanedOldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old values: %w", err)
		}
	}

	if event.NewValues != nil {
		// Clean new values
		cleanedNewValues := make(map[string]interface{})
		for k, v := range event.NewValues {
			if str, ok := v.(string); ok {
				cleanedNewValues[k] = strings.ReplaceAll(str, "\x00", "")
			} else {
				cleanedNewValues[k] = v
			}
		}
		newValuesJSON, err = json.Marshal(cleanedNewValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new values: %w", err)
		}
	}

	// Add success status and error to metadata
	metadata := make(map[string]interface{})
	if event.Metadata != nil {
		// Copy metadata, cleaning any problematic values
		for k, v := range event.Metadata {
			// Ensure string values are UTF-8 safe
			if str, ok := v.(string); ok {
				// Replace any null bytes or invalid UTF-8
				metadata[k] = strings.ReplaceAll(str, "\x00", "")
			} else {
				metadata[k] = v
			}
		}
	}
	metadata["success"] = event.Success
	if event.ErrorMessage != "" {
		// Clean error message as well
		metadata["error"] = strings.ReplaceAll(event.ErrorMessage, "\x00", "")
	}
	metadata["timestamp"] = time.Now().Unix()

	metadataJSON, err = json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Convert net.IP to string for database storage
	var actorIP interface{}
	if event.ActorIP != nil {
		actorIP = event.ActorIP.String()
	}

	// Convert JSON byte slices to strings for PostgreSQL JSONB columns
	var oldValuesStr, newValuesStr, metadataStr interface{}

	if oldValuesJSON != nil {
		oldValuesStr = string(oldValuesJSON)
	}
	if newValuesJSON != nil {
		newValuesStr = string(newValuesJSON)
	}
	if metadataJSON != nil {
		metadataStr = string(metadataJSON)
	}

	_, err = a.db.Exec(
		query,
		event.OrganizationID,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		event.ActorID,
		actorIP,
		oldValuesStr,
		newValuesStr,
		metadataStr,
	)

	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return nil
}

// LogLogin logs a successful login event
func (a *AuditLogger) LogLogin(user *types.User, clientIP net.IP, userAgent string) error {
	return a.LogEvent(&AuditEvent{
		OrganizationID: user.OrganizationID,
		Action:         ActionUserLogin,
		ResourceType:   "user",
		ResourceID:     user.ID,
		ActorID:        user.ID,
		ActorIP:        clientIP,
		Success:        true,
		Metadata: map[string]interface{}{
			"user_agent": userAgent,
			"email":      user.Email,
			"role":       user.Role,
		},
	})
}

// LogLoginFailed logs a failed login attempt
func (a *AuditLogger) LogLoginFailed(email string, organizationID string, clientIP net.IP, userAgent string, reason string) error {
	return a.LogEvent(&AuditEvent{
		OrganizationID: organizationID,
		Action:         ActionUserLoginFailed,
		ResourceType:   "user",
		ResourceID:     "", // No user ID for failed attempts
		ActorID:        email,
		ActorIP:        clientIP,
		Success:        false,
		ErrorMessage:   reason,
		Metadata: map[string]interface{}{
			"user_agent": userAgent,
			"email":      email,
			"reason":     reason,
		},
	})
}

// LogLogout logs a user logout event
func (a *AuditLogger) LogLogout(userID, organizationID string, clientIP net.IP, voluntary bool) error {
	return a.LogEvent(&AuditEvent{
		OrganizationID: organizationID,
		Action:         ActionUserLogout,
		ResourceType:   "user",
		ResourceID:     userID,
		ActorID:        userID,
		ActorIP:        clientIP,
		Success:        true,
		Metadata: map[string]interface{}{
			"voluntary": voluntary, // true for user-initiated, false for forced/expired
		},
	})
}

// LogTokenRefresh logs a token refresh event
func (a *AuditLogger) LogTokenRefresh(userID, organizationID string, clientIP net.IP, success bool, errorMsg string) error {
	return a.LogEvent(&AuditEvent{
		OrganizationID: organizationID,
		Action:         ActionTokenRefresh,
		ResourceType:   "token",
		ResourceID:     userID,
		ActorID:        userID,
		ActorIP:        clientIP,
		Success:        success,
		ErrorMessage:   errorMsg,
	})
}

// LogUserCreated logs user creation event
func (a *AuditLogger) LogUserCreated(user *types.User, creatorID string, clientIP net.IP) error {
	return a.LogEvent(&AuditEvent{
		OrganizationID: user.OrganizationID,
		Action:         ActionUserCreated,
		ResourceType:   "user",
		ResourceID:     user.ID,
		ActorID:        creatorID,
		ActorIP:        clientIP,
		Success:        true,
		NewValues: map[string]interface{}{
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
		Metadata: map[string]interface{}{
			"created_by": creatorID,
		},
	})
}

// LogSuspiciousActivity logs suspicious authentication activity
func (a *AuditLogger) LogSuspiciousActivity(organizationID, actorID string, clientIP net.IP, activity string, details map[string]interface{}) error {
	return a.LogEvent(&AuditEvent{
		OrganizationID: organizationID,
		Action:         ActionSuspiciousActivity,
		ResourceType:   "security",
		ResourceID:     actorID,
		ActorID:        actorID,
		ActorIP:        clientIP,
		Success:        false,
		ErrorMessage:   activity,
		Metadata:       details,
	})
}

// LoginAttemptTracker tracks failed login attempts for rate limiting and security
type LoginAttemptTracker struct {
	db *sql.DB
}

// NewLoginAttemptTracker creates a new login attempt tracker
func NewLoginAttemptTracker(db *sql.DB) *LoginAttemptTracker {
	return &LoginAttemptTracker{
		db: db,
	}
}

// LoginAttempt represents a login attempt record
type LoginAttempt struct {
	CreatedAt time.Time
	ID        string
	Email     string
	ClientIP  net.IP
	Success   bool
}

// RecordLoginAttempt records a login attempt (success or failure)
func (t *LoginAttemptTracker) RecordLoginAttempt(email string, clientIP net.IP, success bool) error {
	// For now, we'll use the audit_logs table to track attempts
	// In a full implementation, you might want a dedicated login_attempts table
	action := ActionUserLogin
	if !success {
		action = ActionUserLoginFailed
	}

	query := `
		INSERT INTO audit_logs (
			organization_id, action, resource_type, actor_id, actor_ip,
			metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	metadata := map[string]interface{}{
		"success":   success,
		"email":     strings.ReplaceAll(email, "\x00", ""), // Clean email
		"client_ip": clientIP.String(),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Use default organization for tracking attempts
	defaultOrgID := "00000000-0000-0000-0000-000000000000"

	// Convert net.IP to string for database storage
	var clientIPStr interface{}
	if clientIP != nil {
		clientIPStr = clientIP.String()
	}

	// Convert JSON byte slice to string for PostgreSQL JSONB column
	var metadataStr interface{}
	if metadataJSON != nil {
		metadataStr = string(metadataJSON)
	}

	_, err = t.db.Exec(
		query,
		defaultOrgID,
		action,
		"login_attempt",
		email,
		clientIPStr,
		metadataStr,
	)

	return err
}

// GetRecentFailedAttempts gets recent failed login attempts for an email or IP
func (t *LoginAttemptTracker) GetRecentFailedAttempts(email string, clientIP net.IP, since time.Duration) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM audit_logs
		WHERE action = $1
		AND (actor_id = $2 OR actor_ip = $3)
		AND created_at > $4
	`

	// Convert net.IP to string for database query
	var clientIPStr interface{}
	if clientIP != nil {
		clientIPStr = clientIP.String()
	}

	var count int
	err := t.db.QueryRow(
		query,
		ActionUserLoginFailed,
		email,
		clientIPStr,
		time.Now().Add(-since),
	).Scan(&count)

	return count, err
}

// IsRateLimited checks if login attempts should be rate limited
func (t *LoginAttemptTracker) IsRateLimited(email string, clientIP net.IP) (bool, time.Duration, error) {
	// Check for failed attempts in the last hour
	count, err := t.GetRecentFailedAttempts(email, clientIP, time.Hour)
	if err != nil {
		return false, 0, err
	}

	// Rate limiting rules:
	// - 5 failed attempts in 1 hour = 15 minute lockout
	// - 10 failed attempts in 1 hour = 1 hour lockout
	// - 20+ failed attempts in 1 hour = 24 hour lockout

	if count >= 20 {
		return true, 24 * time.Hour, nil
	} else if count >= 10 {
		return true, time.Hour, nil
	} else if count >= 5 {
		return true, 15 * time.Minute, nil
	}

	return false, 0, nil
}
