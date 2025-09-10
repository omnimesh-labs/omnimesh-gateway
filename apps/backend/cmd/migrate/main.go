package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnimesh-labs/omnimesh-gateway/apps/backend/internal/config"
)

func main() {
	// Check command
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate [up|down|status|setup-admin]")
	}

	command := os.Args[1]

	// Get config path from environment or use default
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "apps/backend/configs/development.yaml"
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use environment variables if available (Docker environment)
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = cfg.Database.Host
	}
	dbUser := os.Getenv("DB_USERNAME")
	if dbUser == "" {
		dbUser = cfg.Database.User
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = cfg.Database.Password
	}
	dbName := os.Getenv("DB_DATABASE")
	if dbName == "" {
		dbName = cfg.Database.Database
	}
	dbSSLMode := os.Getenv("DB_SSL_MODE")
	if dbSSLMode == "" {
		dbSSLMode = cfg.Database.SSLMode
		if dbSSLMode == "" {
			dbSSLMode = "disable"
		}
	}

	// Build connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbUser,
		dbPassword,
		dbHost,
		cfg.Database.Port,
		dbName,
		dbSSLMode,
	)

	log.Printf("Connecting to database at %s:%d/%s", dbHost, cfg.Database.Port, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	switch command {
	case "up", "down", "status":
		runMigrations(db, command)
	case "setup-admin":
		setupAdminUser(db)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func runMigrations(db *sql.DB, command string) {
	// Get migrations path - handle both root and apps/backend working directories
	var migrationsPath string
	pwd, _ := os.Getwd()

	// Check if we're in the backend directory or root
	if filepath.Base(pwd) == "backend" || strings.HasSuffix(pwd, "apps/backend") {
		migrationsPath = "migrations"
	} else {
		migrationsPath = "apps/backend/migrations"
	}

	if !filepath.IsAbs(migrationsPath) {
		migrationsPath = filepath.Join(pwd, migrationsPath)
	}

	log.Printf("Using migrations path: %s", migrationsPath)

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Fatalf("Migrations directory does not exist: %s", migrationsPath)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	switch command {
	case "up":
		log.Println("Starting database migrations...")
		err := m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		version, _, _ := m.Version()
		if err == migrate.ErrNoChange {
			log.Printf("Database is already up to date (version: %d)", version)
		} else {
			log.Printf("Migrations completed successfully (version: %d)", version)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		log.Println("Migrations rolled back successfully")
	case "status":
		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		log.Printf("Current migration version: %d (dirty: %v)", version, dirty)
	}
}

func setupAdminUser(db *sql.DB) {
	// Check if admin user already exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE email = 'admin@admin.com'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if admin user exists: %v", err)
	}

	if exists {
		log.Println("Admin user already exists")
		return
	}

	// First, ensure we have an organization
	var orgID string
	err = db.QueryRow(`
		SELECT id FROM organizations
		WHERE slug = 'default-org'
	`).Scan(&orgID)

	if err == sql.ErrNoRows {
		// Create default organization
		err = db.QueryRow(`
			INSERT INTO organizations (name, slug, plan_type, max_servers, max_sessions, log_retention_days)
			VALUES ('Default Organization', 'default-org', 'enterprise', 100, 1000, 30)
			RETURNING id
		`).Scan(&orgID)
		if err != nil {
			log.Fatalf("Failed to create default organization: %v", err)
		}
		log.Println("Created default organization")
	} else if err != nil {
		log.Fatalf("Failed to check for organization: %v", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("qwerty123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	_, err = db.Exec(`
		INSERT INTO users (email, name, password_hash, organization_id, role, is_active, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, "admin@admin.com", "Admin User", string(hashedPassword), orgID, "admin", true, true)

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Println("Admin user created successfully")
	log.Println("Email: admin@admin.com")
	log.Println("Password: qwerty123")
}
