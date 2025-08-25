package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// MCPServer represents an MCP server in the database
type MCPServer struct {
	ID             string  `db:"id"`
	OrganizationID string  `db:"organization_id"`
	Name           string  `db:"name"`
	Description    string  `db:"description"`
	Protocol       string  `db:"protocol"`
	URL            *string `db:"url"`
	IsActive       bool    `db:"is_active"`
}

// MCPServerRepository handles MCP server database operations
type MCPServerRepository struct {
	db *sqlx.DB
}

// NewMCPServerRepository creates a new MCP server repository
func NewMCPServerRepository(db *sqlx.DB) *MCPServerRepository {
	return &MCPServerRepository{db: db}
}

// GetByID retrieves an MCP server by ID
func (r *MCPServerRepository) GetByID(ctx context.Context, id string) (*MCPServer, error) {
	server := &MCPServer{}
	
	query := `
		SELECT id, organization_id, name, description, protocol, url, is_active
		FROM mcp_servers 
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&server.ID, &server.OrganizationID, &server.Name, &server.Description,
		&server.Protocol, &server.URL, &server.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	return server, nil
}