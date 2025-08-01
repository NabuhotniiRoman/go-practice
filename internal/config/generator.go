package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ConfigData містить дані для генерації конфігурації з шаблону
type ConfigData struct {
	Server   ServerConfigData   `json:"server"`
	Database DatabaseConfigData `json:"database"`
	OIDC     OIDCConfigData     `json:"oidc"`
	Security SecurityConfigData `json:"security"`
	Redis    RedisConfigData    `json:"redis"`
}

// ServerConfigData містить дані для налаштування сервера
type ServerConfigData struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Environment  string `json:"environment"`
	LogLevel     string `json:"log_level"`
	LogFormat    string `json:"log_format"`
	ReadTimeout  string `json:"read_timeout"`
	WriteTimeout string `json:"write_timeout"`
	IdleTimeout  string `json:"idle_timeout"`
}

// DatabaseConfigData містить дані для налаштування бази даних
type DatabaseConfigData struct {
	Driver                string `json:"driver"`
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	Name                  string `json:"name"`
	User                  string `json:"user"`
	Password              string `json:"password"`
	SSLMode               string `json:"ssl_mode"`
	MaxOpenConnections    int    `json:"max_open_connections"`
	MaxIdleConnections    int    `json:"max_idle_connections"`
	ConnectionMaxLifetime string `json:"connection_max_lifetime"`
}

// OIDCConfigData містить дані для налаштування OIDC
type OIDCConfigData struct {
	Provider OIDCProviderConfigData `json:"provider"`
	Tokens   OIDCTokensConfigData   `json:"tokens"`
	Scopes   []string               `json:"scopes"`
}

// OIDCProviderConfigData містить дані для налаштування OIDC провайдера
type OIDCProviderConfigData struct {
	IssuerURL             string `json:"issuer_url"`
	ClientID              string `json:"client_id"`
	ClientSecret          string `json:"client_secret"`
	RedirectURL           string `json:"redirect_url"`
	PostLogoutRedirectURL string `json:"post_logout_redirect_url"`
}

// OIDCTokensConfigData містить дані для налаштування токенів
type OIDCTokensConfigData struct {
	SigningKey           string `json:"signing_key"`
	SigningMethod        string `json:"signing_method"`
	AccessTokenDuration  string `json:"access_token_duration"`
	RefreshTokenDuration string `json:"refresh_token_duration"`
	IDTokenDuration      string `json:"id_token_duration"`
}

// SecurityConfigData містить дані для налаштування безпеки
type SecurityConfigData struct {
	CORS      CORSConfigData      `json:"cors"`
	RateLimit RateLimitConfigData `json:"rate_limit"`
	Session   SessionConfigData   `json:"session"`
}

// CORSConfigData містить дані для налаштування CORS
type CORSConfigData struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// RateLimitConfigData містить дані для налаштування rate limiting
type RateLimitConfigData struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
	Burst             int  `json:"burst"`
}

// SessionConfigData містить дані для налаштування сесій
type SessionConfigData struct {
	Secret   string `json:"secret"`
	MaxAge   int    `json:"max_age"`
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"http_only"`
}

// RedisConfigData містить дані для налаштування Redis
type RedisConfigData struct {
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Password   string `json:"password"`
	Database   int    `json:"database"`
	MaxRetries int    `json:"max_retries"`
	PoolSize   int    `json:"pool_size"`
}

// GenerateConfig генерує HCL конфігурацію з шаблону
func GenerateConfig(templatePath, outputPath string, data ConfigData) error {
	// Читаємо шаблон
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Створюємо template з додатковими функціями
	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" || value == 0 {
				return defaultValue
			}
			return value
		},
	}).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Генеруємо конфігурацію
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Створюємо директорію якщо не існує
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Записуємо результат
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfigData завантажує дані конфігурації з JSON файлу
func LoadConfigData(dataPath string) (*ConfigData, error) {
	content, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config data file: %w", err)
	}

	var data ConfigData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	return &data, nil
}

// GetDefaultConfigData повертає дефолтні дані конфігурації
func GetDefaultConfigData() ConfigData {
	return ConfigData{
		Server: ServerConfigData{
			Host:         "localhost",
			Port:         8080,
			Environment:  "development",
			LogLevel:     "info",
			LogFormat:    "json",
			ReadTimeout:  "30s",
			WriteTimeout: "30s",
			IdleTimeout:  "120s",
		},
		Database: DatabaseConfigData{
			Driver:                "postgres",
			Host:                  "localhost",
			Port:                  5432,
			Name:                  "go_practice",
			User:                  "oidc_api_user",
			Password:              "",
			SSLMode:               "disable",
			MaxOpenConnections:    25,
			MaxIdleConnections:    5,
			ConnectionMaxLifetime: "5m",
		},
		OIDC: OIDCConfigData{
			Provider: OIDCProviderConfigData{
				IssuerURL:             "",
				ClientID:              "",
				ClientSecret:          "",
				RedirectURL:           "http://localhost:8080/auth/callback",
				PostLogoutRedirectURL: "http://localhost:8080",
			},
			Tokens: OIDCTokensConfigData{
				SigningKey:           "",
				SigningMethod:        "HS256",
				AccessTokenDuration:  "1h",
				RefreshTokenDuration: "24h",
				IDTokenDuration:      "1h",
			},
			Scopes: []string{"openid", "profile", "email"},
		},
		Security: SecurityConfigData{
			CORS: CORSConfigData{
				AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			RateLimit: RateLimitConfigData{
				Enabled:           true,
				RequestsPerMinute: 60,
				Burst:             10,
			},
			Session: SessionConfigData{
				Secret:   "",
				MaxAge:   3600,
				Secure:   false,
				HTTPOnly: true,
			},
		},
		Redis: RedisConfigData{
			Enabled:    false,
			Host:       "localhost",
			Port:       6379,
			Password:   "",
			Database:   0,
			MaxRetries: 3,
			PoolSize:   10,
		},
	}
}
