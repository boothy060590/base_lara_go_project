#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."

# Docker Compose version check
MIN_DOCKER_COMPOSE_VERSION="2.20.0"
DOCKER_COMPOSE_VERSION=$(docker compose version --short 2>/dev/null || docker-compose version --short 2>/dev/null || echo "0.0.0")
if [ "$(printf '%s\n' "$MIN_DOCKER_COMPOSE_VERSION" "$DOCKER_COMPOSE_VERSION" | sort -V | head -n1)" != "$MIN_DOCKER_COMPOSE_VERSION" ]; then
  echo "[WARNING] Your Docker Compose version is $DOCKER_COMPOSE_VERSION. Minimum required is $MIN_DOCKER_COMPOSE_VERSION."
  read -p "Would you like to update Docker Compose now? [Y/n] " yn
  yn=${yn:-Y}
  if [[ $yn =~ ^[Yy]$ ]]; then
    echo "Updating Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/download/v$MIN_DOCKER_COMPOSE_VERSION/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    echo "Docker Compose updated."
  else
    echo "Please update Docker Compose manually soon."
  fi
fi

# Prompt for environment
ENV=${1:-dev}
if [[ ! "$ENV" =~ ^(dev|staging|prod)$ ]]; then
  echo "Unknown environment: $ENV"; exit 1
fi

# Prompt for APP_DOMAIN
APP_DOMAIN=""
if [ -f "$PROJECT_ROOT/.app_domain" ]; then
  APP_DOMAIN=$(cat "$PROJECT_ROOT/.app_domain")
fi
if [ -z "$APP_DOMAIN" ]; then
  read -p "Enter your desired app domain (e.g. baselaragoproject.test): " APP_DOMAIN
  echo "$APP_DOMAIN" > "$PROJECT_ROOT/.app_domain"
fi

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

# Generate nginx config files from templates if they exist
for conf in "$PROJECT_ROOT/docker/api-nginx.conf" "$PROJECT_ROOT/docker/nginx.conf"; do
  if [ -f "$conf.template" ]; then
    cp "$conf.template" "$conf"
    sed -i '' "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" "$conf"
  fi
done

echo "[1/7] Ensuring correct Node.js version with nvm..."
bash "$SCRIPT_DIR/node.sh"

echo "[2/7] Generating and trusting SSL certs..."
bash "$SCRIPT_DIR/certs.sh"
echo -e "\n⚠️  If this is your first time using this domain, you may need to visit each local HTTPS service in your browser and accept the SSL warning for local development."
echo -e "   • https://app.$APP_DOMAIN"
echo -e "   • https://api.$APP_DOMAIN"
echo -e "   • https://mail.$APP_DOMAIN"
echo -e "   • https://sqs.$APP_DOMAIN"

echo "[3/7] Running containerized npm install..."
bash "$SCRIPT_DIR/npm-install.sh"

echo "[4/7] Installing Go dependencies..."
cd "$PROJECT_ROOT/api"
if command -v go &> /dev/null; then
  echo "Checking Go dependencies..."
  if [ -f "go.mod" ]; then
    echo "go.mod found, checking for missing dependencies..."
    if ! go list -m gorm.io/driver/mysql &>/dev/null; then
      echo "Installing gorm.io/driver/mysql..."
      go get gorm.io/driver/mysql
    fi
    if ! go list -m gorm.io/driver/postgres &>/dev/null; then
      echo "Installing gorm.io/driver/postgres..."
      go get gorm.io/driver/postgres
    fi
    if ! go list -m gorm.io/driver/sqlite &>/dev/null; then
      echo "Installing gorm.io/driver/sqlite..."
      go get gorm.io/driver/sqlite
    fi
    if ! go list -m github.com/redis/go-redis/v9 &>/dev/null; then
      echo "Installing github.com/redis/go-redis/v9..."
      go get github.com/redis/go-redis/v9
    fi
    if ! go list -m github.com/getsentry/sentry-go &>/dev/null; then
      echo "Installing github.com/getsentry/sentry-go..."
      go get github.com/getsentry/sentry-go
    fi
    echo "Running go mod tidy..."
    go mod tidy
    echo "Go dependencies are up to date."
  else
    echo "No go.mod found, initializing Go module..."
    go mod init base_lara_go_project
    go get gorm.io/driver/mysql
    go get gorm.io/driver/postgres
    go get gorm.io/driver/sqlite
    go get github.com/redis/go-redis/v9
    go get github.com/getsentry/sentry-go
    go mod tidy
    echo "Go module initialized and dependencies installed."
  fi
else
  echo "[WARNING] Go is not installed. Please install Go and run 'go mod tidy' in the api directory."
fi

# Modular install scripts for queue and logging setup
echo "[5/7] Configuring queue and logging options..."
QUEUE_MODE=$(bash "$SCRIPT_DIR/queue_mode_prompt.sh")
LOG_MODE=$(bash "$SCRIPT_DIR/logging_prompt.sh")

# Conditional Sentry setup
if [ "$LOG_MODE" = "sentry" ]; then
  echo "[6/7] Setting up Sentry logging..."
  echo "   • Generating and injecting Sentry secret key..."
  bash "$SCRIPT_DIR/sentry-secret.sh"
  
  echo "   • Running env-inject logic..."
  bash "$SCRIPT_DIR/env-inject.sh"
  
  # Ensure SENTRY_SECRET_KEY is present before starting the stack
  if ! grep -q '^SENTRY_SECRET_KEY=' "$PROJECT_ROOT/docker/sentry/envs/sentry.env"; then
    echo "[ERROR] SENTRY_SECRET_KEY is missing from docker/sentry/envs/sentry.env. Aborting startup." >&2
    exit 1
  fi
  
  # Generate Sentry docker-compose and nginx config
  echo "   • Setting up Sentry Docker configuration..."
  cp "$PROJECT_ROOT/docker/sentry/docker-compose.yml.template" "$PROJECT_ROOT/docker/sentry/docker-compose.yml"
  sed -i '' "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" "$PROJECT_ROOT/docker/sentry/docker-compose.yml"
  
  if [ -f "$PROJECT_ROOT/docker/sentry/nginx.conf.template" ]; then
    cp "$PROJECT_ROOT/docker/sentry/nginx.conf.template" "$PROJECT_ROOT/docker/sentry/nginx.conf"
    sed -i '' "s|{{APP_DOMAIN}}|$APP_DOMAIN|g" "$PROJECT_ROOT/docker/sentry/nginx.conf"
  fi
  
  echo "   • Sentry setup complete."
  echo -e "   • Sentry will be available at: https://sentry.$APP_DOMAIN"
else
  echo "[6/7] Skipping Sentry setup (using local logging)..."
fi

echo "[7/7] Configuring environment and Docker Compose..."
bash "$SCRIPT_DIR/env_setup.sh" "$QUEUE_MODE" "$LOG_MODE"
bash "$SCRIPT_DIR/docker_compose_setup.sh" "$QUEUE_MODE" "$LOG_MODE"

if [ "$QUEUE_MODE" = "sqs" ]; then
    echo "   • Setting up multi-worker infrastructure..."
    bash "$SCRIPT_DIR/generate-workers.sh"
fi

echo "[8/8] Starting Docker containers..."
cd "$PROJECT_ROOT"
docker-compose up -d

echo ""
echo "Environment setup complete."
echo "Check api/env/.env.worker for logging and Redis config."
echo "Check docker-compose.yaml for correct worker and logging setup."
echo -e "\n🎉 Setup complete! The stack is now running."
echo -e "\n📋 Available services:"
echo "   • Frontend: https://app.$APP_DOMAIN"
echo "   • API: https://api.$APP_DOMAIN"
if [ "$LOG_MODE" = "sentry" ]; then
  echo "   • Sentry: https://sentry.$APP_DOMAIN"
fi
echo "   • MailHog: https://mail.$APP_DOMAIN"
echo "   • SQS UI: https://sqs.$APP_DOMAIN"
echo -e "\n🔧 Useful commands:"
echo "   • View logs: docker-compose logs -f [service_name]"
echo "   • Stop stack: docker-compose down"
echo "   • Clean slate: make clean" 