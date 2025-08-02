#!/bin/bash


# Скрипт для локального встановлення та ініціалізації PostgreSQL бази даних (Linux)


set -e

# Конфігурація
DB_NAME="go_practice"
DB_USER="oidc_api_user"
DB_PASSWORD="oidc_secure_password_2025"
DB_HOST="localhost"
DB_PORT="5432"


echo "🐘 Локальне встановлення PostgreSQL для go_practice OIDC API Server (Linux)"

# Перевірка чи встановлений PostgreSQL
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL не знайдено. Встановіть його через apt або інший пакетний менеджер:"
    echo "   sudo apt update && sudo apt install postgresql postgresql-contrib"
    exit 1
else
    echo "✅ PostgreSQL вже встановлений: $(psql --version)"
fi

# Перевірка чи запущений PostgreSQL
echo "🔍 Перевіряємо статус PostgreSQL..."
if pg_isready -h localhost -p $DB_PORT > /dev/null 2>&1; then
    echo "✅ PostgreSQL вже запущений і готовий"
else
    echo "🚀 Запускаємо PostgreSQL..."
    sudo service postgresql start
    sleep 3
fi

# Функція для перевірки підключення
wait_for_postgres() {
    echo "⏳ Очікуємо готовності PostgreSQL..."
    for i in {1..30}; do
        if pg_isready -h localhost -p $DB_PORT > /dev/null 2>&1; then
            echo "✅ PostgreSQL готовий!"
            return 0
        fi
        sleep 1
    done
    echo "❌ PostgreSQL не відповідає"
    exit 1
}

# Очікуємо готовності
wait_for_postgres


# Створюємо користувача (якщо не існує)
echo "👤 Створюємо користувача $DB_USER..."
if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" | grep -q 1; then
    sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD' CREATEDB;"
    echo "✅ Користувач $DB_USER створений"
else
    echo "✅ Користувач $DB_USER вже існує"
fi

# Створюємо базу даних (якщо не існує)
echo "🗄️ Створюємо базу даних $DB_NAME..."
if ! sudo -u postgres psql -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"
    echo "✅ База даних $DB_NAME створена"
else
    echo "✅ База даних $DB_NAME вже існує"
fi

# Надаємо права користувачу
echo "🔑 Налаштовуємо права доступу..."
sudo -u postgres psql -d $DB_NAME -c "
    GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
    GRANT ALL PRIVILEGES ON SCHEMA public TO $DB_USER;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
"


# Виконуємо SQL скрипт ініціалізації
echo "📊 Виконуємо скрипт ініціалізації бази даних..."
psql -h localhost -U $DB_USER -d $DB_NAME -f "$(dirname "$0")/db/init.sql"

echo ""
echo "🎉 Локальна база даних успішно створена та ініціалізована!"
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
echo "   Підключитись до бази: psql -h localhost -U $DB_USER -d $DB_NAME"
echo "   Зупинити PostgreSQL: sudo service postgresql stop"
echo "   Запустити PostgreSQL: sudo service postgresql start"
echo "   Переглянути статус: sudo service postgresql status"
