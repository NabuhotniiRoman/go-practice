# OIDC API Server Configuration for HTTPS Production

# Налаштування сервера
server {
  host = "0.0.0.0"
  port = 8080
  
  # Режим роботи: development, staging, production
  environment = "production"
  
  # Налаштування логування
  log_level = "info"
  log_format = "json"
  
  # Таймаути
  read_timeout  = "30s"
  write_timeout = "30s"
  idle_timeout  = "120s"
}

# Налаштування бази даних
database {
  driver   = "postgres"
  host     = "postgres-service"
  port     = 5432
  name     = "go_practice"
  user     = "oidc_api_user"
  password = ""
  
  # SSL налаштування
  ssl_mode = "disable"
  
  # Пул з'єднань
  max_open_connections = 25
  max_idle_connections = 5
  connection_max_lifetime = "5m"
}

# OIDC конфігурація
oidc {
  # Налаштування провайдера
  provider {
    issuer_url    = "https://accounts.google.com"
    client_id     = ""
    client_secret = ""
    auth_url      = "https://accounts.google.com/o/oauth2/v2/auth"
    token_url     = "https://oauth2.googleapis.com/token"
    userinfo_url  = "https://openidconnect.googleapis.com/v1/userinfo"
    issuer        = "https://accounts.google.com"

    # Redirect URLs for HTTPS
    redirect_url = "https://api.example.com/auth/callback"
    post_logout_redirect_url = "https://app.example.com"
  }
  
  # Налаштування токенів
  tokens {
    # JWT підпис
    signing_key = "dev-jwt-secret-key-change-in-production"
    signing_method = "HS256"
    
    # Час життя токенів
    access_token_duration  = "1h"
    refresh_token_duration = "24h"
    id_token_duration      = "1h"
  }
  
  # Scopes
  scopes = [
    "openid",
    "profile", 
    "email"
  ]
}

# Налаштування безпеки
security {
  # CORS для HTTPS доменів
  cors {
    allowed_origins = [
      "https://app.example.com",
      "https://api.example.com"
    ]
    
    
    allowed_methods = [
      "GET",
      "POST", 
      "PUT",
      "DELETE",
      "OPTIONS"
    ]
    
    allowed_headers = [
      "Content-Type",
      "Authorization",
      "X-Requested-With"
    ]
    
    allow_credentials = true
    max_age = 3600
  }
  
  # Rate limiting
  rate_limit {
    enabled = true
    requests_per_minute = 60
    burst = 10
  }
  
  # Session
  session {
    secret = "dev-session-secret-change-in-production"
    max_age = 3600
    secure = true
    http_only = true
  }
}

# Налаштування Redis (для сесій та кешування)
redis {
  enabled = false
  host    = "localhost"
  port    = 6379
  password = ""
  database = 0
  
  # Пул з'єднань
  max_retries = 3
  pool_size   = 10
}
