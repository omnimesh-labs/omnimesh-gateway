-- Auth Configuration Tables
-- This migration adds tables for storing authentication configuration per organization

-- Authentication Configuration table - stores auth method settings per organization
CREATE TABLE auth_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- JWT Configuration
    jwt_enabled BOOLEAN DEFAULT true,
    jwt_access_token_expiry INTEGER DEFAULT 900,  -- 15 minutes in seconds
    jwt_refresh_token_expiry INTEGER DEFAULT 86400, -- 24 hours in seconds

    -- API Key Configuration
    api_keys_enabled BOOLEAN DEFAULT true,
    max_api_keys_per_user INTEGER DEFAULT 10,
    api_key_default_expiry INTEGER DEFAULT NULL, -- NULL means no expiry

    -- OAuth2 Configuration
    oauth2_enabled BOOLEAN DEFAULT false,
    oauth2_providers JSONB DEFAULT '[]'::JSONB, -- Array of provider configs

    -- Multi-Factor Authentication
    mfa_required BOOLEAN DEFAULT false,
    mfa_methods JSONB DEFAULT '["totp"]'::JSONB, -- Supported MFA methods

    -- Metadata and tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    -- Ensure one config per organization
    UNIQUE(organization_id)
);

-- Session Configuration table - stores session management settings
CREATE TABLE session_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Session Settings
    session_timeout_seconds INTEGER DEFAULT 3600, -- 1 hour
    refresh_strategy VARCHAR(20) DEFAULT 'sliding' CHECK (refresh_strategy IN ('sliding', 'fixed', 'none')),
    max_concurrent_sessions INTEGER DEFAULT 5,

    -- Cookie Settings
    cookie_secure BOOLEAN DEFAULT true,
    cookie_http_only BOOLEAN DEFAULT true,
    cookie_same_site VARCHAR(10) DEFAULT 'strict' CHECK (cookie_same_site IN ('strict', 'lax', 'none')),

    -- Remember Me Settings
    remember_me_enabled BOOLEAN DEFAULT true,
    remember_me_duration_days INTEGER DEFAULT 30,

    -- Metadata and tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    -- Ensure one config per organization
    UNIQUE(organization_id)
);

-- Security Policy table - stores security requirements and policies
CREATE TABLE security_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Password Requirements
    password_min_length INTEGER DEFAULT 8,
    password_require_uppercase BOOLEAN DEFAULT true,
    password_require_lowercase BOOLEAN DEFAULT true,
    password_require_numbers BOOLEAN DEFAULT true,
    password_require_special BOOLEAN DEFAULT false,
    password_max_age_days INTEGER DEFAULT NULL, -- NULL means no expiry
    password_history_count INTEGER DEFAULT 0, -- Number of previous passwords to remember

    -- Account Security
    account_lockout_enabled BOOLEAN DEFAULT true,
    account_lockout_threshold INTEGER DEFAULT 5, -- Failed attempts before lockout
    account_lockout_duration_minutes INTEGER DEFAULT 30,

    -- Email Verification
    email_verification_required BOOLEAN DEFAULT false,
    email_verification_expiry_hours INTEGER DEFAULT 24,

    -- IP Restrictions
    ip_whitelist TEXT[], -- Array of allowed IP ranges
    geo_blocking_enabled BOOLEAN DEFAULT false,
    allowed_countries TEXT[], -- Array of allowed country codes

    -- Audit and Compliance
    password_change_required BOOLEAN DEFAULT false, -- Force password change on next login
    compliance_mode VARCHAR(20) DEFAULT 'standard' CHECK (compliance_mode IN ('standard', 'strict', 'pci', 'hipaa')),

    -- Metadata and tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    -- Ensure one policy per organization
    UNIQUE(organization_id)
);

-- Create triggers for updated_at
CREATE TRIGGER auth_configurations_updated_at
    BEFORE UPDATE ON auth_configurations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER session_configurations_updated_at
    BEFORE UPDATE ON session_configurations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER security_policies_updated_at
    BEFORE UPDATE ON security_policies
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Create indexes for performance
CREATE INDEX idx_auth_configurations_org_id ON auth_configurations(organization_id);
CREATE INDEX idx_session_configurations_org_id ON session_configurations(organization_id);
CREATE INDEX idx_security_policies_org_id ON security_policies(organization_id);

-- Insert default configurations for existing organizations
INSERT INTO auth_configurations (organization_id, jwt_enabled, api_keys_enabled, oauth2_enabled, mfa_required)
SELECT id, true, true, false, false FROM organizations
ON CONFLICT (organization_id) DO NOTHING;

INSERT INTO session_configurations (organization_id, session_timeout_seconds, refresh_strategy, max_concurrent_sessions)
SELECT id, 3600, 'sliding', 5 FROM organizations
ON CONFLICT (organization_id) DO NOTHING;

INSERT INTO security_policies (
    organization_id,
    password_min_length,
    password_require_uppercase,
    password_require_lowercase,
    password_require_numbers,
    password_require_special,
    account_lockout_enabled,
    account_lockout_threshold,
    email_verification_required
)
SELECT
    id,
    8,
    true,
    true,
    true,
    false,
    true,
    5,
    false
FROM organizations
ON CONFLICT (organization_id) DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE auth_configurations IS 'Authentication method configurations per organization';
COMMENT ON TABLE session_configurations IS 'Session management settings per organization';
COMMENT ON TABLE security_policies IS 'Security policies and requirements per organization';

COMMENT ON COLUMN auth_configurations.jwt_access_token_expiry IS 'JWT access token expiry in seconds';
COMMENT ON COLUMN auth_configurations.jwt_refresh_token_expiry IS 'JWT refresh token expiry in seconds';
COMMENT ON COLUMN auth_configurations.oauth2_providers IS 'JSON array of OAuth2 provider configurations';
COMMENT ON COLUMN session_configurations.refresh_strategy IS 'Token refresh strategy: sliding (extends on use), fixed (fixed expiry), none (no refresh)';
COMMENT ON COLUMN security_policies.ip_whitelist IS 'Array of allowed IP ranges in CIDR notation';
COMMENT ON COLUMN security_policies.compliance_mode IS 'Compliance mode affects default security settings';