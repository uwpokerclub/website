# Architecture Overview

This document describes the high-level architecture, internal backend structure, authentication flow, and deployment pipeline for the UWPSC Website.

## System Overview

```mermaid
flowchart TB
    subgraph Client
        Browser["Browser"]
    end

    subgraph GCP["Google Cloud Platform"]
        CloudRun["Cloud Run"]
        CloudSQL["Cloud SQL - PostgreSQL 17"]
    end

    subgraph CloudRun
        Gin["Go + Gin API Server"]
        Static["Static Assets - React SPA"]
    end

    Browser -->|"HTTPS"| CloudRun
    Gin -->|"GORM"| CloudSQL
    Gin -->|"Serves"| Static

    subgraph CI["GitHub Actions"]
        Build["Docker Build"]
        Tests["Tests"]
        Push["Push Image"]
        Deploy["Deploy"]
    end

    Build --> Tests --> Push -->|"Artifact Registry"| Deploy -->|"gcloud run deploy"| CloudRun
```

The application is a three-tier system: a React single-page application (built with Vite) is compiled into static assets and served by a Go API server running on GCP Cloud Run. The server connects to a Cloud SQL PostgreSQL 17 instance for persistence.

## Backend Architecture

```mermaid
flowchart LR
    Request["HTTP Request"]

    subgraph Middleware
        CORS["CORS"]
        Auth["Authentication"]
        Authz["Authorization"]
    end

    subgraph Handlers
        V1["v1 Legacy Handlers - /api/*"]
        V2["v2 Controllers - /api/v2/*"]
    end

    subgraph Services
        BizLogic["Business Logic"]
    end

    subgraph Data
        GORM["GORM ORM"]
        PG["PostgreSQL"]
    end

    Request --> CORS --> Auth --> Authz
    Authz --> V1
    Authz --> V2
    V1 --> BizLogic
    V2 --> BizLogic
    V1 --> GORM
    V2 --> GORM
    BizLogic --> GORM
    GORM --> PG
```

### Key packages

| Package | Path | Purpose |
|---------|------|---------|
| `server` | `internal/server/` | v1 legacy HTTP handlers and route setup |
| `controller` | `internal/controller/` | v2 controllers implementing `Controller` interface |
| `services` | `internal/services/` | Business logic (membership budget, event state) |
| `models` | `internal/models/` | GORM database models |
| `middleware` | `internal/middleware/` | Auth, CORS, and authorization middleware |
| `authentication` | `internal/authentication/` | Session and credential management |
| `authorization` | `internal/authorization/` | Role-based permission checking |
| `errors` | `internal/errors/` | Structured API error responses |

### v1 vs v2 API

- **v1** (`/api/*`): Handler methods on `apiServer` struct in `internal/server/`. Registered in `SetupRoutes()`.
- **v2** (`/api/v2/*`): Standalone controllers implementing the `Controller` interface with a `LoadRoutes(router *gin.RouterGroup)` method. Registered in `SetupV2Routes()`. New endpoints should use this pattern.

## Authentication and Authorization Flow

### Login

```mermaid
sequenceDiagram
    participant Browser
    participant Gin as Go API Server
    participant Cred as CredentialService
    participant SM as SessionManager
    participant DB as PostgreSQL

    Browser->>Gin: POST /api/v2/session (username, password)
    Gin->>Cred: Validate(username, password)
    Cred->>DB: Look up credentials
    DB-->>Cred: Credential record
    Cred-->>Gin: valid, role
    Gin->>SM: Create(username, role)
    SM->>DB: Insert session row
    DB-->>SM: Session token (UUID)
    SM-->>Gin: Token
    Gin-->>Browser: 201 Created + Set-Cookie (session ID)
```

### Authenticated Request

```mermaid
sequenceDiagram
    participant Browser
    participant CORS as CORS Middleware
    participant AuthN as Authentication Middleware
    participant AuthZ as Authorization Middleware
    participant Ctrl as Controller
    participant DB as PostgreSQL

    Browser->>CORS: Request + session cookie
    CORS->>AuthN: Pass through
    AuthN->>DB: SessionManager.Authenticate(sessionID)
    DB-->>AuthN: Session (username, role)
    AuthN->>AuthZ: Set username and role on context
    AuthZ->>AuthZ: AuthorizationService.IsAuthorized(action)
    alt Authorized
        AuthZ->>Ctrl: Next()
        Ctrl->>DB: Query/Mutate
        DB-->>Ctrl: Result
        Ctrl-->>Browser: JSON response
    else Forbidden
        AuthZ-->>Browser: 403 Forbidden
    end
```

### Frontend Auth

The React app uses an `AuthProvider` that wraps the component tree:

1. On mount, `AuthProvider` calls `GET /api/v2/session` to check for an existing session cookie.
2. If authenticated, the user session (username, role, permissions) is stored in React context.
3. `RequireAuth` wraps admin routes and redirects unauthenticated users to `/admin/login`.
4. `hasPermission(action, resource)` checks the permissions map returned by the server before rendering UI elements.

## Deployment Pipeline

```mermaid
flowchart LR
    subgraph Trigger
        PushEvent["Push to branch"]
        Release["GitHub Release"]
    end

    subgraph CI["CI Workflow"]
        Build["Docker multi-stage build"]
        WebTests["Webapp tests + lint"]
        ServerTests["Go tests + migrations"]
        E2E["Cypress E2E tests"]
    end

    subgraph Publish["Publish Workflow"]
        BuildRelease["Docker build"]
        PushRelease["Push to Artifact Registry"]
        DeployRelease["gcloud run deploy"]
    end

    PushEvent --> Build
    Build --> WebTests
    Build --> ServerTests
    WebTests --> E2E
    ServerTests --> E2E
    E2E -->|"Push to Artifact Registry"| PushImage["Push image"]

    Release --> BuildRelease --> PushRelease --> DeployRelease
```

### CI workflow (`ci.yml`)

Runs on every push and merge group event:

1. **Build** - Multi-stage Docker build producing the combined frontend + backend image.
2. **Webapp tests** - Lint and Jest unit tests (Node.js 22).
3. **Server tests** - Go tests against a PostgreSQL 17 container with Atlas migrations applied.
4. **E2E** - Cypress tests against the built Docker image with a fresh database.
5. **Push** - Tags and pushes the image to GCP Artifact Registry (on push events only).

### Publish workflow (`publish.yml`)

Runs on GitHub release creation:

1. **Push** - Builds and pushes a tagged image plus `latest` to Artifact Registry.
2. **Deploy** - Deploys the tagged image to Cloud Run via `gcloud run deploy`.

### Container startup

The final Docker image runs `./server start --run-migrations`, which applies pending Atlas migrations on startup before serving traffic.
