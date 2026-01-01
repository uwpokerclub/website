# =================== DOCKER COMMANDS ==================
docker-build:
	docker compose build server webapp db

docker-up:
	docker compose up -d server webapp db

docker-down:
	docker compose down

ssh-server:
	docker compose exec -it server /bin/bash

ssh-db:
	docker compose exec -it db psql -U docker -d uwpokerclub_development

test:
	docker compose exec server go test -p=1 ./...

logs:
	docker compose logs -f 

# =================== GENERATORS ==================
generate-api-docs:
	docker compose run --rm generate-api-docs

generate-migration:
	cd server && atlas migrate diff --config "file://atlas/atlas.hcl" --env gorm

# Generate a new empty migration file for custom SQL changes
generate-manual-migration:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Migration name is required. Usage: make generate-manual-migration NAME=your_migration_name"; \
		exit 1; \
	fi
	cd server && atlas migrate new $(NAME) --config "file://atlas/atlas.hcl" --env gorm

migrate-hash:
	cd server && atlas migrate hash --config "file://atlas/atlas.hcl" --env gorm

# =================== MIGRATIONS ==================
# Apply Atlas migrations to both development and test databases
migrate:
	docker compose run --rm atlas-migrate


