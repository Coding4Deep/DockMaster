# DockMaster - Docker Management Platform (Microservices)

.PHONY: help build start stop restart logs clean dev test init-modules update-deps backup restore

# Default target
help:
	@echo "🐳 DockMaster - Available Commands:"
	@echo ""
	@echo "  make start        - Start DockMaster with Docker Compose"
	@echo "  make stop         - Stop all DockMaster services"
	@echo "  make restart      - Restart all services"
	@echo "  make build        - Build all Docker images"
	@echo "  make logs         - View logs from all services"
	@echo "  make clean        - Clean up containers and images"
	@echo "  make dev          - Start in development mode"
	@echo "  make test         - Run health checks"
	@echo "  make init-modules - Initialize Go modules for all services"
	@echo "  make update-deps  - Update dependencies"
	@echo "  make backup       - Backup application data"
	@echo "  make restore      - Show restore instructions"
	@echo "  make help         - Show this help message"
	@echo ""

# Start services with Docker Compose
start:
	@echo "🚀 Starting DockMaster (Microservices)..."
	docker-compose up -d --build
	@echo "✅ DockMaster is running!"
	@echo "📱 Frontend: http://localhost:3000"
	@echo "🔌 API Gateway: http://localhost:8080"
	@echo "🔐 Auth Service: http://localhost:8081"
	@echo "📦 Container Service: http://localhost:8082"
	@echo "🖼️  Image Service: http://localhost:8083"
	@echo "💾 Volume Service: http://localhost:8084"
	@echo "🌐 Network Service: http://localhost:8085"

# Stop all services
stop:
	@echo "🛑 Stopping DockMaster..."
	docker-compose down
	@echo "✅ All services stopped"

# Restart services
restart: stop start

# Build images without starting
build:
	@echo "🔨 Building DockMaster microservices..."
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

# Development mode
dev:
	@echo "🛠️  Starting DockMaster in development mode..."
	docker-compose up --build

# Initialize Go modules for all services
init-modules:
	@echo "📦 Initializing Go modules for all services..."
	cd backend/services/auth-service && go mod tidy
	cd backend/services/container-service && go mod tidy
	cd backend/services/image-service && go mod tidy
	cd backend/services/volume-service && go mod tidy
	cd backend/services/network-service && go mod tidy
	cd backend/services/api-gateway && go mod tidy
	@echo "✅ Go modules initialized"

# Update dependencies
update-deps:
	@echo "🔄 Updating dependencies for all services..."
	cd backend/services/auth-service && go get -u ./...
	cd backend/services/container-service && go get -u ./...
	cd backend/services/image-service && go get -u ./...
	cd backend/services/volume-service && go get -u ./...
	cd backend/services/network-service && go get -u ./...
	cd backend/services/api-gateway && go get -u ./...
	cd frontend && npm update
	@echo "✅ Dependencies updated"

# Run health checks
test:
	@echo "🔍 Running health checks..."
	@curl -s http://localhost:8080/health > /dev/null && echo "✅ API Gateway is healthy" || echo "❌ API Gateway is not responding"
	@curl -s http://localhost:8081/health > /dev/null && echo "✅ Auth Service is healthy" || echo "❌ Auth Service is not responding"
	@curl -s http://localhost:8082/health > /dev/null && echo "✅ Container Service is healthy" || echo "❌ Container Service is not responding"
	@curl -s http://localhost:8083/health > /dev/null && echo "✅ Image Service is healthy" || echo "❌ Image Service is not responding"
	@curl -s http://localhost:8084/health > /dev/null && echo "✅ Volume Service is healthy" || echo "❌ Volume Service is not responding"
	@curl -s http://localhost:8085/health > /dev/null && echo "✅ Network Service is healthy" || echo "❌ Network Service is not responding"
	@curl -s -I http://localhost:3000 > /dev/null && echo "✅ Frontend is healthy" || echo "❌ Frontend is not responding"

# Show status
status:
	@echo "📊 DockMaster Status:"
	@docker-compose ps

# Backup data
backup:
	@echo "💾 Creating backup of DockMaster data..."
	@mkdir -p backups
	@tar -czf backups/dockmaster-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz data/
	@echo "✅ Backup created in backups/ directory"

# Restore instructions
restore:
	@echo "📥 Available backups:"
	@ls -la backups/ 2>/dev/null || echo "No backups found"
	@echo ""
	@echo "To restore from backup:"
	@echo "  tar -xzf backups/[backup-file] -C ."
