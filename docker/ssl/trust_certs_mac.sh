#!/bin/bash

CERT_DIR="./docker/ssl/certs"

echo "🔐 Trusting local certificates on macOS..."

# Check if directory exists
if [ ! -d "$CERT_DIR" ]; then
    echo "❌ Certificate directory '$CERT_DIR' does not exist."
    exit 1
fi

# Find all .crt files
crt_files=$(find "$CERT_DIR" -name "*.crt")

if [ -z "$crt_files" ]; then
    echo "❌ No .crt files found in $CERT_DIR"
    exit 1
fi

for crt in $crt_files; do
    cert_name=$(basename "$crt")
    echo "🔧 Adding $cert_name to System keychain..."

    # Add and trust the cert
    sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "$crt"

    if [ $? -eq 0 ]; then
        echo "✅ $cert_name trusted successfully."
    else
        echo "⚠️ Failed to trust $cert_name"
    fi
done

echo "✅ All available certificates have been processed."
