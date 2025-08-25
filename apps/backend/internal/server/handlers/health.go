package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	startTime        time.Time
	db               *sql.DB
	discoveryService DiscoveryService
	buildVersion     string
}

// DiscoveryService interface for getting server count
type DiscoveryService interface {
	ListServers(orgID string) ([]*interface{}, error) // Using interface{} for now
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, discoveryService DiscoveryService, buildVersion string) *HealthHandler {
	if buildVersion == "" {
		buildVersion = "dev" // Default for development
	}
	return &HealthHandler{
		startTime:        time.Now(),
		db:               db,
		discoveryService: discoveryService,
		buildVersion:     buildVersion,
	}
}

// Health returns the basic health status
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
		"version":   h.buildVersion,
	})
}

// HealthDetailed returns detailed health information
func (h *HealthHandler) HealthDetailed(c *gin.Context) {
	checks := make(map[string]interface{})
	overallStatus := "healthy"

	// Database connectivity check
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck["status"] != "healthy" {
		overallStatus = "unhealthy"
	}

	// System resource checks
	systemCheck := h.checkSystemResources()
	checks["system"] = systemCheck
	if systemCheck["status"] != "healthy" {
		overallStatus = "degraded" // System issues are less critical
	}

	// Discovery service check
	discoveryCheck := h.checkDiscoveryService()
	checks["discovery"] = discoveryCheck
	if discoveryCheck["status"] != "healthy" {
		overallStatus = "degraded"
	}

	// Determine HTTP status code
	status := http.StatusOK
	if overallStatus == "unhealthy" {
		status = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		status = http.StatusOK // Still return 200 for degraded
	}

	c.JSON(status, gin.H{
		"status":    overallStatus,
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
		"version":   h.buildVersion,
		"checks":    checks,
	})
}

// Ready returns readiness status
func (h *HealthHandler) Ready(c *gin.Context) {
	ready := true
	checks := make(map[string]interface{})

	// Database connectivity is required for readiness
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck["status"] != "healthy" {
		ready = false
	}

	// Check if database migrations are complete
	migrationCheck := h.checkMigrations()
	checks["migrations"] = migrationCheck
	if migrationCheck["status"] != "complete" {
		ready = false
	}

	status := http.StatusOK
	statusText := "ready"
	if !ready {
		status = http.StatusServiceUnavailable
		statusText = "not_ready"
	}

	c.JSON(status, gin.H{
		"status":    statusText,
		"timestamp": time.Now(),
		"checks":    checks,
	})
}

// Live returns liveness status
func (h *HealthHandler) Live(c *gin.Context) {
	// Simple liveness check
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase() map[string]interface{} {
	start := time.Now()
	result := map[string]interface{}{
		"status": "healthy",
	}

	if h.db == nil {
		result["status"] = "unhealthy"
		result["error"] = "database connection not available"
		return result
	}

	// Ping database
	err := h.db.Ping()
	if err != nil {
		result["status"] = "unhealthy"
		result["error"] = fmt.Sprintf("database ping failed: %v", err)
		return result
	}

	latency := time.Since(start)
	result["latency_ms"] = latency.Milliseconds()

	// Additional check: simple query
	var count int
	err = h.db.QueryRow("SELECT 1").Scan(&count)
	if err != nil {
		result["status"] = "degraded"
		result["warning"] = "simple query failed"
	}

	return result
}

// checkSystemResources checks system resource usage
func (h *HealthHandler) checkSystemResources() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
	}

	// Memory statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	result["memory"] = map[string]interface{}{
		"alloc_mb":       bToMb(m.Alloc),
		"total_alloc_mb": bToMb(m.TotalAlloc),
		"sys_mb":         bToMb(m.Sys),
		"num_gc":         m.NumGC,
	}

	// Basic disk space check (this is simplified)
	var stat syscall.Statfs_t
	err := syscall.Statfs(".", &stat)
	if err != nil {
		result["disk"] = map[string]interface{}{
			"status": "unknown",
			"error":  "could not get disk info",
		}
	} else {
		available := stat.Bavail * uint64(stat.Bsize)
		total := stat.Blocks * uint64(stat.Bsize)
		used := total - available
		usagePercent := float64(used) / float64(total) * 100

		result["disk"] = map[string]interface{}{
			"total_gb":     bytesToGB(total),
			"available_gb": bytesToGB(available),
			"used_percent": fmt.Sprintf("%.1f%%", usagePercent),
		}

		// Mark as degraded if disk usage is high
		if usagePercent > 90 {
			result["status"] = "degraded"
			result["warning"] = "high disk usage"
		}
	}

	return result
}

// checkDiscoveryService checks discovery service health
func (h *HealthHandler) checkDiscoveryService() map[string]interface{} {
	result := map[string]interface{}{
		"status":  "healthy",
		"servers": 0,
	}

	if h.discoveryService == nil {
		result["status"] = "degraded"
		result["warning"] = "discovery service not available"
		return result
	}

	// Try to get server list to verify service is working
	// Using empty orgID for now - in a real implementation you'd handle multi-tenant properly
	servers, err := h.discoveryService.ListServers("")
	if err != nil {
		result["status"] = "degraded"
		result["error"] = fmt.Sprintf("failed to list servers: %v", err)
		return result
	}

	result["servers"] = len(servers)
	return result
}

// checkMigrations checks if database migrations are complete
func (h *HealthHandler) checkMigrations() map[string]interface{} {
	result := map[string]interface{}{
		"status": "complete",
	}

	if h.db == nil {
		result["status"] = "unknown"
		result["error"] = "database not available"
		return result
	}

	// Check if schema_migrations table exists and has entries
	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_name = 'schema_migrations'
	`).Scan(&count)

	if err != nil || count == 0 {
		result["status"] = "incomplete"
		result["error"] = "schema_migrations table not found"
		return result
	}

	// Check for any dirty migrations
	var dirtyCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*)
		FROM schema_migrations
		WHERE dirty = true
	`).Scan(&dirtyCount)

	if err != nil {
		result["status"] = "unknown"
		result["warning"] = "could not check migration status"
		return result
	}

	if dirtyCount > 0 {
		result["status"] = "incomplete"
		result["error"] = fmt.Sprintf("%d dirty migrations found", dirtyCount)
		return result
	}

	// Get latest migration version
	var latestVersion int
	err = h.db.QueryRow(`
		SELECT COALESCE(MAX(version), 0)
		FROM schema_migrations
	`).Scan(&latestVersion)

	if err == nil {
		result["latest_version"] = latestVersion
	}

	return result
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func bytesToGB(b uint64) float64 {
	return float64(b) / 1024 / 1024 / 1024
}
