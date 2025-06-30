# 🐳 DockMaster - Docker Management Dashboard

A modern, responsive web-based Docker management interface built with React and Go. DockMaster provides an intuitive dashboard to monitor and manage your Docker containers, images, volumes, and networks.

![DockMaster Dashboard](https://img.shields.io/badge/Status-Live-brightgreen) ![Docker](https://img.shields.io/badge/Docker-Ready-blue) ![React](https://img.shields.io/badge/React-18-61dafb) ![Go](https://img.shields.io/badge/Go-1.23-00add8)

## ✨ Features

### 🚀 Container Management
- **Real-time Container Monitoring** - View all containers with live status updates
- **Container Operations** - Start, stop, restart, and delete containers with one click
- **Resource Monitoring** - CPU, memory, network, and disk I/O statistics
- **Log Viewing** - Stream and view container logs in real-time
- **Container Details** - Comprehensive information about each container

### 🖼️ Image Management
- **Image Repository** - Browse and manage Docker images
- **Image Operations** - Pull, delete, and inspect images
- **Size Optimization** - View image sizes and optimize storage
- **Tag Management** - Handle multiple image tags efficiently

### 💾 Volume Management
- **Volume Overview** - Monitor all Docker volumes and their usage
- **Storage Analytics** - Track volume sizes and reference counts
- **Volume Operations** - Create, delete, and manage volumes
- **Mount Point Information** - View detailed mount point data

### 🌐 Network Management
- **Network Topology** - Visualize Docker networks and connections
- **Network Operations** - Create, delete, and configure networks
- **Container Connectivity** - View which containers are connected to each network
- **Network Drivers** - Support for bridge, host, and custom networks

### 📊 Dashboard & Analytics
- **System Overview** - Real-time system resource usage
- **Performance Metrics** - Historical data and trends
- **Health Monitoring** - Service health checks and alerts
- **Responsive Design** - Works perfectly on desktop, tablet, and mobile

## 🏗️ Architecture

DockMaster follows a modern microservices architecture:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Docker API     │    │   Docker        │
│   (React)       │◄──►│   Service (Go)   │◄──►│   Engine        │
│   Port: 3000    │    │   Port: 8080     │    │   Socket        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Frontend (React)
- **Framework**: React 18 with modern hooks
- **Styling**: Tailwind CSS for responsive design
- **Icons**: Heroicons for consistent UI
- **State Management**: React hooks and context
- **HTTP Client**: Axios for API communication
- **Build Tool**: Create React App with optimizations

### Backend (Go)
- **Framework**: Gorilla Mux for routing
- **Docker Integration**: Official Docker Go SDK
- **CORS**: Configured for cross-origin requests
- **Logging**: Structured logging with Logrus
- **Health Checks**: Built-in health monitoring
- **API Design**: RESTful endpoints with JSON responses

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose installed
- Git for cloning the repository
- 4GB+ RAM recommended
- Modern web browser

### 1. Clone and Run
```bash
git clone <repository-url>
cd dockmaster
docker-compose up -d --build
```

### 2. Access the Application
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### 3. Verify Installation
```bash
# Check container status
docker-compose ps

# View logs
docker-compose logs -f

# Test API
curl http://localhost:8080/health
```

## 🛠️ Development Setup

### Running as Individual Services

#### Backend Development
```bash
cd backend/docker-service
go mod tidy
go run main.go
# API available at http://localhost:8080
```

#### Frontend Development
```bash
cd frontend
npm install
npm start
# Development server at http://localhost:3000
```

### Environment Variables
```bash
# Backend
LOG_LEVEL=info
PORT=8080

# Frontend
REACT_APP_API_URL=http://localhost:8080
```

## 📁 Project Structure

```
dockmaster/
├── README.md                 # This comprehensive guide
├── docker-compose.yml        # Multi-service orchestration
├── frontend/                 # React application
│   ├── public/              # Static assets
│   ├── src/
│   │   ├── components/      # Reusable UI components
│   │   ├── pages/          # Main application pages
│   │   ├── services/       # API communication layer
│   │   └── styles/         # CSS and styling
│   ├── package.json        # Node.js dependencies
│   ├── Dockerfile          # Frontend container config
│   └── nginx.conf          # Production web server config
└── backend/
    └── docker-service/      # Unified Go API service
        ├── main.go         # Main application entry point
        ├── go.mod          # Go module dependencies
        └── Dockerfile      # Backend container config
```

## 🔧 Configuration

### Docker Compose Services
- **frontend**: React app served by Nginx (Port 3000)
- **docker-service**: Go API server (Port 8080)
- **dockmaster-network**: Internal bridge network

### Security Features
- CORS protection configured
- Security headers implemented
- Docker socket mounted read-only
- Container isolation enforced

### Performance Optimizations
- Multi-stage Docker builds
- Gzip compression enabled
- Static asset caching
- Health checks implemented
- Resource limits configured

## 🔌 API Endpoints

### Container Management
```
GET    /containers              # List all containers
POST   /containers/{id}/start   # Start container
POST   /containers/{id}/stop    # Stop container
POST   /containers/{id}/restart # Restart container
DELETE /containers/{id}         # Delete container
GET    /containers/{id}/stats   # Get container stats
GET    /containers/{id}/logs    # Get container logs
```

### Image Management
```
GET    /images                  # List all images
DELETE /images/{id}             # Delete image
```

### Volume Management
```
GET    /volumes                 # List all volumes
DELETE /volumes/{name}          # Delete volume
```

### Network Management
```
GET    /networks                # List all networks
DELETE /networks/{id}           # Delete network
```

### System
```
GET    /health                  # Health check endpoint
```

## 🚀 Production Deployment

### Docker Compose (Recommended)
```bash
# Production deployment
docker-compose -f docker-compose.yml up -d

# Scale services if needed
docker-compose up -d --scale docker-service=2
```

### Manual Deployment
```bash
# Build images
docker build -t dockmaster-frontend ./frontend
docker build -t dockmaster-backend ./backend/docker-service

# Run containers
docker run -d -p 3000:3000 dockmaster-frontend
docker run -d -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock:ro dockmaster-backend
```

### Environment-Specific Configurations
- **Development**: Hot reloading, debug logging
- **Staging**: Production builds, monitoring enabled
- **Production**: Optimized builds, security hardened

## 🔍 Monitoring & Troubleshooting

### Health Checks
```bash
# Check service health
curl http://localhost:8080/health

# Container health status
docker-compose ps
```

### Logs
```bash
# View all logs
docker-compose logs -f

# Service-specific logs
docker-compose logs -f frontend
docker-compose logs -f docker-service
```

### Common Issues
1. **Port conflicts**: Ensure ports 3000 and 8080 are available
2. **Docker socket**: Verify Docker daemon is running
3. **Memory issues**: Ensure sufficient RAM (4GB+ recommended)
4. **Network connectivity**: Check firewall and network settings

## 🤝 Contributing

### Development Workflow
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

### Code Standards
- **Go**: Follow Go conventions, use gofmt
- **React**: Use ESLint and Prettier
- **Docker**: Multi-stage builds, minimal base images
- **Documentation**: Update README for new features

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Docker team for the excellent API
- React community for the amazing ecosystem
- Go community for the robust standard library
- All contributors and users of DockMaster

## 📞 Support

- **Issues**: GitHub Issues for bug reports
- **Discussions**: GitHub Discussions for questions
- **Documentation**: This README and inline code comments

---

**DockMaster** - Making Docker management simple, powerful, and accessible. 🐳✨
