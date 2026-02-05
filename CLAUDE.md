# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

UWPSC Website is a full-stack application for the University of Waterloo Poker Studies Club built with Go (backend), React+TypeScript (frontend), and PostgreSQL (database). The app manages poker tournaments, member rankings, event scheduling, and club administration.

## Development Commands

### Backend (Go API Server)
```bash
# Run server locally (from server/ directory)
go run main.go start

# Run all backend tests (disable parallelization due to shared database)
# IMPORTANT: The -C option must come immediately after 'test' and before package paths
go test -C server ./internal/... -v -p=1

# Run specific test suite
go test -C server ./internal/services -v -p=1 --run MembershipService

# Run tests for a specific package (e.g., controller tests)
go test -C server ./internal/controller -v -p=1

# Run a specific test function
go test -C server ./internal/controller -v -p=1 --run TestRestartEvent

# Generate API documentation
make generate-api-docs
# OR: docker compose run --rm generate-api-docs
```

### Frontend (React + TypeScript)
```bash
# From webapp/ directory
npm install
npm run dev          # Development server (port 5173)
npm run build        # Production build
npm run lint         # ESLint check
npm run lint:fix     # ESLint with auto-fix
npm test             # Jest tests
```

### Database Management (Atlas + PostgreSQL)
```bash
# Apply migrations to dev and test databases
make migrate
# OR: docker compose run --rm atlas-migrate

# Generate migration from GORM model changes
make generate-migration
# OR: cd server && atlas migrate diff --config "file://atlas/atlas.hcl" --env gorm

# Generate empty migration file for custom SQL
make generate-manual-migration NAME=your_migration_name
```

### Docker Operations
```bash
# Start all services
docker compose up -d

# Start only database for local development
docker compose up -d db

# View logs
docker compose logs -f [service-name]

# Database shell
make ssh-db
# OR: docker compose exec -it db psql -U docker -d uwpokerclub_development

# Server shell
make ssh-server
```

### Testing & Quality Assurance
```bash
# Root directory - Cypress E2E tests
npm test                # Run Cypress tests
npm run cy:open         # Open Cypress UI
npm run db:seed         # Seed test data
npm run db:reset        # Reset test database
```

## Architecture & Code Structure

### Backend Architecture (Go + Gin)
- **MVC Pattern**: Controllers handle HTTP requests, Services contain business logic
- **Layer Structure**:
  - `server/internal/controller/` - HTTP route handlers (new v2 controller pattern)
  - `server/internal/server/` - Legacy HTTP handlers (v1 API)
  - `server/internal/services/` - Business logic layer
  - `server/internal/models/` - GORM database models
  - `server/internal/authentication/` - Session and credential management
  - `server/internal/authorization/` - Role-based access control
  - `server/internal/middleware/` - HTTP middleware (auth, CORS, etc.)

### API Versioning
The API has two versions:
- **v1** (`/api/*`): Legacy routes in `server.go` using handler methods
- **v2** (`/api/v2/*`): New controller-based routes implementing `Controller` interface

When adding new endpoints, prefer the v2 controller pattern in `internal/controller/`.

### Database Layer
- **ORM**: GORM with PostgreSQL
- **Migrations**: Atlas CLI for schema management (replaces traditional migrations)
- **Models**: Located in `internal/models/`, used by Atlas to auto-generate migrations
- **Testing**: Uses testcontainers for isolated database testing

### Authentication & Authorization
- **Session-based auth**: Managed via `internal/authentication/session_manager.go`
- **Role-based access**: Implemented in `internal/authorization/` with per-resource authorizers
- **Middleware**: Applied at route level with `UseAuthentication` and `UseAuthorization`

### Testing Strategy
- **Backend**: Unit tests with mocks, integration tests with testcontainers
- **Database**: Tests run with `-p=1` flag (no parallelization due to shared DB cleanup)
- **E2E**: Cypress tests from repository root

### Frontend Structure (React + TypeScript)
- **Build Tool**: Vite for development and production builds
- **Routing**: React Router DOM
- **Styling**: CSS modules with responsive design
- **Package Manager**: pnpm (specified in package.json)

## Important Development Notes

### Database Operations
- Always use Atlas for schema migrations, not manual SQL files
- Model changes in `internal/models/` automatically generate migrations
- Test database is completely reset after each test run
- Use `make ssh-db` for direct database access during development

### Code Patterns
- Follow existing error handling patterns in `internal/errors/errors.go`
- Use testutils in `internal/testutils/` for consistent test setup
- Authorization follows resource-based pattern: `[resource].[action]` permissions

### Environment Setup
- Development uses Docker Compose with local development option
- Frontend dev server proxies API calls to backend
- Database runs in Docker even during local development
- Environment variables are defined in docker-compose.yml

## Production Deployment
- Multi-stage Docker build combining frontend and backend
- Frontend built into static assets served by Go server
- PostgreSQL with Atlas migrations applied on deployment

## Go Testing Guidelines
- **CRITICAL**: When running Go tests, always use the `-C` option immediately after `test` and before package paths
- Example: `go test -C server ./internal/controller -v -p=1` (correct)
- Never: `go test ./internal/controller -v -p=1 -C` (incorrect - will fail)
- The `-C` option changes to the specified directory before running tests
- Always disable parallelization with `-p=1` due to shared database in tests

## Jira CLI (jira-cli)
- **Project**: UWPSC at https://uwpokerclub.atlassian.net
- **Config**: `~/.config/.jira/.config.yml`

### Creating Issues
```bash
# Create issue with template file for body
jira issue create --type Task --parent "UWPSC-XX" --priority High --no-input \
  --summary "Issue summary" \
  --template /path/to/body.md

# Create epic
jira epic create --name "Epic Name" --summary "Epic summary" --priority High --no-input \
  --body "Epic description"
```

### Custom Fields
- **IMPORTANT**: Use dash-case field names, not field keys
- Example: `--custom acceptance-criteria="criteria here"` (correct)
- Not: `--custom customfield_10040="criteria here"` (incorrect - will be ignored)

### Editing Issues
```bash
# Update description via pipe
cat description.md | jira issue edit UWPSC-XX --no-input

# Set custom fields
jira issue edit UWPSC-XX --no-input --custom acceptance-criteria="- Criterion 1
- Criterion 2"
```