# Dockerfile для Go API Server
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копіюємо go mod файли
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код
COPY . .

# Будуємо додаток для Kubernetes
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o api-server ./cmd/api-server

# Фінальний образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копіюємо бінарник і конфігурації
COPY --from=builder /app/api-server .
COPY --from=builder /app/configs/ ./configs/

# Відкриваємо порт
EXPOSE 8080

# Команда запуску
CMD ["./api-server", "server"]
