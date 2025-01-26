package server

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	//"errors"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/config"
	"github.com/ZnNr/user-task-reward-controller/internal/handlers"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"github.com/ZnNr/user-task-reward-controller/internal/router"
	"github.com/ZnNr/user-task-reward-controller/internal/service"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
)

const (
	jwtSignKey = "joiQWRtaW4iLCJJc3N1ZXIiOiJJc3N1ZXIiLCJVc2VybmFtZSI"
	tokenTTL   = 120 * time.Minute
)

// App структура приложения
type App struct {
	config     *config.Config
	logger     *zap.Logger
	db         *sql.DB
	httpServer *http.Server
}

// New конструктор нового экземпляра приложения
func New(cfg *config.Config, logger *zap.Logger) *App {
	return &App{
		config: cfg,
		logger: logger,
	}
}

// Initialize инициализирует компоненты приложения
func (a *App) Initialize() error {
	if err := a.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := a.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := a.initHTTPServer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	return nil
}

// initDatabase инициализирует подключение к базе данных
func (a *App) initDatabase() error {
	const op = "server.App.initDatabase"
	logger := a.logger.With(zap.String("op", op))

	connStr := a.config.GetDBConnString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to open database connection", zap.Error(err))
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	a.db = db

	// Проверяем соединение с базой данных
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully")
	return nil
}

// runMigrations запускает миграции базы данных
func (a *App) runMigrations() error {
	const op = "server.App.runMigrations"
	logger := a.logger.With(zap.String("op", op))

	driver, err := postgres.WithInstance(a.db, &postgres.Config{})
	if err != nil {
		logger.Error("Failed to create database driver", zap.Error(err))
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migration",
		"postgres",
		driver,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error("Failed to apply migrations", zap.Error(err))
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("Migrations applied successfully")
	return nil
}

// initHTTPServer инициализирует HTTP сервер
func (a *App) initHTTPServer() error {
	const op = "server.App.initHTTPServer"
	logger := a.logger.With(zap.String("op", op))

	// Инициализируем репозитории
	repos := repository.NewRepositories(a.db, a.logger)

	// Инициализируем сервисы
	services := service.NewService(service.ServicesDependencies{
		Repos:    repos,
		Logger:   a.logger,
		SignKey:  jwtSignKey,
		TokenTTL: tokenTTL,
	})

	// Создаем обработчики
	handler := handlers.NewHandler(services, a.logger)

	// Создаем роутер и добавляем маршруты для всех обработчиков
	router := router.NewRouter(handler, a.logger)

	// Создаем HTTP сервер
	a.httpServer = &http.Server{
		Addr:         ":" + a.config.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("HTTP server initialized successfully")
	return nil
}

// Run запуск приложения
func (a *App) Run() error {
	const op = "server.App.Run"
	logger := a.logger.With(zap.String("op", op))

	logger.Info("Starting server", zap.String("port", a.config.ServerPort))
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Failed to start server", zap.Error(err))
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully останавливает приложение
func (a *App) Shutdown(ctx context.Context) error {
	const op = "server.App.Shutdown"
	logger := a.logger.With(zap.String("op", op))

	logger.Info("Shutting down server...")
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server", zap.Error(err))
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	if err := a.db.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	logger.Info("Server stopped gracefully")
	return nil
}
