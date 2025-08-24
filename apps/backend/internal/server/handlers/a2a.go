package handlers

import (
	"net/http"
	"strconv"
	"time"

	"mcp-gateway/apps/backend/internal/a2a"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// A2AHandler handles A2A (Agent-to-Agent) HTTP endpoints
type A2AHandler struct {
	service *a2a.Service
	client  *a2a.Client
	adapter *a2a.Adapter
}

// NewA2AHandler creates a new A2A handler
func NewA2AHandler(service *a2a.Service, client *a2a.Client, adapter *a2a.Adapter) *A2AHandler {
	return &A2AHandler{
		service: service,
		client:  client,
		adapter: adapter,
	}
}

// ListAgents handles GET /a2a - List agents with filtering
func (h *A2AHandler) ListAgents(c *gin.Context) {
	// Get organization ID from context (set by auth middleware)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Default for single-tenant

	// Parse query parameters for filtering
	filters := make(map[string]interface{})

	if agentType := c.Query("agent_type"); agentType != "" {
		filters["agent_type"] = agentType
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filters["is_active"] = isActive
		}
	}

	if healthStatus := c.Query("health_status"); healthStatus != "" {
		filters["health_status"] = healthStatus
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filters["tags"] = tags
	}

	// Get agents from service
	agents, err := h.service.List(orgID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list agents: " + err.Error(),
		})
		return
	}

	// Convert to specs (hiding sensitive auth data)
	var agentSpecs []*types.A2AAgentSpec
	for _, agent := range agents {
		spec := &types.A2AAgentSpec{
			ID:               agent.ID.String(),
			Name:             agent.Name,
			Description:      agent.Description,
			EndpointURL:      agent.EndpointURL,
			AgentType:        agent.AgentType,
			ProtocolVersion:  agent.ProtocolVersion,
			Capabilities:     agent.CapabilitiesData,
			Config:           agent.ConfigData,
			AuthType:         agent.AuthType,
			Tags:             agent.Tags,
			Metadata:         agent.MetadataData,
			IsActive:         agent.IsActive,
			LastHealthCheck:  agent.LastHealthCheck,
			HealthStatus:     agent.HealthStatus,
			HealthError:      agent.HealthError,
			CreatedAt:        agent.CreatedAt,
			UpdatedAt:        agent.UpdatedAt,
		}
		agentSpecs = append(agentSpecs, spec)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    agentSpecs,
		"count":   len(agentSpecs),
	})
}

// RegisterAgent handles POST /a2a - Register new agent
func (h *A2AHandler) RegisterAgent(c *gin.Context) {
	var spec types.A2AAgentSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if spec.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name is required",
		})
		return
	}

	if spec.EndpointURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Endpoint URL is required",
		})
		return
	}

	// Set defaults
	if spec.AgentType == "" {
		spec.AgentType = types.AgentTypeCustom
	}
	if spec.AuthType == "" {
		spec.AuthType = types.AuthTypeNone
	}
	if spec.ProtocolVersion == "" {
		spec.ProtocolVersion = "1.0"
	}
	spec.IsActive = true // New agents are active by default

	// Create agent
	agent, err := h.service.Create(&spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create agent: " + err.Error(),
		})
		return
	}

	// Convert to spec for response (hide auth value)
	responseSpec := &types.A2AAgentSpec{
		ID:               agent.ID.String(),
		Name:             agent.Name,
		Description:      agent.Description,
		EndpointURL:      agent.EndpointURL,
		AgentType:        agent.AgentType,
		ProtocolVersion:  agent.ProtocolVersion,
		Capabilities:     agent.CapabilitiesData,
		Config:           agent.ConfigData,
		AuthType:         agent.AuthType,
		Tags:             agent.Tags,
		Metadata:         agent.MetadataData,
		IsActive:         agent.IsActive,
		HealthStatus:     agent.HealthStatus,
		CreatedAt:        agent.CreatedAt,
		UpdatedAt:        agent.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    responseSpec,
	})
}

// GetAgent handles GET /a2a/{id} - Get agent details
func (h *A2AHandler) GetAgent(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID format",
		})
		return
	}

	agent, err := h.service.Get(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found: " + err.Error(),
		})
		return
	}

	// Convert to spec (hide auth value)
	spec := &types.A2AAgentSpec{
		ID:               agent.ID.String(),
		Name:             agent.Name,
		Description:      agent.Description,
		EndpointURL:      agent.EndpointURL,
		AgentType:        agent.AgentType,
		ProtocolVersion:  agent.ProtocolVersion,
		Capabilities:     agent.CapabilitiesData,
		Config:           agent.ConfigData,
		AuthType:         agent.AuthType,
		Tags:             agent.Tags,
		Metadata:         agent.MetadataData,
		IsActive:         agent.IsActive,
		LastHealthCheck:  agent.LastHealthCheck,
		HealthStatus:     agent.HealthStatus,
		HealthError:      agent.HealthError,
		CreatedAt:        agent.CreatedAt,
		UpdatedAt:        agent.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spec,
	})
}

// UpdateAgent handles PUT /a2a/{id} - Update agent
func (h *A2AHandler) UpdateAgent(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID format",
		})
		return
	}

	var spec types.A2AAgentSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Update agent
	updatedAgent, err := h.service.Update(agentID, &spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update agent: " + err.Error(),
		})
		return
	}

	// Convert to spec for response (hide auth value)
	responseSpec := &types.A2AAgentSpec{
		ID:               updatedAgent.ID.String(),
		Name:             updatedAgent.Name,
		Description:      updatedAgent.Description,
		EndpointURL:      updatedAgent.EndpointURL,
		AgentType:        updatedAgent.AgentType,
		ProtocolVersion:  updatedAgent.ProtocolVersion,
		Capabilities:     updatedAgent.CapabilitiesData,
		Config:           updatedAgent.ConfigData,
		AuthType:         updatedAgent.AuthType,
		Tags:             updatedAgent.Tags,
		Metadata:         updatedAgent.MetadataData,
		IsActive:         updatedAgent.IsActive,
		LastHealthCheck:  updatedAgent.LastHealthCheck,
		HealthStatus:     updatedAgent.HealthStatus,
		HealthError:      updatedAgent.HealthError,
		CreatedAt:        updatedAgent.CreatedAt,
		UpdatedAt:        updatedAgent.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseSpec,
	})
}

// DeleteAgent handles DELETE /a2a/{id} - Remove agent
func (h *A2AHandler) DeleteAgent(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID format",
		})
		return
	}

	err = h.service.Delete(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete agent: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Agent deleted successfully",
	})
}

// ToggleAgent handles POST /a2a/{id}/toggle - Enable/disable agent
func (h *A2AHandler) ToggleAgent(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID format",
		})
		return
	}

	// Parse request body to get the active status
	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	err = h.service.Toggle(agentID, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to toggle agent: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Agent status updated successfully",
		"is_active": req.IsActive,
	})
}

// InvokeAgent handles POST /a2a/{name}/invoke - Direct invocation
func (h *A2AHandler) InvokeAgent(c *gin.Context) {
	agentName := c.Param("name")
	if agentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Agent name is required",
		})
		return
	}

	// Get agent by name
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Default for single-tenant
	agent, err := h.service.GetByName(orgID, agentName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found: " + err.Error(),
		})
		return
	}

	if !agent.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Agent is not active",
		})
		return
	}

	// Parse request body
	var request types.A2ARequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Set agent ID and protocol version if not provided
	if request.AgentID == "" {
		request.AgentID = agent.ID.String()
	}
	if request.ProtocolVersion == "" {
		request.ProtocolVersion = agent.ProtocolVersion
	}

	// Make the invocation
	start := time.Now()
	response, err := h.client.Invoke(agent, &request)
	duration := int(time.Since(start).Milliseconds())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invocation failed: " + err.Error(),
		})
		return
	}

	// Add timing information
	if response.Usage == nil {
		response.Usage = &types.A2AUsage{}
	}
	response.Usage.Duration = duration

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ChatWithAgent handles POST /a2a/{name}/chat - Chat with an agent
func (h *A2AHandler) ChatWithAgent(c *gin.Context) {
	agentName := c.Param("name")
	if agentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Agent name is required",
		})
		return
	}

	// Get agent by name
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Default for single-tenant
	agent, err := h.service.GetByName(orgID, agentName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found: " + err.Error(),
		})
		return
	}

	if !agent.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Agent is not active",
		})
		return
	}

	// Check if agent supports chat
	if capabilities, ok := agent.CapabilitiesData[types.CapabilityChat].(bool); !ok || !capabilities {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Agent does not support chat functionality",
		})
		return
	}

	// Parse request body
	var request types.A2AChatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate that messages are provided
	if len(request.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least one message is required",
		})
		return
	}

	// Make the chat request
	start := time.Now()
	response, err := h.client.Chat(agent, &request)
	duration := int(time.Since(start).Milliseconds())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Chat request failed: " + err.Error(),
		})
		return
	}

	// Add timing information
	if response.Usage == nil {
		response.Usage = &types.A2AUsage{}
	}
	response.Usage.Duration = duration

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// HealthCheckAgent handles GET /a2a/{id}/health - Check agent health
func (h *A2AHandler) HealthCheckAgent(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID format",
		})
		return
	}

	agent, err := h.service.Get(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found: " + err.Error(),
		})
		return
	}

	// Perform health check
	healthCheck, err := h.client.HealthCheck(agent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Health check failed: " + err.Error(),
		})
		return
	}

	// Update health status in database
	var status types.A2AHealthStatus
	if healthCheck.Status == "healthy" {
		status = types.A2AHealthStatusHealthy
	} else {
		status = types.A2AHealthStatusUnhealthy
	}

	if err := h.service.UpdateHealth(agentID, status, healthCheck.Message); err != nil {
		// Log error but don't fail the request
		// The health check result is still valid
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    healthCheck,
	})
}

// GetAgentTools handles GET /a2a/{id}/tools - Get available tools for an agent
func (h *A2AHandler) GetAgentTools(c *gin.Context) {
	idParam := c.Param("id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID format",
		})
		return
	}

	agent, err := h.service.Get(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found: " + err.Error(),
		})
		return
	}

	// Get tools from adapter
	tools, err := h.adapter.ListTools(agent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list tools: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tools,
		"count":   len(tools),
	})
}

// GetAgentStats handles GET /a2a/stats - Get A2A agent statistics
func (h *A2AHandler) GetAgentStats(c *gin.Context) {
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Default for single-tenant

	stats, err := h.service.Stats(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}