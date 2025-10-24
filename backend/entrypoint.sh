#!/bin/sh
set -e

echo "Waiting for postgres..."
until nc -z db 5432; do
  echo "Postgres is unavailable - sleeping"
  sleep 1
done
echo "Postgres is up"

echo "Running migrations..."
/app/migrate -path /app/migrations -database "$DB_ADDR" -verbose up

echo "Starting server..."
exec "$@"