#!/bin/bash

set -e

echo "Starting Atlas migration process..."

# Baseline version - the last migration before Atlas was adopted
# This is used for databases that were set up before Atlas migration tracking
BASELINE_VERSION="20250817202602"

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

    # Try to apply migrations normally first
    set +e
    output=$(atlas migrate apply \
        --dir "file://atlas/migrations" \
        --url "$db_url" 2>&1)
    exit_code=$?
    set -e

    if [ $exit_code -eq 0 ]; then
        echo "$output"
        echo "Successfully applied migrations to $db_name database"
        return 0
    fi

    # Check if the error is about a dirty database (existing schema without Atlas tracking)
    if echo "$output" | grep -q "not clean"; then
        echo "Database has existing schema without Atlas tracking, applying with baseline $BASELINE_VERSION..."
        atlas migrate apply \
            --dir "file://atlas/migrations" \
            --url "$db_url" \
            --baseline "$BASELINE_VERSION"
        echo "Successfully applied migrations to $db_name database (with baseline)"
    else
        # Some other error occurred
        echo "$output"
        exit $exit_code
    fi
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
