#!/bin/bash
# Usage: bash setup/env_setup.sh <queue_mode> <log_mode>
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."

ENV_TEMPLATE="$PROJECT_ROOT/api/env/.env.template"
ENV_TARGET="$PROJECT_ROOT/api/env/.env.worker"
MAIN_ENV_FILE="$PROJECT_ROOT/api/.env"
QUEUE_MODE="$1"
LOG_MODE="$2"

cp "$ENV_TEMPLATE" "$ENV_TARGET"

# Update worker environment file
if [ "$QUEUE_MODE" = "sqs" ]; then
    sed -i '' 's/^QUEUE_CONNECTION=.*/QUEUE_CONNECTION=sqs/' "$ENV_TARGET"
else
    sed -i '' 's/^QUEUE_CONNECTION=.*/QUEUE_CONNECTION=sync/' "$ENV_TARGET"
fi

# Update main api/.env file
if [ "$QUEUE_MODE" = "sqs" ]; then
    sed -i '' 's/^QUEUE_CONNECTION=.*/QUEUE_CONNECTION=sqs/' "$MAIN_ENV_FILE"
else
    sed -i '' 's/^QUEUE_CONNECTION=.*/QUEUE_CONNECTION=sync/' "$MAIN_ENV_FILE"
fi

if [ "$LOG_MODE" = "sentry" ]; then
    sed -i '' 's/^LOG_CHANNEL=.*/LOG_CHANNEL=sentry/' "$ENV_TARGET"
    grep -q '^SENTRY_DSN=' "$ENV_TARGET" || echo 'SENTRY_DSN=your_sentry_dsn_here' >> "$ENV_TARGET"
    grep -q '^SENTRY_ENVIRONMENT=' "$ENV_TARGET" || echo 'SENTRY_ENVIRONMENT=development' >> "$ENV_TARGET"
    grep -q '^SENTRY_TRACES_SAMPLE_RATE=' "$ENV_TARGET" || echo 'SENTRY_TRACES_SAMPLE_RATE=1.0' >> "$ENV_TARGET"
else
    sed -i '' 's/^LOG_CHANNEL=.*/LOG_CHANNEL=local/' "$ENV_TARGET"
    # Optionally remove Sentry vars if you want
fi

sed -i '' 's/^CACHE_STORE=.*/CACHE_STORE=redis/' "$ENV_TARGET"
