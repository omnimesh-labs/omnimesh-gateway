package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/config"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/database"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/discovery"
	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/transport"
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

	// Initialize transport manager
	transportConfig := cfg.Transport.ToTransportConfig()
	transportManager := transport.NewManager(transportConfig)
	if err := transportManager.Initialize(context.Background()); err != nil {
		log.Printf("Warning: Failed to initialize transport manager: %v", err)
	}

	// Initialize services
	// TODO: Initialize discovery service and other background workers
	discoveryConfig := &discovery.Config{
		Enabled:          cfg.Discovery.Enabled,
		HealthInterval:   cfg.Discovery.HealthInterval,
		FailureThreshold: cfg.Discovery.FailureThreshold,
		RecoveryTimeout:  cfg.Discovery.RecoveryTimeout,
		SingleTenant:     true,
	}
	discoveryService := discovery.NewService(db, discoveryConfig, transportManager)

	log.Println("Background worker started")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping worker...")

	// Cancel context and wait for graceful shutdown
	cancel()

	// Shutdown services gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := discoveryService.Stop(); err != nil {
		log.Printf("Error stopping discovery service: %v", err)
	}

	if err := transportManager.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down transport manager: %v", err)
	}

	log.Println("Worker stopped")
}
