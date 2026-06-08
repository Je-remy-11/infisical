#!/bin/bash
# Development entrypoint script for Infisical
# This script handles dependency installation and starts services in dev mode

set -e

# Function to log messages
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

# Check if we're in the right directory
if [ ! -d "backend" ] || [ ! -d "frontend" ]; then
    log "Error: This script must be run from the Infisical project root directory"
    exit 1
fi

# Function to install backend dependencies
install_backend_deps() {
    log "Installing backend dependencies..."
    cd backend
    if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
        npm install
    else
        log "Backend dependencies are up to date"
    fi
    cd ..
}

# Function to install frontend dependencies
install_frontend_deps() {
    log "Installing frontend dependencies..."
    cd frontend
    if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
        npm install
    else
        log "Frontend dependencies are up to date"
    fi
    cd ..
}

# Function to start backend in dev mode
start_backend_dev() {
    log "Starting backend in development mode..."
    cd backend
    npm run dev &
    BACKEND_PID=$!
    cd ..
    log "Backend started with PID $BACKEND_PID"
}

# Function to start frontend in dev mode
start_frontend_dev() {
    log "Starting frontend in development mode..."
    cd frontend
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    log "Frontend started with PID $FRONTEND_PID"
}

# Function to wait for services
wait_for_services() {
    log "Waiting for services to start..."
    wait
}

# Main logic
case "${1:-all}" in
    "install")
        log "Installing all dependencies..."
        install_backend_deps
        install_frontend_deps
        log "Dependencies installed successfully"
        ;;
    "install-backend")
        install_backend_deps
        ;;
    "install-frontend")
        install_frontend_deps
        ;;
    "backend")
        install_backend_deps
        start_backend_dev
        wait_for_services
        ;;
    "frontend")
        install_frontend_deps
        start_frontend_dev
        wait_for_services
        ;;
    "all")
        log "Starting full development environment..."
        install_backend_deps
        install_frontend_deps
        start_backend_dev
        start_frontend_dev
        wait_for_services
        ;;
    *)
        echo "Usage: $0 [install|install-backend|install-frontend|backend|frontend|all]"
        echo ""
        echo "Commands:"
        echo "  install          - Install all dependencies"
        echo "  install-backend  - Install only backend dependencies"
        echo "  install-frontend - Install only frontend dependencies"
        echo "  backend          - Start backend in dev mode"
        echo "  frontend         - Start frontend in dev mode"
        echo "  all              - Start both backend and frontend (default)"
        exit 1
        ;;
esac
