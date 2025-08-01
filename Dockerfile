### ====================== WEBAPP BUILD IMAGE ======================
FROM node:24.0.1-alpine AS webapp

WORKDIR /usr/app

# Copy package.json and package-lock.json files over
COPY webapp/package.json webapp/package-lock.json ./

# Install dependencies
RUN npm ci --omit-dev

# Copy files into container
COPY webapp .

# Set NODE_ENV to production
ENV NODE_ENV=production

# Build the React app for production
RUN npm run build

### ====================== SETUP IMAGE ======================
FROM golang:1.24.2-bookworm AS setup

WORKDIR /usr/server

COPY server .

### ====================== SERVER BUILD IMAGE ======================
FROM setup AS server

WORKDIR /usr/server

# Copy files over
COPY --from=setup /usr/server .

# Download dependencies
RUN go mod download

# Compile server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /tmp/server .

### ====================== API DOC GENERATION ======================
FROM setup AS generate-api-docs

WORKDIR /usr/server

COPY --from=setup /usr/server .

RUN go install github.com/swaggo/swag/cmd/swag@latest

ENTRYPOINT [ "swag", "init", "--parseDependency" ]

### ====================== ATLAS MIGRATION ======================
FROM setup AS atlas-migrate

WORKDIR /usr/server

COPY --from=setup /usr/server .

RUN curl -sSf https://atlasgo.sh | sh

# Make the migration script executable (it should already be copied from setup stage)
RUN chmod +x /usr/server/scripts/atlas_migrate.sh

ENTRYPOINT [ "/usr/server/scripts/atlas_migrate.sh" ]

### ======================= FINAL IMAGE ========================
FROM debian:bookworm-slim

WORKDIR /server

# Create a non-root user and group
RUN groupadd -r runner && useradd -r -g runner runner

# Change ownership of the working directory
RUN chown runner:runner /server

# Setup certifcates and keys folders
RUN mkdir certs client-cert client-key
RUN chown runner:runner certs client-cert client-key

# Copy the built application
COPY --from=webapp --chown=runner:runner /usr/app/dist ./public
COPY --from=server --chown=runner:runner /tmp/server ./server
COPY --from=server --chown=runner:runner /usr/server/migrations ./migrations

# Switch to non-root user
USER runner

# Start the server
CMD ["./server", "start", "--run-migrations"]
