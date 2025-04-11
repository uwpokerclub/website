### ====================== SERVER BUILD IMAGE ======================
FROM golang:1.24.2-bookworm AS server

WORKDIR /usr/server

# Copy go.mod and go.sum files over
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && \
    go install github.com/pressly/goose/v3/cmd/goose@v3.15.1 && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# Copy all other files
COPY . .

# Compile server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api .

CMD ["./api", "start"]