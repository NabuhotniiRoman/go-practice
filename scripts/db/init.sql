-- Ініціалізація бази даних go_practice для OIDC API Server
-- Створення бази даних та окремих таблиць з правильними зв'язками

-- Створення бази даних go_practice (якщо не існує)
-- Цей блок виконується окремо від основного скрипту
-- CREATE DATABASE go_practice OWNER postgres;

-- Підключення до бази go_practice
-- \c go_practice;

-- Видалення існуючих таблиць (у правильному порядку)
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS auth_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Таблиця користувачів
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255), -- для локальних користувачів
    picture TEXT,
    
    -- OIDC специфічні поля
    sub VARCHAR(255) UNIQUE, -- може бути NULL для локальних користувачів
    iss VARCHAR(255),        -- може бути NULL для локальних користувачів
    aud TEXT[],              -- може бути NULL для локальних користувачів
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    
    -- Статус користувача
    is_active BOOLEAN DEFAULT TRUE,
    is_email_verified BOOLEAN DEFAULT FALSE,
    
    -- Тип автентифікації: 'local' або 'oidc'
    auth_type VARCHAR(10) DEFAULT 'local' CHECK (auth_type IN ('local', 'oidc'))
);

-- Таблиця для OIDC токенів
CREATE TABLE auth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Токени
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    id_token TEXT,
    
    -- Метаданні токена
    token_type VARCHAR(50) DEFAULT 'Bearer',
    scope TEXT,
    
    -- Час життя
    expires_in INTEGER NOT NULL,  -- в секундах
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- OIDC специфічні поля
    exp BIGINT,        -- OIDC expiration time (unix timestamp)
    iat BIGINT,        -- OIDC issued at time (unix timestamp)
    auth_time BIGINT,  -- OIDC authentication time (unix timestamp)
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Статус токена
    is_revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Таблиця сесій користувачів
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    auth_token_id UUID REFERENCES auth_tokens(id) ON DELETE SET NULL,
    
    -- Метаданні сесії
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}',
    
    -- Час життя сесії
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Статус сесії
    is_active BOOLEAN DEFAULT TRUE
);

-- Створення індексів для оптимізації запитів

-- Індекси для таблиці users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_sub ON users(sub);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_last_login ON users(last_login_at);

-- Індекси для таблиці auth_tokens
CREATE INDEX idx_auth_tokens_user_id ON auth_tokens(user_id);
CREATE INDEX idx_auth_tokens_access_token ON auth_tokens(access_token);
CREATE INDEX idx_auth_tokens_refresh_token ON auth_tokens(refresh_token);
CREATE INDEX idx_auth_tokens_expires_at ON auth_tokens(expires_at);
CREATE INDEX idx_auth_tokens_revoked ON auth_tokens(is_revoked);
CREATE INDEX idx_auth_tokens_created_at ON auth_tokens(created_at);

-- Індекси для таблиці user_sessions
CREATE INDEX idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_id ON user_sessions(auth_token_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active);
CREATE INDEX idx_user_sessions_last_accessed ON user_sessions(last_accessed_at);

-- Функція для автоматичного оновлення updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Тригери для автоматичного оновлення updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_auth_tokens_updated_at 
    BEFORE UPDATE ON auth_tokens 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Функція для очищення застарілих токенів
CREATE OR REPLACE FUNCTION cleanup_expired_data()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Помічаємо токени як revoked, якщо вони застаріли
    UPDATE auth_tokens 
    SET is_revoked = TRUE, revoked_at = CURRENT_TIMESTAMP
    WHERE expires_at < CURRENT_TIMESTAMP 
      AND is_revoked = FALSE;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Деактивуємо сесії з застарілими токенами
    UPDATE user_sessions 
    SET is_active = FALSE
    WHERE (expires_at < CURRENT_TIMESTAMP 
           OR auth_token_id IN (
               SELECT id FROM auth_tokens WHERE is_revoked = TRUE
           ))
      AND is_active = TRUE;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Вставка тестових даних
-- OIDC користувачі
INSERT INTO users (email, name, sub, iss, aud, is_email_verified, auth_type) VALUES 
('john.doe@example.com', 'John Doe', 'auth0|user123', 'https://your-domain.auth0.com/', ARRAY['your-api-audience'], TRUE, 'oidc'),
('jane.smith@example.com', 'Jane Smith', 'auth0|user456', 'https://your-domain.auth0.com/', ARRAY['your-api-audience'], TRUE, 'oidc');

-- Локальний користувач (з паролем)
-- Пароль: "password123" -> хеш $2a$10$rJZw8qfUn8QqKqV3NqHhAeH.xTqQJCwJ8J8QqKqV3NqHhAeH.xTqQJC
INSERT INTO users (email, name, password_hash, is_email_verified, auth_type) VALUES 
('local.user@example.com', 'Local User', '$2a$10$rJZw8qfUn8QqKqV3NqHhAeH.xTqQJCwJ8J8QqKqV3NqHhAeH.xTqQJC', TRUE, 'local');

-- Локальний користувач (з паролем)
-- Пароль: "password123" -> хеш $2a$10$rJZw8qfUn8QqKqV3NqHhAeH.xTqQJCwJ8J8QqKqV3NqHhAeH.xTqQJC
INSERT INTO users (email, name, password_hash, is_email_verified, auth_type) VALUES 
('local.user@example.com', 'Local User', '$2a$10$rJZw8qfUn8QqKqV3NqHhAeH.xTqQJCwJ8J8QqKqV3NqHhAeH.xTqQJC', TRUE, 'local');

-- Токени (пов'язані з користувачами)
INSERT INTO auth_tokens (user_id, access_token, refresh_token, token_type, scope, expires_in, expires_at, exp, iat) 
SELECT u.id, 'access_token_123456', 'refresh_token_789012', 'Bearer', 'openid profile email', 3600, 
       CURRENT_TIMESTAMP + INTERVAL '1 hour', 
       EXTRACT(epoch FROM CURRENT_TIMESTAMP + INTERVAL '1 hour')::BIGINT,
       EXTRACT(epoch FROM CURRENT_TIMESTAMP)::BIGINT
FROM users u WHERE u.email = 'john.doe@example.com';

-- Сесії (пов'язані з користувачами та токенами)
INSERT INTO user_sessions (session_id, user_id, auth_token_id, user_agent, ip_address, expires_at, last_accessed_at, device_info)
SELECT 'session_abc123456', u.id, t.id, 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', '192.168.1.100'::INET,
       CURRENT_TIMESTAMP + INTERVAL '24 hours', CURRENT_TIMESTAMP,
       '{"device": "desktop", "os": "macOS", "browser": "Chrome"}'::JSONB
FROM users u 
JOIN auth_tokens t ON t.user_id = u.id
WHERE u.email = 'john.doe@example.com'
LIMIT 1;

-- Надання прав користувачу oidc_api_user (виконується в setup-db.sh)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO oidc_api_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO oidc_api_user;
