package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
)

// SetupTestDatabase creates a test database using testcontainers
func SetupTestDatabase(t *testing.T) (*sql.DB, func(), error) {
	// Check if we should use existing database (for CI/local development)
	if dbURL := os.Getenv("TEST_DATABASE_URL"); dbURL != "" {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to test database: %w", err)
		}

		// Return a no-op teardown function
		return db, func() { db.Close() }, nil
	}

	// Create PostgreSQL container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get container connection details
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get container port: %w", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Wait for database to be ready
	for range 30 {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Teardown function
	teardown := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, teardown, nil
}

// RunMigrations runs database migrations on the test database
func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrationPath := "file://../../migrations"
	if _, err := os.Stat("migrations"); err == nil {
		migrationPath = "file://migrations"
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CleanDatabase removes all data from tables but keeps the schema
func CleanDatabase(t *testing.T, db *sql.DB) {
	tables := []string{
		"oauth_user_consents",
		"oauth_authorization_codes",
		"oauth_tokens",
		"oauth_clients",
		"namespace_tool_mappings",
		"namespace_server_mappings",
		"namespaces",
		"mcp_sessions",
		"virtual_servers",
		"mcp_servers",
		"api_keys",
		"users",
		"organizations",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			// Table might not exist, continue
			continue
		}
	}
}

// CreateTestOrganization creates a test organization with unique ID
func CreateTestOrganization(db *sql.DB) (string, error) {
	// Use a generated UUID for the organization to avoid conflicts in parallel tests
	var orgID string
	err := db.QueryRow("SELECT gen_random_uuid()").Scan(&orgID)
	if err != nil {
		return "", fmt.Errorf("failed to generate organization ID: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO organizations (id, name, slug)
		VALUES ($1, $2, $3)
	`, orgID, "Test Organization", "test-org-"+orgID[:8])

	if err != nil {
		return "", fmt.Errorf("failed to create test organization: %w", err)
	}

	// Verify the organization exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM organizations WHERE id = $1", orgID).Scan(&count)
	if err != nil {
		return "", fmt.Errorf("failed to verify test organization exists: %w", err)
	}

	if count == 0 {
		return "", fmt.Errorf("test organization was not created properly")
	}

	return orgID, nil
}

// CreateTestUser creates a test user with unique ID
func CreateTestUser(db *sql.DB, orgID string) (string, error) {
	// Use a generated UUID for the user to avoid conflicts in parallel tests
	var userID string
	err := db.QueryRow("SELECT gen_random_uuid()").Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate user ID: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, email, name, password_hash, organization_id, role)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "test@example.com", "Test User", "hashed", orgID, "user")

	if err != nil {
		return "", fmt.Errorf("failed to create test user: %w", err)
	}

	return userID, nil
}

// CreateTestUserWithCredentials creates a test user with specific email and password
func CreateTestUserWithCredentials(db *sql.DB, orgID, email, password string) (string, error) {
	// Use a generated UUID for the user to avoid conflicts in parallel tests
	var userID string
	err := db.QueryRow("SELECT gen_random_uuid()").Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Hash the password using bcrypt (same as production)
	passwordHash, err := hashPasswordForTest(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password for test: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, email, name, password_hash, organization_id, role)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, email, "Test User", passwordHash, orgID, "admin") // admin role for API key creation

	if err != nil {
		return "", fmt.Errorf("failed to create test user with credentials: %w", err)
	}

	return userID, nil
}

// CreateTestNamespace creates a test namespace with unique ID
func CreateTestNamespace(db *sql.DB, orgID string) (string, error) {
	// Use a generated UUID for the namespace to avoid conflicts in parallel tests
	var namespaceID string
	err := db.QueryRow("SELECT gen_random_uuid()").Scan(&namespaceID)
	if err != nil {
		return "", fmt.Errorf("failed to generate namespace ID: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO namespaces (id, organization_id, name, slug, description)
		VALUES ($1, $2, $3, $4, $5)
	`, namespaceID, orgID, "test-namespace", "test-namespace-"+namespaceID[:8], "Test namespace")

	if err != nil {
		return "", fmt.Errorf("failed to create test namespace: %w", err)
	}

	return namespaceID, nil
}

// CreateTestMCPServer creates a test MCP server
func CreateTestMCPServer(db *sql.DB, orgID string, name string) (string, error) {
	serverID := fmt.Sprintf("00000000-0000-0000-0000-00000000000%d", len(name))

	_, err := db.Exec(`
		INSERT INTO mcp_servers (id, organization_id, name, description, protocol, url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, serverID, orgID, name, "Test server", "jsonrpc", "http://localhost:8080", true)

	if err != nil {
		return "", fmt.Errorf("failed to create test MCP server: %w", err)
	}

	return serverID, nil
}

// hashPasswordForTest hashes a password using bcrypt with low cost for faster tests
func hashPasswordForTest(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4) // Use lower cost for faster tests
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}
