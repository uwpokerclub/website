build:
	docker compose build

start:
	docker compose up -d

down:
	docker compose down

exec_server:
	docker compose exec server /bin/bash

exec_db:
	docker compose exec db psql -U docker
