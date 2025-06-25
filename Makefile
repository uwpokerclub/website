docker-build:
	docker compose build

docker-up:
	docker compose up -d

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