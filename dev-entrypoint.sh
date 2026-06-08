#!/bin/sh
set -eu

if command -v update-ca-certificates >/dev/null 2>&1; then
  update-ca-certificates >/dev/null 2>&1 || true
fi

if [ -n "${SOFTHSM_TOKEN_DIR:-}" ] && command -v softhsm2-util >/dev/null 2>&1; then
  mkdir -p "$SOFTHSM_TOKEN_DIR"
  if [ ! -f "$SOFTHSM_TOKEN_DIR/auth-app.db" ]; then
    echo "Initializing SoftHSM token..."
    softhsm2-util --init-token --slot 0 --label "auth-app" --pin 1234 --so-pin 0000
  fi
fi

APP_DIR="${APP_DIR:-/app}"
START_COMMAND="${APP_START_CMD:-npm run dev}"
LOCKFILE_PATH="${LOCKFILE_PATH:-package-lock.json}"
STATE_FILE="node_modules/.package-lock.hash"

if [ -z "${INSTALL_COMMAND:-}" ]; then
  if [ -f "$APP_DIR/$LOCKFILE_PATH" ]; then
    INSTALL_COMMAND="npm ci"
  else
    INSTALL_COMMAND="npm install"
  fi
fi

cd "$APP_DIR"

SHOULD_INSTALL=0
if [ ! -d node_modules ]; then
  SHOULD_INSTALL=1
elif [ -f "$LOCKFILE_PATH" ]; then
  CURRENT_HASH="$(sha256sum "$LOCKFILE_PATH" | awk '{print $1}')"
  SAVED_HASH=""
  if [ -f "$STATE_FILE" ]; then
    SAVED_HASH="$(cat "$STATE_FILE")"
  fi
  if [ "$CURRENT_HASH" != "$SAVED_HASH" ]; then
    SHOULD_INSTALL=1
  fi
fi

if [ "$SHOULD_INSTALL" = "1" ]; then
  mkdir -p node_modules
  echo "Installing dependencies in $APP_DIR..."
  sh -lc "$INSTALL_COMMAND"
  if [ -f "$LOCKFILE_PATH" ]; then
    sha256sum "$LOCKFILE_PATH" | awk '{print $1}' > "$STATE_FILE"
  fi
fi

echo "Starting development mode in $APP_DIR..."
exec sh -lc "$START_COMMAND"
