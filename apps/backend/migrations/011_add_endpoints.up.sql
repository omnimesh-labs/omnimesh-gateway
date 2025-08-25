-- Create endpoints table for public-facing namespace URLs
CREATE TABLE endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Authentication settings
    enable_api_key_auth BOOLEAN DEFAULT true,
    enable_oauth BOOLEAN DEFAULT false,
    enable_public_access BOOLEAN DEFAULT false,
    use_query_param_auth BOOLEAN DEFAULT false,

    -- Rate limiting
    rate_limit_requests INTEGER DEFAULT 100,
    rate_limit_window INTEGER DEFAULT 60, -- seconds

    -- CORS settings
    allowed_origins TEXT[] DEFAULT ARRAY['*']::TEXT[],
    allowed_methods TEXT[] DEFAULT ARRAY['GET', 'POST', 'OPTIONS']::TEXT[],

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',

    UNIQUE(name), -- Global uniqueness for URL routing
    CONSTRAINT endpoints_name_url_safe CHECK (name ~ '^[a-zA-Z0-9_-]+$')
);

-- Create indexes
CREATE INDEX idx_endpoints_org ON endpoints(organization_id);
CREATE INDEX idx_endpoints_namespace ON endpoints(namespace_id);
CREATE INDEX idx_endpoints_name ON endpoints(name);
CREATE INDEX idx_endpoints_active ON endpoints(is_active);

-- Create trigger to update updated_at
CREATE TRIGGER update_endpoints_updated_at
    BEFORE UPDATE ON endpoints
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
