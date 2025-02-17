#!/bin/bash

# PostgreSQL configuration variables
DB_NAME="acme_db"
DB_USER="acme_user"
DB_PASSWORD="acme_password"
DB_PORT=5432
CONTAINER_NAME="acme_postgres"

# Get script directory (so we can reference files relative to the script's location)
SCRIPT_DIR=$(dirname "$0")

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Docker is not running. Please start Docker first."
    exit 1
fi

# Stop and remove existing container if it exists
if docker ps -a | grep -q "$CONTAINER_NAME"; then
    echo "Stopping and removing existing container..."
    docker stop "$CONTAINER_NAME"
    docker rm "$CONTAINER_NAME"
    docker volume rm acme_data
fi

# Start PostgreSQL container with mounted schema and migrations
echo "Starting PostgreSQL container..."
docker volume create acme_data
docker run --name "$CONTAINER_NAME" \
    -e POSTGRES_DB="$DB_NAME" \
    -e POSTGRES_USER="$DB_USER" \
    -e POSTGRES_PASSWORD="$DB_PASSWORD" \
    -v acme_data:/var/lib/postgresql/data \
    -v "$SCRIPT_DIR/../internal/database/migrations":/migrations \
    -v "$SCRIPT_DIR/../internal/database/schema.sql":/schema.sql \
    -p "$DB_PORT":5432 \
    -d postgres:15

# Wait for PostgreSQL to initialize
echo "Waiting for PostgreSQL to start..."
sleep 5

# Initialize the database schema from the schema file
echo "Applying initial schema..."
docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -f /schema.sql

# Apply migrations in sorted order
echo "Applying migration files..."
MIGRATIONS_DIR="$SCRIPT_DIR/../internal/database/migrations"
for sql_file in $(ls "$MIGRATIONS_DIR"/*.sql | sort); do
    filename=$(basename "$sql_file")
    echo "Applying migration: $filename"
    docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -f /migrations/"$filename"
done

echo "Database initialized successfully!"

echo "
Connection information:
----------------------
Host: localhost
Port: $DB_PORT
Database: $DB_NAME
Username: $DB_USER
Password: $DB_PASSWORD

Connection URL:
postgresql://$DB_USER:$DB_PASSWORD@localhost:$DB_PORT/$DB_NAME

To connect using psql:
psql postgresql://$DB_USER:$DB_PASSWORD@localhost:$DB_PORT/$DB_NAME
"

# Create config file for the ACME server
cat > "$SCRIPT_DIR/../config/database.json" << EOF
{
  "database": {
    "host": "localhost",
    "port": $DB_PORT,
    "user": "$DB_USER",
    "password": "$DB_PASSWORD",
    "dbname": "$DB_NAME",
    "sslmode": "disable"
  }
}
EOF

echo "Database configuration file created at $SCRIPT_DIR/../config/database.json"