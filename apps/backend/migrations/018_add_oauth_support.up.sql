-- OAuth 2.0 Support Migration
-- Adds OAuth client registration, token management, and authorization code flow

-- OAuth clients table for dynamic client registration
CREATE TABLE oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret_hash VARCHAR(255), -- NULL for public clients
    client_name VARCHAR(255) NOT NULL,
    client_type VARCHAR(50) NOT NULL DEFAULT 'confidential' CHECK (client_type IN ('confidential', 'public')),
    redirect_uris TEXT[] DEFAULT ARRAY[]::TEXT[],
    grant_types TEXT[] DEFAULT ARRAY['client_credentials']::TEXT[],
    response_types TEXT[] DEFAULT ARRAY['token']::TEXT[],
    scope TEXT DEFAULT 'read',
    contacts TEXT[] DEFAULT ARRAY[]::TEXT[],
    logo_uri VARCHAR(500),
    client_uri VARCHAR(500),
    policy_uri VARCHAR(500),
    tos_uri VARCHAR(500),
    jwks_uri VARCHAR(500),
    token_endpoint_auth_method VARCHAR(50) DEFAULT 'client_secret_basic' CHECK (token_endpoint_auth_method IN ('client_secret_basic', 'client_secret_post', 'none', 'private_key_jwt', 'client_secret_jwt')),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- OAuth tokens table (access tokens and refresh tokens)
CREATE TABLE oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(50) NOT NULL CHECK (token_type IN ('access', 'refresh')),
    client_id VARCHAR(255) NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for client_credentials grant
    scope TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    parent_token_id UUID REFERENCES oauth_tokens(id) ON DELETE CASCADE, -- For refresh token relationships
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- OAuth authorization codes for authorization_code grant
CREATE TABLE oauth_authorization_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri VARCHAR(500) NOT NULL,
    scope TEXT,
    code_challenge VARCHAR(255), -- For PKCE
    code_challenge_method VARCHAR(10) CHECK (code_challenge_method IN ('plain', 'S256')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- OAuth consent/grants tracking
CREATE TABLE oauth_user_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, client_id, scope)
);

-- Create indexes for performance
CREATE INDEX idx_oauth_clients_client_id ON oauth_clients(client_id);
CREATE INDEX idx_oauth_clients_org_id ON oauth_clients(organization_id);
CREATE INDEX idx_oauth_clients_active ON oauth_clients(is_active);

CREATE INDEX idx_oauth_tokens_hash ON oauth_tokens(token_hash);
CREATE INDEX idx_oauth_tokens_client_id ON oauth_tokens(client_id);
CREATE INDEX idx_oauth_tokens_user_id ON oauth_tokens(user_id);
CREATE INDEX idx_oauth_tokens_expires ON oauth_tokens(expires_at);
CREATE INDEX idx_oauth_tokens_revoked ON oauth_tokens(revoked_at);
CREATE INDEX idx_oauth_tokens_type ON oauth_tokens(token_type);

CREATE INDEX idx_oauth_codes_hash ON oauth_authorization_codes(code_hash);
CREATE INDEX idx_oauth_codes_client_id ON oauth_authorization_codes(client_id);
CREATE INDEX idx_oauth_codes_expires ON oauth_authorization_codes(expires_at);
CREATE INDEX idx_oauth_codes_used ON oauth_authorization_codes(used_at);

CREATE INDEX idx_oauth_consents_user_client ON oauth_user_consents(user_id, client_id);
CREATE INDEX idx_oauth_consents_expires ON oauth_user_consents(expires_at);

-- Add OAuth configuration to organizations
ALTER TABLE organizations ADD COLUMN oauth_enabled BOOLEAN DEFAULT true;
ALTER TABLE organizations ADD COLUMN oauth_require_consent BOOLEAN DEFAULT false;
ALTER TABLE organizations ADD COLUMN oauth_default_scope TEXT DEFAULT 'read';

-- Add OAuth scopes to users for fine-grained permissions
ALTER TABLE users ADD COLUMN oauth_scopes TEXT[] DEFAULT ARRAY['read']::TEXT[];

-- Add OAuth metadata to endpoints for scope-based access control
ALTER TABLE endpoints ADD COLUMN oauth_scopes TEXT[] DEFAULT ARRAY['read']::TEXT[];
ALTER TABLE endpoints ADD COLUMN require_oauth BOOLEAN DEFAULT false;

-- Update API keys to include OAuth-related fields
ALTER TABLE api_keys ADD COLUMN oauth_client_id VARCHAR(255) REFERENCES oauth_clients(client_id) ON DELETE CASCADE;
ALTER TABLE api_keys ADD COLUMN oauth_scopes TEXT[] DEFAULT ARRAY[]::TEXT[];

-- Create view for active OAuth tokens (non-expired, non-revoked)
CREATE VIEW active_oauth_tokens AS
SELECT 
    t.*,
    c.client_name,
    c.organization_id,
    u.email as user_email,
    u.role as user_role
FROM oauth_tokens t
JOIN oauth_clients c ON t.client_id = c.client_id
LEFT JOIN users u ON t.user_id = u.id
WHERE 
    t.expires_at > NOW()
    AND t.revoked_at IS NULL
    AND c.is_active = true;