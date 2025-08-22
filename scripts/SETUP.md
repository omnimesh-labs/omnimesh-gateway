# MCP Gateway Development Setup Script

This document describes the local development setup script for MCP Gateway, which helps you quickly set up your development environment with test data and users.

## Quick Start

The fastest way to get started is to run:

```bash
# Ensure database is running and migrations are applied
make docker-run
make migrate

# Create admin user
make setup-admin
```

This creates an admin user with:
- **Email**: `team@wraithscan.com`
- **Password**: `qwerty123`
- **Role**: `admin`

## Available Commands

### Makefile Shortcuts

```bash
# Interactive setup (shows menu)
make setup

# Create admin user only
make setup-admin

# Create default organization only
make setup-org

# Add dummy test data
make setup-dummy

# Reset database (WARNING: deletes all data)
make setup-reset
```

### Direct Script Usage

You can also run the setup script directly:

```bash
# Interactive mode
go run apps/backend/cmd/setup/main.go

# Specific functions
go run apps/backend/cmd/setup/main.go admin
go run apps/backend/cmd/setup/main.go org
go run apps/backend/cmd/setup/main.go dummy
go run apps/backend/cmd/setup/main.go reset
```

## Setup Functions

### 🔑 `admin` - Create Admin User
Creates the primary admin user for local development.

**Details:**
- Email: `team@wraithscan.com`
- Password: `qwerty123`
- Role: `admin`
- Automatically creates default organization if needed

**Safe to run multiple times** - will skip if user already exists.

### 🏢 `org` - Create Default Organization
Creates the default organization that all users belong to.

**Details:**
- ID: `00000000-0000-0000-0000-000000000000`
- Name: "Default Organization"
- Slug: "default"
- Plan: Free tier (10 servers, 100 sessions, 7 days log retention)

**Safe to run multiple times** - will skip if organization already exists.

### 📊 `dummy` - Add Dummy Data
Creates additional test data for development and testing.

**Creates:**
- Test users:
  - `user1@example.com` (user role)
  - `user2@example.com` (user role)
  - `viewer@example.com` (viewer role)
- All dummy users have password: `password123`

**Future Extensions:** Will be extended to include:
- Sample MCP servers
- Test sessions
- Sample log entries
- Rate limiting policies

### 🗑️ `reset` - Reset Database
**⚠️ WARNING:** This completely wipes all data from the database.

**What it does:**
- Truncates all tables in dependency order
- Preserves table structure
- Requires explicit confirmation (`yes`)

**Use case:** Clean slate for testing or when data becomes corrupted.

## Development Workflow

### Initial Setup
```bash
# 1. Start database
make docker-run

# 2. Run migrations
make migrate

# 3. Create basic data
make setup-admin
```

### Daily Development
```bash
# Quick admin user creation if needed
make setup-admin

# Add test data when needed
make setup-dummy
```

### Clean Slate Testing
```bash
# Reset everything
make setup-reset

# Recreate basic setup
make setup-admin
```

## Extending the Setup Script

The setup script is designed to be easily extensible. To add new setup functions:

### 1. Add Function to Map

In `apps/backend/cmd/setup/main.go`, add to the `setupFunctions` map:

```go
var setupFunctions = map[string]SetupFunction{
    // ... existing functions ...
    "your-function": {
        Name:        "Your Function Name",
        Description: "Description of what it does",
        Execute:     (*SetupManager).yourFunction,
    },
}
```

### 2. Implement the Function

```go
func (s *SetupManager) yourFunction() error {
    fmt.Println("Doing your setup task...")
    
    // Your setup logic here
    // You have access to:
    // - s.db (database connection)
    // - s.config (configuration)
    
    return nil
}
```

### 3. Add Makefile Target (Optional)

```makefile
setup-your-function:
	@echo "Running your function..."
	@go run apps/backend/cmd/setup/main.go your-function
```

### Example Extension Ideas

- **`servers`** - Create sample MCP servers
- **`sessions`** - Create active test sessions
- **`logs`** - Generate sample log entries
- **`policies`** - Create rate limiting policies
- **`apikeys`** - Generate test API keys
- **`backup`** - Create database backup before reset
- **`restore`** - Restore from backup

## Database Requirements

The setup script requires:

1. **PostgreSQL database** running and accessible
2. **Migrations applied** - all tables must exist
3. **Configuration file** at `apps/backend/configs/development.yaml`

### Environment Variables

Ensure your `.env` file has database configuration:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=your_user
DB_PASSWORD=your_password
DB_DATABASE=mcp_gateway
DB_SCHEMA=public
```

## Error Handling

The setup script includes robust error handling:

- **Database connection failures** - Clear error messages
- **Missing migrations** - Checks table existence before proceeding
- **Duplicate data** - Safely skips existing records
- **Configuration errors** - Validates config file loading

## Security Notes

⚠️ **Development Only**: This script creates users with known passwords and should **never be used in production**.

- Default passwords are intentionally simple for development
- Admin credentials are hardcoded for convenience
- No password complexity requirements

For production deployment, use proper user management and strong authentication.

## Troubleshooting

### "table does not exist" Error
```bash
# Run migrations first
make migrate
```

### "Failed to connect to database"
```bash
# Start database container
make docker-run

# Check if database is accessible
psql -h localhost -p 5432 -U your_user -d mcp_gateway
```

### "Config file not found"
```bash
# Ensure development config exists
ls apps/backend/configs/development.yaml

# Copy from example if needed
cp apps/backend/configs/development.yaml.example apps/backend/configs/development.yaml
```

### Reset Everything
```bash
# Nuclear option - restart everything
make docker-down
make docker-run
make migrate
make setup-admin
```