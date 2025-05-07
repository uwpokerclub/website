### ====================== WEBAPP BUILD IMAGE ======================
FROM node:24.0.0-alpine AS webapp

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

### ====================== SERVER BUILD IMAGE ======================
FROM golang:1.24.2-bookworm AS server

WORKDIR /usr/server

# Copy go.mod and go.sum files over
COPY server/go.mod server/go.sum ./

# Download dependencies
RUN go mod download

# Copy all other files
COPY server .

# Compile server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

### ======================= FINAL IMAGE ========================
FROM debian:bookworm-slim

WORKDIR /server

# Create a non-root user and group
RUN groupadd -r runner && useradd -r -g runner runner

# Setup certifcates and keys folders
RUN mkdir certs client-cert client-key
RUN chown runner:runner certs client-cert client-key

# Copy the built application
COPY --from=webapp --chown=runner:runner /usr/app/dist ./public
COPY --from=server --chown=runner:runner /usr/server/server .
COPY --from=server --chown=runner:runner /usr/server/migrations ./migrations

# Switch to non-root user
USER runner

# Start the server
CMD ["./server", "start", "--run-migrations"]
