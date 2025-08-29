package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OAuth2Provider represents an OAuth2 provider configuration
type OAuth2Provider struct {
	Name         string `json:"name"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"` // Omit from JSON responses
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"user_info_url"`
	Scopes       []string `json:"scopes"`
	Enabled      bool   `json:"enabled"`
}

// OAuth2Providers represents a slice of OAuth2Provider
type OAuth2Providers []OAuth2Provider

// Value implements the driver.Valuer interface for database storage
func (p OAuth2Providers) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for database retrieval
func (p *OAuth2Providers) Scan(value interface{}) error {
	if value == nil {
		*p = OAuth2Providers{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into OAuth2Providers", value)
	}

	return json.Unmarshal(bytes, p)
}

// MFAMethods represents supported MFA methods
type MFAMethods []string

// Value implements the driver.Valuer interface for database storage
func (m MFAMethods) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database retrieval
func (m *MFAMethods) Scan(value interface{}) error {
	if value == nil {
		*m = MFAMethods{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into MFAMethods", value)
	}

	return json.Unmarshal(bytes, m)
}

// AuthConfiguration represents the auth_configurations table
type AuthConfiguration struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// JWT Configuration
	JWTEnabled              bool `db:"jwt_enabled" json:"jwt_enabled"`
	JWTAccessTokenExpiry    int  `db:"jwt_access_token_expiry" json:"jwt_access_token_expiry"`
	JWTRefreshTokenExpiry   int  `db:"jwt_refresh_token_expiry" json:"jwt_refresh_token_expiry"`

	// API Key Configuration
	APIKeysEnabled         bool `db:"api_keys_enabled" json:"api_keys_enabled"`
	MaxAPIKeysPerUser      int  `db:"max_api_keys_per_user" json:"max_api_keys_per_user"`
	APIKeyDefaultExpiry    *int `db:"api_key_default_expiry" json:"api_key_default_expiry,omitempty"`

	// OAuth2 Configuration
	OAuth2Enabled   bool             `db:"oauth2_enabled" json:"oauth2_enabled"`
	OAuth2Providers OAuth2Providers `db:"oauth2_providers" json:"oauth2_providers"`

	// Multi-Factor Authentication
	MFARequired bool       `db:"mfa_required" json:"mfa_required"`
	MFAMethods  MFAMethods `db:"mfa_methods" json:"mfa_methods"`

	// Metadata
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	CreatedBy *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `db:"updated_by" json:"updated_by,omitempty"`
}

// SessionConfiguration represents the session_configurations table
type SessionConfiguration struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// Session Settings
	SessionTimeoutSeconds  int    `db:"session_timeout_seconds" json:"session_timeout_seconds"`
	RefreshStrategy        string `db:"refresh_strategy" json:"refresh_strategy"`
	MaxConcurrentSessions  int    `db:"max_concurrent_sessions" json:"max_concurrent_sessions"`

	// Cookie Settings
	CookieSecure   bool   `db:"cookie_secure" json:"cookie_secure"`
	CookieHTTPOnly bool   `db:"cookie_http_only" json:"cookie_http_only"`
	CookieSameSite string `db:"cookie_same_site" json:"cookie_same_site"`

	// Remember Me Settings
	RememberMeEnabled      bool `db:"remember_me_enabled" json:"remember_me_enabled"`
	RememberMeDurationDays int  `db:"remember_me_duration_days" json:"remember_me_duration_days"`

	// Metadata
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	CreatedBy *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `db:"updated_by" json:"updated_by,omitempty"`
}

// SecurityPolicy represents the security_policies table
type SecurityPolicy struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// Password Requirements
	PasswordMinLength        int  `db:"password_min_length" json:"password_min_length"`
	PasswordRequireUppercase bool `db:"password_require_uppercase" json:"password_require_uppercase"`
	PasswordRequireLowercase bool `db:"password_require_lowercase" json:"password_require_lowercase"`
	PasswordRequireNumbers   bool `db:"password_require_numbers" json:"password_require_numbers"`
	PasswordRequireSpecial   bool `db:"password_require_special" json:"password_require_special"`
	PasswordMaxAgeDays       *int `db:"password_max_age_days" json:"password_max_age_days,omitempty"`
	PasswordHistoryCount     int  `db:"password_history_count" json:"password_history_count"`

	// Account Security
	AccountLockoutEnabled         bool `db:"account_lockout_enabled" json:"account_lockout_enabled"`
	AccountLockoutThreshold       int  `db:"account_lockout_threshold" json:"account_lockout_threshold"`
	AccountLockoutDurationMinutes int  `db:"account_lockout_duration_minutes" json:"account_lockout_duration_minutes"`

	// Email Verification
	EmailVerificationRequired    bool `db:"email_verification_required" json:"email_verification_required"`
	EmailVerificationExpiryHours int  `db:"email_verification_expiry_hours" json:"email_verification_expiry_hours"`

	// IP Restrictions
	IPWhitelist       pq.StringArray `db:"ip_whitelist" json:"ip_whitelist"`
	GeoBlockingEnabled bool          `db:"geo_blocking_enabled" json:"geo_blocking_enabled"`
	AllowedCountries   pq.StringArray `db:"allowed_countries" json:"allowed_countries"`

	// Audit and Compliance
	PasswordChangeRequired bool   `db:"password_change_required" json:"password_change_required"`
	ComplianceMode        string `db:"compliance_mode" json:"compliance_mode"`

	// Metadata
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	CreatedBy *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `db:"updated_by" json:"updated_by,omitempty"`
}

// AuthConfigurationModel handles database operations for auth configurations
type AuthConfigurationModel struct {
	db Database
}

// NewAuthConfigurationModel creates a new auth configuration model
func NewAuthConfigurationModel(db Database) *AuthConfigurationModel {
	return &AuthConfigurationModel{db: db}
}

// GetByOrganizationID retrieves auth configuration by organization ID
func (m *AuthConfigurationModel) GetByOrganizationID(orgID uuid.UUID) (*AuthConfiguration, error) {
	query := `
		SELECT id, organization_id, jwt_enabled, jwt_access_token_expiry, jwt_refresh_token_expiry,
			   api_keys_enabled, max_api_keys_per_user, api_key_default_expiry,
			   oauth2_enabled, oauth2_providers, mfa_required, mfa_methods,
			   created_at, updated_at, created_by, updated_by
		FROM auth_configurations
		WHERE organization_id = $1
	`

	config := &AuthConfiguration{}
	err := m.db.QueryRow(query, orgID).Scan(
		&config.ID, &config.OrganizationID, &config.JWTEnabled,
		&config.JWTAccessTokenExpiry, &config.JWTRefreshTokenExpiry,
		&config.APIKeysEnabled, &config.MaxAPIKeysPerUser, &config.APIKeyDefaultExpiry,
		&config.OAuth2Enabled, &config.OAuth2Providers, &config.MFARequired, &config.MFAMethods,
		&config.CreatedAt, &config.UpdatedAt, &config.CreatedBy, &config.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	return config, nil
}

// Create inserts a new auth configuration
func (m *AuthConfigurationModel) Create(config *AuthConfiguration) error {
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	query := `
		INSERT INTO auth_configurations (
			id, organization_id, jwt_enabled, jwt_access_token_expiry, jwt_refresh_token_expiry,
			api_keys_enabled, max_api_keys_per_user, api_key_default_expiry,
			oauth2_enabled, oauth2_providers, mfa_required, mfa_methods,
			created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		RETURNING created_at, updated_at
	`

	return m.db.QueryRow(query,
		config.ID, config.OrganizationID, config.JWTEnabled,
		config.JWTAccessTokenExpiry, config.JWTRefreshTokenExpiry,
		config.APIKeysEnabled, config.MaxAPIKeysPerUser, config.APIKeyDefaultExpiry,
		config.OAuth2Enabled, config.OAuth2Providers, config.MFARequired, config.MFAMethods,
		config.CreatedBy, config.UpdatedBy,
	).Scan(&config.CreatedAt, &config.UpdatedAt)
}

// Update modifies an existing auth configuration
func (m *AuthConfigurationModel) Update(config *AuthConfiguration) error {
	query := `
		UPDATE auth_configurations SET
			jwt_enabled = $2, jwt_access_token_expiry = $3, jwt_refresh_token_expiry = $4,
			api_keys_enabled = $5, max_api_keys_per_user = $6, api_key_default_expiry = $7,
			oauth2_enabled = $8, oauth2_providers = $9, mfa_required = $10, mfa_methods = $11,
			updated_by = $12
		WHERE id = $1
		RETURNING updated_at
	`

	return m.db.QueryRow(query,
		config.ID, config.JWTEnabled,
		config.JWTAccessTokenExpiry, config.JWTRefreshTokenExpiry,
		config.APIKeysEnabled, config.MaxAPIKeysPerUser, config.APIKeyDefaultExpiry,
		config.OAuth2Enabled, config.OAuth2Providers, config.MFARequired, config.MFAMethods,
		config.UpdatedBy,
	).Scan(&config.UpdatedAt)
}

// SessionConfigurationModel handles database operations for session configurations
type SessionConfigurationModel struct {
	db Database
}

// NewSessionConfigurationModel creates a new session configuration model
func NewSessionConfigurationModel(db Database) *SessionConfigurationModel {
	return &SessionConfigurationModel{db: db}
}

// GetByOrganizationID retrieves session configuration by organization ID
func (m *SessionConfigurationModel) GetByOrganizationID(orgID uuid.UUID) (*SessionConfiguration, error) {
	query := `
		SELECT id, organization_id, session_timeout_seconds, refresh_strategy, max_concurrent_sessions,
			   cookie_secure, cookie_http_only, cookie_same_site,
			   remember_me_enabled, remember_me_duration_days,
			   created_at, updated_at, created_by, updated_by
		FROM session_configurations
		WHERE organization_id = $1
	`

	config := &SessionConfiguration{}
	err := m.db.QueryRow(query, orgID).Scan(
		&config.ID, &config.OrganizationID, &config.SessionTimeoutSeconds,
		&config.RefreshStrategy, &config.MaxConcurrentSessions,
		&config.CookieSecure, &config.CookieHTTPOnly, &config.CookieSameSite,
		&config.RememberMeEnabled, &config.RememberMeDurationDays,
		&config.CreatedAt, &config.UpdatedAt, &config.CreatedBy, &config.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	return config, nil
}

// Update modifies an existing session configuration
func (m *SessionConfigurationModel) Update(config *SessionConfiguration) error {
	query := `
		UPDATE session_configurations SET
			session_timeout_seconds = $2, refresh_strategy = $3, max_concurrent_sessions = $4,
			cookie_secure = $5, cookie_http_only = $6, cookie_same_site = $7,
			remember_me_enabled = $8, remember_me_duration_days = $9,
			updated_by = $10
		WHERE id = $1
		RETURNING updated_at
	`

	return m.db.QueryRow(query,
		config.ID, config.SessionTimeoutSeconds, config.RefreshStrategy, config.MaxConcurrentSessions,
		config.CookieSecure, config.CookieHTTPOnly, config.CookieSameSite,
		config.RememberMeEnabled, config.RememberMeDurationDays,
		config.UpdatedBy,
	).Scan(&config.UpdatedAt)
}

// SecurityPolicyModel handles database operations for security policies
type SecurityPolicyModel struct {
	db Database
}

// NewSecurityPolicyModel creates a new security policy model
func NewSecurityPolicyModel(db Database) *SecurityPolicyModel {
	return &SecurityPolicyModel{db: db}
}

// GetByOrganizationID retrieves security policy by organization ID
func (m *SecurityPolicyModel) GetByOrganizationID(orgID uuid.UUID) (*SecurityPolicy, error) {
	query := `
		SELECT id, organization_id, password_min_length, password_require_uppercase,
			   password_require_lowercase, password_require_numbers, password_require_special,
			   password_max_age_days, password_history_count,
			   account_lockout_enabled, account_lockout_threshold, account_lockout_duration_minutes,
			   email_verification_required, email_verification_expiry_hours,
			   ip_whitelist, geo_blocking_enabled, allowed_countries,
			   password_change_required, compliance_mode,
			   created_at, updated_at, created_by, updated_by
		FROM security_policies
		WHERE organization_id = $1
	`

	policy := &SecurityPolicy{}
	err := m.db.QueryRow(query, orgID).Scan(
		&policy.ID, &policy.OrganizationID, &policy.PasswordMinLength,
		&policy.PasswordRequireUppercase, &policy.PasswordRequireLowercase,
		&policy.PasswordRequireNumbers, &policy.PasswordRequireSpecial,
		&policy.PasswordMaxAgeDays, &policy.PasswordHistoryCount,
		&policy.AccountLockoutEnabled, &policy.AccountLockoutThreshold,
		&policy.AccountLockoutDurationMinutes,
		&policy.EmailVerificationRequired, &policy.EmailVerificationExpiryHours,
		&policy.IPWhitelist, &policy.GeoBlockingEnabled, &policy.AllowedCountries,
		&policy.PasswordChangeRequired, &policy.ComplianceMode,
		&policy.CreatedAt, &policy.UpdatedAt, &policy.CreatedBy, &policy.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	return policy, nil
}

// Update modifies an existing security policy
func (m *SecurityPolicyModel) Update(policy *SecurityPolicy) error {
	query := `
		UPDATE security_policies SET
			password_min_length = $2, password_require_uppercase = $3,
			password_require_lowercase = $4, password_require_numbers = $5, password_require_special = $6,
			password_max_age_days = $7, password_history_count = $8,
			account_lockout_enabled = $9, account_lockout_threshold = $10, account_lockout_duration_minutes = $11,
			email_verification_required = $12, email_verification_expiry_hours = $13,
			ip_whitelist = $14, geo_blocking_enabled = $15, allowed_countries = $16,
			password_change_required = $17, compliance_mode = $18,
			updated_by = $19
		WHERE id = $1
		RETURNING updated_at
	`

	return m.db.QueryRow(query,
		policy.ID, policy.PasswordMinLength,
		policy.PasswordRequireUppercase, policy.PasswordRequireLowercase,
		policy.PasswordRequireNumbers, policy.PasswordRequireSpecial,
		policy.PasswordMaxAgeDays, policy.PasswordHistoryCount,
		policy.AccountLockoutEnabled, policy.AccountLockoutThreshold,
		policy.AccountLockoutDurationMinutes,
		policy.EmailVerificationRequired, policy.EmailVerificationExpiryHours,
		policy.IPWhitelist, policy.GeoBlockingEnabled, policy.AllowedCountries,
		policy.PasswordChangeRequired, policy.ComplianceMode,
		policy.UpdatedBy,
	).Scan(&policy.UpdatedAt)
}