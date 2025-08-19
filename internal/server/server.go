package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"mcp-gateway/internal/config"
	"mcp-gateway/internal/database"
)

type Server struct {
	port int
	cfg  *config.Config
	db   database.Service
}

func NewServer(cfg *config.Config) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = cfg.Server.Port
	}

	NewServer := &Server{
		port: port,
		cfg:  cfg,
		db:   database.New(),
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
