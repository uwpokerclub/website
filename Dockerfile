FROM golang:1.18.3-stretch

# Set Golang build environment variables
ENV GO111MODULE=on

# Set directory to run app in
WORKDIR /usr/app

# Copy go.mod and go.sum files over
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Install development dependencies
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.6.1

RUN go mod verify
RUN go mod vendor

# Copy all other files
COPY . .

# Build executable
RUN go build -mod=vendor -o api .

CMD [ "./api" ]