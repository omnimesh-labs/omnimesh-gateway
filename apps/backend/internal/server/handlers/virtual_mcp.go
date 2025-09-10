package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/types"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/virtual"

	"github.com/gin-gonic/gin"
)

// VirtualMCPHandler handles MCP JSON-RPC requests for virtual servers
type VirtualMCPHandler struct {
	virtualService *virtual.Service
	servers        map[string]*virtual.VirtualServer // Cache of virtual server instances
}

// NewVirtualMCPHandler creates a new virtual MCP handler
func NewVirtualMCPHandler(virtualService *virtual.Service) *VirtualMCPHandler {
	return &VirtualMCPHandler{
		virtualService: virtualService,
		servers:        make(map[string]*virtual.VirtualServer),
	}
}

// HandleMCPRPC handles MCP JSON-RPC 2.0 requests
func (h *VirtualMCPHandler) HandleMCPRPC(c *gin.Context) {
	// Parse JSON-RPC request
	var rpcRequest types.VirtualMCPRequest
	if err := c.ShouldBindJSON(&rpcRequest); err != nil {
		h.sendErrorResponse(c, nil, types.ParseError, "Parse error", err.Error())
		return
	}

	// Validate JSON-RPC format
	if rpcRequest.JSONRPC != "2.0" {
		h.sendErrorResponse(c, rpcRequest.ID, types.InvalidRequest, "Invalid Request", "JSONRPC version must be 2.0")
		return
	}

	if rpcRequest.Method == "" {
		h.sendErrorResponse(c, rpcRequest.ID, types.InvalidRequest, "Invalid Request", "Method is required")
		return
	}

	// Handle the request
	h.handleMethod(c, &rpcRequest)
}

// handleMethod dispatches the request to the appropriate handler
func (h *VirtualMCPHandler) handleMethod(c *gin.Context, req *types.VirtualMCPRequest) {
	switch req.Method {
	case "initialize":
		h.handleInitialize(c, req)
	case "notifications/initialized":
		h.handleInitialized(c, req)
	case "tools/list":
		h.handleToolsList(c, req)
	case "tools/call":
		h.handleToolsCall(c, req)
	default:
		h.sendErrorResponse(c, req.ID, types.MethodNotFound, "Method not found", fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

// handleInitialize handles the initialize method
func (h *VirtualMCPHandler) handleInitialize(c *gin.Context, req *types.VirtualMCPRequest) {
	// Parse initialize parameters
	var params types.InitializeParams
	if req.Params != nil {
		paramsBytes, err := json.Marshal(req.Params)
		if err != nil {
			h.sendErrorResponse(c, req.ID, types.InvalidParams, "Invalid params", err.Error())
			return
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			h.sendErrorResponse(c, req.ID, types.InvalidParams, "Invalid params", err.Error())
			return
		}
	}

	// Get virtual server by ID
	virtualServer, err := h.getVirtualServer(params.ServerID)
	if err != nil {
		h.sendErrorResponse(c, req.ID, types.ServerError, "Server error", err.Error())
		return
	}

	// Call initialize on the virtual server
	result, err := virtualServer.Initialize(params)
	if err != nil {
		h.sendErrorResponse(c, req.ID, types.ServerError, "Initialize failed", err.Error())
		return
	}

	h.sendSuccessResponse(c, req.ID, result)
}

// handleInitialized handles the notifications/initialized method
func (h *VirtualMCPHandler) handleInitialized(c *gin.Context, req *types.VirtualMCPRequest) {
	// This is a notification, so no response is expected
	c.Status(http.StatusOK)
}

// handleToolsList handles the tools/list method
func (h *VirtualMCPHandler) handleToolsList(c *gin.Context, req *types.VirtualMCPRequest) {
	// Parse parameters to get server ID
	var params types.ListToolsParams
	if req.Params != nil {
		paramsBytes, err := json.Marshal(req.Params)
		if err != nil {
			h.sendErrorResponse(c, req.ID, types.InvalidParams, "Invalid params", err.Error())
			return
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			h.sendErrorResponse(c, req.ID, types.InvalidParams, "Invalid params", err.Error())
			return
		}
	}

	// Get virtual server by ID
	virtualServer, err := h.getVirtualServer(params.ServerID)
	if err != nil {
		h.sendErrorResponse(c, req.ID, types.ServerError, "Server error", err.Error())
		return
	}

	// Call ListTools on the virtual server
	result, err := virtualServer.ListTools()
	if err != nil {
		h.sendErrorResponse(c, req.ID, types.ServerError, "ListTools failed", err.Error())
		return
	}

	h.sendSuccessResponse(c, req.ID, result)
}

// handleToolsCall handles the tools/call method
func (h *VirtualMCPHandler) handleToolsCall(c *gin.Context, req *types.VirtualMCPRequest) {
	// Parse call tool parameters
	var params types.CallToolParams
	if req.Params != nil {
		paramsBytes, err := json.Marshal(req.Params)
		if err != nil {
			h.sendErrorResponse(c, req.ID, types.InvalidParams, "Invalid params", err.Error())
			return
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			h.sendErrorResponse(c, req.ID, types.InvalidParams, "Invalid params", err.Error())
			return
		}
	}

	// Validate required parameters
	if params.Name == "" {
		h.sendErrorResponse(c, req.ID, types.InvalidParams, "Invalid params", "Tool name is required")
		return
	}

	// Get virtual server by ID
	virtualServer, err := h.getVirtualServer(params.ServerID)
	if err != nil {
		h.sendErrorResponse(c, req.ID, types.ServerError, "Server error", err.Error())
		return
	}

	// Call the tool on the virtual server
	result, err := virtualServer.CallTool(params.Name, params.Arguments)
	if err != nil {
		h.sendErrorResponse(c, req.ID, types.ServerError, "Tool call failed", err.Error())
		return
	}

	h.sendSuccessResponse(c, req.ID, result)
}

// getVirtualServer retrieves or creates a virtual server instance
func (h *VirtualMCPHandler) getVirtualServer(serverID string) (*virtual.VirtualServer, error) {
	// If no server ID specified, try to get the default one
	if serverID == "" {
		// List servers and use the first one as default
		specs, err := h.virtualService.List()
		if err != nil {
			return nil, fmt.Errorf("failed to list virtual servers: %w", err)
		}
		if len(specs) == 0 {
			return nil, fmt.Errorf("no virtual servers available")
		}
		serverID = specs[0].ID
	}

	// Check if we already have an instance
	if server, exists := h.servers[serverID]; exists {
		return server, nil
	}

	// Get server spec
	spec, err := h.virtualService.Get(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual server spec: %w", err)
	}

	// Create new virtual server instance
	server := virtual.NewVirtualServer(spec)
	h.servers[serverID] = server

	return server, nil
}

// sendSuccessResponse sends a successful JSON-RPC response
func (h *VirtualMCPHandler) sendSuccessResponse(c *gin.Context, id interface{}, result interface{}) {
	response := types.VirtualMCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	c.JSON(http.StatusOK, response)
}

// sendErrorResponse sends an error JSON-RPC response
func (h *VirtualMCPHandler) sendErrorResponse(c *gin.Context, id interface{}, code int, message, data string) {
	response := types.VirtualMCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &types.VirtualMCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	// Determine HTTP status code based on JSON-RPC error code
	httpStatus := http.StatusOK // JSON-RPC errors are typically sent as 200 OK
	if code == types.ParseError {
		httpStatus = http.StatusBadRequest
	}

	c.JSON(httpStatus, response)
}
