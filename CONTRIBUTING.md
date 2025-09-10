# Contributing to Omnimesh Gateway

Thank you for your interest in contributing to Omnimesh Gateway! We welcome contributions from the community and are excited to work with you.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)

## 📜 Code of Conduct

This project and everyone participating in it is governed by the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have:

- **Go 1.25+** installed ([Installation Guide](https://golang.org/doc/install))
- **PostgreSQL 12+** for database backend
- **Docker & Docker Compose** for development environment
- **Git** for version control
- **Make** for running build commands

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/omnimesh-gateway.git
   cd omnimesh-gateway
   ```
3. Add the original repository as a remote:
   ```bash
   git remote add upstream https://github.com/omnimesh-labs/omnimesh-gateway.git
   ```

## 🛠️ Development Setup

### 1. Environment Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your local settings
# At minimum, set a secure JWT_SECRET
```

### 2. Install Development Tools

```bash
# Install pre-commit hooks (recommended)
make setup-precommit

# Install Go tools (if not already installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/github-action-gosec@latest
```

### 3. Start Development Environment

```bash
# Start PostgreSQL and Redis
make docker-run

# Run database migrations
make migrate

# Create admin user for testing
make setup-admin
```

### 4. Verify Setup

```bash
# Build the project
make build

# Run tests
make test

# Run linting
make lint

# Start the development server
make watch
```

## 🔧 Making Changes

### Branch Naming

Use descriptive branch names with prefixes:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements

Examples:
```bash
git checkout -b feature/add-graphql-support
git checkout -b fix/rate-limiting-bug
git checkout -b docs/update-api-guide
```

### Development Workflow

1. **Create a new branch** from `main`:
   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our code style guidelines

3. **Write or update tests** for your changes

4. **Run the test suite**:
   ```bash
   make test
   make test-transport  # For transport layer changes
   make test-integration  # For integration changes
   ```

5. **Run code quality checks**:
   ```bash
   make lint          # Check for issues
   make lint-fix      # Auto-fix issues
   make security      # Security scan
   ```

6. **Update documentation** if needed

## 📝 Submitting Changes

### Before Submitting

Ensure your changes pass all checks:

```bash
# Run all tests
make test

# Run linting
make lint

# Run security checks
make security

# Run pre-commit on all files
make precommit-all
```

### Pull Request Process

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub with:
   - Clear title describing the change
   - Detailed description of what you changed and why
   - Reference any related issues
   - Screenshots if applicable (for UI changes)

3. **Fill out the PR template** completely

4. **Ensure CI passes** - all GitHub Actions must be green

5. **Respond to feedback** - address reviewer comments promptly

### PR Requirements

- [ ] Tests pass locally and in CI
- [ ] Code follows project style guidelines
- [ ] Security scan passes
- [ ] Documentation updated if needed
- [ ] PR description is clear and complete
- [ ] No merge conflicts with main branch

## 🎨 Code Style

### Go Code Style

We follow standard Go conventions with some additions:

- **gofmt** - Code must be formatted
- **golangci-lint** - Must pass linting checks
- **Comments** - Public functions must have comments
- **Error handling** - Always handle errors appropriately
- **Package naming** - Use clear, lowercase package names

### Key Guidelines

```go
// Good: Clear function with proper error handling
func GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
    if id == uuid.Nil {
        return nil, errors.New("user ID cannot be empty")
    }
    
    // Implementation...
    return user, nil
}

// Bad: No error handling or validation
func GetUser(id string) *User {
    // Implementation...
    return user
}
```

### Configuration

- Use environment variables for secrets
- Provide sensible defaults
- Document all configuration options

### Database

- Use parameterized queries to prevent SQL injection
- Follow migration naming conventions
- Include both up and down migrations

## 🧪 Testing

### Test Categories

1. **Unit Tests**: Test individual functions and methods
   ```bash
   make test-unit
   ```

2. **Integration Tests**: Test service interactions
   ```bash
   make test-integration
   ```

3. **Transport Tests**: Test all transport layers
   ```bash
   make test-transport
   ```

4. **Coverage**: Run tests with coverage reporting
   ```bash
   make test-coverage
   ```

### Writing Tests

- Write tests for new functionality
- Update tests when modifying existing code
- Use testcontainers for database testing
- Mock external dependencies
- Test both success and error cases

Example:
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserRequest
        want    *User
        wantErr bool
    }{
        {
            name: "valid user creation",
            input: CreateUserRequest{
                Email: "test@example.com",
                Name:  "Test User",
            },
            want: &User{
                Email: "test@example.com",
                Name:  "Test User",
            },
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation...
        })
    }
}
```

## 📚 Documentation

### What to Document

- **API Changes**: Update API documentation
- **Configuration**: Document new config options
- **Features**: Update feature documentation
- **Architecture**: Update architecture docs for significant changes

### Documentation Standards

- Use clear, concise language
- Include code examples
- Update relevant README sections
- Add inline comments for complex logic

## 🏷️ Issue Labels

When creating issues, use appropriate labels:

- `bug` - Something isn't working
- `enhancement` - New feature or request
- `documentation` - Improvements or additions to docs
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention is needed
- `security` - Security-related issues

## 🔍 Review Process

### For Contributors

- Be responsive to feedback
- Make requested changes promptly
- Ask questions if feedback is unclear
- Keep PRs focused and atomic

### For Reviewers

- Be constructive and helpful
- Explain the "why" behind suggestions
- Approve when ready, request changes when needed
- Consider security implications

## 📞 Getting Help

- **Questions**: Open a GitHub issue with the `question` label
- **Security Issues**: Follow our [Security Policy](SECURITY.md)
- **General Discussion**: Start a GitHub Discussion

## 🙏 Recognition

Contributors are recognized in:
- GitHub contributor list
- Release notes for significant contributions
- Special mentions for first-time contributors

---

Thank you for contributing to Omnimesh Gateway! Your efforts help make this project better for everyone. 🚀
