#!/bin/bash

set -e
set -u

function create_user_and_database() {
	for dbenv in $(echo $DATABASE_ENVS | tr ',' ' '); do
		local database="$1_$dbenv"
		echo "Creating user and database '$database'"
		psql -v --username "$POSTGRES_USER" -U "$POSTGRES_USER" <<-EOSQL
		    CREATE USER $database;
		    CREATE DATABASE $database;
		    GRANT ALL PRIVILEGES ON DATABASE $database TO $database;
		EOSQL
	done
}

if [ -n "$DATABASES" ]; then
	echo "Multiple database creation requested: $DATABASES"
	for db in $(echo $DATABASES | tr ',' ' '); do
		create_user_and_database $db
	done
	echo "Multiple databases created"
fi