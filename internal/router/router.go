package router

import (
	"github.com/ZnNr/user-task-reward-controller/internal/handlers"
	"github.com/ZnNr/user-task-reward-controller/internal/logging"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewRouter создает и конфигурирует новый маршрутизатор для обработки HTTP-запросов.
func NewRouter(
	handler *handlers.Handler,
	logger *zap.Logger,
) *mux.Router {
	// Создаем новый роутер
	r := mux.NewRouter()

	// Добавляем middleware для логирования
	r.Use(logging.LoggingMiddleware(logger))

	// Создаем подмаршрутизатор для группировки маршрутов аутентификации
	authRouter := r.PathPrefix("/auth").Subrouter()

	// Настраиваем маршруты для аутентификации
	setupAuthRoutes(authRouter, handler)

	// Добавляем middleware для авторизации к маршрутам
	r.Use(handlers.JWTMiddleware(handler.Services.Auth))

	// Настраиваем маршруты для задач и пользователей
	setupAPIRoutes(r, handler)

	return r
}

// setupAuthRoutes настраивает маршруты для аутентификации
func setupAuthRoutes(router *mux.Router, handler *handlers.Handler) {
	router.HandleFunc("/register", handler.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", handler.LoginHandler).Methods("POST")
}

// setupAPIRoutes настраивает маршруты для API
func setupAPIRoutes(router *mux.Router, handler *handlers.Handler) {
	// Регистрируем маршруты для задач
	router.HandleFunc("/task/create", handler.TaskCreate).Methods("POST")
	router.HandleFunc("/task/all", handler.TaskGetAll).Methods("GET")

	// Регистрируем маршруты для пользователей
	router.HandleFunc("/users/{id}/status", handler.UserInfo).Methods("GET")
}
