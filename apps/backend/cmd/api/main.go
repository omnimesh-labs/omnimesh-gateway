package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"mcp-gateway/apps/backend/internal/config"
	"mcp-gateway/apps/backend/internal/server"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	var (
		configPath = flag.String("config", "", "Path to configuration file")
	)
	flag.Parse()

	// Determine config path
	var finalConfigPath string
	if *configPath != "" {
		finalConfigPath = *configPath
	} else {
		// Try both possible locations
		candidates := []string{
			"configs/development.yaml",              // When running from apps/backend/
			"apps/backend/configs/development.yaml", // When running from repo root
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				finalConfigPath = candidate
				break
			}
		}

		if finalConfigPath == "" {
			log.Fatal("Could not find development.yaml config file. Tried: " +
				filepath.Join(candidates[0]) + ", " + filepath.Join(candidates[1]))
		}
	}

	log.Printf("Loading config from: %s", finalConfigPath)

	// Load configuration
	cfg, err := config.Load(finalConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set defaults if config loading is not fully implemented
	if cfg == nil {
		cfg = &config.Config{
			MCPDiscovery: config.MCPDiscoveryConfig{
				Enabled: true,
				BaseURL: "https://metatool-service.jczstudio.workers.dev/search",
			},
			Server: config.ServerConfig{
				Port: 8080,
			},
		}
	}

	server := server.NewServer(cfg)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
