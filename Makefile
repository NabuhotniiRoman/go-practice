# Makefile для OIDC API Server

APP_PATH := $(shell pwd)
MODULE_NAME := go-practice
BUILD_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_NUMBER ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "local")

.PHONY: help build test clean run-api dev setup-db-local db-connect dependencies configure _local.hcl k8s-start k8s-stop k8s-status k8s-logs docker-build docker-push k8s-deploy

# Показати допомогу
help:
	@echo "Available commands:"
	@echo "  run-api         - Run API server with _local.hcl config"
	@echo "  configure       - Generate _local.hcl from template"
	@echo "  dev             - Generate config and run API server"
	@echo "  build           - Build API server"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  dependencies    - Install dependencies"
	@echo "  setup-db-local  - Setup PostgreSQL database locally"
	@echo "  db-connect      - Connect to the database"
	@echo ""
	@echo "Kubernetes commands:"
	@echo "  k8s-start       - Start port forwarding (API: 8080, Frontend: 3000)"
	@echo "  k8s-stop        - Stop all port forwarding"
	@echo "  k8s-status      - Show Kubernetes status"
	@echo "  k8s-logs        - Show API logs"
	@echo "  docker-build    - Build and push Docker images"
	@echo "  k8s-deploy      - Deploy to Kubernetes via ArgoCD"

# Залежності
dependencies:
	go mod tidy
	go mod download

# Генерація конфігурації
configure:
	go run ./cmd/api-server configure \
		-t ${APP_PATH}/configs/oidc-api.hcl.tmpl \
		-o ${APP_PATH}/_local.hcl \
		-v ${BUILD_VERSION} \
		-m local

# Генерація _local.hcl файлу
_local.hcl: dependencies configs/oidc-api.hcl.tmpl
	@$(MAKE) configure

# Запуск API сервера
run-api: dependencies _local.hcl
	go run -ldflags "-X ${MODULE_NAME}/internal/build.Version=${BUILD_VERSION} -X ${MODULE_NAME}/internal/build.Number=${BUILD_NUMBER}" \
		./cmd/api-server server -c _local.hcl

# Development mode (генерація конфігу + запуск)
dev: configure run-api

# Збірка API сервера
build:
	@echo "Building API server..."
	go build -ldflags "-X ${MODULE_NAME}/internal/build.Version=${BUILD_VERSION} -X ${MODULE_NAME}/internal/build.Number=${BUILD_NUMBER}" \
		-o bin/api-server ./cmd/api-server

# Запуск тестів
test:
	go test -v ./...

# Очистка артефактів
clean:
	rm -rf bin/
	rm -f _local.hcl
	go clean

# Налаштування бази даних (локально на macOS)
setup-db-local:
	@echo "🔧 Setting up PostgreSQL database locally..."
	chmod +x ./scripts/setup-db-local.sh
	./scripts/setup-db-local.sh

# Підключення до бази даних
db-connect:
	@echo "🔗 Connecting to go_practice database..."
	psql -h localhost -U oidc_api_user -d go_practice

swagger:
	swag init -g cmd/api-server/main.go

# Kubernetes команди
k8s-start:
	@echo "🚀 Starting port forwarding..."
	@echo "   - Go API: https://api.example.com"
	@echo "   - React Frontend: http://localhost:3000"
	@echo ""
	@echo "Press Ctrl+C to stop"
	@./scripts/port-forward.sh

k8s-stop:
	@echo "🛑 Stopping all port forwarding..."
	@pkill -f "kubectl port-forward" || true
	@echo "✅ Port forwarding stopped"

k8s-status:
	@echo "📊 Kubernetes Status:"
	@echo ""
	@echo "Pods:"
	@kubectl get pods -l app=go-api
	@kubectl get pods -l app=react-frontend
	@echo ""
	@echo "Services:"
	@kubectl get svc go-api-service react-frontend-service
	@echo ""
	@echo "ArgoCD Applications:"
	@kubectl get applications -n argocd

k8s-logs:
	@echo "📋 API Server Logs (last 50 lines):"
	@kubectl logs -l app=go-api --tail=50

docker-build:
	@echo "🐳 Building Docker images..."
	@cd .. && docker build -t nabuhotnii/go-practice:latest -f go-practice/Dockerfile go-practice/
	@cd ../go-practice-ui && docker build -t nabuhotnii/go-practice-ui:latest .
	@echo "✅ Docker images built"

docker-push: docker-build
	@echo "📤 Pushing Docker images..."
	@docker push nabuhotnii/go-practice:latest
	@docker push nabuhotnii/go-practice-ui:latest
	@echo "✅ Docker images pushed"

k8s-deploy: docker-push
	@echo "🚀 Triggering ArgoCD sync..."
	@kubectl patch app my-go-api -n argocd --type merge -p '{"operation":{"sync":{"revision":"HEAD"}}}'
	@kubectl patch app my-react-frontend -n argocd --type merge -p '{"operation":{"sync":{"revision":"HEAD"}}}'
	@echo "✅ ArgoCD sync triggered"