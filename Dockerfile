FROM golang:1.18.3-stretch

# Set Golang build environment variables
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

# Set directory to run app in
WORKDIR /usr/app

# Copy go.mod and go.sum files over
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify
RUN go mod vendor

# Copy all other files
COPY . .

# Build executable
RUN go build -o api .

CMD [ "./api" ]