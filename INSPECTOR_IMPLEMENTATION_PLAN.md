# MCP Inspector Implementation Plan

Based on analysis of MetaMCP's inspector implementation and the existing MCP Gateway architecture, here's the comprehensive plan to implement the MCP Inspector feature:

## Backend Implementation

### 1. Inspector Service (`internal/inspector/service.go`)
- Create core inspector service that manages MCP connections and sessions
- Handle session lifecycle (create, execute requests, cleanup)
- Support all MCP methods: tools/list, tools/call, resources/list, resources/read, prompts/list, prompts/get, ping, completion/complete
- Implement real-time event streaming via SSE/WebSocket

### 2. Session Management (`internal/inspector/session.go`)
- Create inspector sessions with unique IDs
- Track server connections and capabilities
- Manage session cleanup and timeouts
- Store active sessions in memory with proper cleanup

### 3. API Handlers (`internal/server/handlers/inspector_handlers.go`)
- REST endpoints for session management (create, get, delete)
- Execute inspector requests endpoint
- SSE endpoint for real-time event streaming
- WebSocket handler for bidirectional communication
- Server capabilities endpoint

### 4. Integration with Existing Transport Layer
- Leverage existing transport implementations (JSON-RPC, SSE, WebSocket, STDIO)
- Use TransportManager to coordinate connections
- Integrate with existing session management

## Frontend Implementation

### 1. Main Inspector Page (`src/app/inspector/page.tsx`)
- Server selection dropdown
- Connection status management
- Integration with existing server list
- Session management UI

### 2. Inspector Components (`src/components/inspector/`)
- **Inspector.tsx**: Main tabbed interface
- **InspectorTools.tsx**: Tool discovery and execution with dynamic forms
- **InspectorResources.tsx**: Resource browsing and reading
- **InspectorPrompts.tsx**: Prompt management
- **InspectorPing.tsx**: Connection testing
- **InspectorRoots.tsx**: Root directory exploration
- **InspectorSampling.tsx**: Sampling/completion testing

### 3. Connection Hook (`src/hooks/useInspectorConnection.ts`)
- WebSocket/SSE connection management
- Request/response handling
- Event streaming
- Error handling and reconnection logic

### 4. UI Components
- Execution history display
- JSON visualization for complex data
- Real-time status indicators
- Error handling displays

## Routing & Middleware

### 1. Backend Routes
```go
// Inspector routes
inspector := r.Group("/api/inspector")
inspector.Use(RequireAuth())
{
    inspector.POST("/sessions", h.CreateInspectorSession)
    inspector.GET("/sessions/:id", h.GetInspectorSession)
    inspector.DELETE("/sessions/:id", h.CloseInspectorSession)
    inspector.POST("/sessions/:id/request", h.ExecuteInspectorRequest)
    inspector.GET("/sessions/:id/events", h.StreamInspectorEvents)
    inspector.GET("/sessions/:id/ws", h.HandleInspectorWebSocket)
    inspector.GET("/servers/:id/capabilities", h.GetServerCapabilities)
}
```

### 2. Frontend Routes
- Add `/inspector` route to the main navigation
- Protected route requiring authentication

## Key Features to Implement

### 1. Dynamic Tool Execution
- Parse tool schemas to generate input forms
- Support various data types (string, number, boolean, object, array)
- Validate inputs before execution
- Display execution results with proper formatting

### 2. Real-time Updates
- SSE/WebSocket for live event streaming
- Display notifications and stderr output
- Show connection status changes
- Track request/response pairs

### 3. Session Management
- Create sessions per server connection
- Automatic cleanup on disconnect
- Session timeout handling
- Multiple concurrent sessions support

### 4. Error Handling
- Graceful error recovery
- User-friendly error messages
- Automatic reconnection attempts
- Timeout management

## Implementation Details from MetaMCP Analysis

### Key Components from MetaMCP

1. **Frontend Connection Hook** (`useConnection.ts`)
   - Uses MCP SDK client libraries
   - Supports SSE and StreamableHTTP transports
   - Manages session lifecycle
   - Handles real-time notifications

2. **Inspector UI Components**
   - Tabbed interface for different MCP capabilities
   - Dynamic form generation based on tool schemas
   - Execution history tracking
   - Real-time status updates

3. **Backend Proxy Layer** (`metamcp.ts`)
   - Session-based connection management
   - Transport multiplexing
   - Automatic cleanup on disconnect
   - Support for multiple concurrent sessions

### MCP Protocol Methods to Support

1. **Tools**
   - `tools/list` - List available tools
   - `tools/call` - Execute a tool with arguments

2. **Resources**
   - `resources/list` - List available resources
   - `resources/read` - Read resource content

3. **Prompts**
   - `prompts/list` - List available prompts
   - `prompts/get` - Get prompt details

4. **Core**
   - `ping` - Test connection
   - `initialize` - Initialize connection with capabilities

5. **Sampling** (if supported)
   - `completion/complete` - Request completions

## Database Considerations
- No database changes required
- Inspector uses in-memory session management
- Leverages existing MCP server configurations from database
- Session data is ephemeral and not persisted

## Testing Strategy

### 1. Unit Tests
- Inspector service methods
- Session management logic
- Request/response handling
- Error scenarios

### 2. Integration Tests
- Session creation and cleanup
- MCP method execution
- Event streaming
- Transport layer integration

### 3. Frontend Tests
- Component rendering
- User interactions
- Connection management
- Error handling

### 4. End-to-End Tests
- Complete inspector workflow
- Tool execution flow
- Resource browsing
- Real-time updates

## Security Considerations

1. **Authentication**
   - Require authentication for all inspector endpoints
   - Validate user permissions for server access

2. **Input Validation**
   - Sanitize tool arguments
   - Prevent injection attacks
   - Validate JSON schemas

3. **Session Security**
   - Use secure session tokens
   - Implement session expiration
   - Prevent session hijacking

## Performance Considerations

1. **Session Management**
   - Limit concurrent sessions per user
   - Automatic cleanup after inactivity
   - Connection pooling for MCP servers

2. **Event Streaming**
   - Buffer events to prevent memory issues
   - Implement backpressure handling
   - Clean up old events periodically

3. **Request Handling**
   - Implement request timeouts
   - Queue management for concurrent requests
   - Rate limiting per session

## Success Metrics

- [ ] Inspector can connect to any MCP server
- [ ] All MCP methods are supported
- [ ] Real-time updates work reliably
- [ ] Tool execution with complex arguments works
- [ ] <200ms latency for most operations
- [ ] Graceful error handling and recovery
- [ ] Support 100+ concurrent inspector sessions

## Implementation Priority

1. **Phase 1: Core Backend**
   - Inspector service
   - Session management
   - Basic API endpoints

2. **Phase 2: Frontend UI**
   - Main inspector page
   - Connection management
   - Basic tool execution

3. **Phase 3: Full Features**
   - All MCP methods
   - Real-time streaming
   - Advanced UI components

4. **Phase 4: Polish**
   - Error handling
   - Performance optimization
   - Testing coverage