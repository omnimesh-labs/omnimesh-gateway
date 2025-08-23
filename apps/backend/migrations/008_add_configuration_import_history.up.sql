-- Add Configuration Import History Table
-- This migration adds support for tracking configuration import operations

-- Create import status enum
CREATE TYPE import_status_enum AS ENUM ('pending', 'running', 'completed', 'failed', 'partial', 'validating');

-- Create conflict strategy enum  
CREATE TYPE conflict_strategy_enum AS ENUM ('update', 'skip', 'rename', 'fail');

-- Configuration Import History - Track all import operations
CREATE TABLE config_imports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Import metadata
    filename VARCHAR(255) NOT NULL,
    entity_types TEXT[] NOT NULL DEFAULT '{}',
    status import_status_enum DEFAULT 'pending',
    conflict_strategy conflict_strategy_enum DEFAULT 'update',
    dry_run BOOLEAN DEFAULT false,
    
    -- Import results and statistics
    summary JSONB NOT NULL DEFAULT '{
        "total_items": 0,
        "processed_items": 0, 
        "created_items": 0,
        "updated_items": 0,
        "skipped_items": 0,
        "failed_items": 0,
        "entity_counts": {}
    }',
    
    -- Error and warning counts for quick filtering
    error_count INTEGER DEFAULT 0,
    warning_count INTEGER DEFAULT 0,
    
    -- Performance tracking
    duration INTERVAL,
    
    -- Audit information
    imported_by UUID NOT NULL REFERENCES users(id),
    imported_by_name VARCHAR(255), -- Denormalized for performance
    imported_by_email VARCHAR(255), -- Denormalized for performance
    
    -- Additional metadata and context
    metadata JSONB DEFAULT '{}',
    
    -- Detailed results (stored separately for performance)
    details_file_path VARCHAR(500), -- Path to detailed results file
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_filename CHECK (LENGTH(filename) > 0),
    CONSTRAINT valid_entity_types CHECK (array_length(entity_types, 1) > 0),
    CONSTRAINT valid_error_count CHECK (error_count >= 0),
    CONSTRAINT valid_warning_count CHECK (warning_count >= 0),
    CONSTRAINT completed_at_after_created CHECK (completed_at IS NULL OR completed_at >= created_at)
);

-- Add updated_at trigger (reuse existing function)
CREATE TRIGGER config_imports_updated_at 
    BEFORE UPDATE ON config_imports 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Performance indexes for config_imports
CREATE INDEX idx_config_imports_org_status ON config_imports(organization_id, status);
CREATE INDEX idx_config_imports_org_created_at ON config_imports(organization_id, created_at DESC);
CREATE INDEX idx_config_imports_imported_by ON config_imports(imported_by);
CREATE INDEX idx_config_imports_entity_types ON config_imports USING gin(entity_types);
CREATE INDEX idx_config_imports_status ON config_imports(status);
CREATE INDEX idx_config_imports_dry_run ON config_imports(dry_run);
CREATE INDEX idx_config_imports_conflict_strategy ON config_imports(conflict_strategy);
CREATE INDEX idx_config_imports_filename ON config_imports(organization_id, filename);
CREATE INDEX idx_config_imports_errors ON config_imports(error_count) WHERE error_count > 0;
CREATE INDEX idx_config_imports_warnings ON config_imports(warning_count) WHERE warning_count > 0;

-- Add index for filtering by completion status and date range
CREATE INDEX idx_config_imports_completed_status ON config_imports(completed_at DESC, status) 
    WHERE completed_at IS NOT NULL;

-- Configuration Export History - Track export operations (optional, for audit)
CREATE TABLE config_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Export metadata
    entity_types TEXT[] NOT NULL DEFAULT '{}',
    filters JSONB DEFAULT '{}',
    
    -- Export statistics
    total_entities INTEGER DEFAULT 0,
    file_size_bytes BIGINT,
    
    -- Export file information
    filename VARCHAR(255),
    file_path VARCHAR(500), -- Path to exported file
    expires_at TIMESTAMP WITH TIME ZONE, -- When export file expires
    
    -- Audit information
    exported_by UUID NOT NULL REFERENCES users(id),
    exported_by_name VARCHAR(255), -- Denormalized for performance
    exported_by_email VARCHAR(255), -- Denormalized for performance
    
    -- Additional context
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_export_entity_types CHECK (array_length(entity_types, 1) > 0),
    CONSTRAINT valid_total_entities CHECK (total_entities >= 0),
    CONSTRAINT valid_file_size CHECK (file_size_bytes IS NULL OR file_size_bytes >= 0),
    CONSTRAINT valid_expires_at CHECK (expires_at IS NULL OR expires_at > created_at)
);

-- Performance indexes for config_exports
CREATE INDEX idx_config_exports_org_created_at ON config_exports(organization_id, created_at DESC);
CREATE INDEX idx_config_exports_exported_by ON config_exports(exported_by);
CREATE INDEX idx_config_exports_entity_types ON config_exports USING gin(entity_types);
CREATE INDEX idx_config_exports_expires_at ON config_exports(expires_at) WHERE expires_at IS NOT NULL;

-- Add comments for documentation
COMMENT ON TABLE config_imports IS 'Tracks configuration import operations with detailed results and statistics';
COMMENT ON TABLE config_exports IS 'Tracks configuration export operations for audit and cleanup purposes';

COMMENT ON COLUMN config_imports.summary IS 'JSON summary of import results including counts by entity type';
COMMENT ON COLUMN config_imports.metadata IS 'Additional context and configuration options used during import';
COMMENT ON COLUMN config_imports.details_file_path IS 'Path to detailed import results file for large operations';
COMMENT ON COLUMN config_imports.imported_by_name IS 'Cached user name for performance (avoids joins)';
COMMENT ON COLUMN config_imports.imported_by_email IS 'Cached user email for performance (avoids joins)';

COMMENT ON COLUMN config_exports.filters IS 'JSON filters applied during export (tags, date ranges, etc.)';
COMMENT ON COLUMN config_exports.file_path IS 'Path to generated export file';
COMMENT ON COLUMN config_exports.expires_at IS 'When export file should be cleaned up';