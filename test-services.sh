#!/bin/bash

# DockMaster - Service Health Check Script

echo "🔍 Testing DockMaster Microservices..."
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to test service health
test_service() {
    local service_name=$1
    local port=$2
    local url="http://localhost:${port}/health"
    
    echo -n "Testing ${service_name} (${port}): "
    
    if curl -s -f "${url}" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ Unhealthy${NC}"
        return 1
    fi
}

# Test all services
services=(
    "API Gateway:8080"
    "Auth Service:8081"
    "Container Service:8082"
    "Image Service:8083"
    "Volume Service:8084"
    "Network Service:8085"
)

healthy_count=0
total_count=${#services[@]}

for service in "${services[@]}"; do
    IFS=':' read -r name port <<< "$service"
    if test_service "$name" "$port"; then
        ((healthy_count++))
    fi
done

echo "======================================"
echo -e "Health Check Summary: ${healthy_count}/${total_count} services healthy"

# Test frontend
echo -n "Testing Frontend (3000): "
if curl -s -I "http://localhost:3000" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Accessible${NC}"
    ((healthy_count++))
    ((total_count++))
else
    echo -e "${RED}❌ Not accessible${NC}"
    ((total_count++))
fi

echo "======================================"
echo -e "Overall Status: ${healthy_count}/${total_count} services operational"

if [ $healthy_count -eq $total_count ]; then
    echo -e "${GREEN}🎉 All services are running correctly!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some services are not responding${NC}"
    exit 1
fi
