# UWPSC Admin API

## Table of Contents
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Clone the repository](#clone-the-repository)
  - [Building](#building)
- [Usage](#testing)
  - [Starting the server](#starting-the-server)
- [Migrations](#migrations)
  - [Creating a new migration](#creating-a-new-migration)
  - [Running migrations](#running-migrations)
- [Testing](#testing)
  - [Running all tests](#running-all-tests)
  - [Running a subset of tests](#running-a-subset-of-tests)
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
git clone https://github.com/uwpokerclub/api.git
# SSH
git clone git@github.com:uwpokerclub/api.git
# GH CLI
gh repo clone uwpokerclub/api

# Change into project directory
cd api
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

### Creating a new migration
To create a new migration, run the following command.
```
docker-compose exec server ./scripts/migrate.sh create add_column_to_database sql
```

### Running migrations
To run the most recent migrations, you can run the command
```
docker-compose exec server ./scripts/migrate.sh up
```
This will run the migrations for the development database. To run migrations on the test database:
```
docker-compose exec server ./scripts/migrate.sh --test up
```
To view all possible migration options, you can list them by not specifying an option.
```
docker-compose exec server ./scripts/migrate.sh
```
## Testing

### Running all tests
To run the entire unit test suite, use:
```
docker-compose exec server go test ./internal/... -v -p=1
```
Ensure that `-p=1` is set. This disables parallelization of the tests. This is crucial since the database is wiped after every test run.

### Running a subset of tests
If you only want to run a subset of the test suite, you can use the `--run` option and specify text that matches a test name. For example:
```
docker-compose exec server go test ./internal/... -v -p=1 --run MembershipService
```
This will only run the membership service test suite.

## Contributing
TBD

## License
The UWPSC Admin API is licensed under the terms of the Apache 2.0 License and can be found [here](LICENSE).