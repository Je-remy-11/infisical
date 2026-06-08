# Infisical Development Environment Setup

This document explains how to use the improved development environment with hot-reloading capabilities.

## Quick Start

### Using Docker Compose (Recommended)

1. Copy the development environment example:
   ```bash
   cp .env.dev.example .env
   ```

2. Start the development environment:
   ```bash
   docker-compose -f docker-compose.dev.yml up --build
   ```

3. Access the application:
   - Web UI: http://localhost:8080
   - Backend API: http://localhost:4000
   - Frontend dev server: http://localhost:3000

### Using Local Development Script

If you prefer to run services locally:

```bash
# Make the script executable
chmod +x dev-entrypoint.sh

# Install dependencies and start both backend and frontend
./dev-entrypoint.sh all

# Or start only backend
./dev-entrypoint.sh backend

# Or start only frontend
./dev-entrypoint.sh frontend

# Just install dependencies
./dev-entrypoint.sh install
```

## Features

### Hot Reloading

- **Backend**: Uses Nodemon to watch for changes in `backend/src/` and automatically restarts
- **Frontend**: Uses Vite's built-in HMR (Hot Module Replacement) for instant updates
- **Volumes**: Source code is mounted with `:cached` flag for better performance

### Customizable Ports

All ports can be customized in your `.env` file. See `.env.dev.example` for all available options.

Key ports (default values):
- `NGINX_PORT=8080` - Main web interface
- `BACKEND_PORT=4000` - Backend API
- `FRONTEND_PORT=3000` - Frontend dev server
- `DB_PORT=5432` - Postgres database
- `REDIS_PORT=6379` - Redis cache

## Docker Compose Services

The `docker-compose.dev.yml` includes:
- **nginx**: Reverse proxy
- **db**: Postgres database
- **redis**: Redis cache
- **clickhouse**: Analytics database
- **backend**: Node.js API with hot-reload
- **frontend**: Vite dev server with HMR
- **pgadmin**: Database management UI
- **smtp-server**: MailHog for email testing
- And more optional services (LDAP, Keycloak, etc.)

## Development Workflow

1. Make changes to backend code in `backend/src/`
2. Changes are automatically picked up by Nodemon
3. Make changes to frontend code in `frontend/src/`
4. Vite HMR updates the browser instantly
5. Test your changes at http://localhost:8080

## Troubleshooting

### Frontend not reloading
- Ensure `CHOKIDAR_USEPOLLING=true` is set in your `.env`
- Check that Docker volume mounts are working correctly

### Backend not restarting
- Verify `nodemon.json` is properly mounted
- Check backend logs for errors

### Port conflicts
- Modify the port mappings in your `.env` file
- Ensure no other services are using the ports

## Dev Entrypoint Script Commands

The `dev-entrypoint.sh` script provides:
- `install` - Install all dependencies
- `install-backend` - Install only backend dependencies
- `install-frontend` - Install only frontend dependencies
- `backend` - Start backend in dev mode
- `frontend` - Start frontend in dev mode
- `all` - Start both (default)
