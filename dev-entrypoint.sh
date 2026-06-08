#!/bin/sh
set -e

# Initialize SoftHSM token if it doesn't exist and softhsm2-util is available (for backend)
if command -v softhsm2-util >/dev/null 2>&1; then
  if [ ! -f /etc/softhsm2/tokens/auth-app.db ]; then
    echo "Initializing SoftHSM token..."
    mkdir -p /etc/softhsm2/tokens
    softhsm2-util --init-token --slot 0 --label "auth-app" --pin 1234 --so-pin 0000 || true
    echo "SoftHSM token initialized"
  else
    echo "SoftHSM token already exists, skipping initialization"
  fi
fi

echo "Installing dependencies..."
npm install

echo "Starting dev mode..."
exec "$@"
