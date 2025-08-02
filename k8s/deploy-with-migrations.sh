#!/bin/bash

# Скрипт для деплоювання з автоматичними міграціями
set -e

echo "🚀 Deploying Go API with database migrations..."

# Застосовуємо PostgreSQL
echo "📦 Applying PostgreSQL..."
kubectl apply -f postgres.yaml

# Чекаємо поки PostgreSQL буде готовий
echo "⏳ Waiting for PostgreSQL to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres --timeout=300s

# Застосовуємо ConfigMap і Secrets
echo "🔧 Applying ConfigMap and Secrets..."
kubectl apply -f configmap.yaml

# Запускаємо міграції
echo "🗃️ Running database migrations..."
kubectl delete job db-migration --ignore-not-found=true
kubectl apply -f db-migration-job.yaml

# Чекаємо завершення міграцій
echo "⏳ Waiting for migrations to complete..."
kubectl wait --for=condition=complete job/db-migration --timeout=300s

# Перевіряємо статус міграцій
migration_status=$(kubectl get job db-migration -o jsonpath='{.status.succeeded}')
if [ "$migration_status" != "1" ]; then
    echo "❌ Migration failed!"
    kubectl logs job/db-migration
    exit 1
fi

echo "✅ Migrations completed successfully!"

# Застосовуємо основний деплоймент
echo "🚀 Deploying Go API..."
kubectl apply -f go-api-deployment.yaml
kubectl apply -f service.yaml

# Застосовуємо Ingress (якщо існує)
if [ -f "ingress.yaml" ]; then
    echo "🌐 Applying Ingress..."
    kubectl apply -f ingress.yaml
fi

echo "✅ Deployment completed successfully!"
echo ""
echo "📊 Checking deployment status..."
kubectl get pods -l app=go-api
kubectl get pods -l app=postgres
kubectl get svc
