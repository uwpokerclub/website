### ====================== SERVER BUILD IMAGE ======================
FROM golang:1.26.0-bookworm AS server

WORKDIR /usr/server

# Copy go.mod and go.sum files over
COPY go.mod go.sum ./

# Install atlas CLI and delve debugger
RUN apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/* && \
    curl -sSf https://atlasgo.sh | sh && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# Download dependencies
RUN go mod download

# Copy all other files
COPY . .

# Compile server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api .

CMD ["./api", "start"]