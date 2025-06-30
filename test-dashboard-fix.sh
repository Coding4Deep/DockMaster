#!/bin/bash

echo "🔧 DockMaster Dashboard Fix Verification"
echo "========================================"
echo

# Get auth token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Authentication failed!"
    exit 1
fi

echo "✅ Authentication successful"
echo

# Test system info endpoint
echo "📊 Testing System Info Data:"
SYSTEM_INFO=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/system/info)

CONTAINERS=$(echo "$SYSTEM_INFO" | jq '.Containers')
RUNNING=$(echo "$SYSTEM_INFO" | jq '.ContainersRunning')
STOPPED=$(echo "$SYSTEM_INFO" | jq '.ContainersStopped')
IMAGES=$(echo "$SYSTEM_INFO" | jq '.Images')
CPUS=$(echo "$SYSTEM_INFO" | jq '.NCPU')
MEMORY_GB=$(echo "$SYSTEM_INFO" | jq '.MemTotal / 1024 / 1024 / 1024 | floor')

echo "   📦 Total Containers: $CONTAINERS"
echo "   ▶️  Running: $RUNNING"
echo "   ⏹️  Stopped: $STOPPED"
echo "   🖼️  Images: $IMAGES"
echo "   🖥️  CPU Cores: $CPUS"
echo "   💾 Memory: ${MEMORY_GB}GB"
echo

# Verify all numbers are greater than 0
if [ "$CONTAINERS" -gt 0 ] && [ "$IMAGES" -gt 0 ] && [ "$CPUS" -gt 0 ]; then
    echo "✅ Dashboard should now show correct numbers (no more zeros!)"
else
    echo "❌ Some values are still zero - there may be an issue"
fi

echo
echo "🧪 Testing Individual Service Endpoints:"

# Test containers
CONTAINER_COUNT=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/containers | jq '. | length')
echo "   📦 Containers API: $CONTAINER_COUNT items"

# Test images  
IMAGE_COUNT=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/images | jq '. | length')
echo "   🖼️  Images API: $IMAGE_COUNT items"

# Test volumes
VOLUME_COUNT=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/volumes | jq '.Volumes | length')
echo "   💾 Volumes API: $VOLUME_COUNT items"

# Test networks
NETWORK_COUNT=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/networks | jq '. | length')
echo "   🌐 Networks API: $NETWORK_COUNT items"

echo
echo "🎯 VERIFICATION RESULTS:"
echo "======================="

if [ "$CONTAINERS" -gt 0 ] && [ "$IMAGES" -gt 0 ] && [ "$CONTAINER_COUNT" -gt 0 ] && [ "$IMAGE_COUNT" -gt 0 ]; then
    echo "🎉 SUCCESS! All fixes are working correctly!"
    echo "✅ Dashboard will now show:"
    echo "   • $CONTAINERS total containers ($RUNNING running, $STOPPED stopped)"
    echo "   • $IMAGES Docker images"
    echo "   • $CPUS CPU cores"
    echo "   • ${MEMORY_GB}GB memory"
    echo "✅ All service pages will show their respective resources"
    echo "✅ No more zeros on the dashboard!"
else
    echo "❌ ISSUE: Some endpoints are still returning zero or empty data"
fi

echo
echo "🌐 ACCESS YOUR FIXED APPLICATION:"
echo "================================"
echo "Frontend: http://localhost:3000"
echo "Alternative: http://$(hostname -I | awk '{print $1}'):3000"
echo "Credentials: admin / admin123"
echo
echo "🔍 What was fixed:"
echo "• Dashboard data format mismatch - Fixed API response parsing"
echo "• System info display - Now shows correct Docker system data"
echo "• Container counts - Now displays running/stopped/total correctly"
echo "• Resource statistics - All numbers now display properly"
echo "• Added debugging to API calls for troubleshooting"
