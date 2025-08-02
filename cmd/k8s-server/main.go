// API Server для Kubernetes - читає конфігурацію зі змінних середовища
package main

import (
	"log"
	"os"
	"strconv"

	"go-practice/internal/config"
)

func main() {
	// Читаємо конфігурацію зі змінних середовища
	cfg := loadConfigFromEnv()

	// Запускаємо сервер
	if err := config.StartServer(cfg); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loadConfigFromEnv() *config.Config {
	// Парсимо порт
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		port = 8080
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		dbPort = 5432
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         getEnv("HOST", "0.0.0.0"),
			Port:         port,
			Environment:  getEnv("MODE", "production"),
			LogLevel:     getEnv("LOG_LEVEL", "info"),
			LogFormat:    getEnv("LOG_FORMAT", "text"),
			ReadTimeout:  getEnv("READ_TIMEOUT", "30s"),
			WriteTimeout: getEnv("WRITE_TIMEOUT", "30s"),
			IdleTimeout:  getEnv("IDLE_TIMEOUT", "120s"),
		},

		Database: config.DatabaseConfig{
			Driver:                getEnv("DB_DRIVER", "postgres"),
			Host:                  getEnv("DB_HOST", "postgres-service"),
			Port:                  dbPort,
			Name:                  getEnv("DB_NAME", "go_practice"),
			User:                  getEnv("DB_USER", "oidc_api_user"),
			Password:              getEnv("DB_PASSWORD", ""),
			SSLMode:               getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConnections:    10,
			MaxIdleConnections:    5,
			ConnectionMaxLifetime: getEnv("DB_CONN_MAX_LIFETIME", "5m"),
		},

		OIDC: config.OIDCConfig{
			Provider: config.OIDCProviderConfig{
				IssuerURL:             getEnv("OIDC_ISSUER_URL", ""),
				ClientID:              getEnv("OIDC_CLIENT_ID", ""),
				ClientSecret:          getEnv("OIDC_CLIENT_SECRET", ""),
				RedirectURL:           getEnv("OIDC_REDIRECT_URL", "http://localhost:8080/auth/callback"),
				PostLogoutRedirectURL: getEnv("OIDC_POST_LOGOUT_URL", "http://localhost:3000"),
				AuthURL:               getEnv("OIDC_AUTH_URL", ""),
				TokenURL:              getEnv("OIDC_TOKEN_URL", ""),
				UserInfoURL:           getEnv("OIDC_USERINFO_URL", ""),
				Issuer:                getEnv("OIDC_ISSUER", ""),
			},
			Tokens: config.OIDCTokensConfig{
				SigningKey:           getEnv("JWT_SIGNING_KEY", "dev-jwt-secret-key"),
				SigningMethod:        getEnv("JWT_SIGNING_METHOD", "HS256"),
				AccessTokenDuration:  getEnv("JWT_ACCESS_DURATION", "15m"),
				RefreshTokenDuration: getEnv("JWT_REFRESH_DURATION", "24h"),
				IDTokenDuration:      getEnv("JWT_ID_DURATION", "1h"),
			},
			Scopes: []string{"openid", "profile", "email"},
		},

		Security: config.SecurityConfig{
			CORS: config.CORSConfig{
				AllowedOrigins: []string{
					"http://192.168.49.2:30090", // Frontend URL
					"http://localhost:3000",     // Local development
					"http://127.0.0.1:3000",     // Local development alternative
				},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{
					"Content-Type",
					"Authorization",
					"X-Requested-With",
					"Accept",
					"Origin",
					"Access-Control-Request-Method",
					"Access-Control-Request-Headers",
				},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			RateLimit: config.RateLimitConfig{
				Enabled:           false,
				RequestsPerMinute: 100,
				Burst:             10,
			},
			Session: config.SessionConfig{
				Secret:   getEnv("SESSION_SECRET", "dev-session-secret"),
				MaxAge:   3600,
				Secure:   false,
				HTTPOnly: true,
			},
		},

		Redis: config.RedisConfig{
			Enabled: false,
		},
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
