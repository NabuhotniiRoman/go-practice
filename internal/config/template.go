package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// generateConfigWithVars генерує конфігурацію з шаблону з використанням змінних
func generateConfigWithVars(templatePath, outputPath string, vars map[string]interface{}) error {
	// Читаємо шаблон
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Обробляємо {{var}} теги в шаблоні
	processedContent := processVarTags(string(content), vars)

	// Створюємо template з додатковими функціями
	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"duration": func(d string) string {
			return d // Просто повертаємо рядок як є
		},
	}).Parse(processedContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Генеруємо конфігурацію
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
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

// processVarTags обробляє {{var "name" default_value required}} теги
func processVarTags(content string, vars map[string]interface{}) string {
	// Регулярний вираз для пошуку {{var "name" default_value required}} тегів
	varRegex := regexp.MustCompile(`\{\{var\s+"([^"]+)"\s+([^\s}]+)\s+(true|false)\s*\}\}`)

	return varRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := varRegex.FindStringSubmatch(match)
		if len(matches) != 4 {
			return match
		}

		varName := matches[1]
		defaultValue := matches[2]
		required := matches[3] == "true"

		// Перевіряємо чи є значення в змінних
		if value, exists := vars[varName]; exists {
			return formatValue(value)
		}

		// Якщо обов'язково і немає значення
		if required && (defaultValue == "" || defaultValue == `""`) {
			return `"REQUIRED_VALUE_NOT_SET"`
		}

		// Повертаємо дефолтне значення через formatValue для правильного форматування
		return formatValue(parseDefaultValue(defaultValue))
	})
}

// formatValue форматує значення для HCL
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Якщо це список через кому, обробляємо як масив
		if strings.Contains(v, ",") {
			parts := strings.Split(v, ",")
			var quoted []string
			for _, part := range parts {
				quoted = append(quoted, fmt.Sprintf(`"%s"`, strings.TrimSpace(part)))
			}
			return strings.Join(quoted, ",\n      ")
		}
		return fmt.Sprintf(`"%s"`, v)
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}

// parseDefaultValue парсить дефолтне значення з template
func parseDefaultValue(defaultValue string) interface{} {
	// Видаляємо лапки якщо є
	if strings.HasPrefix(defaultValue, `"`) && strings.HasSuffix(defaultValue, `"`) {
		return strings.Trim(defaultValue, `"`)
	}

	// Спробуємо парсити як число
	if intVal, err := strconv.Atoi(defaultValue); err == nil {
		return intVal
	}

	// Спробуємо парсити як float
	if floatVal, err := strconv.ParseFloat(defaultValue, 64); err == nil {
		return floatVal
	}

	// Спробуємо парсити як bool
	if boolVal, err := strconv.ParseBool(defaultValue); err == nil {
		return boolVal
	}

	// Повертаємо як рядок
	return defaultValue
}
