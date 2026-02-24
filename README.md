# UWPSC Website

A modern web application for the University of Waterloo Poker Studies Club (UWPSC) built with Go, React, and PostgreSQL.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Documentation](#documentation)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
  - [Development Setup](#development-setup)
- [Database Management](#database-management)
  - [Atlas Migration System](#atlas-migration-system)
  - [Running Migrations](#running-migrations)
  - [Creating New Migrations](#creating-new-migrations)
- [API Documentation](#api-documentation)
- [Frontend Development](#frontend-development)
  - [Development Server](#development-server)
  - [Frontend Testing](#frontend-testing)
  - [Production Build](#production-build)
- [Backend Development](#backend-development)
  - [Server Testing](#server-testing)
  - [Environment Variables](#environment-variables)
- [Docker Services](#docker-services)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Overview

The UWPSC Website is a full-stack application that manages poker tournaments, member rankings, event scheduling, and administrative functions for the University of Waterloo Poker Studies Club.

### Key Features

- **Tournament Management**: Create and manage poker events with custom structures
- **Member Management**: Handle memberships, payments, and club administration
- **Rankings System**: Track player performance and semester rankings
- **Authentication**: Role-based access control for different user types
- **Financial Tracking**: Manage club budget and transaction history

## Architecture

The application follows a modern microservices architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React WebApp  │    │   Go API Server │    │   PostgreSQL    │
│   (Frontend)    │◄──►│   (Backend)     │◄──►│   Database      │
│   Port: 5173    │    │   Port: 5000    │    │   Port: 5432    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Frontend (React + TypeScript)**
- Modern React application with TypeScript
- Vite for fast development and building
- Responsive design for desktop and mobile

**Backend (Go + Gin)**
- RESTful API built with Go and Gin framework
- GORM for database ORM
- Structured logging and error handling
- OpenAPI/Swagger documentation

**Database (PostgreSQL)**
- PostgreSQL 17 for data persistence
- Atlas for schema migrations
- Optimized indexing and constraints

## Documentation

Additional documentation is available in the [`docs/`](docs/) directory:

- [Database Schema](docs/database-schema.md) - ERD, table definitions, relationships, and cascade behavior

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- [Git](https://git-scm.com/)
- [Node.js](https://nodejs.org/en/) (version 18+) - for local frontend development
- [Go](https://go.dev/) (version 1.24+) - for local backend development
- [Atlas CLI](https://atlasgo.io/getting-started#installation) - for database migrations

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/uwpokerclub/website.git
   cd website
   ```

2. **Create the Docker network**
   ```bash
   docker network create uwpokerclub_services_network
   ```

3. **Start all services**
   ```bash
   docker compose up -d
   ```

4. **Access the application**
   - Frontend: http://localhost:5173
   - API: http://localhost:5000
   - API Documentation: http://localhost:5000/swagger/index.html

### Development Setup

For active development, you can run services individually:

```bash
# Start database only
docker compose up -d db

# Run API server in development mode
cd server
go run main.go start

# Run frontend in development mode (in another terminal)
cd webapp
npm install
npm run dev
```

## Database Management

### Atlas Migration System

The project uses [Atlas](https://atlasgo.io/) for database schema management, providing:

- **Declarative Migrations**: Define your desired schema state
- **Automatic Migration Generation**: Atlas generates SQL migrations from GORM models
- **Safety Checks**: Validates migrations before applying
- **Version Control**: Track schema changes with version control

### Running Migrations

**Using Docker (Recommended)**
```bash
# Run migrations on both development and test databases
docker compose run --rm atlas-migrate
```

**Manual Migration**
```bash
# Apply migrations to development database (run from server directory)
cd server
atlas migrate apply \
  --config "file://atlas/atlas.hcl" \
  --env gorm \
  --url "postgres://docker:password@localhost:5432/uwpokerclub_development?sslmode=disable"
```

### Creating New Migrations

**Automatic Migrations (from GORM model changes)**

When you modify GORM models, generate migrations using:

```bash
# Generate migration from GORM model changes
make generate-migration
```

Or manually:
```bash
cd server
atlas migrate diff \
  --config "file://atlas/atlas.hcl" \
  --env gorm
```

**Manual Migrations (for custom SQL)**

For custom database changes that can't be expressed through GORM models (like data migrations, custom indexes, etc.):

```bash
# Using Make (recommended)
make generate-manual-migration NAME=your_migration_name

# Or run Atlas directly
cd server
atlas migrate new your_migration_name \
  --config "file://atlas/atlas.hcl" \
  --env gorm
```

This creates a new migration file in `server/atlas/migrations/` that you can edit with custom SQL.

**Running Commands Without Make**

All migration commands can be run directly using Atlas CLI:

```bash
# Generate migration from GORM model changes
cd server
atlas migrate diff --config "file://atlas/atlas.hcl" --env gorm

# Generate manual migration
cd server
atlas migrate new your_migration_name --config "file://atlas/atlas.hcl" --env gorm

# Apply migrations to development database
cd server
atlas migrate apply \
  --config "file://atlas/atlas.hcl" \
  --env gorm \
  --url "postgres://docker:password@localhost:5432/uwpokerclub_development?sslmode=disable"
```

**Note**: Atlas automatically generates migrations from GORM model changes. The traditional `migrate.sh` script has been replaced with Atlas for better schema management.

## API Documentation

The API is documented using OpenAPI/Swagger specifications:

- **Local**: http://localhost:5000/swagger/index.html
- **Generate Docs**: `docker compose run --rm generate-api-docs`

The API follows RESTful conventions with endpoints for:
- Authentication and session management
- User and membership management
- Event and tournament operations
- Rankings and statistics
- Financial transactions

## Frontend Development

### Development Server

**Using Docker:**
```bash
# Start complete development environment
docker compose up -d
```

**Local Development:**
```bash
cd webapp
npm install
npm run dev
```

The development server will start on [http://localhost:5173](http://localhost:5173) and includes hot reloading for code changes.

### Frontend Testing

**Unit and Component Tests:**
```bash
cd webapp
npm test
```

**Interactive Testing:**
```bash
cd webapp
npm run test:watch
```

### Production Build

**Complete Application Build:**
```bash
# Build the complete application (from root directory)
docker build -t uwpsc-website .
```

**Frontend Only Build:**
```bash
cd webapp
npm run build
```

The build process:
1. Installs dependencies with `npm ci --omit-dev`
2. Builds the React app with `npm run build`
3. Creates an optimized production bundle
4. Serves static files through the Go API server

## Backend Development

### Server Testing

**Run All Tests:**
```bash
# Using Docker
docker compose exec server go test ./internal/... -v -p=1

# Local development
cd server
go test ./internal/... -v -p=1
```

**Run Specific Test Suites:**
```bash
docker compose exec server go test ./internal/... -v -p=1 --run MembershipService
```

**Note**: The `-p=1` flag disables test parallelization, which is crucial since the database is wiped after every test run.

### Environment Variables

The following environment variables are required for backend development:

```bash
# Database Configuration
DATABASE_URL=postgres://user:password@host:port/database
TEST_DATABASE_URL=postgres://user:password@host:port/test_database

# Server Configuration
PORT=5000
HOSTNAME=localhost:5000
ENVIRONMENT=development

# Database SSL (optional)
DATABASE_TLS_PARAMETERS=?sslmode=disable
```

## Docker Services

The application is containerized with multiple services:

| Service | Description | Port |
|---------|-------------|------|
| `webapp` | React frontend application | 5173 |
| `server` | Go API server | 5000 |
| `db` | PostgreSQL database | 5432 |
| `generate-api-docs` | Swagger documentation generator | - |
| `atlas-migrate` | Database migration service | - |

**Useful Docker Commands**
```bash
# View logs
docker compose logs -f [service-name]

# Rebuild services
docker compose build [service-name]

# Execute commands in containers
docker compose exec server bash
docker compose exec db psql -U docker -d uwpokerclub_development

# Clean up
docker compose down -v  # Remove containers and volumes
```

## Testing

### Backend Tests
```bash
# Run all tests
cd server
go test ./internal/... -v -p=1

# Run specific test suites
go test ./internal/services/... -v -p=1 --run MembershipService

# Run tests in Docker
docker compose exec server go test ./internal/... -v -p=1
```

### End-to-End Tests
```bash
# Run Cypress tests
npm test

# Open Cypress UI
npm run cy:open

# Reset test database
npm run db:reset
```

### Database Utilities
```bash
# Seed test data
npm run db:seed

# Reset database
npm run db:reset
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

---

**University of Waterloo Poker Studies Club**  
Contact: uwaterloopoker@gmail.com
