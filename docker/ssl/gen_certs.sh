#!/bin/bash

set -e

CERT_DIR="./docker/ssl/certs"
DAYS_VALID=825
FORCE=false

# Allow --force flag
if [[ "$1" == "--force" ]]; then
  FORCE=true
fi

mkdir -p "$CERT_DIR"

echo "🔍 Parsing VIRTUAL_HOST entries from docker-compose.yaml..."
VIRTUAL_HOSTS=$(grep -h "VIRTUAL_HOST=" ./docker-compose.yaml | cut -d'=' -f2 | tr -d '"' | tr ',' '\n' | sort -u)

if [[ -z "$VIRTUAL_HOSTS" ]]; then
  echo "❌ No VIRTUAL_HOST entries found in docker-compose.yaml."
  exit 1
fi

for DOMAIN in $VIRTUAL_HOSTS; do
  CERT_PATH="$CERT_DIR/$DOMAIN.crt"
  KEY_PATH="$CERT_DIR/$DOMAIN.key"

  if [[ -f "$CERT_PATH" && -f "$KEY_PATH" && "$FORCE" == "false" ]]; then
    echo "✅ Certificate already exists for $DOMAIN — skipping."
    continue
  fi

  echo "🔧 Generating self-signed certificate for $DOMAIN..."
  openssl req -x509 -nodes -days $DAYS_VALID \
    -newkey rsa:2048 \
    -keyout "$KEY_PATH" \
    -out "$CERT_PATH" \
    -subj "/CN=$DOMAIN"

  echo "🆕 Created cert and key for $DOMAIN"
done

echo "✅ All certificates generated in $CERT_DIR"
