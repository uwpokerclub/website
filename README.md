# UWPSC Admin API

## Table of Contents
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Clone the repository](#clone-the-repository)
  - [Building](#building)
- [Usage](#testing)
  - [Starting the server](#starting-the-server)
- [Migrations](#migrations)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Installation
Before starting development or usage of the API, ensure you have all of the prerequiste software installed below.

### Prerequisites
- [Docker](https://www.docker.com/) (with Docker Compose)
- [Golang](https://go.dev/) (version 1.18+)

After installing both Docker and Golang, you are ready to begin.

### Clone the repository
Clone the repository to your local development environment using:
```sh
# HTTPS
git clone https://github.com/uwpokerclub/api-golang.git
# SSH
git clone git@github.com:uwpokerclub/api-golang.git
# GH CLI
gh repo clone uwpokerclub/api-golang

# Change into project directory
cd api-golang
```

### Building
After cloning the repository, you will now need to build the container image with Docker Compose.
```sh
docker-compose build
```
This will build the image to run the API server in.

## Usage

### Starting the server
To start the server, simply start the container using Docker Compose.
```sh
docker-compose up -d
```
This will start both the API server, and a PostgreSQL database along side.

## Migrations

## Testing

## Contributing

## License
The UWPSC Admin API is licensed under the terms of the Apache 2.0 License and can be found [here](LICENSE).