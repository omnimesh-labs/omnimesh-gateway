package discovery

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// HealthChecker manages health checking for MCP servers
type HealthChecker struct {
	registry      *Registry
	config        *Config
	client        *http.Client
	stopCh        chan struct{}
	healthModel   *models.HealthCheckModel
	failureCounts map[string]int
	wg            sync.WaitGroup
	mu            sync.RWMutex
	running       bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(registry *Registry, config *Config, healthModel *models.HealthCheckModel) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		config:   config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh:        make(chan struct{}),
		healthModel:   healthModel,
		failureCounts: make(map[string]int),
	}
}

// Start starts the health checking process
func (h *HealthChecker) Start() error {
	if h.running {
		return nil
	}

	h.running = true
	h.wg.Add(1)

	go h.healthCheckLoop()

	return nil
}

// Stop stops the health checking process
func (h *HealthChecker) Stop() error {
	if !h.running {
		return nil
	}

	close(h.stopCh)
	h.wg.Wait()
	h.running = false

	return nil
}

// CheckHealth performs a health check on a specific server
func (h *HealthChecker) CheckHealth(serverID string) (*types.HealthCheck, error) {
	server, exists := h.registry.GetServer(serverID)
	if !exists {
		return nil, types.NewNotFoundError("Server not found")
	}

	return h.performHealthCheck(server)
}

// healthCheckLoop runs the periodic health checking
func (h *HealthChecker) healthCheckLoop() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.config.HealthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkAllServers()
		case <-h.stopCh:
			return
		}
	}
}

// checkAllServers performs health checks on all registered servers
func (h *HealthChecker) checkAllServers() {
	// Get all servers from registry
	servers := h.registry.getAllServers()

	// Use a worker pool to perform health checks in parallel
	maxWorkers := 10
	serverCh := make(chan *types.MCPServer, len(servers))
	var wg sync.WaitGroup

	// Start worker goroutines
	for range maxWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for server := range serverCh {
				_, err := h.performHealthCheck(server)
				if err != nil {
					log.Printf("Health check failed for server %s (%s): %v", server.Name, server.ID, err)
				}
			}
		}()
	}

	// Send servers to workers
	for _, server := range servers {
		serverCh <- server
	}
	close(serverCh)

	// Wait for all health checks to complete
	wg.Wait()
}

// performHealthCheck performs a single health check
func (h *HealthChecker) performHealthCheck(server *types.MCPServer) (*types.HealthCheck, error) {
	startTime := time.Now()

	healthCheck := &types.HealthCheck{
		ServerID:  server.ID,
		CheckedAt: startTime,
	}

	// Skip health check if server has no URL (e.g., STDIO servers)
	if server.URL == "" && server.HealthCheckURL == "" {
		healthCheck.Status = types.HealthStatusHealthy // Assume STDIO/local servers are healthy
		healthCheck.Latency = 0
		h.updateServerHealth(server.ID, healthCheck)
		return healthCheck, nil
	}

	// Determine health check URL
	healthCheckURL := server.HealthCheckURL
	if healthCheckURL == "" {
		healthCheckURL = server.URL + "/health"
	}

	// Create HTTP request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), server.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthCheckURL, http.NoBody)
	if err != nil {
		healthCheck.Status = types.HealthStatusError
		healthCheck.Error = err.Error()
		healthCheck.Latency = time.Since(startTime).Milliseconds()
		h.updateServerHealth(server.ID, healthCheck)
		h.saveHealthCheck(healthCheck)
		return healthCheck, nil
	}

	// Add User-Agent header
	req.Header.Set("User-Agent", "MCP-Gateway-HealthChecker/1.0")
	req.Header.Set("Accept", "application/json")

	// Perform HTTP request
	resp, err := h.client.Do(req)
	if err != nil {
		// Check if it's a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			healthCheck.Status = types.HealthStatusTimeout
			healthCheck.Error = "Request timeout"
		} else {
			healthCheck.Status = types.HealthStatusError
			healthCheck.Error = err.Error()
		}
		healthCheck.Latency = time.Since(startTime).Milliseconds()
		h.updateServerHealth(server.ID, healthCheck)
		h.saveHealthCheck(healthCheck)
		return healthCheck, nil
	}
	defer resp.Body.Close()

	// Calculate latency
	healthCheck.Latency = time.Since(startTime).Milliseconds()

	// Read response body for additional context
	var responseBody string
	if resp.ContentLength != 0 && resp.ContentLength < 1024 { // Only read small responses
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		responseBody = string(buf[:n])
	}

	// Determine health status based on HTTP status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		healthCheck.Status = types.HealthStatusHealthy
		healthCheck.Response = fmt.Sprintf("%d %s", resp.StatusCode, resp.Status)
	} else if resp.StatusCode >= 500 {
		healthCheck.Status = types.HealthStatusError
		healthCheck.Response = fmt.Sprintf("%d %s", resp.StatusCode, resp.Status)
		healthCheck.Error = responseBody
	} else {
		healthCheck.Status = types.HealthStatusUnhealthy
		healthCheck.Response = fmt.Sprintf("%d %s", resp.StatusCode, resp.Status)
		healthCheck.Error = responseBody
	}

	// Update server status based on health check result
	h.updateServerHealth(server.ID, healthCheck)

	// Save health check to database
	h.saveHealthCheck(healthCheck)

	return healthCheck, nil
}

// updateServerHealth updates server health status
func (h *HealthChecker) updateServerHealth(serverID string, healthCheck *types.HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch healthCheck.Status {
	case types.HealthStatusHealthy:
		// Reset failure count on successful health check
		h.failureCounts[serverID] = 0
		h.registry.UpdateServerStatus(serverID, types.ServerStatusActive)

	case types.HealthStatusUnhealthy, types.HealthStatusTimeout, types.HealthStatusError:
		// Increment failure count
		h.failureCounts[serverID]++

		// Check failure threshold before marking as unhealthy
		if h.failureCounts[serverID] >= h.config.FailureThreshold {
			log.Printf("Server %s marked as unhealthy after %d consecutive failures", serverID, h.failureCounts[serverID])
			h.registry.UpdateServerStatus(serverID, types.ServerStatusUnhealthy)
		} else {
			log.Printf("Server %s health check failed (%d/%d), status unchanged", serverID, h.failureCounts[serverID], h.config.FailureThreshold)
		}
	}
}

// GetHealthHistory returns health check history for a server
func (h *HealthChecker) GetHealthHistory(serverID string, limit int) ([]*types.HealthCheck, error) {
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %w", err)
	}

	// Retrieve health checks from database
	checks, err := h.healthModel.GetHistoryByServerID(serverUUID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve health history: %w", err)
	}

	// Convert from models.HealthCheck to types.HealthCheck
	result := make([]*types.HealthCheck, len(checks))
	for i, check := range checks {
		result[i] = &types.HealthCheck{
			ID:        check.ID.String(),
			ServerID:  check.ServerID.String(),
			Status:    check.Status,
			Response:  check.ResponseBody.String,
			Error:     check.ErrorMessage.String,
			Latency:   int64(check.ResponseTimeMS.Int32),
			CheckedAt: check.CheckedAt,
		}
	}

	return result, nil
}

// GetServerHealth returns current health status for a server
func (h *HealthChecker) GetServerHealth(serverID string) (*types.HealthCheck, error) {
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %w", err)
	}

	// Retrieve latest health check from database
	check, err := h.healthModel.GetLatestByServerID(serverUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no health checks found for server %s", serverID)
		}
		return nil, fmt.Errorf("failed to retrieve current health status: %w", err)
	}

	// Convert from models.HealthCheck to types.HealthCheck
	return &types.HealthCheck{
		ID:        check.ID.String(),
		ServerID:  check.ServerID.String(),
		Status:    check.Status,
		Response:  check.ResponseBody.String,
		Error:     check.ErrorMessage.String,
		Latency:   int64(check.ResponseTimeMS.Int32),
		CheckedAt: check.CheckedAt,
	}, nil
}

// saveHealthCheck saves a health check result to the database
func (h *HealthChecker) saveHealthCheck(healthCheck *types.HealthCheck) {
	serverUUID, err := uuid.Parse(healthCheck.ServerID)
	if err != nil {
		log.Printf("Invalid server ID in health check: %v", err)
		return
	}

	// Convert types.HealthCheck to models.HealthCheck
	check := &models.HealthCheck{
		ServerID:  serverUUID,
		Status:    healthCheck.Status,
		CheckedAt: healthCheck.CheckedAt,
	}

	// Set optional fields if they exist
	if healthCheck.Latency > 0 {
		check.ResponseTimeMS = sql.NullInt32{Int32: int32(healthCheck.Latency), Valid: true}
	}
	if healthCheck.Response != "" {
		check.ResponseBody = sql.NullString{String: healthCheck.Response, Valid: true}
	}
	if healthCheck.Error != "" {
		check.ErrorMessage = sql.NullString{String: healthCheck.Error, Valid: true}
	}

	// Save to database
	err = h.healthModel.Create(check)
	if err != nil {
		log.Printf("Failed to save health check for server %s: %v", healthCheck.ServerID, err)
	}
}
