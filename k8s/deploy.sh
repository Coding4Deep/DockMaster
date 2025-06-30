#!/bin/bash

echo "🚀 Deploying DockMaster to Kubernetes..."

# Apply namespace first
kubectl apply -f namespace.yaml

# Apply storage
kubectl apply -f storage.yaml

# Apply configmap and secrets
kubectl apply -f configmap.yaml

# Apply deployments
kubectl apply -f backend-deployment.yaml
kubectl apply -f frontend-deployment.yaml

# Apply ingress
kubectl apply -f ingress.yaml

echo "✅ DockMaster deployed successfully!"
echo ""
echo "📊 Check deployment status:"
echo "kubectl get pods -n dockmaster"
echo ""
echo "🌐 Access application:"
echo "kubectl port-forward -n dockmaster svc/dockmaster-frontend 3000:3000"
echo "kubectl port-forward -n dockmaster svc/dockmaster-backend 8081:8081"
