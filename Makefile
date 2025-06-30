# DockMaster - Makefile for common operations

.PHONY: help build start stop restart logs clean dev test

# Default target
help:
	@echo "🐳 DockMaster - Available Commands:"
	@echo ""
	@echo "  make start     - Start DockMaster with Docker Compose"
	@echo "  make stop      - Stop all DockMaster services"
	@echo "  make restart   - Restart all services"
	@echo "  make build     - Build all Docker images"
	@echo "  make logs      - View logs from all services"
	@echo "  make clean     - Clean up containers and images"
	@echo "  make dev       - Start in development mode (processes)"
	@echo "  make test      - Run health checks"
	@echo "  make help      - Show this help message"
	@echo ""

# Start services with Docker Compose
start:
	@echo "🚀 Starting DockMaster..."
	docker-compose up -d --build
	@echo "✅ DockMaster is running!"
	@echo "📱 Frontend: http://localhost:3000"
	@echo "🔌 Backend: http://localhost:8080"

# Stop all services
stop:
	@echo "🛑 Stopping DockMaster..."
	docker-compose down
	@echo "✅ All services stopped"

# Restart services
restart: stop start

# Build images without starting
build:
	@echo "🔨 Building DockMaster images..."
	docker-compose build
	@echo "✅ Build complete"

# View logs
logs:
	@echo "📋 Viewing DockMaster logs..."
	docker-compose logs -f

# Clean up everything
clean:
	@echo "🧹 Cleaning up DockMaster..."
	docker-compose down -v --rmi all --remove-orphans
	docker system prune -f
	@echo "✅ Cleanup complete"

# Development mode (run as processes)
dev:
	@echo "🛠️  Starting DockMaster in development mode..."
	./start.sh

# Run health checks
test:
	@echo "🔍 Running health checks..."
	@curl -s http://localhost:8080/health > /dev/null && echo "✅ Backend is healthy" || echo "❌ Backend is not responding"
	@curl -s -I http://localhost:3000 > /dev/null && echo "✅ Frontend is healthy" || echo "❌ Frontend is not responding"

# Show status
status:
	@echo "📊 DockMaster Status:"
	@docker-compose ps
