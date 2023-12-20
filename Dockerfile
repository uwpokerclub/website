FROM golang:1.18.3-stretch as build

# Set Golang build environment variables
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

# Set directory to run app in
WORKDIR /app

# Copy go.mod and go.sum files over
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Install development dependencies
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.6.1
RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go mod verify

# Copy all other files
COPY . .

# Build executable
RUN go build -o /app/server .

EXPOSE 5000

# ====================== THIN IMAGE ======================
FROM alpine:3.19.0

WORKDIR /app

# Install bash for easy ssh
RUN apk --no-cache add bash

# Copy compiled executable
COPY --from=build /app/server .

# Copy migration files and goose binary
COPY --from=build /go/bin/goose /bin
COPY --from=build /app/migrations ./migrations

RUN mkdir certs client-cert client-key

CMD [ "/app/server", "start" ]