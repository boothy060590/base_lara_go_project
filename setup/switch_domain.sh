#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."

# Prompt for environment
ENV=${1:-dev}
if [[ ! "$ENV" =~ ^(dev|staging|prod)$ ]]; then
  echo "Unknown environment: $ENV"; exit 1
fi

# Prompt for new APP_DOMAIN
read -p "Enter your new app domain (e.g. baselaragoproject.test): " APP_DOMAIN
if [ -z "$APP_DOMAIN" ]; then
  echo "No domain entered. Aborting."; exit 1
fi

echo "$APP_DOMAIN" > "$PROJECT_ROOT/.app_domain"

# Remove any old static .env files
rm -f "$PROJECT_ROOT/api/.env" "$PROJECT_ROOT/frontend/.env"

# Set environment-specific values
if [ "$ENV" = "dev" ]; then
  DB_HOST="db"
  DB_DRIVER="mysql"
  DB_USER="api_user"
  DB_PASSWORD="b4s3L4r4G0212!"
  DB_NAME="dev_base_lara_go"
  DB_PORT="3306"
  API_SECRET="yoursecretstring"
  TOKEN_HOUR_LIFESPAN="1"
  MAIL_PORT="1025"
  SQS_ENDPOINT="http://sqs.${APP_DOMAIN}:9324"
  APP_PORT="8080"
  REDIS_HOST="redis"
  REDIS_PASSWORD="null"
  REDIS_PORT="6379"
  CACHE_STORE="redis"
elif [ "$ENV" = "staging" ]; then
  DB_HOST="db"
  DB_DRIVER="mysql"
  DB_USER="api_user"
  DB_PASSWORD="b4s3L4r4G0212!"
  DB_NAME="staging_base_lara_go"
  DB_PORT="3306"
  API_SECRET="stagingsecret"
  TOKEN_HOUR_LIFESPAN="1"
  MAIL_PORT="1025"
  SQS_ENDPOINT="http://sqs.${APP_DOMAIN}:9324"
  APP_PORT="8080"
  REDIS_HOST="redis"
  REDIS_PASSWORD="null"
  REDIS_PORT="6379"
  CACHE_STORE="redis"
else # prod
  DB_HOST=""
  DB_DRIVER=""
  DB_USER=""
  DB_PASSWORD=""
  DB_NAME=""
  DB_PORT=""
  API_SECRET="prodsecret"
  TOKEN_HOUR_LIFESPAN="1"
  MAIL_PORT="1025"
  SQS_ENDPOINT="http://sqs.${APP_DOMAIN}:9324"
  APP_PORT="8080"
  REDIS_HOST="redis"
  REDIS_PASSWORD="null"
  REDIS_PORT="6379"
  CACHE_STORE="redis"
fi

# Generate api/.env from template
cp "$PROJECT_ROOT/api/env/.env.template" "$PROJECT_ROOT/api/.env"
sed -i '' \
  -e "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" \
  -e "s|{{DB_HOST}}|$DB_HOST|g" \
  -e "s|{{DB_DRIVER}}|$DB_DRIVER|g" \
  -e "s|{{DB_USER}}|$DB_USER|g" \
  -e "s|{{DB_PASSWORD}}|$DB_PASSWORD|g" \
  -e "s|{{DB_NAME}}|$DB_NAME|g" \
  -e "s|{{DB_PORT}}|$DB_PORT|g" \
  -e "s|{{API_SECRET}}|$API_SECRET|g" \
  -e "s|{{TOKEN_HOUR_LIFESPAN}}|$TOKEN_HOUR_LIFESPAN|g" \
  -e "s|{{MAIL_PORT}}|$MAIL_PORT|g" \
  -e "s|{{SQS_ENDPOINT}}|$SQS_ENDPOINT|g" \
  -e "s|{{APP_PORT}}|$APP_PORT|g" \
  -e "s|{{REDIS_HOST}}|$REDIS_HOST|g" \
  -e "s|{{REDIS_PASSWORD}}|$REDIS_PASSWORD|g" \
  -e "s|{{REDIS_PORT}}|$REDIS_PORT|g" \
  -e "s|{{CACHE_STORE}}|$CACHE_STORE|g" \
  "$PROJECT_ROOT/api/.env"

# Generate frontend/.env from template
cp "$PROJECT_ROOT/frontend/config/.env.template" "$PROJECT_ROOT/frontend/.env"
sed -i '' "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" "$PROJECT_ROOT/frontend/.env"

# Generate docker-compose.yaml from template
cp "$PROJECT_ROOT/docker-compose.yaml.template" "$PROJECT_ROOT/docker-compose.yaml"
sed -i '' "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" "$PROJECT_ROOT/docker-compose.yaml"

# Generate docker/sentry/docker-compose.yml from template
cp "$PROJECT_ROOT/docker/sentry/docker-compose.yml.template" "$PROJECT_ROOT/docker/sentry/docker-compose.yml"
sed -i '' "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" "$PROJECT_ROOT/docker/sentry/docker-compose.yml"

# Generate nginx config files from templates if they exist
for conf in "$PROJECT_ROOT/docker/api-nginx.conf" "$PROJECT_ROOT/docker/sentry/nginx.conf" "$PROJECT_ROOT/docker/nginx.conf"; do
  if [ -f "$conf.template" ]; then
    cp "$conf.template" "$conf"
    sed -i '' "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" "$conf"
  fi
done

echo -e "\n✅ Domain switched to $APP_DOMAIN."
echo -e "\n⚠️  Please run: make clean && make install_dev (or install_staging/install_prod) to apply changes." 