# Cypress E2E Tests

## Prerequisites

- Docker and Docker Compose
- Node.js
- `GITHUB_PACKAGE_TOKEN` environment variable set (required for building the webapp Docker image, which pulls `@uwpokerclub/components` from GitHub Packages)

## Setup

### 1. Create the external Docker network (one-time)

```bash
docker network create uwpokerclub_services_network
```

### 2. Build and start services

```bash
docker compose -f cypress/docker-compose.yml build --build-arg GITHUB_PACKAGE_TOKEN=${GITHUB_PACKAGE_TOKEN}
docker compose -f cypress/docker-compose.yml up -d
```

## Running Tests

```bash
# Headless
npm test

# Interactive UI
npm run cy:open
```

## Resetting the Database

```bash
npm run db:reset
```

## Stopping Services

```bash
docker compose -f cypress/docker-compose.yml down
```
