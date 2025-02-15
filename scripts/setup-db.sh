#!/bin/bash

# PostgreSQL configuration
DB_NAME="acme_db"
DB_USER="acme_user"
DB_PASSWORD="acme_password"
DB_PORT=5432
CONTAINER_NAME="acme_postgres"

cd $(dirname $0)

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Docker is not running. Please start Docker first."
    exit 1
fi

# Stop and remove existing container if it exists
if docker ps -a | grep -q $CONTAINER_NAME; then
    echo "Stopping and removing existing container..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
fi

# Start PostgreSQL container
echo "Starting PostgreSQL container..."
docker run --name $CONTAINER_NAME \
    -e POSTGRES_DB=$DB_NAME \
    -e POSTGRES_USER=$DB_USER \
    -e POSTGRES_PASSWORD=$DB_PASSWORD \
    -v $(pwd)/../internal/database/migrations:/migrations \
    -v $(pwd)/../internal/database/schema.sql:/schema.sql \
    -p $DB_PORT:5432 \
    -d postgres:15

# Wait for PostgreSQL to start
echo "Waiting for PostgreSQL to start..."
sleep 5

# Copy schema file into container
echo "Copying schema file into container..."

# Execute schema file
echo "Initializing database schema..."
docker exec -i $CONTAINER_NAME psql \
    -U $DB_USER \
    -d $DB_NAME \
    -f /schema.sql \
    -f /migrations/*
# Print connection information
echo "
Database initialized successfully!

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
cat > ./config/database.json << EOF
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

echo "Database configuration file created at ./config/database.json" 