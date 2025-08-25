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
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// SetupManager handles all setup operations
type SetupManager struct {
	db     *sql.DB
	config *config.Config
}

// SetupFunction represents a setup function that can be executed
type SetupFunction struct {
	Execute     func(*SetupManager) error
	Name        string
	Description string
}

// Available setup functions
var setupFunctions = map[string]SetupFunction{
	"admin": {
		Name:        "Create Admin User",
		Description: "Creates a default admin user (admin@admin.com)",
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
	"defaults": {
		Name:        "Add Default MCP Data",
		Description: "Adds default resources, tools, and prompts for testing",
		Execute:     (*SetupManager).addDefaultMCPData,
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
	fmt.Println("Creating admin user: admin@admin.com")

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
	existingUser, err := userModel.GetByEmail("admin@admin.com")
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
		Email:          "admin@admin.com",
		Name:           "Admin User",
		PasswordHash:   string(passwordHash),
		OrganizationID: org.ID.String(),
		Role:           types.RoleSystemAdmin,
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

func (s *SetupManager) addDefaultMCPData() error {
	fmt.Println("Adding default MCP resources, tools, and prompts...")

	// Ensure default org exists
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

	// Add default resources
	if err := s.addDefaultResources(org.ID); err != nil {
		return fmt.Errorf("failed to add default resources: %w", err)
	}

	// Add default tools
	if err := s.addDefaultTools(org.ID); err != nil {
		return fmt.Errorf("failed to add default tools: %w", err)
	}

	// Add default prompts
	if err := s.addDefaultPrompts(org.ID); err != nil {
		return fmt.Errorf("failed to add default prompts: %w", err)
	}

	fmt.Println("✅ Default MCP data added successfully!")
	return nil
}

func (s *SetupManager) addDefaultResources(orgID uuid.UUID) error {
	resourceModel := models.NewMCPResourceModel(s.db)

	resources := []struct {
		name         string
		description  string
		resourceType string
		uri          string
		mimeType     string
		metadata     map[string]interface{}
		tags         []string
	}{
		{
			name:         "System Documentation",
			description:  "Default system documentation resource",
			resourceType: "url",
			uri:          "https://docs.example.com/mcp-gateway",
			mimeType:     "text/html",
			metadata:     map[string]interface{}{"version": "1.0", "public": true},
			tags:         []string{"documentation", "system", "help"},
		},
		{
			name:         "API Reference",
			description:  "MCP Gateway API reference documentation",
			resourceType: "url",
			uri:          "https://api.example.com/docs",
			mimeType:     "application/json",
			metadata:     map[string]interface{}{"version": "1.0", "swagger": true},
			tags:         []string{"api", "reference", "documentation"},
		},
	}

	for _, r := range resources {
		// Check if resource already exists
		_, err := resourceModel.GetByName(orgID, r.name)
		if err == nil {
			fmt.Printf("⚠️  Resource already exists: %s\n", r.name)
			continue
		}

		resource := &models.MCPResource{
			OrganizationID:    orgID,
			Name:              r.name,
			ResourceType:      r.resourceType,
			URI:               r.uri,
			IsActive:          true,
			Metadata:          r.metadata,
			Tags:              r.tags,
			AccessPermissions: map[string]interface{}{"read": []string{"*"}, "write": []string{"admin"}},
		}

		if r.description != "" {
			resource.Description = sql.NullString{String: r.description, Valid: true}
		}
		if r.mimeType != "" {
			resource.MimeType = sql.NullString{String: r.mimeType, Valid: true}
		}

		if err := resourceModel.Create(resource); err != nil {
			fmt.Printf("❌ Failed to create resource %s: %v\n", r.name, err)
		} else {
			fmt.Printf("✅ Created resource: %s\n", r.name)
		}
	}

	return nil
}

func (s *SetupManager) addDefaultTools(orgID uuid.UUID) error {
	toolModel := models.NewMCPToolModel(s.db)

	tools := []struct {
		name               string
		description        string
		functionName       string
		schema             map[string]interface{}
		category           string
		implementationType string
		examples           []interface{}
		documentation      string
		metadata           map[string]interface{}
		tags               []string
	}{
		{
			name:               "Echo Tool",
			description:        "Simple echo tool that returns the input message",
			functionName:       "echo",
			schema:             map[string]interface{}{"type": "object", "properties": map[string]interface{}{"message": map[string]interface{}{"type": "string", "description": "Message to echo back"}}, "required": []string{"message"}},
			category:           "general",
			implementationType: "internal",
			examples:           []interface{}{map[string]interface{}{"input": map[string]interface{}{"message": "Hello World"}, "output": map[string]interface{}{"result": "Hello World"}}},
			documentation:      "The echo tool simply returns whatever message you send to it. Useful for testing and debugging.",
			metadata:           map[string]interface{}{"version": "1.0", "simple": true},
			tags:               []string{"test", "debug", "simple"},
		},
		{
			name:               "Get Current Time",
			description:        "Returns the current server timestamp",
			functionName:       "get_current_time",
			schema:             map[string]interface{}{"type": "object", "properties": map[string]interface{}{"format": map[string]interface{}{"type": "string", "description": "Time format (iso, unix, human)", "default": "iso"}}, "required": []string{}},
			category:           "system",
			implementationType: "internal",
			examples:           []interface{}{map[string]interface{}{"input": map[string]interface{}{"format": "iso"}, "output": map[string]interface{}{"timestamp": "2024-01-01T12:00:00Z"}}, map[string]interface{}{"input": map[string]interface{}{"format": "unix"}, "output": map[string]interface{}{"timestamp": 1704110400}}},
			documentation:      "Returns the current server time in various formats. Supports ISO 8601, Unix timestamp, or human-readable format.",
			metadata:           map[string]interface{}{"version": "1.0", "timezone": "UTC"},
			tags:               []string{"time", "system", "utility"},
		},
		{
			name:               "Generate UUID",
			description:        "Generates a new UUID v4",
			functionName:       "generate_uuid",
			schema:             map[string]interface{}{"type": "object", "properties": map[string]interface{}{"version": map[string]interface{}{"type": "integer", "description": "UUID version (4 or 1)", "default": 4}}, "required": []string{}},
			category:           "dev",
			implementationType: "internal",
			examples:           []interface{}{map[string]interface{}{"input": map[string]interface{}{"version": 4}, "output": map[string]interface{}{"uuid": "550e8400-e29b-41d4-a716-446655440000"}}},
			documentation:      "Generates a new UUID (Universally Unique Identifier). Supports version 4 (random) and version 1 (timestamp-based).",
			metadata:           map[string]interface{}{"version": "1.0", "secure": true},
			tags:               []string{"uuid", "generate", "utility", "dev"},
		},
		{
			name:               "Base64 Encode",
			description:        "Encodes text to Base64 format",
			functionName:       "base64_encode",
			schema:             map[string]interface{}{"type": "object", "properties": map[string]interface{}{"text": map[string]interface{}{"type": "string", "description": "Text to encode"}, "url_safe": map[string]interface{}{"type": "boolean", "description": "Use URL-safe encoding", "default": false}}, "required": []string{"text"}},
			category:           "data",
			implementationType: "internal",
			examples:           []interface{}{map[string]interface{}{"input": map[string]interface{}{"text": "Hello World", "url_safe": false}, "output": map[string]interface{}{"encoded": "SGVsbG8gV29ybGQ="}}},
			documentation:      "Encodes text to Base64 format. Supports both standard and URL-safe encoding.",
			metadata:           map[string]interface{}{"version": "1.0", "encoding": "base64"},
			tags:               []string{"encode", "base64", "data", "utility"},
		},
	}

	for _, t := range tools {
		// Check if tool already exists
		_, err := toolModel.GetByName(orgID, t.name)
		if err == nil {
			fmt.Printf("⚠️  Tool already exists: %s\n", t.name)
			continue
		}

		tool := &models.MCPTool{
			OrganizationID:     orgID,
			Name:               t.name,
			FunctionName:       t.functionName,
			Schema:             t.schema,
			Category:           t.category,
			ImplementationType: t.implementationType,
			TimeoutSeconds:     30,
			MaxRetries:         3,
			UsageCount:         0,
			AccessPermissions:  map[string]interface{}{"execute": []string{"*"}, "read": []string{"*"}},
			IsActive:           true,
			IsPublic:           false,
			Metadata:           t.metadata,
			Tags:               t.tags,
			Examples:           t.examples,
		}

		if t.description != "" {
			tool.Description = sql.NullString{String: t.description, Valid: true}
		}
		if t.documentation != "" {
			tool.Documentation = sql.NullString{String: t.documentation, Valid: true}
		}

		if err := toolModel.Create(tool); err != nil {
			fmt.Printf("❌ Failed to create tool %s: %v\n", t.name, err)
		} else {
			fmt.Printf("✅ Created tool: %s\n", t.name)
		}
	}

	return nil
}

func (s *SetupManager) addDefaultPrompts(orgID uuid.UUID) error {
	promptModel := models.NewMCPPromptModel(s.db)

	prompts := []struct {
		name           string
		description    string
		promptTemplate string
		parameters     []interface{}
		category       string
		metadata       map[string]interface{}
		tags           []string
	}{
		{
			name:           "Code Review",
			description:    "Template for conducting thorough code reviews",
			promptTemplate: "Please review the following code for:\n1. Correctness and logic\n2. Performance considerations\n3. Security issues\n4. Code style and maintainability\n\nCode:\n{{code}}\n\nPlease provide specific feedback and suggestions for improvement.",
			parameters:     []interface{}{map[string]interface{}{"name": "code", "type": "string", "required": true, "description": "Code to be reviewed"}},
			category:       "coding",
			metadata:       map[string]interface{}{"version": "1.0", "language_agnostic": true},
			tags:           []string{"code-review", "development", "quality"},
		},
		{
			name:           "Documentation Generation",
			description:    "Generate comprehensive documentation from code",
			promptTemplate: "Generate clear, comprehensive documentation for the following code. Include:\n1. Purpose and functionality\n2. Parameters and return values\n3. Usage examples\n4. Any important notes or limitations\n\nCode:\n{{code}}\n\nFormat: {{format}}\n\nPlease provide well-structured documentation.",
			parameters:     []interface{}{map[string]interface{}{"name": "code", "type": "string", "required": true, "description": "Code to document"}, map[string]interface{}{"name": "format", "type": "string", "required": false, "description": "Documentation format (markdown, rst, etc.)", "default": "markdown"}},
			category:       "coding",
			metadata:       map[string]interface{}{"version": "1.0", "supports_multiple_formats": true},
			tags:           []string{"documentation", "code-generation", "development"},
		},
		{
			name:           "Data Analysis",
			description:    "Template for structured data analysis tasks",
			promptTemplate: "Analyze the following data and provide insights:\n\nData: {{data}}\n\nAnalysis Requirements:\n{{requirements}}\n\nPlease provide:\n1. Summary statistics\n2. Key patterns and trends\n3. Actionable insights\n4. Visualizations if applicable",
			parameters:     []interface{}{map[string]interface{}{"name": "data", "type": "string", "required": true, "description": "Data to analyze"}, map[string]interface{}{"name": "requirements", "type": "string", "required": false, "description": "Specific analysis requirements", "default": "General analysis"}},
			category:       "analysis",
			metadata:       map[string]interface{}{"version": "1.0", "supports_visualizations": true},
			tags:           []string{"data-analysis", "statistics", "insights"},
		},
	}

	for _, p := range prompts {
		// Check if prompt already exists
		_, err := promptModel.GetByName(orgID, p.name)
		if err == nil {
			fmt.Printf("⚠️  Prompt already exists: %s\n", p.name)
			continue
		}

		prompt := &models.MCPPrompt{
			OrganizationID: orgID,
			Name:           p.name,
			PromptTemplate: p.promptTemplate,
			Parameters:     p.parameters,
			Category:       p.category,
			UsageCount:     0,
			IsActive:       true,
			Metadata:       p.metadata,
			Tags:           p.tags,
		}

		if p.description != "" {
			prompt.Description = sql.NullString{String: p.description, Valid: true}
		}

		if err := promptModel.Create(prompt); err != nil {
			fmt.Printf("❌ Failed to create prompt %s: %v\n", p.name, err)
		} else {
			fmt.Printf("✅ Created prompt: %s\n", p.name)
		}
	}

	return nil
}

func (s *SetupManager) resetDatabase() error {
	fmt.Println("⚠️  WARNING: This will delete ALL data in the database!")
	fmt.Print("Are you sure you want to continue? (type 'yes' to confirm): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	confirmation := strings.TrimSpace(scanner.Text())

	if confirmation != "yes" {
		fmt.Println("❌ Database reset canceled")
		return nil
	}

	fmt.Println("🗑️  Resetting database...")

	// List of tables to truncate in dependency order
	tables := []string{
		"mcp_sessions",
		"mcp_resources",
		"mcp_prompts",
		"mcp_tools",
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
