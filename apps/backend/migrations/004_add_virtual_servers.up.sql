-- Add virtual servers table for virtualizing non-MCP services as MCP-compatible servers
-- This migration adds support for wrapping REST APIs and other services as virtual MCP servers

-- Add adapter type enum for virtual servers
CREATE TYPE adapter_type_enum AS ENUM ('REST', 'GraphQL', 'gRPC', 'SOAP');

-- Virtual Servers - Configuration for virtual MCP servers that wrap non-MCP services
CREATE TABLE virtual_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Basic information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    adapter_type adapter_type_enum NOT NULL DEFAULT 'REST',
    
    -- Tool definitions stored as JSONB for flexibility
    tools JSONB NOT NULL DEFAULT '[]',
    
    -- Status and metadata
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(organization_id, name),
    CONSTRAINT valid_tools_format CHECK (
        jsonb_typeof(tools) = 'array'
    )
);

-- Trigger for updated_at
CREATE TRIGGER virtual_servers_updated_at 
    BEFORE UPDATE ON virtual_servers 
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Performance indexes
CREATE INDEX idx_virtual_servers_org_active ON virtual_servers(organization_id, is_active);
CREATE INDEX idx_virtual_servers_adapter_type ON virtual_servers(adapter_type);
CREATE INDEX idx_virtual_servers_name ON virtual_servers(organization_id, name);

-- Add virtual server support to audit logs
-- Allow tracking virtual server operations
ALTER TABLE audit_logs ADD COLUMN virtual_server_id UUID REFERENCES virtual_servers(id) ON DELETE SET NULL;
CREATE INDEX idx_audit_logs_virtual_server ON audit_logs(virtual_server_id, created_at DESC) WHERE virtual_server_id IS NOT NULL;

-- Insert example Slack virtual server for testing
INSERT INTO virtual_servers (
    organization_id, 
    name, 
    description, 
    adapter_type, 
    tools,
    metadata
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    'Slack Server',
    'Slack REST API wrapped as MCP server',
    'REST',
    '[
        {
            "name": "list_channels",
            "description": "List public channels",
            "inputSchema": {
                "type": "object",
                "properties": {},
                "additionalProperties": false
            },
            "REST": {
                "method": "GET",
                "URLTemplate": "https://slack.com/api/conversations.list",
                "headers": {
                    "Accept": "application/json"
                },
                "auth": {
                    "type": "Bearer",
                    "token": "${SECRET:SLACK_BOT_TOKEN}"
                },
                "timeoutSec": 15
            }
        },
        {
            "name": "send_message",
            "description": "Post a message to a channel",
            "inputSchema": {
                "type": "object",
                "required": ["channel", "text"],
                "properties": {
                    "channel": {
                        "type": "string"
                    },
                    "text": {
                        "type": "string"
                    }
                }
            },
            "REST": {
                "method": "POST",
                "URLTemplate": "https://slack.com/api/chat.postMessage",
                "headers": {
                    "Content-Type": "application/json"
                },
                "auth": {
                    "type": "Bearer",
                    "token": "${SECRET:SLACK_BOT_TOKEN}"
                },
                "bodyMap": {
                    "channel": "channel",
                    "text": "text"
                },
                "timeoutSec": 15
            }
        }
    ]'::jsonb,
    '{"example": true}'::jsonb
) ON CONFLICT (organization_id, name) DO NOTHING;
