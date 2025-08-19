package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-gateway/internal/config"
	"mcp-gateway/internal/database"
	"mcp-gateway/internal/discovery"
)

func main() {
	var (
		configPath = flag.String("config", "configs/development.yaml", "Path to configuration file")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := database.NewWithConfig(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create context for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	// TODO: Initialize discovery service and other background workers
	discoveryConfig := &discovery.Config{
		Enabled:          cfg.Discovery.Enabled,
		HealthInterval:   cfg.Discovery.HealthInterval,
		FailureThreshold: cfg.Discovery.FailureThreshold,
		RecoveryTimeout:  cfg.Discovery.RecoveryTimeout,
	}
	_ = discovery.NewService(db, discoveryConfig)

	log.Println("Background worker started")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping worker...")

	// Cancel context and wait for graceful shutdown
	cancel()
	time.Sleep(5 * time.Second)

	log.Println("Worker stopped")
}
