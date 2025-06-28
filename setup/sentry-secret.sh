#!/bin/bash

# Generate a secure secret key for Sentry
echo "Generating Sentry secret key..."
SECRET_KEY=$(openssl rand -base64 32)

echo "Sentry Secret Key: $SECRET_KEY"
echo ""
echo "Update the following files with this key:"
echo "1. docker/sentry/envs/sentry.env (SENTRY_SECRET_KEY)"
echo "2. api/.env (SENTRY_SECRET_KEY)"
echo "3. api/env/.env.local (SENTRY_SECRET_KEY)"
echo "4. api/env/.env.staging (SENTRY_SECRET_KEY)"
echo "5. api/env/.env.production (SENTRY_SECRET_KEY)"
echo ""
echo "Example .env entry:"
echo "SENTRY_SECRET_KEY=$SECRET_KEY" 