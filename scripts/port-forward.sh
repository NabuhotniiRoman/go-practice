#!/bin/bash

# Скрипт для автоматичного port forwarding
set -e

echo "🚀 Starting automatic port forwarding..."

# Функція для очищення процесів при виході
cleanup() {
    echo "🛑 Stopping port forwarding..."
    pkill -f "kubectl port-forward" || true
    exit 0
}

# Встановлюємо обробник сигналів
trap cleanup SIGINT SIGTERM

# Функція для запуску port forwarding з перевіркою
start_port_forward() {
    local service=$1
    local local_port=$2
    local target_port=$3
    local service_name=$4
    
    echo "🔗 Starting port forwarding for $service_name ($service:$target_port -> localhost:$local_port)"
    
    while true; do
        kubectl port-forward service/$service $local_port:$target_port &
        local pid=$!
        
        # Чекаємо поки процес запуститься або впаде
        sleep 5
        
        if kill -0 $pid 2>/dev/null; then
            echo "✅ Port forwarding for $service_name is running (PID: $pid)"
            wait $pid
        else
            echo "❌ Port forwarding for $service_name failed, restarting in 5 seconds..."
            sleep 5
        fi
    done
}

# Перевіряємо чи доступний кластер
if ! kubectl cluster-info >/dev/null 2>&1; then
    echo "❌ Kubernetes cluster is not accessible"
    exit 1
fi

echo "✅ Kubernetes cluster is accessible"

# Запускаємо port forwarding для обох сервісів в фоні
start_port_forward "go-api-service" "8080" "8080" "Go API" &
start_port_forward "react-frontend-service" "3000" "80" "React Frontend" &

echo "🎯 Port forwarding started:"
echo "   - Go API: https://api.example.com"
echo "   - React Frontend: http://localhost:3000"
echo ""
echo "Press Ctrl+C to stop all port forwarding"

# Чекаємо поки всі фонові процеси не завершаться
wait
