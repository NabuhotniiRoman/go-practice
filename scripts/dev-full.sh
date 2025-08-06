#!/bin/bash

# Dev Full Environment Setup Script
echo "🚀 Starting full development environment..."

# Кольори для виводу
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функція для чекання порту
wait_for_port() {
    local port=$1
    local service=$2
    echo "⏳ Waiting for $service on port $port..."
    while ! nc -z localhost $port > /dev/null 2>&1; do
        sleep 1
    done
    echo "✅ $service is ready on port $port"
}

# Функція для зупинки всіх port-forward процесів
cleanup() {
    echo -e "\n${RED}🛑 Stopping all port forwarding...${NC}"
    pkill -f "kubectl port-forward" || true
    # Видаляємо запис з /etc/hosts якщо він був доданий
    if grep -q "127.0.0.1.*api.example.com" /etc/hosts 2>/dev/null; then
        echo "🧹 Removing api.example.com from /etc/hosts..."
        sudo sed -i '/127.0.0.1.*api.example.com/d' /etc/hosts || true
    fi
    echo -e "${GREEN}✅ Cleanup completed${NC}"
    exit 0
}

# Встановлюємо обробник сигналів
trap cleanup SIGINT SIGTERM

# Перевіряємо чи все працює в k8s
echo "📊 Checking Kubernetes status..."
kubectl get pods -l app=go-api --no-headers | head -1
kubectl get pods -l app=react-frontend --no-headers | head -1

# Перевіряємо чи встановлений netcat
if ! command -v nc &> /dev/null; then
    echo "⚠️  netcat not found, installing..."
    sudo apt-get update && sudo apt-get install -y netcat-openbsd
fi

# Налаштовуємо локальний домен api.example.com
echo "🌐 Setting up local domain api.example.com..."
if ! grep -q "127.0.0.1.*api.example.com" /etc/hosts 2>/dev/null; then
    echo "Adding api.example.com to /etc/hosts..."
    echo "127.0.0.1 api.example.com" | sudo tee -a /etc/hosts > /dev/null
    echo "✅ Domain api.example.com added to /etc/hosts"
else
    echo "✅ Domain api.example.com already configured"
fi

echo -e "\n${BLUE}🚀 Starting port forwarding...${NC}"

# Запускаємо port-forwarding для API в фоновому режимі
echo "🔗 Starting API port forwarding (8080 -> go-api-service:8080)..."
kubectl port-forward svc/go-api-service 8080:8080 > /dev/null 2>&1 &
API_PID=$!

# Запускаємо port-forwarding для фронтенду в фоновому режимі  
echo "🔗 Starting Frontend port forwarding (3000 -> react-frontend-service:80)..."
kubectl port-forward svc/react-frontend-service 3000:80 > /dev/null 2>&1 &
FRONTEND_PID=$!

# Чекаємо поки сервіси будуть готові
wait_for_port 8080 "API"
wait_for_port 3000 "Frontend"

echo -e "\n${GREEN}✅ Development environment is ready!${NC}"
echo -e "\n📋 ${BLUE}Access URLs:${NC}"
echo -e "   🔧 API Server:     ${YELLOW}http://api.example.com:8080${NC}"
echo -e "   🌐 Frontend:       ${YELLOW}http://localhost:3000${NC}"
echo -e "   📚 API Docs:       ${YELLOW}http://api.example.com:8080/docs${NC}"
echo -e "   🔍 Health Check:   ${YELLOW}http://api.example.com:8080/health${NC}"

echo -e "\n💡 ${BLUE}Tips:${NC}"
echo "   - Frontend will automatically connect to http://api.example.com:8080"
echo "   - API is accessible both locally and from frontend"
echo "   - Press Ctrl+C to stop all services"

# Відкриваємо фронтенд в браузері
echo -e "\n🌐 Opening frontend in browser..."
sleep 2
google-chrome http://localhost:3000 2>/dev/null || chromium-browser http://localhost:3000 2>/dev/null || xdg-open http://localhost:3000 2>/dev/null || echo "❌ Could not open browser automatically"

echo -e "\n${GREEN}🎉 Ready to develop! Press Ctrl+C to stop.${NC}"

# Чекаємо поки користувач не зупинить
wait
