#!/bin/bash

# DockMaster - Single Command Startup Script
# This script allows running DockMaster as individual processes for development

set -e

echo "🐳 Starting DockMaster..."

# Function to check if port is available
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "❌ Port $1 is already in use"
        exit 1
    fi
}

# Function to start backend
start_backend() {
    echo "🚀 Starting backend service..."
    cd backend/docker-service
    
    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        echo "📦 Initializing Go module..."
        go mod init docker-service
        go mod tidy
    fi
    
    # Build and run
    go build -o dockmaster-api main.go
    ./dockmaster-api &
    BACKEND_PID=$!
    echo "✅ Backend started (PID: $BACKEND_PID)"
    cd ../..
}

# Function to start frontend
start_frontend() {
    echo "🎨 Starting frontend service..."
    cd frontend
    
    # Check if node_modules exists
    if [ ! -d "node_modules" ]; then
        echo "📦 Installing dependencies..."
        npm install
    fi
    
    # Build and serve
    npm run build
    npx serve -s build -l 3000 &
    FRONTEND_PID=$!
    echo "✅ Frontend started (PID: $FRONTEND_PID)"
    cd ..
}

# Function to cleanup on exit
cleanup() {
    echo "🛑 Shutting down DockMaster..."
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    echo "✅ Cleanup complete"
    exit 0
}

# Set trap for cleanup
trap cleanup SIGINT SIGTERM

# Main execution
main() {
    echo "🔍 Checking prerequisites..."
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        echo "❌ Docker is not running. Please start Docker first."
        exit 1
    fi
    
    # Check if required tools are installed
    command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed"; exit 1; }
    command -v node >/dev/null 2>&1 || { echo "❌ Node.js is not installed"; exit 1; }
    command -v npm >/dev/null 2>&1 || { echo "❌ npm is not installed"; exit 1; }
    
    # Check ports
    check_port 3000
    check_port 8080
    
    echo "✅ Prerequisites check passed"
    
    # Start services
    start_backend
    sleep 2  # Give backend time to start
    start_frontend
    
    echo ""
    echo "🎉 DockMaster is now running!"
    echo "📱 Frontend: http://localhost:3000"
    echo "🔌 Backend API: http://localhost:8080"
    echo "❤️  Health Check: http://localhost:8080/health"
    echo ""
    echo "Press Ctrl+C to stop all services"
    
    # Wait for processes
    wait
}

# Run main function
main "$@"
