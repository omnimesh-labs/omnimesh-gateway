package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"mcp-gateway/apps/backend/internal/config"
	"mcp-gateway/apps/backend/internal/database"
	"mcp-gateway/apps/backend/internal/logging"
	"mcp-gateway/apps/backend/internal/logging/plugins/file"
)

type Server struct {
	db      database.Service
	logging logging.LogService
	cfg     *config.Config
	port    int
}

func NewServer(cfg *config.Config) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = cfg.Server.Port
	}

	// Register the file logging plugin (skip if environment variable is set for tests)
	if os.Getenv("SKIP_PLUGIN_REGISTRATION") == "" {
		if err := logging.RegisterPlugin(file.NewFilePlugin()); err != nil {
			panic(fmt.Sprintf("failed to register file plugin: %v", err))
		}
	}

	// Initialize logging service
	loggingConfig := &logging.LoggingConfig{
		Backend:       cfg.Logging.Backend,
		Level:         logging.LogLevel(cfg.Logging.Level),
		Environment:   cfg.Logging.Environment,
		BufferSize:    cfg.Logging.BufferSize,
		BatchSize:     cfg.Logging.BatchSize,
		FlushInterval: cfg.Logging.FlushInterval,
		Async:         cfg.Logging.Async,
		Config:        cfg.Logging.Config,
	}

	loggingService, err := logging.NewService(loggingConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logging service: %v", err))
	}

	NewServer := &Server{
		port:    port,
		cfg:     cfg,
		db:      database.New(),
		logging: loggingService,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
