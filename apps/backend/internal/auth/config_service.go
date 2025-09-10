package auth

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/database/models"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// ConfigService handles authentication configuration management
type ConfigService struct {
	db                  models.Database
	authConfigModel     *models.AuthConfigurationModel
	sessionConfigModel  *models.SessionConfigurationModel
	securityPolicyModel *models.SecurityPolicyModel
}

// NewConfigService creates a new auth configuration service
func NewConfigService(db models.Database) *ConfigService {
	return &ConfigService{
		db:                  db,
		authConfigModel:     models.NewAuthConfigurationModel(db),
		sessionConfigModel:  models.NewSessionConfigurationModel(db),
		securityPolicyModel: models.NewSecurityPolicyModel(db),
	}
}

// GetConfiguration retrieves the complete auth configuration for an organization
func (s *ConfigService) GetConfiguration(orgID uuid.UUID) (*types.AuthConfigurationResponse, error) {
	// Get auth configuration
	authConfig, err := s.authConfigModel.GetByOrganizationID(orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return defaults if no configuration exists
			return s.getDefaultConfiguration()
		}
		return nil, fmt.Errorf("failed to get auth configuration: %w", err)
	}

	// Get session configuration
	sessionConfig, err := s.sessionConfigModel.GetByOrganizationID(orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return defaults if no configuration exists
			return s.getDefaultConfiguration()
		}
		return nil, fmt.Errorf("failed to get session configuration: %w", err)
	}

	// Get security policy
	securityPolicy, err := s.securityPolicyModel.GetByOrganizationID(orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return defaults if no configuration exists
			return s.getDefaultConfiguration()
		}
		return nil, fmt.Errorf("failed to get security policy: %w", err)
	}

	// Convert OAuth2 providers from database model to API response
	oauth2Providers := make([]types.OAuth2ProviderConfig, len(authConfig.OAuth2Providers))
	for i, provider := range authConfig.OAuth2Providers {
		oauth2Providers[i] = types.OAuth2ProviderConfig{
			Name:         provider.Name,
			ClientID:     provider.ClientID,
			ClientSecret: "", // Never return secrets in API responses
			AuthURL:      provider.AuthURL,
			TokenURL:     provider.TokenURL,
			UserInfoURL:  provider.UserInfoURL,
			Scopes:       provider.Scopes,
			Enabled:      provider.Enabled,
		}
	}

	// Build response
	response := &types.AuthConfigurationResponse{
		Methods: types.AuthMethodsResponse{
			JWTEnabled:            authConfig.JWTEnabled,
			JWTAccessTokenExpiry:  authConfig.JWTAccessTokenExpiry,
			JWTRefreshTokenExpiry: authConfig.JWTRefreshTokenExpiry,
			APIKeysEnabled:        authConfig.APIKeysEnabled,
			MaxAPIKeysPerUser:     authConfig.MaxAPIKeysPerUser,
			APIKeyDefaultExpiry:   authConfig.APIKeyDefaultExpiry,
			OAuth2Enabled:         authConfig.OAuth2Enabled,
			OAuth2Providers:       oauth2Providers,
			MFARequired:           authConfig.MFARequired,
			MFAMethods:            []string(authConfig.MFAMethods),
		},
		Session: types.SessionConfigResponse{
			SessionTimeoutSeconds:  sessionConfig.SessionTimeoutSeconds,
			RefreshStrategy:        sessionConfig.RefreshStrategy,
			MaxConcurrentSessions:  sessionConfig.MaxConcurrentSessions,
			CookieSecure:           sessionConfig.CookieSecure,
			CookieHTTPOnly:         sessionConfig.CookieHTTPOnly,
			CookieSameSite:         sessionConfig.CookieSameSite,
			RememberMeEnabled:      sessionConfig.RememberMeEnabled,
			RememberMeDurationDays: sessionConfig.RememberMeDurationDays,
		},
		Security: types.SecurityPolicyResponse{
			PasswordRequirements: types.PasswordRequirementsResponse{
				MinLength:        securityPolicy.PasswordMinLength,
				RequireUppercase: securityPolicy.PasswordRequireUppercase,
				RequireLowercase: securityPolicy.PasswordRequireLowercase,
				RequireNumbers:   securityPolicy.PasswordRequireNumbers,
				RequireSpecial:   securityPolicy.PasswordRequireSpecial,
				MaxAgeDays:       securityPolicy.PasswordMaxAgeDays,
				HistoryCount:     securityPolicy.PasswordHistoryCount,
			},
			AccountSecurity: types.AccountSecurityResponse{
				LockoutEnabled:         securityPolicy.AccountLockoutEnabled,
				LockoutThreshold:       securityPolicy.AccountLockoutThreshold,
				LockoutDurationMinutes: securityPolicy.AccountLockoutDurationMinutes,
			},
			EmailVerification: types.EmailVerificationResponse{
				Required:    securityPolicy.EmailVerificationRequired,
				ExpiryHours: securityPolicy.EmailVerificationExpiryHours,
			},
			IPRestrictions: types.IPRestrictionsResponse{
				Whitelist:          []string(securityPolicy.IPWhitelist),
				GeoBlockingEnabled: securityPolicy.GeoBlockingEnabled,
				AllowedCountries:   []string(securityPolicy.AllowedCountries),
			},
			ComplianceMode:         securityPolicy.ComplianceMode,
			PasswordChangeRequired: securityPolicy.PasswordChangeRequired,
		},
		LastUpdated: authConfig.UpdatedAt,
		UpdatedBy:   authConfig.UpdatedBy,
	}

	return response, nil
}

// UpdateConfiguration updates the authentication configuration for an organization
func (s *ConfigService) UpdateConfiguration(orgID uuid.UUID, req *types.CompleteAuthConfigurationRequest, updatedBy uuid.UUID) (*types.AuthConfigurationResponse, error) {
	// Get existing configuration
	authConfig, err := s.authConfigModel.GetByOrganizationID(orgID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get auth configuration: %w", err)
	}

	sessionConfig, err := s.sessionConfigModel.GetByOrganizationID(orgID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get session configuration: %w", err)
	}

	securityPolicy, err := s.securityPolicyModel.GetByOrganizationID(orgID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get security policy: %w", err)
	}

	// Update auth methods configuration
	if req.Methods != nil {
		if req.Methods.JWTEnabled != nil {
			authConfig.JWTEnabled = *req.Methods.JWTEnabled
		}
		if req.Methods.JWTAccessTokenExpiry != nil {
			authConfig.JWTAccessTokenExpiry = *req.Methods.JWTAccessTokenExpiry
		}
		if req.Methods.JWTRefreshTokenExpiry != nil {
			authConfig.JWTRefreshTokenExpiry = *req.Methods.JWTRefreshTokenExpiry
		}
		if req.Methods.APIKeysEnabled != nil {
			authConfig.APIKeysEnabled = *req.Methods.APIKeysEnabled
		}
		if req.Methods.MaxAPIKeysPerUser != nil {
			authConfig.MaxAPIKeysPerUser = *req.Methods.MaxAPIKeysPerUser
		}
		if req.Methods.APIKeyDefaultExpiry != nil {
			authConfig.APIKeyDefaultExpiry = req.Methods.APIKeyDefaultExpiry
		}
		if req.Methods.OAuth2Enabled != nil {
			authConfig.OAuth2Enabled = *req.Methods.OAuth2Enabled
		}
		if req.Methods.OAuth2Providers != nil {
			// Convert OAuth2 providers from API request to database model
			providers := make(models.OAuth2Providers, len(*req.Methods.OAuth2Providers))
			for i, provider := range *req.Methods.OAuth2Providers {
				providers[i] = models.OAuth2Provider{
					Name:         provider.Name,
					ClientID:     provider.ClientID,
					ClientSecret: provider.ClientSecret,
					AuthURL:      provider.AuthURL,
					TokenURL:     provider.TokenURL,
					UserInfoURL:  provider.UserInfoURL,
					Scopes:       provider.Scopes,
					Enabled:      provider.Enabled,
				}
			}
			authConfig.OAuth2Providers = providers
		}
		if req.Methods.MFARequired != nil {
			authConfig.MFARequired = *req.Methods.MFARequired
		}
		if req.Methods.MFAMethods != nil {
			authConfig.MFAMethods = models.MFAMethods(*req.Methods.MFAMethods)
		}

		authConfig.UpdatedBy = &updatedBy
		if err := s.authConfigModel.Update(authConfig); err != nil {
			return nil, fmt.Errorf("failed to update auth configuration: %w", err)
		}
	}

	// Update session configuration
	if req.Session != nil {
		if req.Session.SessionTimeoutSeconds != nil {
			sessionConfig.SessionTimeoutSeconds = *req.Session.SessionTimeoutSeconds
		}
		if req.Session.RefreshStrategy != nil {
			sessionConfig.RefreshStrategy = *req.Session.RefreshStrategy
		}
		if req.Session.MaxConcurrentSessions != nil {
			sessionConfig.MaxConcurrentSessions = *req.Session.MaxConcurrentSessions
		}
		if req.Session.CookieSecure != nil {
			sessionConfig.CookieSecure = *req.Session.CookieSecure
		}
		if req.Session.CookieHTTPOnly != nil {
			sessionConfig.CookieHTTPOnly = *req.Session.CookieHTTPOnly
		}
		if req.Session.CookieSameSite != nil {
			sessionConfig.CookieSameSite = *req.Session.CookieSameSite
		}
		if req.Session.RememberMeEnabled != nil {
			sessionConfig.RememberMeEnabled = *req.Session.RememberMeEnabled
		}
		if req.Session.RememberMeDurationDays != nil {
			sessionConfig.RememberMeDurationDays = *req.Session.RememberMeDurationDays
		}

		sessionConfig.UpdatedBy = &updatedBy
		if err := s.sessionConfigModel.Update(sessionConfig); err != nil {
			return nil, fmt.Errorf("failed to update session configuration: %w", err)
		}
	}

	// Update security policy
	if req.Security != nil {
		if req.Security.PasswordMinLength != nil {
			securityPolicy.PasswordMinLength = *req.Security.PasswordMinLength
		}
		if req.Security.PasswordRequireUppercase != nil {
			securityPolicy.PasswordRequireUppercase = *req.Security.PasswordRequireUppercase
		}
		if req.Security.PasswordRequireLowercase != nil {
			securityPolicy.PasswordRequireLowercase = *req.Security.PasswordRequireLowercase
		}
		if req.Security.PasswordRequireNumbers != nil {
			securityPolicy.PasswordRequireNumbers = *req.Security.PasswordRequireNumbers
		}
		if req.Security.PasswordRequireSpecial != nil {
			securityPolicy.PasswordRequireSpecial = *req.Security.PasswordRequireSpecial
		}
		if req.Security.PasswordMaxAgeDays != nil {
			securityPolicy.PasswordMaxAgeDays = req.Security.PasswordMaxAgeDays
		}
		if req.Security.PasswordHistoryCount != nil {
			securityPolicy.PasswordHistoryCount = *req.Security.PasswordHistoryCount
		}
		if req.Security.AccountLockoutEnabled != nil {
			securityPolicy.AccountLockoutEnabled = *req.Security.AccountLockoutEnabled
		}
		if req.Security.AccountLockoutThreshold != nil {
			securityPolicy.AccountLockoutThreshold = *req.Security.AccountLockoutThreshold
		}
		if req.Security.AccountLockoutDurationMinutes != nil {
			securityPolicy.AccountLockoutDurationMinutes = *req.Security.AccountLockoutDurationMinutes
		}
		if req.Security.EmailVerificationRequired != nil {
			securityPolicy.EmailVerificationRequired = *req.Security.EmailVerificationRequired
		}
		if req.Security.EmailVerificationExpiryHours != nil {
			securityPolicy.EmailVerificationExpiryHours = *req.Security.EmailVerificationExpiryHours
		}
		if req.Security.IPWhitelist != nil {
			securityPolicy.IPWhitelist = *req.Security.IPWhitelist
		}
		if req.Security.GeoBlockingEnabled != nil {
			securityPolicy.GeoBlockingEnabled = *req.Security.GeoBlockingEnabled
		}
		if req.Security.AllowedCountries != nil {
			securityPolicy.AllowedCountries = *req.Security.AllowedCountries
		}
		if req.Security.PasswordChangeRequired != nil {
			securityPolicy.PasswordChangeRequired = *req.Security.PasswordChangeRequired
		}
		if req.Security.ComplianceMode != nil {
			securityPolicy.ComplianceMode = *req.Security.ComplianceMode
		}

		securityPolicy.UpdatedBy = &updatedBy
		if err := s.securityPolicyModel.Update(securityPolicy); err != nil {
			return nil, fmt.Errorf("failed to update security policy: %w", err)
		}
	}

	// Return updated configuration
	return s.GetConfiguration(orgID)
}

// ValidateConfiguration validates auth configuration before applying
func (s *ConfigService) ValidateConfiguration(req *types.CompleteAuthConfigurationRequest) error {
	if req.Methods != nil {
		if req.Methods.JWTAccessTokenExpiry != nil && *req.Methods.JWTAccessTokenExpiry <= 0 {
			return fmt.Errorf("JWT access token expiry must be positive")
		}
		if req.Methods.JWTRefreshTokenExpiry != nil && *req.Methods.JWTRefreshTokenExpiry <= 0 {
			return fmt.Errorf("JWT refresh token expiry must be positive")
		}
		if req.Methods.MaxAPIKeysPerUser != nil && *req.Methods.MaxAPIKeysPerUser <= 0 {
			return fmt.Errorf("max API keys per user must be positive")
		}
	}

	if req.Session != nil {
		if req.Session.SessionTimeoutSeconds != nil && *req.Session.SessionTimeoutSeconds <= 0 {
			return fmt.Errorf("session timeout must be positive")
		}
		if req.Session.RefreshStrategy != nil {
			strategy := *req.Session.RefreshStrategy
			if strategy != "sliding" && strategy != "fixed" && strategy != "none" {
				return fmt.Errorf("refresh strategy must be 'sliding', 'fixed', or 'none'")
			}
		}
		if req.Session.MaxConcurrentSessions != nil && *req.Session.MaxConcurrentSessions <= 0 {
			return fmt.Errorf("max concurrent sessions must be positive")
		}
		if req.Session.CookieSameSite != nil {
			sameSite := *req.Session.CookieSameSite
			if sameSite != "strict" && sameSite != "lax" && sameSite != "none" {
				return fmt.Errorf("cookie same site must be 'strict', 'lax', or 'none'")
			}
		}
	}

	if req.Security != nil {
		if req.Security.PasswordMinLength != nil && *req.Security.PasswordMinLength < 6 {
			return fmt.Errorf("password minimum length must be at least 6")
		}
		if req.Security.AccountLockoutThreshold != nil && *req.Security.AccountLockoutThreshold <= 0 {
			return fmt.Errorf("account lockout threshold must be positive")
		}
		if req.Security.AccountLockoutDurationMinutes != nil && *req.Security.AccountLockoutDurationMinutes <= 0 {
			return fmt.Errorf("account lockout duration must be positive")
		}
		if req.Security.ComplianceMode != nil {
			mode := *req.Security.ComplianceMode
			if mode != "standard" && mode != "strict" && mode != "pci" && mode != "hipaa" {
				return fmt.Errorf("compliance mode must be 'standard', 'strict', 'pci', or 'hipaa'")
			}
		}
	}

	return nil
}

// GetDefaults returns the default auth configuration
func (s *ConfigService) GetDefaults() *types.AuthConfigDefaults {
	return types.GetAuthConfigDefaults()
}

// getDefaultConfiguration returns the default configuration as a response
func (s *ConfigService) getDefaultConfiguration() (*types.AuthConfigurationResponse, error) {
	defaults := s.GetDefaults()
	return &types.AuthConfigurationResponse{
		Methods:     defaults.Methods,
		Session:     defaults.Session,
		Security:    defaults.Security,
		LastUpdated: time.Now(), // Use current time
	}, nil
}
