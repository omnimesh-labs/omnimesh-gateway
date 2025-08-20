package gateway

import (
	"math/rand"
	"sync"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

// LoadBalancer interface defines load balancing operations
type LoadBalancer interface {
	SelectServer(servers []*types.MCPServer) (*types.MCPServer, error)
	UpdateStats(serverID string, success bool, latency time.Duration)
}

// RoundRobinBalancer implements round-robin load balancing
type RoundRobinBalancer struct {
	mu      sync.Mutex
	current int
}

// NewRoundRobinBalancer creates a new round-robin load balancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{
		current: 0,
	}
}

// SelectServer selects a server using round-robin algorithm
func (rb *RoundRobinBalancer) SelectServer(servers []*types.MCPServer) (*types.MCPServer, error) {
	if len(servers) == 0 {
		return nil, types.NewServiceUnavailableError("No servers available")
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	server := servers[rb.current%len(servers)]
	rb.current++

	return server, nil
}

// UpdateStats updates server statistics (no-op for round-robin)
func (rb *RoundRobinBalancer) UpdateStats(serverID string, success bool, latency time.Duration) {
	// Round-robin doesn't use stats
}

// LeastConnectionsBalancer implements least connections load balancing
type LeastConnectionsBalancer struct {
	mu          sync.RWMutex
	connections map[string]int
}

// NewLeastConnectionsBalancer creates a new least connections load balancer
func NewLeastConnectionsBalancer() *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{
		connections: make(map[string]int),
	}
}

// SelectServer selects a server with the least connections
func (lb *LeastConnectionsBalancer) SelectServer(servers []*types.MCPServer) (*types.MCPServer, error) {
	if len(servers) == 0 {
		return nil, types.NewServiceUnavailableError("No servers available")
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var selected *types.MCPServer
	minConnections := int(^uint(0) >> 1) // Max int

	for _, server := range servers {
		connections := lb.connections[server.ID]
		if connections < minConnections {
			minConnections = connections
			selected = server
		}
	}

	if selected != nil {
		lb.mu.RUnlock()
		lb.mu.Lock()
		lb.connections[selected.ID]++
		lb.mu.Unlock()
		lb.mu.RLock()
	}

	return selected, nil
}

// UpdateStats updates connection count for a server
func (lb *LeastConnectionsBalancer) UpdateStats(serverID string, success bool, latency time.Duration) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.connections[serverID] > 0 {
		lb.connections[serverID]--
	}
}

// WeightedBalancer implements weighted load balancing
type WeightedBalancer struct {
	mu    sync.Mutex
	stats map[string]*ServerStats
}

// ServerStats holds server statistics for weighted balancing
type ServerStats struct {
	Weight      int
	Connections int
	AvgLatency  time.Duration
	ErrorRate   float64
}

// NewWeightedBalancer creates a new weighted load balancer
func NewWeightedBalancer() *WeightedBalancer {
	return &WeightedBalancer{
		stats: make(map[string]*ServerStats),
	}
}

// SelectServer selects a server based on weights and performance
func (wb *WeightedBalancer) SelectServer(servers []*types.MCPServer) (*types.MCPServer, error) {
	if len(servers) == 0 {
		return nil, types.NewServiceUnavailableError("No servers available")
	}

	wb.mu.Lock()
	defer wb.mu.Unlock()

	// Calculate total weight
	totalWeight := 0
	for _, server := range servers {
		weight := wb.getEffectiveWeight(server)
		totalWeight += weight
	}

	if totalWeight == 0 {
		// Fallback to random selection
		return servers[rand.Intn(len(servers))], nil
	}

	// Select based on weight
	target := rand.Intn(totalWeight)
	current := 0

	for _, server := range servers {
		weight := wb.getEffectiveWeight(server)
		current += weight
		if current > target {
			return server, nil
		}
	}

	// Fallback
	return servers[0], nil
}

// UpdateStats updates server statistics
func (wb *WeightedBalancer) UpdateStats(serverID string, success bool, latency time.Duration) {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	stats, exists := wb.stats[serverID]
	if !exists {
		stats = &ServerStats{
			Weight: 100, // Default weight
		}
		wb.stats[serverID] = stats
	}

	// Update error rate
	if success {
		stats.ErrorRate = stats.ErrorRate * 0.9 // Exponential decay
	} else {
		stats.ErrorRate = stats.ErrorRate*0.9 + 0.1 // Add error weight
	}

	// Update average latency
	if stats.AvgLatency == 0 {
		stats.AvgLatency = latency
	} else {
		stats.AvgLatency = time.Duration(float64(stats.AvgLatency)*0.9 + float64(latency)*0.1)
	}
}

// getEffectiveWeight calculates effective weight based on server performance
func (wb *WeightedBalancer) getEffectiveWeight(server *types.MCPServer) int {
	stats, exists := wb.stats[server.ID]
	if !exists {
		return server.Weight
	}

	effectiveWeight := float64(server.Weight)

	// Reduce weight based on error rate
	effectiveWeight *= (1.0 - stats.ErrorRate)

	// Reduce weight based on latency (assuming 100ms is baseline)
	if stats.AvgLatency > 100*time.Millisecond {
		latencyPenalty := float64(stats.AvgLatency) / float64(100*time.Millisecond)
		effectiveWeight /= latencyPenalty
	}

	if effectiveWeight < 1 {
		effectiveWeight = 1
	}

	return int(effectiveWeight)
}

// RandomBalancer implements random load balancing
type RandomBalancer struct{}

// NewRandomBalancer creates a new random load balancer
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{}
}

// SelectServer selects a random server
func (r *RandomBalancer) SelectServer(servers []*types.MCPServer) (*types.MCPServer, error) {
	if len(servers) == 0 {
		return nil, types.NewServiceUnavailableError("No servers available")
	}

	return servers[rand.Intn(len(servers))], nil
}

// UpdateStats updates server statistics (no-op for random)
func (r *RandomBalancer) UpdateStats(serverID string, success bool, latency time.Duration) {
	// Random balancer doesn't use stats
}

// CreateLoadBalancer creates a load balancer based on algorithm
func CreateLoadBalancer(algorithm string) LoadBalancer {
	switch algorithm {
	case types.LoadBalancerLeastConn:
		return NewLeastConnectionsBalancer()
	case types.LoadBalancerWeighted:
		return NewWeightedBalancer()
	case types.LoadBalancerRandom:
		return NewRandomBalancer()
	default:
		return NewRoundRobinBalancer()
	}
}
