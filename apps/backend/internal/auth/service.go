package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication and user management
type Service struct {
	db             *sql.DB
	jwtManager     *JWTManager
	config         *Config
	auditLogger    *AuditLogger
	attemptTracker *LoginAttemptTracker
}

// Config holds authentication service configuration
type Config struct {
	JWTSecret          string
	Cache              CacheConfig
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	BCryptCost         int
}

// NewService creates a new authentication service
func NewService(db *sql.DB, config *Config) *Service {
	// Create token cache
	cache, err := NewTokenCache(config.Cache)
	if err != nil {
		// Fallback to memory cache if Redis fails
		cache = NewMemoryTokenCache()
	}

	jwtManager := NewJWTManagerWithCache(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry, cache)
	auditLogger := NewAuditLogger(db)
	attemptTracker := NewLoginAttemptTracker(db)

	return &Service{
		db:             db,
		jwtManager:     jwtManager,
		config:         config,
		auditLogger:    auditLogger,
		attemptTracker: attemptTracker,
	}
}

// GetJWTManager returns the JWT manager instance
func (s *Service) GetJWTManager() *JWTManager {
	return s.jwtManager
}

// GetAuditLogger returns the audit logger instance
func (s *Service) GetAuditLogger() *AuditLogger {
	return s.auditLogger
}

// GetAttemptTracker returns the login attempt tracker instance
func (s *Service) GetAttemptTracker() *LoginAttemptTracker {
	return s.attemptTracker
}

// LoginContext contains additional context for login attempts
type LoginContext struct {
	UserAgent string
	ClientIP  net.IP
}

// Login authenticates a user with email and password
func (s *Service) Login(email, password string) (*types.LoginResponse, error) {
	return s.LoginWithContext(email, password, nil)
}

// LoginWithContext authenticates a user with email and password including security context
func (s *Service) LoginWithContext(email, password string, ctx *LoginContext) (*types.LoginResponse, error) {
	// Default context if none provided
	if ctx == nil {
		ctx = &LoginContext{
			ClientIP:  net.IPv4(127, 0, 0, 1),
			UserAgent: "unknown",
		}
	}

	// Check if login attempts are rate limited
	isRateLimited, lockoutDuration, err := s.attemptTracker.IsRateLimited(email, ctx.ClientIP)
	if err != nil {
		// Log the error but don't fail the login attempt
		fmt.Printf("Warning: failed to check rate limiting: %v\n", err)
	}

	if isRateLimited {
		// Log suspicious activity
		s.auditLogger.LogSuspiciousActivity(
			"00000000-0000-0000-0000-000000000000", // Default org for rate limiting
			email,
			ctx.ClientIP,
			"rate_limited_login_attempt",
			map[string]interface{}{
				"lockout_duration_seconds": int(lockoutDuration.Seconds()),
				"user_agent":               ctx.UserAgent,
			},
		)

		return nil, types.NewRateLimitExceededError(fmt.Sprintf("too many failed attempts, try again in %v", lockoutDuration))
	}

	// Get user by email
	user, err := s.GetUserByEmail(email)
	if err != nil {
		// Record failed attempt
		s.attemptTracker.RecordLoginAttempt(email, ctx.ClientIP, false)

		// Log failed login attempt
		s.auditLogger.LogLoginFailed(
			email,
			"00000000-0000-0000-0000-000000000000", // Default org
			ctx.ClientIP,
			ctx.UserAgent,
			"user_not_found",
		)

		return nil, types.NewUnauthorizedError("invalid credentials")
	}

	// Check if user account is active
	if !user.IsActive {
		// Record failed attempt
		s.attemptTracker.RecordLoginAttempt(email, ctx.ClientIP, false)

		// Log failed login attempt
		s.auditLogger.LogLoginFailed(
			email,
			user.OrganizationID,
			ctx.ClientIP,
			ctx.UserAgent,
			"account_inactive",
		)

		return nil, types.NewUnauthorizedError("account is inactive")
	}

	// Validate password
	if !s.validatePassword(password, user.PasswordHash) {
		// Record failed attempt
		s.attemptTracker.RecordLoginAttempt(email, ctx.ClientIP, false)

		// Log failed login attempt
		s.auditLogger.LogLoginFailed(
			email,
			user.OrganizationID,
			ctx.ClientIP,
			ctx.UserAgent,
			"invalid_password",
		)

		return nil, types.NewUnauthorizedError("invalid credentials")
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Record successful login attempt
	s.attemptTracker.RecordLoginAttempt(email, ctx.ClientIP, true)

	// Log successful login
	err = s.auditLogger.LogLogin(user, ctx.ClientIP, ctx.UserAgent)
	if err != nil {
		// Log the error but don't fail the login
		fmt.Printf("Warning: failed to log successful login: %v\n", err)
	}

	// Create response (don't include password hash)
	response := &types.LoginResponse{
		User: &types.User{
			ID:             user.ID,
			Email:          user.Email,
			Name:           user.Name,
			OrganizationID: user.OrganizationID,
			Role:           user.Role,
			IsActive:       user.IsActive,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}

	return response, nil
}

// RefreshTokenResponse extends TokenResponse with new refresh token
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshToken generates new access token from refresh token
func (s *Service) RefreshToken(refreshToken string) (*types.LoginResponse, error) {
	return s.RefreshTokenWithContext(refreshToken, nil)
}

// RefreshTokenWithContext generates new access token with security context and optional rotation
func (s *Service) RefreshTokenWithContext(refreshToken string, ctx *LoginContext) (*types.LoginResponse, error) {
	// Default context if none provided
	if ctx == nil {
		ctx = &LoginContext{
			ClientIP:  net.IPv4(127, 0, 0, 1),
			UserAgent: "unknown",
		}
	}

	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		// Log failed token refresh attempt
		s.auditLogger.LogTokenRefresh(
			"unknown",                              // userID not available due to invalid token
			"00000000-0000-0000-0000-000000000000", // default org
			ctx.ClientIP,
			false,
			"invalid_refresh_token",
		)
		return nil, types.NewUnauthorizedError("invalid refresh token")
	}

	// Ensure it's a refresh token
	if claims.TokenType != "refresh" {
		// Log suspicious activity - using access token as refresh token
		s.auditLogger.LogSuspiciousActivity(
			claims.OrganizationID,
			claims.UserID,
			ctx.ClientIP,
			"access_token_used_as_refresh",
			map[string]interface{}{
				"token_type": claims.TokenType,
				"user_agent": ctx.UserAgent,
			},
		)
		return nil, types.NewUnauthorizedError("invalid token type")
	}

	// Get current user data
	user, err := s.GetUserByID(claims.UserID)
	if err != nil {
		// Log failed token refresh attempt
		s.auditLogger.LogTokenRefresh(
			claims.UserID,
			claims.OrganizationID,
			ctx.ClientIP,
			false,
			"user_not_found",
		)
		return nil, types.NewUnauthorizedError("user not found")
	}

	// Check if user account is still active
	if !user.IsActive {
		// Log failed token refresh attempt
		s.auditLogger.LogTokenRefresh(
			user.ID,
			user.OrganizationID,
			ctx.ClientIP,
			false,
			"user_inactive",
		)
		return nil, types.NewUnauthorizedError("user account is inactive")
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		// Log failed token refresh attempt
		s.auditLogger.LogTokenRefresh(
			user.ID,
			user.OrganizationID,
			ctx.ClientIP,
			false,
			"failed_to_generate_access_token",
		)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Log successful token refresh
	s.auditLogger.LogTokenRefresh(
		user.ID,
		user.OrganizationID,
		ctx.ClientIP,
		true,
		"",
	)

	// Generate new refresh token to maintain security
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	response := &types.LoginResponse{
		User: &types.User{
			ID:             user.ID,
			Email:          user.Email,
			Name:           user.Name,
			OrganizationID: user.OrganizationID,
			Role:           user.Role,
			IsActive:       user.IsActive,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}

	return response, nil
}

// RefreshTokenWithRotation generates new access and refresh tokens, invalidating the old refresh token
func (s *Service) RefreshTokenWithRotation(refreshToken string, ctx *LoginContext) (*RefreshTokenResponse, error) {
	// Default context if none provided
	if ctx == nil {
		ctx = &LoginContext{
			ClientIP:  net.IPv4(127, 0, 0, 1),
			UserAgent: "unknown",
		}
	}

	// Validate current refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, types.NewUnauthorizedError("invalid refresh token")
	}

	if claims.TokenType != "refresh" {
		return nil, types.NewUnauthorizedError("invalid token type")
	}

	// Get current user data
	user, err := s.GetUserByID(claims.UserID)
	if err != nil || !user.IsActive {
		return nil, types.NewUnauthorizedError("user not found or inactive")
	}

	// Invalidate the old refresh token first
	err = s.jwtManager.InvalidateToken(context.Background(), refreshToken)
	if err != nil {
		// Log but don't fail - continue with new token generation
		fmt.Printf("Warning: failed to invalidate old refresh token: %v\n", err)
	}

	// Generate new tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log token rotation event
	s.auditLogger.LogEvent(&AuditEvent{
		OrganizationID: user.OrganizationID,
		Action:         "user.token.rotated",
		ResourceType:   "token",
		ResourceID:     user.ID,
		ActorID:        user.ID,
		ActorIP:        ctx.ClientIP,
		Success:        true,
		Metadata: map[string]interface{}{
			"user_agent": ctx.UserAgent,
			"rotation":   true,
		},
	})

	response := &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}

	return response, nil
}

// Logout invalidates user tokens
func (s *Service) Logout(accessToken string) error {
	return s.LogoutWithContext(accessToken, nil, true)
}

// LogoutWithContext invalidates user tokens with security context
func (s *Service) LogoutWithContext(accessToken string, ctx *LoginContext, voluntary bool) error {
	// Default context if none provided
	if ctx == nil {
		ctx = &LoginContext{
			ClientIP:  net.IPv4(127, 0, 0, 1),
			UserAgent: "unknown",
		}
	}

	// Validate the token first to ensure it's legitimate
	claims, err := s.jwtManager.ValidateToken(accessToken)
	if err != nil {
		// Token is already invalid, but still log the logout attempt
		s.auditLogger.LogEvent(&AuditEvent{
			OrganizationID: "00000000-0000-0000-0000-000000000000", // default
			Action:         ActionUserLogout,
			ResourceType:   "token",
			ResourceID:     "unknown",
			ActorID:        "unknown",
			ActorIP:        ctx.ClientIP,
			Success:        true,
			Metadata: map[string]interface{}{
				"voluntary":    voluntary,
				"token_status": "already_invalid",
				"user_agent":   ctx.UserAgent,
			},
		})
		return nil
	}

	// Ensure it's an access token
	if claims.TokenType != "access" {
		// Log suspicious activity - trying to logout with refresh token
		s.auditLogger.LogSuspiciousActivity(
			claims.OrganizationID,
			claims.UserID,
			ctx.ClientIP,
			"logout_with_refresh_token",
			map[string]interface{}{
				"token_type": claims.TokenType,
				"user_agent": ctx.UserAgent,
			},
		)
		return types.NewValidationError("invalid token type")
	}

	// Add to blacklist
	err = s.jwtManager.InvalidateToken(context.Background(), accessToken)
	if err != nil {
		// Log failed logout attempt
		s.auditLogger.LogEvent(&AuditEvent{
			OrganizationID: claims.OrganizationID,
			Action:         ActionUserLogout,
			ResourceType:   "token",
			ResourceID:     claims.UserID,
			ActorID:        claims.UserID,
			ActorIP:        ctx.ClientIP,
			Success:        false,
			ErrorMessage:   "failed to invalidate token",
			Metadata: map[string]interface{}{
				"voluntary":  voluntary,
				"user_agent": ctx.UserAgent,
			},
		})
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	// Log successful logout
	err = s.auditLogger.LogLogout(claims.UserID, claims.OrganizationID, ctx.ClientIP, voluntary)
	if err != nil {
		// Log the error but don't fail the logout
		fmt.Printf("Warning: failed to log logout: %v\n", err)
	}

	return nil
}

// GetUserByID retrieves user by ID
func (s *Service) GetUserByID(userID string) (*types.User, error) {
	query := `
		SELECT id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var user types.User
	err := s.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.OrganizationID,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves user by email
func (s *Service) GetUserByEmail(email string) (*types.User, error) {
	query := `
		SELECT id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	var user types.User
	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.OrganizationID,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(req *types.CreateUserRequest) (*types.User, error) {
	// Hash the password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Insert user into database
	query := `
		INSERT INTO users (email, name, password_hash, organization_id, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
		RETURNING id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at
	`

	var user types.User
	err = s.db.QueryRow(
		query,
		req.Email,
		req.Name,
		hashedPassword,
		req.OrganizationID,
		req.Role,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.OrganizationID,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(userID string, req *types.UpdateUserRequest) (*types.User, error) {
	// Build dynamic update query based on provided fields
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.Role != "" {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, req.Role)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil, types.NewValidationError("no fields to update")
	}

	// Add updated_at and user ID
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d AND is_active = true
		RETURNING id, email, name, password_hash, organization_id, role, is_active, created_at, updated_at
	`,
		strings.Join(setClauses, ", "),
		argIndex,
	)

	var user types.User
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.OrganizationID,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// DeleteUser soft deletes a user
func (s *Service) DeleteUser(userID string) error {
	// TODO: Implement user deletion (soft delete)
	return nil
}

// CreateAPIKey creates a new API key for a user
func (s *Service) CreateAPIKey(userID string, req *types.CreateAPIKeyRequest) (*types.CreateAPIKeyResponse, error) {
	// Get user to validate it exists and get organization ID
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Generate a secure random API key
	keyString := generateAPIKey()
	keyHash := hashAPIKey(keyString)
	prefix := keyString[:8] // Store first 8 chars as prefix for identification

	// Create API key record
	query := `
		INSERT INTO api_keys (
			user_id, organization_id, name, key_hash, prefix,
			key_type, permissions, expires_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	var apiKey types.APIKey
	var expiresAt sql.NullTime
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, types.NewValidationError("invalid expiration date format")
		}
		expiresAt = sql.NullTime{Time: t, Valid: true}
	}

	// Map role to permissions
	permissions := getPermissionsForRole(req.Role)

	err = s.db.QueryRow(
		query,
		userID,
		user.OrganizationID,
		req.Name,
		keyHash,
		prefix,
		"user",
		pq.Array(permissions),
		expiresAt,
		true,
	).Scan(&apiKey.ID, &apiKey.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Build response with the actual key (only returned once)
	apiKey.Name = req.Name
	apiKey.KeyHash = prefix + "..." // Only show prefix in response
	apiKey.Role = req.Role
	apiKey.IsActive = true
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}

	return &types.CreateAPIKeyResponse{
		APIKey: &apiKey,
		Key:    keyString, // Return the actual key only once
	}, nil
}

// ListAPIKeys lists all API keys for a user
func (s *Service) ListAPIKeys(userID string) ([]*types.APIKey, error) {
	query := `
		SELECT id, name, prefix || '...' as key_hash, permissions,
		       is_active, expires_at, created_at, last_used_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var keys []*types.APIKey
	for rows.Next() {
		var key types.APIKey
		var expiresAt, lastUsedAt sql.NullTime
		var permissions []string

		err := rows.Scan(
			&key.ID,
			&key.Name,
			&key.KeyHash,
			&permissions,
			&key.IsActive,
			&expiresAt,
			&key.CreatedAt,
			&lastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		// Map permissions back to role
		key.Role = getRoleFromPermissions(permissions)

		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}

		keys = append(keys, &key)
	}

	return keys, nil
}

// ListAllAPIKeys lists all API keys for an organization (admin only)
func (s *Service) ListAllAPIKeys(organizationID string) ([]*types.APIKey, error) {
	query := `
		SELECT ak.id, ak.name, ak.prefix || '...' as key_hash, ak.permissions,
		       ak.is_active, ak.expires_at, ak.created_at, ak.last_used_at,
		       ak.user_id, ak.organization_id, u.email as user_email
		FROM api_keys ak
		LEFT JOIN users u ON ak.user_id = u.id
		WHERE ak.organization_id = $1
		ORDER BY ak.created_at DESC
	`

	rows, err := s.db.Query(query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var keys []*types.APIKey
	for rows.Next() {
		var key types.APIKey
		var expiresAt, lastUsedAt sql.NullTime
		var permissions []string
		var userEmail sql.NullString

		err := rows.Scan(
			&key.ID,
			&key.Name,
			&key.KeyHash,
			pq.Array(&permissions),
			&key.IsActive,
			&expiresAt,
			&key.CreatedAt,
			&lastUsedAt,
			&key.UserID,
			&key.OrganizationID,
			&userEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		// Set the role based on permissions
		key.Role = getRoleFromPermissions(permissions)

		// Set optional fields
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}

		// Add user email to metadata for display
		if userEmail.Valid {
			if key.Metadata == nil {
				key.Metadata = make(map[string]interface{})
			}
			key.Metadata["user_email"] = userEmail.String
		}

		keys = append(keys, &key)
	}

	return keys, nil
}

// DeleteAPIKey deletes an API key
func (s *Service) DeleteAPIKey(userID, keyID string) error {
	// Verify the key belongs to the user
	var ownerID string
	err := s.db.QueryRow("SELECT user_id FROM api_keys WHERE id = $1", keyID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.NewNotFoundError("API key not found")
		}
		return fmt.Errorf("failed to verify API key ownership: %w", err)
	}

	if ownerID != userID {
		return types.NewForbiddenError("you do not have permission to delete this API key")
	}

	// Delete the key
	result, err := s.db.Exec("DELETE FROM api_keys WHERE id = $1", keyID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deletion result: %w", err)
	}

	if rowsAffected == 0 {
		return types.NewNotFoundError("API key not found")
	}

	return nil
}

// DeleteAPIKeyByAdmin deletes any API key in the organization (admin only)
func (s *Service) DeleteAPIKeyByAdmin(organizationID, keyID string) error {
	// Verify the key belongs to the organization
	var keyOrgID string
	err := s.db.QueryRow("SELECT organization_id FROM api_keys WHERE id = $1", keyID).Scan(&keyOrgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.NewNotFoundError("API key not found")
		}
		return fmt.Errorf("failed to verify API key organization: %w", err)
	}

	if keyOrgID != organizationID {
		return types.NewForbiddenError("API key does not belong to your organization")
	}

	// Delete the key
	result, err := s.db.Exec("DELETE FROM api_keys WHERE id = $1", keyID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deletion result: %w", err)
	}

	if rowsAffected == 0 {
		return types.NewNotFoundError("API key not found")
	}

	return nil
}

// ValidateAPIKey validates an API key
func (s *Service) ValidateAPIKey(keyString string) (*types.APIKey, error) {
	keyHash := hashAPIKey(keyString)

	query := `
		SELECT id, user_id, organization_id, name, permissions,
		       is_active, expires_at, created_at, last_used_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true
	`

	var apiKey types.APIKey
	var expiresAt, lastUsedAt sql.NullTime
	var permissions []string

	err := s.db.QueryRow(query, keyHash).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.OrganizationID,
		&apiKey.Name,
		pq.Array(&permissions),
		&apiKey.IsActive,
		&expiresAt,
		&apiKey.CreatedAt,
		&lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.NewUnauthorizedError("invalid API key")
		}
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}

	// Check if expired
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return nil, types.NewUnauthorizedError("API key has expired")
	}

	// Update last used timestamp
	go func() {
		_, _ = s.db.Exec("UPDATE api_keys SET last_used_at = NOW() WHERE id = $1", apiKey.ID)
	}()

	// Map permissions to role
	apiKey.Role = getRoleFromPermissions(permissions)

	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		apiKey.LastUsedAt = &lastUsedAt.Time
	}

	return &apiKey, nil
}

// RevokeAPIKey revokes an API key
func (s *Service) RevokeAPIKey(keyID string) error {
	result, err := s.db.Exec("UPDATE api_keys SET is_active = false WHERE id = $1", keyID)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check revocation result: %w", err)
	}

	if rowsAffected == 0 {
		return types.NewNotFoundError("API key not found")
	}

	return nil
}

// hashPassword hashes a password using bcrypt
func (s *Service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BCryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// validatePassword validates a password against hash
func (s *Service) validatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateAPIKey generates a secure random API key
func generateAPIKey() string {
	// Generate 32 random bytes
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}

	// Return hex-encoded string with "mgw_" prefix for identification
	return "mgw_" + hex.EncodeToString(b)
}

// hashAPIKey creates a SHA256 hash of an API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// getPermissionsForRole maps a role to permissions
func getPermissionsForRole(role string) []string {
	switch role {
	case "admin":
		return []string{"read", "write", "delete", "execute", "admin"}
	case "user":
		return []string{"read", "write", "execute"}
	case "viewer":
		return []string{"read"}
	case "api_user":
		return []string{"read", "write", "execute"}
	default:
		return []string{"read"}
	}
}

// getRoleFromPermissions maps permissions back to a role
func getRoleFromPermissions(permissions []string) string {
	permSet := make(map[string]bool)
	for _, p := range permissions {
		permSet[p] = true
	}

	if permSet["admin"] {
		return "admin"
	}
	if permSet["write"] && permSet["execute"] {
		return "user"
	}
	if permSet["read"] && !permSet["write"] {
		return "viewer"
	}

	return "user" // Default
}
