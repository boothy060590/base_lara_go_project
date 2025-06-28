#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

NODE_VERSION=20

# Check if nvm is installed
if ! command -v nvm >/dev/null 2>&1 && [ -z "$NVM_DIR" ]; then
  echo "[node.sh] nvm (Node Version Manager) is not installed."
  read -p "Would you like to install nvm now? [Y/n] " yn
  yn=${yn:-Y}
  if [[ $yn =~ ^[Yy]$ ]]; then
    echo "[node.sh] Installing nvm..."
    # Install nvm (official install script)
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
  else
    echo "[node.sh] nvm is required for this project. Aborting."
    exit 1
  fi
else
  # Load nvm if not already loaded
  export NVM_DIR="$HOME/.nvm"
  [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
fi

# Check if Node.js 20 is installed
if ! nvm ls $NODE_VERSION | grep -q "v$NODE_VERSION"; then
  echo "[node.sh] Node.js $NODE_VERSION not found. Installing..."
  nvm install $NODE_VERSION
fi

echo "[node.sh] Using Node.js $NODE_VERSION via nvm."
nvm use $NODE_VERSION

# Print node and npm version for confirmation
echo "[node.sh] Node version: $(node -v)"
echo "[node.sh] NPM version: $(npm -v)" 