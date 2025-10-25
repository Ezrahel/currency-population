#!/bin/sh

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
for i in $(seq 1 30); do
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" >/dev/null 2>&1; then
        break
    fi
    echo "Waiting for MySQL to be ready... ($i/30)"
    sleep 1
done

# Initialize database
echo "Initializing database..."
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" < /app/init.sql

# Start the application
echo "Starting application..."
exec ./main