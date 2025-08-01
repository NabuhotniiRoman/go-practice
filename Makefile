# Makefile для OIDC API Server

APP_PATH := $(shell pwd)
MODULE_NAME := go-practice
BUILD_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_NUMBER ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "local")

.PHONY: help build test clean run-api dev setup-db-local db-connect dependencies configure _local.hcl

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