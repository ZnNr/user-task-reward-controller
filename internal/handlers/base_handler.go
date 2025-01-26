package handlers

import (
	"context"
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/service"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// UserIDResponse структура для возврата идентификатора пользователя
type UserIDResponse struct {
	Id int64 `json:"user_id"`
}

// Список маршрутов, которые не требуют авторизации
var noAuthRoutes = map[string]map[string]bool{
	"/auth/register": {http.MethodPost: true},
	"/auth/login":    {http.MethodPost: true},
}

// Проверяет, является ли маршрут свободным от авторизации
func isNoAuthRoute(path, method string) bool {
	for route, methods := range noAuthRoutes {
		if strings.HasPrefix(path, route) && methods[method] {
			return true
		}
	}
	return false
}

// JWTMiddleware создает middleware для проверки JWT токена
func JWTMiddleware(authService service.Auth, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "handlers.JWTMiddleware"
			logger := logger.With(zap.String("op", op))

			path := r.URL.Path
			method := r.Method

			// Проверяем маршрут в noAuthRoutes
			if isNoAuthRoute(path, method) {
				next.ServeHTTP(w, r)
				return
			}

			// Извлечение токена из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			tokenString := ""
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}

			// Если токен не найден в заголовке, проверяем куки
			if tokenString == "" {
				cookie, err := r.Cookie("token")
				if err == nil {
					tokenString = cookie.Value
				} else {
					logger.Info("JWTMiddleware: missing token from Authorization header and cookie", zap.String("path", path))
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			// Верификация токена
			userId, err := authService.ParseToken(tokenString)
			if err != nil {
				logger.Warn("JWTMiddleware: invalid token", zap.String("path", path), zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Добавляем userId в контекст
			ctx := context.WithValue(r.Context(), "userID", userId)
			r = r.WithContext(ctx)
			// Передаем управление дальше
			next.ServeHTTP(w, r)
		})
	}
}

// Handler структура для работы с HTTP-запросами
type Handler struct {
	Services *service.Service
	logger   *zap.Logger
}

// NewHandler создает новый экземпляр Handler
func NewHandler(services *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		Services: services,
		logger:   logger,
	}
}

// handleServiceError Унифицированная обработка ошибок сервиса
func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.IsBadRequest(err):
		h.httpError(w, errors.NewBadRequest(errors.ErrorMessage[errors.BadRequest], err))
	case errors.IsNotFound(err):
		h.httpError(w, errors.NewNotFound(errors.ErrorMessage[errors.NotFound], err))
	case errors.IsInvalidToken(err):
		h.httpError(w, errors.NewInvalidToken(errors.ErrorMessage[errors.InvalidToken], err))
	case errors.IsUnauthorized(err):
		h.httpError(w, errors.NewUnauthorized(errors.ErrorMessage[errors.Unauthorized], err))
	case errors.IsAlreadyExists(err):
		h.httpError(w, errors.NewAlreadyExists(errors.ErrorMessage[errors.AlreadyExists], err))
	default:
		h.httpError(w, errors.NewInternal(errors.ErrorMessage[errors.Internal], err))
	}
}

// httpError отправляет ошибку клиенту
func (h *Handler) httpError(w http.ResponseWriter, err *errors.Error) {
	h.logger.Error("Handling error", zap.Error(err))
	//var status int
	var errorResponse map[string]string
	switch err.Type {
	case errors.InvalidToken:
		errorResponse = map[string]string{"error": err.Message}
	case errors.Unauthorized:
		errorResponse = map[string]string{"error": err.Message}
	case errors.BadRequest:
		errorResponse = map[string]string{"error": err.Message}
	case errors.NotFound:
		errorResponse = map[string]string{"error": err.Message}
	default:
		errorResponse = map[string]string{"error": err.Message}
	}
	h.jsonResponse(w, err.Status(), errorResponse)
}

// jsonResponse отправляет JSON-ответ клиенту
func (h *Handler) jsonResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
