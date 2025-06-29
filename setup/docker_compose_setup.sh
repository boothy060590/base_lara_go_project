#!/bin/bash
# Usage: bash setup/docker_compose_setup.sh <queue_mode> <log_mode>
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."

TEMPLATE="$PROJECT_ROOT/docker-compose.yaml.template"
TARGET="$PROJECT_ROOT/docker-compose.yaml"
QUEUE_MODE="$1"
LOG_MODE="$2"

cp "$TEMPLATE" "$TARGET"

if [ "$QUEUE_MODE" = "sync" ]; then
    # Remove worker service from docker-compose.yaml
    sed -i '' '/worker:/,/^$/d' "$TARGET"
fi

if [ "$LOG_MODE" = "local" ]; then
    # Remove sentry service from docker-compose.yaml
    sed -i '' '/sentry:/,/^$/d' "$TARGET"
    # Optionally remove sentry include from docker/sentry/docker-compose.yml
    if [ -f "$PROJECT_ROOT/docker/sentry/docker-compose.yml" ]; then
        rm "$PROJECT_ROOT/docker/sentry/docker-compose.yml"
    fi
fi
