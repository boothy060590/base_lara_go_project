#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Generating local SSL certificates..."
bash "$SCRIPT_DIR/../docker/ssl/gen_certs.sh"

echo "Trusting local SSL certificates (macOS only, may require sudo password)..."
echo "If you do not want to trust the certs system-wide, you can skip this step."
bash "$SCRIPT_DIR/../docker/ssl/trust_certs_mac.sh" || true

echo "Copying generated certs to frontend/certs..."
rm -rf "$SCRIPT_DIR/../frontend/certs"
cp -R "$SCRIPT_DIR/../docker/ssl/certs" "$SCRIPT_DIR/../frontend/certs"

echo "SSL certificate setup complete. Certs are available in frontend/certs." 