# Dockerfile для Go API Server
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копіюємо go mod файли
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код
COPY . .

# Будуємо додаток для Kubernetes
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o k8s-server ./cmd/k8s-server

# Фінальний образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копіюємо бінарник
COPY --from=builder /app/k8s-server .

# Відкриваємо порт
EXPOSE 8080

# Команда запуску
CMD ["./k8s-server"]
