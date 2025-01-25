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
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Настраиваем маршруты для аутентификации
	setupAuthRoutes(authRouter, handler)

	// Настраиваем маршруты для API и добавляем middleware для авторизации к маршрутам
	setupAPIRoutes(apiRouter, handler)

	// Применяем JWT Middleware только к защищенным маршрутам
	protectedAPIRouter := apiRouter.PathPrefix("").Subrouter()
	protectedAPIRouter.Use(handlers.JWTMiddleware(handler.Services.Auth, logger))
	setupProtectedAPIRoutes(protectedAPIRouter, handler)

	return r
}

// setupAuthRoutes настраивает маршруты для аутентификации
func setupAuthRoutes(router *mux.Router, handler *handlers.Handler) {

	/*
			curl -X POST "http://localhost:8080/auth/register" \
			-H "Content-Type: application/json" \
			-d '{
			"username": "john_doe",
				"password": "securepassword123",
				"email": "john.doe@example.com"
		}'
	*/
	//Пример успешного ответа
	/*
		{
		  "message": "User registered successfully",
		  "user_id": 123
		}
	*/
	router.HandleFunc("/register", handler.RegisterHandler).Methods("POST")

	/*
		curl -X POST "http://localhost:8080/auth/login" \
		-H "Content-Type: application/json" \
		-d '{
		  "username": "john_doe",
		  "password": "securepassword123"
		}'
	*/
	//Пример ответа от сервера

	/*
		{
		  "message": "Login successful",
		  "token": "your_jwt_token_here"
		}
	*/

	router.HandleFunc("/login", handler.LoginHandler).Methods("POST")

}

// setupAPIRoutes настраивает общие маршруты для API
func setupAPIRoutes(router *mux.Router, handler *handlers.Handler) {
	// Регистрируем маршруты, которые могут быть открытыми или имеют специальные условия
	router.HandleFunc("/task/all", handler.TaskGetAll).Methods("GET") // Предположим, что этот маршрут открыт для всех
}

// setupProtectedAPIRoutes настраивает защищенные маршруты для API
func setupProtectedAPIRoutes(router *mux.Router, handler *handlers.Handler) {
	// Регистрируем маршруты для задач

	/*
			curl -X POST "http://localhost:8080/api/task/create" \
			-H "Content-Type: application/json" \
			-d '{
			"title": "New Task",
				"description": "This is a new task description.",
				"price": 50
		}'
	*/

	router.HandleFunc("/task/create", handler.TaskCreate).Methods("POST")
	//curl -X GET "http://localhost:8080/api/task/all"
	router.HandleFunc("/task/all", handler.TaskGetAll).Methods("GET")
	/*
			curl -X POST "http://localhost:8080/api/task/123/complete" \
			-H "Content-Type: application/json" \
			-d '{
			"task_id": 456
		}'
	*/

	router.HandleFunc("/task/{user_id}/complete", handler.TaskComplete).Methods("POST")
	// Регистрируем маршруты для пользователей
	/*

		curl -X POST "http://localhost:8080/api/users/123/refferer" \
		-H "Content-Type: application/json" \
		-d '{
		  "ref_code": "ABC123"
		}'
	*/

	router.HandleFunc("/users/{user_id}/refferer", handler.UserReferrerCode).Methods("POST")

	//curl -X GET "http://localhost:8080/api/users/123/status"
	router.HandleFunc("/users/{user_id}/status", handler.UserInfo).Methods("GET")
	//curl -X GET "http://localhost:8080/api/users/leaderboard"
	router.HandleFunc("/users/leaderboard", handler.UsersLeaderboard).Methods("GET")

	//примеры запросов
	//curl -X GET "http://localhost:8080/auth/public/john_doe"
	//curl -X GET "http://localhost:8080/auth/public/example@example.com"

	//пример ответа {"user_id": 123 }
	router.HandleFunc("/users/{username_or_email}", handler.GetUserIDbyUsernameOrEmailHandler).Methods("GET")
}
