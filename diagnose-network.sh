#!/bin/bash

echo "🔍 DockMaster Network Diagnostics"
echo "=================================="
echo

echo "1. Docker Services Status:"
docker-compose ps
echo

echo "2. Port Bindings:"
echo "Frontend (port 3000):"
ss -tlnp | grep :3000 || echo "Port 3000 not found"
echo "API Gateway (port 8080):"
ss -tlnp | grep :8080 || echo "Port 8080 not found"
echo

echo "3. Network Interfaces:"
ip addr show | grep -E "inet.*scope global" || ifconfig | grep -E "inet.*broadcast"
echo

echo "4. API Health Check:"
echo "Testing from localhost:"
curl -s -w "Status: %{http_code}, Time: %{time_total}s\n" http://localhost:8080/health || echo "Failed to connect to localhost:8080"
echo

echo "5. API Login Test:"
echo "Testing login from localhost:"
curl -s -w "Status: %{http_code}, Time: %{time_total}s\n" \
  -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | head -1
echo

echo "6. CORS Test:"
echo "Testing CORS preflight:"
curl -s -w "Status: %{http_code}\n" \
  -X OPTIONS http://localhost:8080/auth/login \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type"
echo

echo "7. Frontend Accessibility:"
echo "Testing frontend:"
curl -s -w "Status: %{http_code}, Time: %{time_total}s\n" http://localhost:3000 | head -1
echo

echo "🔧 Troubleshooting Tips:"
echo "========================"
echo "1. If you're accessing from a different machine, use the host machine's IP instead of localhost"
echo "2. Make sure no firewall is blocking ports 3000 and 8080"
echo "3. Check browser console for detailed error messages"
echo "4. Try accessing: http://$(hostname -I | awk '{print $1}'):3000"
echo "5. Use the network test page: http://localhost:3000/../network-test.html"
echo

echo "✅ Diagnostics completed!"
