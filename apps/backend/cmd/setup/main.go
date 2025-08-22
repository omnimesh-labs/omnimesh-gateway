package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"mcp-gateway/apps/backend/internal/config"
	"mcp-gateway/apps/backend/internal/database/models"
	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

// SetupManager handles all setup operations
type SetupManager struct {
	db     *sql.DB
	config *config.Config
}

// SetupFunction represents a setup function that can be executed
type SetupFunction struct {
	Name        string
	Description string
	Execute     func(*SetupManager) error
}

// Available setup functions
var setupFunctions = map[string]SetupFunction{
	"admin": {
		Name:        "Create Admin User",
		Description: "Creates a default admin user (team@wraithscan.com)",
		Execute:     (*SetupManager).createAdminUser,
	},
	"org": {
		Name:        "Create Default Organization",
		Description: "Creates the default organization",
		Execute:     (*SetupManager).createDefaultOrganization,
	},
	"dummy": {
		Name:        "Add Dummy Data",
		Description: "Adds sample data for testing (users, servers, etc.)",
		Execute:     (*SetupManager).addDummyData,
	},
	"reset": {
		Name:        "Reset Database",
		Description: "WARNING: Removes all data and recreates tables",
		Execute:     (*SetupManager).resetDatabase,
	},
}

func main() {
	fmt.Println("🚀 MCP Gateway Local Development Setup")
	fmt.Println("======================================")

	// Load configuration
	cfg, err := config.Load("apps/backend/configs/development.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	setupManager := &SetupManager{
		db:     db,
		config: cfg,
	}

	// Check if migrations are up to date
	if err := setupManager.checkMigrations(); err != nil {
		log.Fatalf("Migration check failed: %v", err)
	}

	if len(os.Args) > 1 {
		// Run specific function from command line
		functionName := os.Args[1]
		if setupFunc, exists := setupFunctions[functionName]; exists {
			fmt.Printf("\n🔧 Running: %s\n", setupFunc.Name)
			if err := setupFunc.Execute(setupManager); err != nil {
				log.Fatalf("Setup failed: %v", err)
			}
			fmt.Println("✅ Setup completed successfully!")
		} else {
			fmt.Printf("❌ Unknown setup function: %s\n", functionName)
			printAvailableFunctions()
		}
	} else {
		// Interactive mode
		runInteractiveSetup(setupManager)
	}
}

func connectDB(cfg *config.Config) (*sql.DB, error) {
	// Build DSN from config
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("✅ Database connection established")
	return db, nil
}

func (s *SetupManager) checkMigrations() error {
	// Check if main tables exist
	tables := []string{"organizations", "users", "mcp_servers", "mcp_sessions"}
	
	for _, table := range tables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`
		err := s.db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		
		if !exists {
			return fmt.Errorf("table %s does not exist - please run migrations first", table)
		}
	}
	
	fmt.Println("✅ Database migrations are up to date")
	return nil
}

func runInteractiveSetup(setupManager *SetupManager) {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Println("\n📋 Available Setup Functions:")
		printAvailableFunctions()
		
		fmt.Print("\nEnter function name (or 'quit' to exit): ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		
		if input == "quit" || input == "q" || input == "exit" {
			fmt.Println("👋 Goodbye!")
			break
		}
		
		if setupFunc, exists := setupFunctions[input]; exists {
			fmt.Printf("\n🔧 Running: %s\n", setupFunc.Name)
			if err := setupFunc.Execute(setupManager); err != nil {
				fmt.Printf("❌ Setup failed: %v\n", err)
			} else {
				fmt.Println("✅ Setup completed successfully!")
			}
		} else {
			fmt.Printf("❌ Unknown function: %s\n", input)
		}
	}
}

func printAvailableFunctions() {
	for key, fn := range setupFunctions {
		fmt.Printf("  %s - %s\n", key, fn.Description)
	}
}

// Setup Functions Implementation

func (s *SetupManager) createAdminUser() error {
	fmt.Println("Creating admin user: team@wraithscan.com")
	
	// First ensure default organization exists
	orgModel := models.NewOrganizationModel(s.db)
	org, err := orgModel.GetDefault()
	if err != nil {
		fmt.Println("Default organization not found, creating it first...")
		if err := s.createDefaultOrganization(); err != nil {
			return fmt.Errorf("failed to create default organization: %w", err)
		}
		org, err = orgModel.GetDefault()
		if err != nil {
			return fmt.Errorf("failed to get default organization after creation: %w", err)
		}
	}
	
	// Check if admin user already exists
	userModel := models.NewUserModel(s.db)
	existingUser, err := userModel.GetByEmail("team@wraithscan.com")
	if err == nil {
		fmt.Printf("⚠️  Admin user already exists: %s (ID: %s)\n", existingUser.Email, existingUser.ID)
		return nil
	}
	
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("qwerty123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create admin user
	adminUser := &types.User{
		ID:             uuid.New().String(),
		Email:          "team@wraithscan.com",
		Name:           "Admin User",
		PasswordHash:   string(passwordHash),
		OrganizationID: org.ID.String(),
		Role:           types.RoleAdmin,
		IsActive:       true,
	}
	
	if err := userModel.Create(adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}
	
	fmt.Printf("✅ Admin user created successfully!\n")
	fmt.Printf("   Email: %s\n", adminUser.Email)
	fmt.Printf("   Password: qwerty123\n")
	fmt.Printf("   Role: %s\n", adminUser.Role)
	fmt.Printf("   Organization: %s\n", org.Name)
	
	return nil
}

func (s *SetupManager) createDefaultOrganization() error {
	fmt.Println("Creating default organization...")
	
	orgModel := models.NewOrganizationModel(s.db)
	
	// Check if default org already exists
	_, err := orgModel.GetDefault()
	if err == nil {
		fmt.Println("⚠️  Default organization already exists")
		return nil
	}
	
	// Create default organization
	defaultOrg := &models.Organization{
		ID:               uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		Name:             "Default Organization",
		Slug:             "default",
		IsActive:         true,
		PlanType:         "free",
		MaxServers:       10,
		MaxSessions:      100,
		LogRetentionDays: 7,
	}
	
	if err := orgModel.Create(defaultOrg); err != nil {
		return fmt.Errorf("failed to create default organization: %w", err)
	}
	
	fmt.Printf("✅ Default organization created: %s\n", defaultOrg.Name)
	return nil
}

func (s *SetupManager) addDummyData() error {
	fmt.Println("Adding dummy data for testing...")
	
	// Ensure default org and admin user exist
	if err := s.createDefaultOrganization(); err != nil {
		return err
	}
	if err := s.createAdminUser(); err != nil {
		return err
	}
	
	// Get default organization
	orgModel := models.NewOrganizationModel(s.db)
	org, err := orgModel.GetDefault()
	if err != nil {
		return fmt.Errorf("failed to get default organization: %w", err)
	}
	
	// Create dummy users
	userModel := models.NewUserModel(s.db)
	dummyUsers := []struct {
		email string
		name  string
		role  string
	}{
		{"user1@example.com", "Test User 1", types.RoleUser},
		{"user2@example.com", "Test User 2", types.RoleUser},
		{"viewer@example.com", "Test Viewer", types.RoleViewer},
	}
	
	for _, userData := range dummyUsers {
		// Check if user already exists
		_, err := userModel.GetByEmail(userData.email)
		if err == nil {
			fmt.Printf("⚠️  User already exists: %s\n", userData.email)
			continue
		}
		
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		
		user := &types.User{
			ID:             uuid.New().String(),
			Email:          userData.email,
			Name:           userData.name,
			PasswordHash:   string(passwordHash),
			OrganizationID: org.ID.String(),
			Role:           userData.role,
			IsActive:       true,
		}
		
		if err := userModel.Create(user); err != nil {
			fmt.Printf("❌ Failed to create user %s: %v\n", userData.email, err)
		} else {
			fmt.Printf("✅ Created user: %s (%s)\n", userData.email, userData.role)
		}
	}
	
	// TODO: Add dummy MCP servers, sessions, etc.
	fmt.Println("📝 Note: Additional dummy data (servers, sessions) can be added in future iterations")
	
	return nil
}

func (s *SetupManager) resetDatabase() error {
	fmt.Println("⚠️  WARNING: This will delete ALL data in the database!")
	fmt.Print("Are you sure you want to continue? (type 'yes' to confirm): ")
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	confirmation := strings.TrimSpace(scanner.Text())
	
	if confirmation != "yes" {
		fmt.Println("❌ Database reset cancelled")
		return nil
	}
	
	fmt.Println("🗑️  Resetting database...")
	
	// List of tables to truncate in dependency order
	tables := []string{
		"mcp_sessions",
		"log_index", 
		"audit_logs",
		"mcp_servers",
		"users",
		"organizations",
	}
	
	// Truncate tables
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := s.db.Exec(query); err != nil {
			// If table doesn't exist, continue
			fmt.Printf("⚠️  Could not truncate %s (may not exist): %v\n", table, err)
		} else {
			fmt.Printf("✅ Truncated table: %s\n", table)
		}
	}
	
	fmt.Println("✅ Database reset completed!")
	fmt.Println("💡 You may want to run the 'admin' and 'org' functions to recreate basic data")
	
	return nil
}