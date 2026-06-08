#!/bin/sh
set -e

# Detect the service context from the WORKDIR or a marker file
if [ -f /app/package.json ] && [ -f /app/nodemon.json ]; then
  SERVICE="backend"
elif [ -f /app/package.json ] && [ -f /app/vite.config.ts ]; then
  SERVICE="frontend"
elif [ -f /app/go.mod ]; then
  SERVICE="backend-go"
else
  SERVICE="unknown"
fi
echo "[dev-entrypoint] Detected service: ${SERVICE}"

# ---- Dependency installation ----
install_deps() {
  if [ -f package.json ]; then
    if [ ! -d node_modules ] || [ ! -f node_modules/.package-lock.json ]; then
      echo "[dev-entrypoint] Installing npm dependencies..."
      npm install
      echo "[dev-entrypoint] npm install complete"
    elif [ package.json -nt node_modules/.package-lock.json ]; then
      echo "[dev-entrypoint] package.json changed, reinstalling dependencies..."
      npm install
      echo "[dev-entrypoint] npm install complete"
    else
      echo "[dev-entrypoint] Dependencies are up to date"
    fi
  fi
}

# ---- Service-specific setup ----
setup_backend() {
  echo "[dev-entrypoint] Running backend setup..."

  update-ca-certificates

  # Initialize SoftHSM token if it doesn't exist
  if [ ! -f /etc/softhsm2/tokens/auth-app.db ]; then
    echo "[dev-entrypoint] Initializing SoftHSM token..."
    mkdir -p /etc/softhsm2/tokens
    softhsm2-util --init-token --slot 0 --label "auth-app" --pin 1234 --so-pin 0000
    echo "[dev-entrypoint] SoftHSM token initialized"
  else
    echo "[dev-entrypoint] SoftHSM token already exists, skipping initialization"
  fi
}

setup_frontend() {
  echo "[dev-entrypoint] Running frontend setup..."
}

setup_backend_go() {
  echo "[dev-entrypoint] Running backend-go setup..."
}

# Run setup based on service
case "${SERVICE}" in
  backend)
    install_deps
    setup_backend
    ;;
  frontend)
    install_deps
    setup_frontend
    ;;
  backend-go)
    setup_backend_go
    ;;
  *)
    echo "[dev-entrypoint] Unknown service, attempting dependency installation..."
    if [ -f package.json ]; then
      install_deps
    fi
    ;;
esac

echo "[dev-entrypoint] Starting: $@"
exec "$@"