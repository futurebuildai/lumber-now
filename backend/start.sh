#!/bin/sh
set -e

# Run database migrations before starting the service
if [ -d /migrations ] && [ -n "$DATABASE_URL" ]; then
  echo "Running database migrations..."
  migrate -path /migrations -database "$DATABASE_URL" up || true
  echo "Migrations complete."
fi

# Start the requested binary (default: api)
exec "${1:-api}"
