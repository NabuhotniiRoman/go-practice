#!/bin/bash

# Скрипт для створення та ініціалізації PostgreSQL бази даних для OIDC API Server

set -e

# Конфігурація
DB_NAME="go_practice"
DB_USER="oidc_api_user"
DB_PASSWORD="oidc_secure_password_2025"
DB_HOST="localhost"
DB_PORT="5432"
CONTAINER_NAME="go-practice-postgres"

echo "🐘 Ініціалізація PostgreSQL бази даних go_practice для OIDC API Server"

# Функція для перевірки чи PostgreSQL готовий
wait_for_postgres() {
    echo "⏳ Очікуємо готовності PostgreSQL..."
    until docker exec $CONTAINER_NAME pg_isready -U postgres -h localhost > /dev/null 2>&1; do
        sleep 1
    done
    echo "✅ PostgreSQL готовий!"
}

# Перевіряємо чи контейнер вже запущений
if docker ps -a | grep -q $CONTAINER_NAME; then
    echo "📦 Знайдено існуючий контейнер $CONTAINER_NAME"
    
    # Якщо контейнер зупинений, запускаємо його
    if ! docker ps | grep -q $CONTAINER_NAME; then
        echo "🔄 Запускаємо зупинений контейнер..."
        docker start $CONTAINER_NAME
    else
        echo "✅ Контейнер вже запущений"
    fi
else
    echo "🆕 Створюємо новий PostgreSQL контейнер..."
    docker run --name $CONTAINER_NAME \
        -e POSTGRES_DB=$DB_NAME \
        -e POSTGRES_USER=postgres \
        -e POSTGRES_PASSWORD=postgres \
        -p $DB_PORT:5432 \
        -d postgres:15-alpine
fi

# Очікуємо готовності бази
wait_for_postgres

# Створюємо користувача та базу даних
echo "👤 Створюємо користувача та базу даних..."
docker exec $CONTAINER_NAME psql -U postgres -c "
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
            CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
        END IF;
    END
    \$\$;
"

docker exec $CONTAINER_NAME psql -U postgres -c "
    SELECT 'CREATE DATABASE $DB_NAME OWNER $DB_USER;' 
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME');
" | docker exec -i $CONTAINER_NAME psql -U postgres

# Надаємо права користувачу
echo "🔑 Налаштовуємо права доступу..."
docker exec $CONTAINER_NAME psql -U postgres -d $DB_NAME -c "
    GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
    GRANT ALL PRIVILEGES ON SCHEMA public TO $DB_USER;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
"

# Виконуємо SQL скрипт ініціалізації
echo "📊 Виконуємо скрипт ініціалізації бази даних..."
docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME < "$(dirname "$0")/db/init.sql"

echo ""
echo "🎉 База даних успішно створена та ініціалізована!"
echo ""
echo "📋 Інформація для підключення:"
echo "   Host: $DB_HOST"
echo "   Port: $DB_PORT" 
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"
echo "   Password: $DB_PASSWORD"
echo ""
echo "🔗 Connection String:"
echo "   postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
echo ""
echo "🛠️ Корисні команди:"
echo "   Підключитись до бази: docker exec -it $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME"
echo "   Зупинити контейнер: docker stop $CONTAINER_NAME"
echo "   Видалити контейнер: docker rm $CONTAINER_NAME"
echo "   Переглянути логи: docker logs $CONTAINER_NAME"
