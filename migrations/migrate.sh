#!/bin/bash

# This script runs database migrations

# Check if psql is installed
if ! command -v psql &> /dev/null
then
    echo "psql could not be found. Please install PostgreSQL client."
    exit 1
fi

# Source environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Set default values if not in environment
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5000}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_NAME=${DB_NAME:-belimang}

echo "Running database migration..."

# Connect to database and run migration
psql "host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME" -f scripts/init.sql

if [ $? -eq 0 ]; then
    echo "Database migration completed successfully."
else
    echo "Database migration failed."
    exit 1
fi