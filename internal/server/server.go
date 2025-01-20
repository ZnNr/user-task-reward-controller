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
	schema     = "migration/000001_init_schema.up.sql"
	jwtSignKey = "joiQWRtaW4iLCJJc3N1ZXIiOiJJc3N1ZXIiLCJVc2VybmFtZSI"
	tokenTTL   = 120 * time.Minute
	hasherSalt = "Jc3N1ZXI"
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
	connStr := a.config.GetDBConnString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	a.db = db
	return nil
}

// runMigrations запускает миграции базы данных
func (a *App) runMigrations() error {
	driver, err := postgres.WithInstance(a.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migration",
		"postgres", driver)
	if err != nil {
		a.logger.Error("Failed to create migration instance", zap.Error(err))
		// Продолжим работу, даже если есть проблемы с миграциями
		return nil
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		a.logger.Error("Failed to apply migrations", zap.Error(err))
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	a.logger.Info("Migrations applied successfully")
	return nil
}

// initHTTPServer инициализирует HTTP сервер
func (a *App) initHTTPServer() error {
	// Инициализируем репозитории
	repos := repository.NewRepositories(a.db)

	// Инициализируем сервисы - остановился здесь
	services := service.NewService(service.ServicesDependencies{
		Repos:    repos,
		Logger:   a.logger,
		SignKey:  jwtSignKey, // Используем переданный jwt_sign_key
		TokenTTL: tokenTTL,   // Используем переданный token_ttl
		Salt:     hasherSalt, // Используем переданный hashes_salt

	})

	// Создаем обработчики
	handler := handlers.NewHandler(services, a.logger)

	// Создаем роутер и добавляем маршруты для всех обработчиков
	r := router.NewRouter(handler, a.logger) // Импортируйте новый роутер без хендлеров

	// Создаем HTTP сервер
	a.httpServer = &http.Server{
		Addr:         ":" + a.config.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return nil
}

// Run запуск приложения
func (a *App) Run() error {
	a.logger.Info("Starting server", zap.String("port", a.config.ServerPort))
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Shutdown gracefully останавливает приложение
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down server...")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
