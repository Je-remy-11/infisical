#!/bin/sh

echo "Starting frontend development environment..."

# Check if node_modules exists and has content
if [ ! -d "/app/node_modules" ] || [ -z "$(ls -A /app/node_modules 2>/dev/null)" ]; then
  echo "Installing frontend dependencies..."
  npm install --ignore-scripts
  echo "Frontend dependencies installed successfully."
else
  echo "Node modules already exist, skipping installation."
  echo "If you need to reinstall, run: rm -rf /app/node_modules && npm install"
fi

# Execute the main command
exec "$@"
