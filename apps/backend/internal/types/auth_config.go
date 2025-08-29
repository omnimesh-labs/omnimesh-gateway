package types

import (
	"time"

	"github.com/google/uuid"
)

// AuthConfigurationRequest represents a request to update authentication configuration
type AuthConfigurationRequest struct {
	// JWT Configuration
	JWTEnabled              *bool `json:"jwt_enabled,omitempty"`
	JWTAccessTokenExpiry    *int  `json:"jwt_access_token_expiry,omitempty"`
	JWTRefreshTokenExpiry   *int  `json:"jwt_refresh_token_expiry,omitempty"`

	// API Key Configuration
	APIKeysEnabled         *bool `json:"api_keys_enabled,omitempty"`
	MaxAPIKeysPerUser      *int  `json:"max_api_keys_per_user,omitempty"`
	APIKeyDefaultExpiry    *int  `json:"api_key_default_expiry,omitempty"`

	// OAuth2 Configuration
	OAuth2Enabled   *bool                    `json:"oauth2_enabled,omitempty"`
	OAuth2Providers *[]OAuth2ProviderConfig `json:"oauth2_providers,omitempty"`

	// Multi-Factor Authentication
	MFARequired *bool     `json:"mfa_required,omitempty"`
	MFAMethods  *[]string `json:"mfa_methods,omitempty"`
}

// SessionConfigurationRequest represents a request to update session configuration
type SessionConfigurationRequest struct {
	SessionTimeoutSeconds  *int    `json:"session_timeout_seconds,omitempty"`
	RefreshStrategy        *string `json:"refresh_strategy,omitempty"`
	MaxConcurrentSessions  *int    `json:"max_concurrent_sessions,omitempty"`
	CookieSecure           *bool   `json:"cookie_secure,omitempty"`
	CookieHTTPOnly         *bool   `json:"cookie_http_only,omitempty"`
	CookieSameSite         *string `json:"cookie_same_site,omitempty"`
	RememberMeEnabled      *bool   `json:"remember_me_enabled,omitempty"`
	RememberMeDurationDays *int    `json:"remember_me_duration_days,omitempty"`
}

// SecurityPolicyRequest represents a request to update security policy
type SecurityPolicyRequest struct {
	// Password Requirements
	PasswordMinLength        *int  `json:"password_min_length,omitempty"`
	PasswordRequireUppercase *bool `json:"password_require_uppercase,omitempty"`
	PasswordRequireLowercase *bool `json:"password_require_lowercase,omitempty"`
	PasswordRequireNumbers   *bool `json:"password_require_numbers,omitempty"`
	PasswordRequireSpecial   *bool `json:"password_require_special,omitempty"`
	PasswordMaxAgeDays       *int  `json:"password_max_age_days,omitempty"`
	PasswordHistoryCount     *int  `json:"password_history_count,omitempty"`

	// Account Security
	AccountLockoutEnabled         *bool `json:"account_lockout_enabled,omitempty"`
	AccountLockoutThreshold       *int  `json:"account_lockout_threshold,omitempty"`
	AccountLockoutDurationMinutes *int  `json:"account_lockout_duration_minutes,omitempty"`

	// Email Verification
	EmailVerificationRequired    *bool `json:"email_verification_required,omitempty"`
	EmailVerificationExpiryHours *int  `json:"email_verification_expiry_hours,omitempty"`

	// IP Restrictions
	IPWhitelist        *[]string `json:"ip_whitelist,omitempty"`
	GeoBlockingEnabled *bool     `json:"geo_blocking_enabled,omitempty"`
	AllowedCountries   *[]string `json:"allowed_countries,omitempty"`

	// Audit and Compliance
	PasswordChangeRequired *bool   `json:"password_change_required,omitempty"`
	ComplianceMode        *string `json:"compliance_mode,omitempty"`
}

// CompleteAuthConfigurationRequest represents a complete auth configuration update request
type CompleteAuthConfigurationRequest struct {
	Methods  *AuthConfigurationRequest  `json:"methods,omitempty"`
	Session  *SessionConfigurationRequest `json:"session,omitempty"`
	Security *SecurityPolicyRequest     `json:"security,omitempty"`
}

// OAuth2ProviderConfig represents OAuth2 provider configuration
type OAuth2ProviderConfig struct {
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret,omitempty"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
	Scopes       []string `json:"scopes"`
	Enabled      bool     `json:"enabled"`
}

// AuthConfigurationResponse represents the complete auth configuration response
type AuthConfigurationResponse struct {
	Methods  AuthMethodsResponse  `json:"methods"`
	Session  SessionConfigResponse `json:"session"`
	Security SecurityPolicyResponse `json:"security"`

	// Metadata
	LastUpdated time.Time  `json:"last_updated"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
}

// AuthMethodsResponse represents the authentication methods configuration
type AuthMethodsResponse struct {
	JWTEnabled              bool                    `json:"jwt_enabled"`
	JWTAccessTokenExpiry    int                     `json:"jwt_access_token_expiry"`
	JWTRefreshTokenExpiry   int                     `json:"jwt_refresh_token_expiry"`
	APIKeysEnabled          bool                    `json:"api_keys_enabled"`
	MaxAPIKeysPerUser       int                     `json:"max_api_keys_per_user"`
	APIKeyDefaultExpiry     *int                    `json:"api_key_default_expiry,omitempty"`
	OAuth2Enabled           bool                    `json:"oauth2_enabled"`
	OAuth2Providers         []OAuth2ProviderConfig `json:"oauth2_providers"`
	MFARequired             bool                    `json:"mfa_required"`
	MFAMethods              []string                `json:"mfa_methods"`
}

// SessionConfigResponse represents the session configuration
type SessionConfigResponse struct {
	SessionTimeoutSeconds  int    `json:"session_timeout_seconds"`
	RefreshStrategy        string `json:"refresh_strategy"`
	MaxConcurrentSessions  int    `json:"max_concurrent_sessions"`
	CookieSecure           bool   `json:"cookie_secure"`
	CookieHTTPOnly         bool   `json:"cookie_http_only"`
	CookieSameSite         string `json:"cookie_same_site"`
	RememberMeEnabled      bool   `json:"remember_me_enabled"`
	RememberMeDurationDays int    `json:"remember_me_duration_days"`
}

// SecurityPolicyResponse represents the security policy configuration
type SecurityPolicyResponse struct {
	PasswordRequirements    PasswordRequirementsResponse `json:"password_requirements"`
	AccountSecurity        AccountSecurityResponse       `json:"account_security"`
	EmailVerification      EmailVerificationResponse     `json:"email_verification"`
	IPRestrictions         IPRestrictionsResponse        `json:"ip_restrictions"`
	ComplianceMode         string                        `json:"compliance_mode"`
	PasswordChangeRequired bool                          `json:"password_change_required"`
}

// PasswordRequirementsResponse represents password requirements
type PasswordRequirementsResponse struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSpecial   bool `json:"require_special"`
	MaxAgeDays       *int `json:"max_age_days,omitempty"`
	HistoryCount     int  `json:"history_count"`
}

// AccountSecurityResponse represents account security settings
type AccountSecurityResponse struct {
	LockoutEnabled         bool `json:"lockout_enabled"`
	LockoutThreshold       int  `json:"lockout_threshold"`
	LockoutDurationMinutes int  `json:"lockout_duration_minutes"`
}

// EmailVerificationResponse represents email verification settings
type EmailVerificationResponse struct {
	Required    bool `json:"required"`
	ExpiryHours int  `json:"expiry_hours"`
}

// IPRestrictionsResponse represents IP restriction settings
type IPRestrictionsResponse struct {
	Whitelist          []string `json:"whitelist"`
	GeoBlockingEnabled bool     `json:"geo_blocking_enabled"`
	AllowedCountries   []string `json:"allowed_countries"`
}

// AuthConfigDefaults represents default values for auth configuration
type AuthConfigDefaults struct {
	Methods  AuthMethodsResponse   `json:"methods"`
	Session  SessionConfigResponse `json:"session"`
	Security SecurityPolicyResponse `json:"security"`
}

// GetAuthConfigDefaults returns the default auth configuration values
func GetAuthConfigDefaults() *AuthConfigDefaults {
	return &AuthConfigDefaults{
		Methods: AuthMethodsResponse{
			JWTEnabled:              true,
			JWTAccessTokenExpiry:    900,  // 15 minutes
			JWTRefreshTokenExpiry:   86400, // 24 hours
			APIKeysEnabled:          true,
			MaxAPIKeysPerUser:       10,
			APIKeyDefaultExpiry:     nil, // No default expiry
			OAuth2Enabled:           false,
			OAuth2Providers:         []OAuth2ProviderConfig{},
			MFARequired:             false,
			MFAMethods:              []string{"totp"},
		},
		Session: SessionConfigResponse{
			SessionTimeoutSeconds:  3600, // 1 hour
			RefreshStrategy:        "sliding",
			MaxConcurrentSessions:  5,
			CookieSecure:           true,
			CookieHTTPOnly:         true,
			CookieSameSite:         "strict",
			RememberMeEnabled:      true,
			RememberMeDurationDays: 30,
		},
		Security: SecurityPolicyResponse{
			PasswordRequirements: PasswordRequirementsResponse{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSpecial:   false,
				MaxAgeDays:       nil,
				HistoryCount:     0,
			},
			AccountSecurity: AccountSecurityResponse{
				LockoutEnabled:         true,
				LockoutThreshold:       5,
				LockoutDurationMinutes: 30,
			},
			EmailVerification: EmailVerificationResponse{
				Required:    false,
				ExpiryHours: 24,
			},
			IPRestrictions: IPRestrictionsResponse{
				Whitelist:          []string{},
				GeoBlockingEnabled: false,
				AllowedCountries:   []string{},
			},
			ComplianceMode:         "standard",
			PasswordChangeRequired: false,
		},
	}
}