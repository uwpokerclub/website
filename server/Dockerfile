### ====================== SERVER BUILD IMAGE ======================
FROM golang:1.26.0-bookworm AS server

WORKDIR /usr/server

# Copy go.mod and go.sum files over
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && \
    go install ariga.io/atlas/cmd/atlas@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# Copy all other files
COPY . .

# Compile server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api .

CMD ["./api", "start"]