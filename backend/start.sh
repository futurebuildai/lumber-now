#!/bin/sh
set -e

# Run database migrations before starting the API
if [ -d /migrations ]; then
  echo "Running database migrations..."
  migrate -path /migrations -database "$DATABASE_URL" up
  echo "Migrations complete."
fi

exec api
