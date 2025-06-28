#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Stopping all containers..."
docker stop $(docker ps -a -q) 2>/dev/null || true

echo "Removing all containers..."
docker rm $(docker ps -a -q) 2>/dev/null || true

echo "Removing all images..."
docker rmi $(docker images -q) -f 2>/dev/null || true

echo "Removing all volumes..."
docker volume rm $(docker volume ls -q) 2>/dev/null || true

echo "Removing frontend/node_modules..."
rm -rf "$SCRIPT_DIR/../frontend/node_modules"

echo "Removing api/.env..."
rm -f "$SCRIPT_DIR/../api/.env"

echo "Removing frontend/.env..."
rm -f "$SCRIPT_DIR/../frontend/.env"

echo "Removing frontend/certs..."
rm -rf "$SCRIPT_DIR/../frontend/certs"

echo "Removing generated docker-compose and nginx config files..."
rm -f "$SCRIPT_DIR/../docker-compose.yaml" \
      "$SCRIPT_DIR/../docker/sentry/docker-compose.yml" \
      "$SCRIPT_DIR/../docker/api-nginx.conf" \
      "$SCRIPT_DIR/../docker/sentry/nginx.conf" \
      "$SCRIPT_DIR/../docker/nginx.conf"

echo "Removing generated SSL certs in docker/ssl/certs..."
rm -rf "$SCRIPT_DIR/../docker/ssl/certs"

echo "Docker and workspace clean slate complete." 