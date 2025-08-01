# OIDC Practice Project

Практичний проект для вивчення OpenID Connect (OIDC) з використанням бібліотеки Zitadel OIDC v3.

## Структура проекту

```
go-practice/
├── cmd/                      # Точки входу для різних додатків
│   ├── web-client/          # Веб-клієнт з Authorization Code Flow
│   ├── api-server/          # API сервер з перевіркою токенів
│   ├── cli-client/          # CLI клієнт для desktop додатків
│   ├── device-client/       # Device Flow для IoT пристроїв
│   └── service-client/      # Service-to-service автентифікація
├── pkg/                     # Публічні бібліотеки/пакети
│   ├── auth/               # Логіка автентифікації
│   ├── middleware/         # HTTP middleware для автентифікації
│   ├── handlers/           # HTTP handlers
│   ├── models/             # Структури даних
│   └── utils/              # Утилітарні функції
├── internal/               # Приватна логіка додатку
│   ├── config/            # Конфігурація
│   ├── storage/           # Зберігання даних
│   └── server/            # HTTP сервер
├── web/                   # Веб ресурси
│   ├── templates/         # HTML шаблони
│   └── static/           # CSS, JS, зображення
│       ├── css/
│       └── js/
├── configs/               # Файли конфігурації
├── scripts/              # Скрипти для збірки та деплоя
├── docs/                 # Документація
├── test/                 # Тести
├── go.mod                # Go модуль
└── main.go              # Головний файл для демонстрації
```

## Компоненти проекту

### 1. Web Client (`cmd/web-client/`)
- Повноцінний веб-додаток з HTML формами
- Authorization Code Flow
- PKCE підтримка для SPA
- Cookie-based session management

### 2. API Server (`cmd/api-server/`)
- REST API з JWT токен валідацією
- Middleware для перевірки автентифікації
- Protected endpoints
- Token introspection

### 3. CLI Client (`cmd/cli-client/`)
- Desktop додаток
- Browser-based authentication flow
- Local callback server

### 4. Device Client (`cmd/device-client/`)
- IoT/Smart TV flow
- Device authorization grant
- Polling для токенів

### 5. Service Client (`cmd/service-client/`)
- Machine-to-machine автентифікація
- Client Credentials Flow
- JWT Profile

## Загальні пакети

### `pkg/auth/`
- OIDC клієнт обгортки
- Token management
- Validation utilities

### `pkg/middleware/`
- HTTP middleware для автентифікації
- CORS handling
- Logging

### `pkg/handlers/`
- HTTP handlers для різних endpoints
- Login/logout handlers
- Callback handlers

### `pkg/models/`
- User структури
- Token claims
- Configuration models

## Конфігурація

Всі додатки підтримують конфігурацію через:
- Environment variables
- Configuration files в `configs/`
- Command line flags

## Використання

```bash
# Веб клієнт
go run cmd/web-client/main.go

# API сервер
go run cmd/api-server/main.go

# CLI клієнт
go run cmd/cli-client/main.go

# Device flow
go run cmd/device-client/main.go

# Service клієнт
go run cmd/service-client/main.go
```

## Залежності

- `github.com/zitadel/oidc/v3` - OIDC бібліотека
- `github.com/gorilla/mux` - HTTP router
- `github.com/sirupsen/logrus` - Logging
