#!/bin/bash
set -e

# Wait for DB and run migrations
until sentry upgrade --noinput; do
  echo "Waiting for Sentry DB to be ready..."
  sleep 3
done

# Create superuser if none exists
if ! sentry shell -c "from sentry.models.user import User; exit(0) if User.objects.filter(is_superuser=True).exists() else exit(1)"; then
  sentry createuser --superuser --email "$SENTRY_SU_EMAIL" --password "$SENTRY_SU_PASSWORD" --no-input
fi

# Start Sentry as usual
exec sentry "$@" 