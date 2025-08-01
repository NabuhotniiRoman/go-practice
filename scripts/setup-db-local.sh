#!/bin/bash

# Скрипт для локального встановлення PostgreSQL на macOS та ініціалізації бази даних

set -e

# Конфігурація
DB_NAME="go_practice"
DB_USER="oidc_api_user"
DB_PASSWORD="oidc_secure_password_2025"
DB_HOST="localhost"
DB_PORT="5432"

echo "🐘 Локальне встановлення PostgreSQL для go_practice OIDC API Server"

# Перевірка чи встановлений Homebrew
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew не знайдено. Встановіть спочатку Homebrew:"
    echo "   /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
    exit 1
fi

# Перевірка чи встановлений PostgreSQL
if ! command -v psql &> /dev/null; then
    echo "📦 Встановлюємо PostgreSQL через Homebrew..."
    brew install postgresql@14
    
    # Додаємо в PATH
    echo 'export PATH="/opt/homebrew/opt/postgresql@14/bin:$PATH"' >> ~/.zshrc
    export PATH="/opt/homebrew/opt/postgresql@14/bin:$PATH"
    
    echo "✅ PostgreSQL встановлено"
else
    echo "✅ PostgreSQL вже встановлений: $(psql --version)"
fi

# Перевірка чи запущений PostgreSQL
echo "🔍 Перевіряємо статус PostgreSQL..."
if pg_isready -h localhost -p $DB_PORT > /dev/null 2>&1; then
    echo "✅ PostgreSQL вже запущений і готовий"
elif ! brew services list | grep postgresql@14 | grep started > /dev/null; then
    echo "🚀 Запускаємо PostgreSQL сервіс..."
    brew services start postgresql@14 2>/dev/null || {
        echo "⚠️  Помилка з brew services, спробуємо запустити вручну..."
        pg_ctl -D /opt/homebrew/var/postgresql@14 start -l /opt/homebrew/var/postgresql@14/server.log
    }
    sleep 3
else
    echo "✅ PostgreSQL сервіс запущений"
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
if ! psql -h localhost -U $(whoami) -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" | grep -q 1; then
    psql -h localhost -U $(whoami) -d postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD' CREATEDB;"
    echo "✅ Користувач $DB_USER створений"
else
    echo "✅ Користувач $DB_USER вже існує"
fi

# Створюємо базу даних (якщо не існує)
echo "🗄️ Створюємо базу даних $DB_NAME..."
if ! psql -h localhost -U $(whoami) -d postgres -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    psql -h localhost -U $(whoami) -d postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"
    echo "✅ База даних $DB_NAME створена"
else
    echo "✅ База даних $DB_NAME вже існує"
fi

# Надаємо права користувачу
echo "🔑 Налаштовуємо права доступу..."
psql -h localhost -U $(whoami) -d $DB_NAME -c "
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
echo "   Зупинити PostgreSQL: brew services stop postgresql@14"
echo "   Запустити PostgreSQL: brew services start postgresql@14"
echo "   Переглянути статус: brew services list | grep postgresql"
echo ""
echo "📱 Додатково:"
echo "   Додайте в ~/.zshrc: export PATH=\"/opt/homebrew/opt/postgresql@14/bin:\$PATH\""
echo "   Або перезапустіть термінал для застосування змін"
