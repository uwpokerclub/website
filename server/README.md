# UWPSC API Server

> **📖 For complete documentation, please refer to the main [README.md](../README.md)**

This directory contains the Go API server for the UWPSC Website.

## Quick Start

From the root directory:
```bash
# Start all services including the API server
docker compose up -d

# Or run the server locally for development
cd server
go run main.go start
```

## Key Information

- **Port**: 5000
- **API Documentation**: http://localhost:5000/swagger/index.html
- **Technology**: Go + Gin framework + GORM + PostgreSQL
- **Testing**: `docker compose exec server go test ./internal/... -v -p=1`

## Directory Structure

- `atlas/` - Database migration configuration (Atlas)
- `cmd/` - CLI commands and application entry points
- `docs/` - Generated API documentation (Swagger)
- `internal/` - Private application code
  - `models/` - GORM database models
  - `services/` - Business logic layer
  - `server/` - HTTP handlers and middleware
  - `database/` - Database connection and utilities
- `migrations/` - Legacy migrations (deprecated, use Atlas)
- `scripts/` - Utility scripts

## Environment Variables

Key environment variables for local development:

```bash
DATABASE_URL=postgres://docker:password@localhost:5432/uwpokerclub_development
PORT=5000
ENVIRONMENT=development
```

## Database Migrations

The server uses [Atlas](https://atlasgo.io/) for database schema management.

### Generating Migrations

**Automatic migrations from GORM model changes:**
```bash
# From root directory
make generate-migration

# Or run Atlas directly (from server directory)
cd server
atlas migrate diff --config "file://atlas/atlas.hcl" --env gorm
```

**Manual migrations for custom SQL:**
```bash
# From root directory
make generate-manual-migration NAME=your_migration_name

# Or run Atlas directly (from server directory)
cd server
atlas migrate new your_migration_name --config "file://atlas/atlas.hcl" --env gorm
```

**Applying migrations:**
```bash
# From root directory
make migrate

# Or run Atlas directly (from server directory)
cd server
atlas migrate apply \
  --config "file://atlas/atlas.hcl" \
  --env gorm \
  --url "postgres://docker:password@localhost:5432/uwpokerclub_development?sslmode=disable"
```

For complete environment setup and database management, see the main [README.md](../README.md).

## License

Licensed under the Apache 2.0 License - see the [LICENSE](../LICENSE) file for details.

