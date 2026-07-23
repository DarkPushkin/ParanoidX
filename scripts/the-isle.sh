#!/bin/bash
# Launch The Isle desktop client with Boss keys
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="$HOME/.local/share/the-isle"
BINARY="isle_app"
SERVER="${1:-http://127.0.0.1:8080}"
PUBKEY="f03bb8d0cfa3405827883060ed85f862f07212b2a37191ae9f218babf0994bf7"

if [ -f "$INSTALL_DIR/$BINARY" ]; then
  exec "$INSTALL_DIR/$BINARY" --server "$SERVER" --pubkey "$PUBKEY"
elif [ -f "$SCRIPT_DIR/$BINARY" ]; then
  exec "$SCRIPT_DIR/$BINARY" --server "$SERVER" --pubkey "$PUBKEY"
else
  echo "Error: $BINARY not found."
  echo "Looked in: $INSTALL_DIR"
  echo "Looked in: $SCRIPT_DIR"
  echo ""
  echo "Install first: ./install.sh"
  exit 1
fi
