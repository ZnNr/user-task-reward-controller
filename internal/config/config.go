package config

import (
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения, включая настройки базы данных и сервера.
type Config struct {
	DBHost     string // Хост базы данных
	DBPort     string // Порт базы данных
	DBUser     string // Пользователь базы данных
	DBPassword string // Пароль базы данных
	DBName     string // Имя базы данных
	ServerPort string // Порт сервера приложения
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadConfig инициализирует конфигурацию из переменных окружения с значениями по умолчанию.
func LoadConfig() (*Config, error) {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "user_reward_db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}, nil
}

// GetDBConnString формирует строку подключения к базе данных.
func (c *Config) GetDBConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

// getEnv возвращает значение переменной окружения или значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate проверяет, что важные параметры конфигурации заполнены.
func (c *Config) Validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DBHost cannot be empty")
	}
	if c.DBPort == "" {
		return fmt.Errorf("DBPort cannot be empty")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DBUser cannot be empty")
	}
	if c.DBName == "" {
		return fmt.Errorf("DBName cannot be empty")
	}
	if c.ServerPort == "" {
		return fmt.Errorf("ServerPort cannot be empty")
	}
	return nil
}
