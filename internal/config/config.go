package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-practice/internal/handlers"
	"go-practice/internal/middleware"
	"go-practice/internal/services"
	"go-practice/migrations"

	_ "go-practice/docs"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config представляє повну конфігурацію додатку
type Config struct {
	Server   ServerConfig   `hcl:"server,block"`
	Database DatabaseConfig `hcl:"database,block"`
	OIDC     OIDCConfig     `hcl:"oidc,block"`
	Security SecurityConfig `hcl:"security,block"`
	Redis    RedisConfig    `hcl:"redis,block"`
}

// ServerConfig містить налаштування HTTP сервера
type ServerConfig struct {
	Host         string `hcl:"host"`
	Port         int    `hcl:"port"`
	Environment  string `hcl:"environment"`
	LogLevel     string `hcl:"log_level"`
	LogFormat    string `hcl:"log_format"`
	ReadTimeout  string `hcl:"read_timeout"`
	WriteTimeout string `hcl:"write_timeout"`
	IdleTimeout  string `hcl:"idle_timeout"`
}

// DatabaseConfig містить налаштування бази даних
type DatabaseConfig struct {
	Driver                string `hcl:"driver"`
	Host                  string `hcl:"host"`
	Port                  int    `hcl:"port"`
	Name                  string `hcl:"name"`
	User                  string `hcl:"user"`
	Password              string `hcl:"password"`
	SSLMode               string `hcl:"ssl_mode"`
	MaxOpenConnections    int    `hcl:"max_open_connections"`
	MaxIdleConnections    int    `hcl:"max_idle_connections"`
	ConnectionMaxLifetime string `hcl:"connection_max_lifetime"`
}

// OIDCConfig містить налаштування OpenID Connect
type OIDCConfig struct {
	Provider OIDCProviderConfig `hcl:"provider,block"`
	Tokens   OIDCTokensConfig   `hcl:"tokens,block"`
	Scopes   []string           `hcl:"scopes"`
}

// OIDCProviderConfig містить налаштування OIDC провайдера
type OIDCProviderConfig struct {
	IssuerURL             string `hcl:"issuer_url"`
	ClientID              string `hcl:"client_id"`
	ClientSecret          string `hcl:"client_secret"`
	RedirectURL           string `hcl:"redirect_url"`
	PostLogoutRedirectURL string `hcl:"post_logout_redirect_url"`
	AuthURL               string `hcl:"auth_url"`
	TokenURL              string `hcl:"token_url"`
	UserInfoURL           string `hcl:"userinfo_url"`
	Issuer                string `hcl:"issuer"`
}

// OIDCTokensConfig містить налаштування токенів
type OIDCTokensConfig struct {
	SigningKey           string `hcl:"signing_key"`
	SigningMethod        string `hcl:"signing_method"`
	AccessTokenDuration  string `hcl:"access_token_duration"`
	RefreshTokenDuration string `hcl:"refresh_token_duration"`
	IDTokenDuration      string `hcl:"id_token_duration"`
}

// SecurityConfig містить налаштування безпеки
type SecurityConfig struct {
	CORS      CORSConfig      `hcl:"cors,block"`
	RateLimit RateLimitConfig `hcl:"rate_limit,block"`
	Session   SessionConfig   `hcl:"session,block"`
}

// CORSConfig містить налаштування CORS
type CORSConfig struct {
	AllowedOrigins   []string `hcl:"allowed_origins"`
	AllowedMethods   []string `hcl:"allowed_methods"`
	AllowedHeaders   []string `hcl:"allowed_headers"`
	AllowCredentials bool     `hcl:"allow_credentials"`
	MaxAge           int      `hcl:"max_age"`
}

// RateLimitConfig містить налаштування rate limiting
type RateLimitConfig struct {
	Enabled           bool `hcl:"enabled"`
	RequestsPerMinute int  `hcl:"requests_per_minute"`
	Burst             int  `hcl:"burst"`
}

// SessionConfig містить налаштування сесій
type SessionConfig struct {
	Secret   string `hcl:"secret"`
	MaxAge   int    `hcl:"max_age"`
	Secure   bool   `hcl:"secure"`
	HTTPOnly bool   `hcl:"http_only"`
}

// RedisConfig містить налаштування Redis
type RedisConfig struct {
	Enabled    bool   `hcl:"enabled"`
	Host       string `hcl:"host"`
	Port       int    `hcl:"port"`
	Password   string `hcl:"password"`
	Database   int    `hcl:"database"`
	MaxRetries int    `hcl:"max_retries"`
	PoolSize   int    `hcl:"pool_size"`
}

// LoadConfig завантажує конфігурацію з HCL файлу
func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	var config Config
	err := hclsimple.DecodeFile(configPath, nil, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Валідація конфігурації
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Validate перевіряє валідність конфігурації
func (c *Config) Validate() error {
	// Перевірка обов'язкових полів сервера
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Перевірка бази даних
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	// Перевірка OIDC (якщо використовується)
	if c.OIDC.Provider.IssuerURL != "" {
		if c.OIDC.Provider.ClientID == "" {
			return fmt.Errorf("OIDC client ID is required when issuer URL is set")
		}
		if c.OIDC.Provider.ClientSecret == "" {
			return fmt.Errorf("OIDC client secret is required when issuer URL is set")
		}
	}

	// Перевірка секрету сесії
	if c.Security.Session.Secret == "" {
		return fmt.Errorf("session secret is required")
	}

	return nil
}

// GetAddress повертає адресу для прослуховування сервера
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDatabaseDSN повертає DSN для підключення до бази даних
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// IsDevelopment перевіряє чи додаток працює в режимі розробки
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction перевіряє чи додаток працює в продакшн режимі
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// GenerateConfigFromTemplate генерує HCL конфігурацію з шаблону використовуючи змінні
func GenerateConfigFromTemplate(templatePath, outputPath string, vars map[string]interface{}) error {
	return generateConfigWithVars(templatePath, outputPath, vars)
}

// StartServer запускає HTTP сервер з конфігурацією
func StartServer(cfg *Config) error {
	// Налаштування логування
	setupLogging(cfg)

	// Підключення до бази даних
	db, err := connectToDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Налаштування Gin режиму
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Створення Gin роутера
	r := gin.New()

	// Swagger UI route
	// Імпорти: _ "go-practice/docs", ginSwagger "github.com/swaggo/gin-swagger", swaggerFiles "github.com/swaggo/files"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Додавання middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg))

	// Реєстрація routes (передаємо db для використання в handlers)
	setupRoutes(r, cfg, db)

	// Парсинг таймаутів
	readTimeout, err := time.ParseDuration(cfg.Server.ReadTimeout)
	if err != nil {
		logrus.Warnf("Invalid read timeout, using default: %v", err)
		readTimeout = 30 * time.Second
	}

	writeTimeout, err := time.ParseDuration(cfg.Server.WriteTimeout)
	if err != nil {
		logrus.Warnf("Invalid write timeout, using default: %v", err)
		writeTimeout = 30 * time.Second
	}

	idleTimeout, err := time.ParseDuration(cfg.Server.IdleTimeout)
	if err != nil {
		logrus.Warnf("Invalid idle timeout, using default: %v", err)
		idleTimeout = 120 * time.Second
	}

	// Створення HTTP сервера
	srv := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Канал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера в goroutine
	go func() {
		logrus.Infof("🚀 Starting OIDC API Server on %s", cfg.GetAddress())
		logrus.Infof("Environment: %s", cfg.Server.Environment)
		logrus.Infof("Log Level: %s", cfg.Server.LogLevel)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Очікування сигналу для graceful shutdown
	<-quit
	logrus.Info("🛑 Shutting down server...")

	// Graceful shutdown з таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
		return err
	}

	logrus.Info("✅ Server exited gracefully")
	return nil
}

// setupLogging налаштовує логування
func setupLogging(cfg *Config) {
	// Встановлення рівня логування
	level, err := logrus.ParseLevel(cfg.Server.LogLevel)
	if err != nil {
		logrus.Warnf("Invalid log level '%s', using info", cfg.Server.LogLevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Встановлення формату логування
	if cfg.Server.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

// corsMiddleware налаштовує CORS middleware
func corsMiddleware(cfg *Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Встановлення CORS заголовків
		origin := c.Request.Header.Get("Origin")
		if isAllowedOrigin(origin, cfg.Security.CORS.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.Security.CORS.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.Security.CORS.AllowedHeaders, ", "))

		if cfg.Security.CORS.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if cfg.Security.CORS.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.Security.CORS.MaxAge))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// setupRoutes налаштовує маршрути
func setupRoutes(r *gin.Engine, cfg *Config, db *gorm.DB) {
	// Ініціалізуємо сервіси
	userService := services.NewUserService(db)

	// Створюємо JWT сервіс з секретами з конфігурації
	jwtService := services.NewJWTService(
		cfg.OIDC.Tokens.SigningKey+"_access",
		cfg.OIDC.Tokens.SigningKey+"_id",
		cfg.OIDC.Tokens.SigningKey+"_refresh",
	)

	// Створюємо State сервіс для CSRF захисту (TTL 10 хвилин)
	stateService := services.NewStateService(10 * time.Minute)

	// Створюємо OIDC Provider сервіс для роботи з зовнішнім провайдером
	oidcProviderService := services.NewOIDCProviderService(
		cfg.OIDC.Provider.ClientID,
		cfg.OIDC.Provider.ClientSecret,
		cfg.OIDC.Provider.TokenURL,
		cfg.OIDC.Provider.UserInfoURL,
		cfg.OIDC.Provider.Issuer,
	)

	// Створюємо Session Manager для відстеження сесій (TTL 1 година)
	sessionManager := services.NewSessionManager(1 * time.Hour)

	// Створюємо Auth сервіс який об'єднує всі інші сервіси
	authService := services.NewAuthService(userService, jwtService, stateService, oidcProviderService, sessionManager)

	// Ініціалізуємо handlers з усіма сервісами
	authHandler := handlers.NewAuthHandler(authService, cfg.OIDC.Provider.PostLogoutRedirectURL) // Передаємо postLogoutRedirectURL з конфігурації
	apiHandler := handlers.NewAPIHandler(userService)                                            // Health endpoint з інформацією про базу даних
	r.GET("/health", func(c *gin.Context) {
		// Перевірка підключення до БД
		sqlDB, err := db.DB()
		dbStatus := "healthy"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "unhealthy"
		}

		c.JSON(200, gin.H{
			"status":   "healthy",
			"service":  "oidc-api-server",
			"version":  "dev",
			"database": dbStatus,
		})
	})

	// Додаємо дублікат health endpoint для API маршруту
	r.GET("/api/health", func(c *gin.Context) {
		// Перевірка підключення до БД
		sqlDB, err := db.DB()
		dbStatus := "healthy"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "unhealthy"
		}

		c.JSON(200, gin.H{
			"status":   "healthy",
			"service":  "oidc-api-server",
			"version":  "dev",
			"database": dbStatus,
		})
	})

	// API group
	api := r.Group("/api/v1")
	{
		// Public endpoints
		api.GET("/public", apiHandler.PublicData)

		// Protected endpoints з middleware аутентифікації
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtService, userService))
		{
			protected.GET("/protected", apiHandler.ProtectedData)
			protected.GET("/profile", apiHandler.UserProfile)
			protected.PUT("/profile", apiHandler.UpdateProfile)
			protected.GET("/user-data", apiHandler.UserData)
			protected.GET("/users", apiHandler.Users)
			protected.GET("/users/:id", apiHandler.GetUserByID)
			protected.POST("/users/search", apiHandler.SearchUsers)
		}

		// Database test endpoint
		api.GET("/db-test", func(c *gin.Context) {
			// Простий тест підключення до БД
			var result struct {
				Version string
				Now     time.Time
			}

			if err := db.Raw("SELECT version() as version, now() as now").Scan(&result).Error; err != nil {
				c.JSON(500, gin.H{
					"error":   "Database query failed",
					"details": err.Error(),
				})
				return
			}

			c.JSON(200, gin.H{
				"message":          "Database connection successful",
				"database_version": result.Version,
				"server_time":      result.Now,
			})
		})
	}

	// OIDC endpoints
	oidc := r.Group("/auth")
	{
		oidc.POST("/default/login", authHandler.DefaultLogin)
		oidc.POST("/login", authHandler.Login)       // Resource Owner Password Grant
		oidc.GET("/callback", authHandler.Callback)  // Authorization Code Flow callback
		oidc.POST("/logout", authHandler.Logout)     // End Session
		oidc.POST("/refresh", authHandler.Refresh)   // Token Refresh
		oidc.GET("/userinfo", authHandler.UserInfo)  // UserInfo endpoint
		oidc.POST("/register", authHandler.Register) // User Registration
	}
}

// Helper functions
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// connectToDatabase підключається до PostgreSQL бази даних через GORM
func connectToDatabase(cfg *Config) (*gorm.DB, error) {
	dsn := cfg.GetDatabaseDSN()
	logrus.Infof("🔌 Connecting to PostgreSQL database: %s@%s:%d/%s",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	// Налаштування GORM конфігурації
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// В debug режимі включаємо логування SQL запитів
	if cfg.IsDevelopment() {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	// Підключення до бази даних
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Отримання sqlDB для налаштування connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Парсинг connection max lifetime
	connectionMaxLifetime, err := time.ParseDuration(cfg.Database.ConnectionMaxLifetime)
	if err != nil {
		logrus.Warnf("Invalid connection max lifetime, using default 5m: %v", err)
		connectionMaxLifetime = 5 * time.Minute
	}

	// Налаштування connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(connectionMaxLifetime)

	// Тест підключення
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Infof("📊 Database connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		cfg.Database.MaxOpenConnections, cfg.Database.MaxIdleConnections, connectionMaxLifetime)

	// Автоматична міграція тільки для моделей, які мають GORM-структури
	logrus.Info("🛠️  Running AutoMigrate for User and Friendship...")
	if err := db.AutoMigrate(
		&services.User{},
		&migrations.Friendship{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	logrus.Info("✅ Database connection established and migrated")
	return db, nil
}

// RunMigrations виконує тільки міграції без запуску сервера
func RunMigrations(cfg *Config) error {
	dsn := cfg.GetDatabaseDSN()
	logrus.Infof("🔌 Connecting to PostgreSQL database for migrations: %s@%s:%d/%s",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	// Налаштування GORM конфігурації
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// В debug режимі включаємо логування SQL запитів
	if cfg.IsDevelopment() {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	// Підключення до бази даних
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Отримання sqlDB для тестування підключення
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Тест підключення
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("🛠️  Running migrations for new tables only...")

	// Перевіряємо чи існує таблиця friendships
	var exists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'friendships')").Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check if friendships table exists: %w", err)
	}

	if !exists {
		logrus.Info("Creating friendships table...")
		// Створюємо тільки таблицю friendships, не чіпаємо users
		if err := db.AutoMigrate(&migrations.Friendship{}); err != nil {
			return fmt.Errorf("failed to create friendships table: %w", err)
		}
		logrus.Info("✅ Friendships table created successfully")
	} else {
		logrus.Info("Friendships table already exists, skipping...")
	}

	logrus.Info("✅ Database migrations completed successfully")

	// Закриваємо з'єднання
	sqlDB.Close()

	return nil
}
