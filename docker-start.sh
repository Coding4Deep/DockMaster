#!/bin/bash

echo "🐳 Starting DockMaster with Docker Compose..."
echo "=============================================="

# Load environment variables from .env file
if [ -f ".env" ]; then
    echo "📋 Loading configuration from .env file..."
    export $(grep -v '^#' .env | xargs)
else
    echo "⚠️  No .env file found, using default values"
    export BACKEND_PORT=8081
    export FRONTEND_PORT=3000
    export JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
    export ADMIN_USERNAME=admin
    export ADMIN_PASSWORD=admin123
    export LOG_LEVEL=info
fi

# Display configuration
echo ""
echo "🔧 Configuration:"
echo "   Backend Port: ${BACKEND_PORT:-8081}"
echo "   Frontend Port: ${FRONTEND_PORT:-3000}"
echo "   Admin Username: ${ADMIN_USERNAME:-admin}"
echo "   Admin Password: ${ADMIN_PASSWORD:-admin123}"
echo ""

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
    echo "❌ Docker Compose is not available. Please install Docker Compose."
    exit 1
fi

# Stop any existing containers
echo "🧹 Stopping existing containers..."
docker-compose down 2>/dev/null || docker compose down 2>/dev/null || true

# Build and start the services
echo "🔨 Building and starting services..."
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose up --build -d
else
    docker compose up --build -d
fi

# Wait for services to be healthy
echo "⏳ Waiting for services to start..."
sleep 10

# Check backend health
echo "🔍 Checking backend health..."
for i in {1..30}; do
    if curl -s http://localhost:${BACKEND_PORT:-8081}/health >/dev/null; then
        echo "✅ Backend is healthy"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ Backend health check failed"
        echo "📋 Backend logs:"
        docker logs dockmaster-backend
        exit 1
    fi
    sleep 2
done

# Check frontend
echo "🔍 Checking frontend..."
for i in {1..30}; do
    if curl -s http://localhost:${FRONTEND_PORT:-3000} >/dev/null; then
        echo "✅ Frontend is accessible"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ Frontend check failed"
        echo "📋 Frontend logs:"
        docker logs dockmaster-frontend
        exit 1
    fi
    sleep 2
done

echo ""
echo "🎉 DockMaster is running successfully!"
echo ""
echo "📊 Backend API: http://localhost:${BACKEND_PORT:-8081}"
echo "🌐 Frontend UI: http://localhost:${FRONTEND_PORT:-3000}"
echo ""
echo "🔐 Default credentials:"
echo "   Username: ${ADMIN_USERNAME:-admin}"
echo "   Password: ${ADMIN_PASSWORD:-admin123}"
echo ""
echo "⚠️  IMPORTANT: Change the default password after first login!"
echo ""
echo "✨ Features available:"
echo "   • Image search and pull from Docker Hub"
echo "   • Run containers from UI"
echo "   • Real-time system metrics"
echo "   • Change password functionality"
echo "   • SQLite database persistence"
echo "   • Configurable ports via .env file"
echo ""
echo "🛑 To stop the application:"
echo "   ./docker-stop.sh"
echo "   or: docker-compose down"
echo ""
echo "📝 To view logs:"
echo "   docker logs dockmaster-backend"
echo "   docker logs dockmaster-frontend"
echo ""
echo "🔧 To change configuration, edit .env file and restart"
