# 🚀 DockMaster Deployment Guide

## Quick Start (Docker Compose)

```bash
# Clone and start
git clone https://github.com/yourusername/dockmaster.git
cd dockmaster
./docker-start.sh

# Access
# Frontend: http://localhost:4000
# Backend: http://localhost:9090
# Login: admin/admin123
```

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (1.19+)
- kubectl configured
- Ingress controller (nginx recommended)

### Deploy to Kubernetes
```bash
# Deploy all components
cd k8s
./deploy.sh

# Check status
kubectl get pods -n dockmaster

# Access via port-forward
kubectl port-forward -n dockmaster svc/dockmaster-frontend 3000:3000
```

### Production Configuration
```bash
# Update secrets
kubectl create secret generic dockmaster-secrets \
  --from-literal=JWT_SECRET="your-production-secret" \
  --from-literal=ADMIN_PASSWORD="secure-password" \
  -n dockmaster

# Configure ingress domain
# Edit k8s/ingress.yaml with your domain
```

## Environment Configuration

### Docker Compose (.env)
```bash
BACKEND_PORT=9090
FRONTEND_PORT=4000
JWT_SECRET=your-secret-key
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

### Kubernetes (ConfigMap)
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dockmaster-config
data:
  BACKEND_PORT: "8081"
  LOG_LEVEL: "info"
```

## Monitoring Setup

### Prometheus Integration
```yaml
# Add to backend deployment
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8081"
  prometheus.io/path: "/metrics"
```

### Health Checks
- Backend: `GET /health`
- Frontend: `GET /`
- Database: Connection check included

## Security Considerations

### Production Checklist
- [ ] Change default admin password
- [ ] Use strong JWT secret
- [ ] Enable HTTPS/TLS
- [ ] Configure CORS properly
- [ ] Set up network policies
- [ ] Enable audit logging
- [ ] Configure resource limits

### Network Security
```yaml
# Network Policy example
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: dockmaster-netpol
spec:
  podSelector:
    matchLabels:
      app: dockmaster-backend
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: dockmaster-frontend
```

## Scaling Configuration

### Horizontal Pod Autoscaler
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: dockmaster-backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: dockmaster-backend
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Backup Strategy

### Database Backup
```bash
# Backup SQLite database
kubectl exec -n dockmaster deployment/dockmaster-backend -- \
  cp /app/data/dockmaster.db /tmp/backup-$(date +%Y%m%d).db

# Restore from backup
kubectl cp backup-20231201.db dockmaster/dockmaster-backend-pod:/app/data/dockmaster.db
```

## Troubleshooting

### Common Issues
1. **Container won't start**: Check Docker socket permissions
2. **Database errors**: Verify persistent volume claims
3. **Network issues**: Check service and ingress configuration
4. **Authentication fails**: Verify JWT secret configuration

### Debug Commands
```bash
# Check logs
kubectl logs -n dockmaster deployment/dockmaster-backend
kubectl logs -n dockmaster deployment/dockmaster-frontend

# Check resources
kubectl describe pod -n dockmaster
kubectl get events -n dockmaster

# Test connectivity
kubectl exec -n dockmaster deployment/dockmaster-backend -- curl localhost:8081/health
```

## Performance Tuning

### Resource Optimization
```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "256Mi"
    cpu: "200m"
```

### Database Optimization
- Enable WAL mode for SQLite
- Configure connection pooling
- Set up read replicas for scaling

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Deploy to Kubernetes
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Build and push Docker images
      run: |
        docker build -t dockmaster-backend:${{ github.sha }} ./backend
        docker build -t dockmaster-frontend:${{ github.sha }} ./frontend
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/dockmaster-backend backend=dockmaster-backend:${{ github.sha }} -n dockmaster
        kubectl set image deployment/dockmaster-frontend frontend=dockmaster-frontend:${{ github.sha }} -n dockmaster
```
