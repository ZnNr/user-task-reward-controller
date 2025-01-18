package main

import (
	"context"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/config"
	"github.com/ZnNr/user-task-reward-controller/internal/server"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Загружаем переменные окружения из файла .env
	if err := loadEnv(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Инициализируем логгер
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer syncLogger(logger)

	// Загружаем конфигурацию приложения
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Создаем и инициализируем приложение
	application := server.New(cfg, logger)
	if err := application.Initialize(); err != nil {
		logger.Fatal("Failed to initialize application", zap.Error(err))
	}

	// Запускаем приложение в горутине
	go runApplication(application, logger)

	// Ожидаем сигнал для graceful shutdown
	waitForShutdown(application, logger)
}

// loadEnv загружает переменные окружения из файла .env.
func loadEnv() error {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return fmt.Errorf(".env file does not exist")
	}
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// вывод переменных .env для отладки
	fmt.Println("Environment variables loaded:")
	fmt.Printf("DB_USER: %s\n", os.Getenv("DB_USER"))
	fmt.Printf("DB_PASSWORD: %s\n", os.Getenv("DB_PASSWORD"))

	return nil
}

// initLogger инициализирует логгер.
func initLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

// syncLogger синхронизирует логгер для безопасного завершения работы.
func syncLogger(logger *zap.Logger) {
	if err := logger.Sync(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
	}
}

// runApplication запускает приложение и обрабатывает ошибки.
func runApplication(application *server.App, logger *zap.Logger) {
	if err := application.Run(); err != nil {
		logger.Fatal("Failed to run application", zap.Error(err))
	}
}

// waitForShutdown ожидает сигнал завершения и выполняет graceful shutdown.
func waitForShutdown(application *server.App, logger *zap.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		logger.Fatal("Failed to shutdown application", zap.Error(err))
	}

	logger.Info("Application stopped")
}
