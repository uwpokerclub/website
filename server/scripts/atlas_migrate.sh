#!/bin/bash

set -e

echo "Starting Atlas migration process..."

# Function to apply migrations to a database
apply_migrations() {
    local db_url="$1"
    local db_name="$2"
    
    # Add sslmode=disable if not already present
    if [[ "$db_url" != *"sslmode="* ]]; then
        if [[ "$db_url" == *"?"* ]]; then
            db_url="${db_url}&sslmode=disable"
        else
            db_url="${db_url}?sslmode=disable"
        fi
    fi
    
    echo "Applying migrations to $db_name database..."
    echo "Database URL: $db_url"
    
    atlas migrate apply \
        --dir "file://atlas/migrations" \
        --url "$db_url"
    
    echo "Successfully applied migrations to $db_name database"
}

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable is not set"
    exit 1
fi

# Apply migrations to development database
apply_migrations "$DATABASE_URL" "development"

# Apply migrations to test database (if TEST_DATABASE_URL is provided)
if [ -n "$TEST_DATABASE_URL" ]; then
    apply_migrations "$TEST_DATABASE_URL" "test"
else
    echo "TEST_DATABASE_URL not provided, skipping test database migration"
fi

echo "All migrations applied successfully!"
