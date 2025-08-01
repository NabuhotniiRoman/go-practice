package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"go-practice/internal/build"
	"go-practice/internal/config"
)

// configureAction генерує конфігурацію з шаблону
func configureAction(c *cli.Context) error {
	templatePath := c.String("template")
	outputPath := c.String("output")
	version := c.String("version")
	mode := c.String("mode")

	fmt.Printf("🔧 Configuring OIDC API Server\n")
	fmt.Printf("Template: %s\n", templatePath)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Mode: %s\n", mode)

	// Використовуємо шляхи як є, якщо вони абсолютні
	templatePathAbs := templatePath
	outputPathAbs := outputPath

	if !filepath.IsAbs(templatePath) {
		workDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		templatePathAbs = filepath.Join(workDir, templatePath)
	}

	if !filepath.IsAbs(outputPath) {
		workDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		outputPathAbs = filepath.Join(workDir, outputPath)
	}

	// Перевіряємо що шаблон існує
	if _, err := os.Stat(templatePathAbs); os.IsNotExist(err) {
		return fmt.Errorf("template file does not exist: %s", templatePathAbs)
	}

	// Генеруємо конфігурацію з дефолтними значеннями та змінними оточення
	vars := getConfigVars(mode, version)

	if err := config.GenerateConfigFromTemplate(templatePathAbs, outputPathAbs, vars); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Printf("✅ Configuration generated successfully: %s\n", outputPathAbs)
	return nil
}

// serverAction запускає API сервер
func serverAction(c *cli.Context) error {
	configPath := c.String("config")

	fmt.Printf("🚀 Starting OIDC API Server\n")
	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Version: %s\n", build.Version)

	// Перевіряємо що конфіг існує
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s. Run 'configure' command first", configPath)
	}

	// Завантажуємо конфігурацію
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Запускаємо сервер
	return config.StartServer(cfg)
}

// versionAction показує інформацію про версію
func versionAction(c *cli.Context) error {
	info := build.Info()

	fmt.Printf("OIDC API Server\n")
	fmt.Printf("Version: %s\n", info["version"])
	fmt.Printf("Build Number: %s\n", info["number"])
	fmt.Printf("Git Commit: %s\n", info["git_commit"])
	fmt.Printf("Build Time: %s\n", info["build_time"])

	return nil
}

// getConfigVars повертає мапу змінних для конфігурації
func getConfigVars(mode, version string) map[string]interface{} {
	vars := map[string]interface{}{
		"build_version": version,
		"environment":   mode,
	}

	// Загальні змінні
	setVarFromEnv(vars, "api_server_host", "API_SERVER_HOST", "localhost")
	setVarFromEnv(vars, "api_server_port", "API_SERVER_PORT", 8080)
	setVarFromEnv(vars, "log_level", "LOG_LEVEL", getLogLevelForMode(mode))

	// База даних
	setVarFromEnv(vars, "db_host", "DB_HOST", "localhost")
	setVarFromEnv(vars, "db_port", "DB_PORT", 5432)
	setVarFromEnv(vars, "db_name", "DB_NAME", "go_practice")
	setVarFromEnv(vars, "db_user", "DB_USER", "oidc_api_user")
	setVarFromEnv(vars, "db_password", "DB_PASSWORD", "oidc_secure_password_2025")

	// OIDC
	setVarFromEnv(vars, "oidc_issuer_url", "OIDC_ISSUER_URL", "https://accounts.google.com")
	setVarFromEnv(vars, "oidc_client_id", "OIDC_CLIENT_ID", "your_client_id_here")
	setVarFromEnv(vars, "oidc_client_secret", "OIDC_CLIENT_SECRET", "your_client_secret_here")
	setVarFromEnv(vars, "oidc_auth_url", "OIDC_AUTH_URL", "https://accounts.google.com/o/oauth2/v2/auth")
	setVarFromEnv(vars, "oidc_token_url", "OIDC_TOKEN_URL", "https://oauth2.googleapis.com/token")
	setVarFromEnv(vars, "oidc_userinfo_url", "OIDC_USERINFO_URL", "https://openidconnect.googleapis.com/v1/userinfo")
	setVarFromEnv(vars, "oidc_issuer", "OIDC_ISSUER", "https://accounts.google.com")	// Безпека
	setVarFromEnv(vars, "jwt_signing_key", "JWT_SIGNING_KEY", "dev-jwt-secret-key-change-in-production")
	setVarFromEnv(vars, "session_secret", "SESSION_SECRET", "dev-session-secret-change-in-production")

	return vars
}

// setVarFromEnv встановлює змінну з оточення або дефолтне значення
func setVarFromEnv(vars map[string]interface{}, key, envKey string, defaultValue interface{}) {
	if envValue := os.Getenv(envKey); envValue != "" {
		vars[key] = envValue
	} else {
		vars[key] = defaultValue
	}
}

// getLogLevelForMode повертає рівень логування для режиму
func getLogLevelForMode(mode string) string {
	switch mode {
	case "production":
		return "warn"
	case "staging":
		return "info"
	default:
		return "debug"
	}
}
