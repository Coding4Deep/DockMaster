#!/bin/bash

echo "🛑 Stopping DockMaster..."
echo "========================"

# Stop and remove containers
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose down
else
    docker compose down
fi

echo "✅ DockMaster stopped successfully!"
echo ""
echo "💡 To start again, run: ./docker-start.sh"
