# MCP Gateway - Virtual Servers Testing Guide

This guide provides examples for testing the Virtual Server feature in the MCP Gateway, which allows virtualizing non-MCP services as MCP-compatible servers.

## Overview

The Virtual Server feature enables you to:
- Wrap REST APIs as MCP-compatible servers
- Expose tools, prompts, and resources from non-MCP services
- Use JSON-RPC 2.0 over HTTP to interact with virtual servers
- Manage virtual servers through admin REST endpoints

## Prerequisites

1. Start the MCP Gateway server:
```bash
make run
```

2. Ensure the database migrations are up to date:
```bash
make migrate
```

3. The server should be running on `http://localhost:8080`

## 1. Admin REST API - Virtual Server Management

### List All Virtual Servers
```bash
curl -X GET http://localhost:8080/api/admin/virtual-servers
```

### Get Specific Virtual Server
```bash
curl -X GET http://localhost:8080/api/admin/virtual-servers/{server_id}
```

### Create a Simple Test Virtual Server
```bash
curl -X POST http://localhost:8080/api/admin/virtual-servers \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_server",
    "name": "Test Server",
    "description": "Test virtual server",
    "adapterType": "REST",
    "tools": [
      {
        "name": "test_tool",
        "description": "A test tool",
        "inputSchema": {
          "type": "object",
          "properties": {
            "message": {
              "type": "string"
            }
          }
        }
      }
    ]
  }'
```

### Create a Slack Virtual Server with REST Configuration
```bash
curl -X POST http://localhost:8080/api/admin/virtual-servers \
  -H "Content-Type: application/json" \
  -d '{
    "id": "slack_server",
    "name": "Slack API Server",
    "description": "Slack REST API wrapped as MCP server",
    "adapterType": "REST",
    "tools": [
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
    ]
  }'
```

### Update Virtual Server
```bash
curl -X PUT http://localhost:8080/api/admin/virtual-servers/slack_server \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Slack Server",
    "description": "Updated Slack REST API wrapper",
    "adapterType": "REST",
    "tools": [
      {
        "name": "ping",
        "description": "Ping the Slack API",
        "inputSchema": {
          "type": "object",
          "properties": {}
        }
      }
    ]
  }'
```

### Delete Virtual Server
```bash
curl -X DELETE http://localhost:8080/api/admin/virtual-servers/test_server
```

### Get Virtual Server Tools
```bash
curl -X GET http://localhost:8080/api/admin/virtual-servers/slack_server/tools
```

### Test Virtual Server Tool
```bash
curl -X POST http://localhost:8080/api/admin/virtual-servers/slack_server/tools/list_channels/test \
  -H "Content-Type: application/json" \
  -d '{}'
```

## 2. MCP JSON-RPC 2.0 API

### Initialize Virtual Server
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "initialize",
    "params": {
      "server_id": "slack_server",
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "MCP Test Client",
        "version": "1.0.0"
      }
    }
  }'
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {
        "listChanged": false
      }
    },
    "serverInfo": {
      "name": "Slack API Server",
      "version": "1.0.0"
    }
  }
}
```

### List Tools
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "2",
    "method": "tools/list",
    "params": {
      "server_id": "slack_server"
    }
  }'
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": "2",
  "result": {
    "tools": [
      {
        "name": "list_channels",
        "description": "List public channels",
        "inputSchema": {
          "type": "object",
          "properties": {},
          "additionalProperties": false
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
        }
      }
    ]
  }
}
```

### List Tools Without Server ID (uses default)
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "3",
    "method": "tools/list",
    "params": {}
  }'
```

### Call Tool - List Channels
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "4",
    "method": "tools/call",
    "params": {
      "server_id": "slack_server",
      "name": "list_channels",
      "arguments": {}
    }
  }'
```

Expected response (mock data):
```json
{
  "jsonrpc": "2.0",
  "id": "4",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "map[data:map[channels:[map[id:C1234567890 name:general type:public] map[id:C0987654321 name:random type:public]]] mock:true status:success url:https://slack.com/api/conversations.list]"
      }
    ]
  }
}
```

### Call Tool - Send Message
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "5",
    "method": "tools/call",
    "params": {
      "server_id": "slack_server",
      "name": "send_message",
      "arguments": {
        "channel": "#general",
        "text": "Hello from MCP Gateway!"
      }
    }
  }'
```

Expected response (mock data):
```json
{
  "jsonrpc": "2.0",
  "id": "5",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "map[data:map[channel:#general message_id:1234567890.123456 text:Hello from MCP Gateway! timestamp:1755827948] mock:true status:success url:https://slack.com/api/chat.postMessage]"
      }
    ]
  }
}
```

### Send Initialized Notification
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "notifications/initialized",
    "params": {
      "server_id": "slack_server"
    }
  }'
```

## 3. Error Handling Examples

### Invalid JSON-RPC Request
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "invalid": "request"
  }'
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32700,
    "message": "Parse error",
    "data": "error details..."
  }
}
```

### Method Not Found
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "6",
    "method": "unknown/method",
    "params": {}
  }'
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": "6",
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": "Unknown method: unknown/method"
  }
}
```

### Invalid Parameters
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "7",
    "method": "tools/call",
    "params": {
      "server_id": "slack_server"
    }
  }'
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": "7",
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Tool name is required"
  }
}
```

### Server Not Found
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "8",
    "method": "tools/list",
    "params": {
      "server_id": "nonexistent_server"
    }
  }'
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": "8",
  "error": {
    "code": -32000,
    "message": "Server error",
    "data": "failed to get virtual server spec: virtual server not found: nonexistent_server"
  }
}
```

## 4. Advanced Examples

### Create GitHub Virtual Server
```bash
curl -X POST http://localhost:8080/api/admin/virtual-servers \
  -H "Content-Type: application/json" \
  -d '{
    "id": "github_server",
    "name": "GitHub API Server",
    "description": "GitHub REST API wrapped as MCP server",
    "adapterType": "REST",
    "tools": [
      {
        "name": "list_repositories",
        "description": "List user repositories",
        "inputSchema": {
          "type": "object",
          "properties": {
            "username": {
              "type": "string",
              "description": "GitHub username"
            }
          },
          "required": ["username"]
        },
        "REST": {
          "method": "GET",
          "URLTemplate": "https://api.github.com/users/{username}/repos",
          "headers": {
            "Accept": "application/vnd.github.v3+json",
            "User-Agent": "MCP-Gateway/1.0"
          },
          "timeoutSec": 30
        }
      },
      {
        "name": "create_issue",
        "description": "Create a new issue in a repository",
        "inputSchema": {
          "type": "object",
          "properties": {
            "owner": {
              "type": "string"
            },
            "repo": {
              "type": "string"
            },
            "title": {
              "type": "string"
            },
            "body": {
              "type": "string"
            }
          },
          "required": ["owner", "repo", "title"]
        },
        "REST": {
          "method": "POST",
          "URLTemplate": "https://api.github.com/repos/{owner}/{repo}/issues",
          "headers": {
            "Accept": "application/vnd.github.v3+json",
            "Authorization": "token ${SECRET:GITHUB_TOKEN}",
            "Content-Type": "application/json"
          },
          "bodyMap": {
            "title": "title",
            "body": "body"
          },
          "timeoutSec": 30
        }
      }
    ]
  }'
```

### Test GitHub List Repositories
```bash
curl -X POST http://localhost:8080/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "9",
    "method": "tools/call",
    "params": {
      "server_id": "github_server",
      "name": "list_repositories",
      "arguments": {
        "username": "octocat"
      }
    }
  }'
```

## 5. Database Schema

The virtual servers are stored in PostgreSQL with the following structure:

```sql
-- View virtual servers table
SELECT id, name, description, adapter_type, tools, is_active, created_at, updated_at 
FROM virtual_servers;

-- Check tools for a specific server
SELECT name, tools::jsonb 
FROM virtual_servers 
WHERE name = 'Slack API Server';
```

## 6. Architecture Notes

### Components Created:
1. **Types** (`apps/backend/internal/types/virtual.go`):
   - `VirtualServerSpec` - Configuration for virtual servers
   - `ToolDef` - Tool definitions with REST specifications
   - `VirtualMCPRequest/Response` - JSON-RPC 2.0 structures

2. **Database Models** (`apps/backend/internal/database/models/virtual_server.go`):
   - CRUD operations for virtual servers
   - JSON marshaling/unmarshaling for tools

3. **Service Layer** (`apps/backend/internal/virtual/service.go`):
   - In-memory caching with database persistence
   - Virtual server registry management

4. **Adapters** (`apps/backend/internal/virtual/adapter.go`):
   - REST adapter for HTTP API calls
   - Pluggable design for other protocols (GraphQL, gRPC, SOAP)

5. **Handlers**:
   - **Admin API** (`apps/backend/internal/server/handlers/virtual_admin.go`): REST endpoints for management
   - **MCP RPC** (`apps/backend/internal/server/handlers/virtual_mcp.go`): JSON-RPC 2.0 MCP protocol

### Endpoints:
- **Admin REST API**: `/api/admin/virtual-servers/*`
- **MCP JSON-RPC**: `/mcp/rpc` (POST)

### Features:
- ✅ JSON-RPC 2.0 over HTTP
- ✅ MCP methods: `initialize`, `notifications/initialized`, `tools/list`, `tools/call`
- ✅ In-memory registry with database persistence
- ✅ REST adapter with mock responses
- ✅ Error mapping to JSON-RPC error codes
- ✅ Server selection by `server_id` parameter
- ✅ PostgreSQL storage with migration
- ✅ Admin CRUD operations

### Next Steps for Production:
1. Implement real HTTP calls in REST adapter
2. Add secret management for tokens
3. Add authentication/authorization
4. Implement other adapter types (GraphQL, gRPC)
5. Add resource and prompt support
6. Add comprehensive logging and metrics

## Tips for Testing

1. **Server IDs**: Use the returned server ID from the create operation or list all servers to get IDs.

2. **Mock vs Real**: The current implementation returns mock data for tool calls. This is indicated by `"mock": true` in responses.

3. **Error Codes**: JSON-RPC 2.0 error codes:
   - `-32700`: Parse error
   - `-32600`: Invalid Request
   - `-32601`: Method not found
   - `-32602`: Invalid params
   - `-32603`: Internal error
   - `-32000`: Server error

4. **Database**: Virtual servers are persisted in PostgreSQL and cached in memory for performance.

5. **Tools Configuration**: Tools without REST configuration will return errors when called.
