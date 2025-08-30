package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// MCPServer represents an MCP server in the database
type MCPServer struct {
	ID             string   `db:"id"`
	OrganizationID string   `db:"organization_id"`
	Name           string   `db:"name"`
	Description    string   `db:"description"`
	Protocol       string   `db:"protocol"`
	URL            *string  `db:"url"`
	Command        *string  `db:"command"`
	Args           []string `db:"args"`
	Environment    []string `db:"environment"`
	WorkingDir     *string  `db:"working_dir"`
	IsActive       bool     `db:"is_active"`
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
		SELECT id, organization_id, name, description, protocol, url, command, args, environment, working_dir, is_active
		FROM mcp_servers
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&server.ID, &server.OrganizationID, &server.Name, &server.Description,
		&server.Protocol, &server.URL, &server.Command, (*pq.StringArray)(&server.Args),
		(*pq.StringArray)(&server.Environment), &server.WorkingDir, &server.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	return server, nil
}
