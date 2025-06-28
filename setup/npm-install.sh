#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Running containerized npm install for frontend..."
docker run --rm -v "$SCRIPT_DIR/../frontend:/app" -w /app node:20 npm install

echo "npm install complete." 