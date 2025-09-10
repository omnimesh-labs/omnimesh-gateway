# Omnimesh Gateway - Claude Context

## Project Overview
This is a production-ready API gateway for Model Context Protocol (MCP) servers, built with Go and PostgreSQL.

## Key Files for Development
- `CLAUDE.md` - Comprehensive development guide (main reference)
- `.claude/architecture.md` - Detailed system architecture
- `.claude/database-schema.md` - Complete database schema and design
- `IMPLEMENTATION_GUIDE.md` - Step-by-step implementation roadmap

## Quick Reference

### Build Commands
- `make build` - Build the application
- `make run` - Run the application  
- `make test` - Run all tests
- `make watch` - Live reload for development
- `make docker-run` - Start PostgreSQL container
- `make migrate` - Run database migrations

### Key Directories
- `apps/backend/internal/` - Core Go business logic
- `apps/backend/cmd/api/` - Main application entrypoint
- `apps/backend/migrations/` - Database migrations
- `apps/frontend/` - Next.js dashboard
- `tests/` - Test suites

### Architecture Status
- ✅ Complete: Scaffolding, database schema, transport interfaces
- 🔄 Needs Implementation: Business logic in service layers

### Development Approach
1. Read `CLAUDE.md` for comprehensive project understanding
2. Follow `IMPLEMENTATION_GUIDE.md` for step-by-step development
3. Refer to `.claude/architecture.md` for system design details
4. Use `.claude/database-schema.md` for database queries and schema
5. Use `bun` over `npm` for frontend development
