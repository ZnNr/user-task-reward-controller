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

// Список маршрутов, которые не требуют авторизации
var noAuthRoutes = map[string][]string{
	"/auth/register":    {"POST"},
	"/auth/login":       {"POST"},
	"/public/{user_id}": {"GET"}, // С параметрами (например, /public/123), обработаем с помощью префиксов.
}

// Проверяет, является ли маршрут свободным от авторизации
func isNoAuthRoute(path, method string) bool {
	for route, methods := range noAuthRoutes {
		if path == strings.TrimSuffix(route, "{user_id}") {
			for _, m := range methods {
				if m == method {
					return true
				}
			}
		}

		// Для роутов с переменной частью, например "/public/{user_id}"
		if strings.HasPrefix(route, "/public/") && strings.HasPrefix(path, "/public/") {
			for _, m := range methods {
				if m == method {
					return true
				}
			}
		}
	}
	return false
}

func JWTMiddleware(authService service.Auth) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					//logger.Info("JWTMiddleware: missing token from Authorization header and cookie", zap.String("path", path))
					http.Error(w, "Unauthorized", http.StatusUnauthorized)

					return
				}
			}

			// Верификация токена
			userId, err := authService.ParseToken(tokenString)
			if err != nil {
				//logger.Warn("JWTMiddleware: invalid token", zap.String("path", path), zap.Error(err))
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

type Handler struct {
	logger   *zap.Logger
	Services *service.Service
}

func NewHandler(services *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		logger:   logger,
		Services: services,
	}
}

// Унифицированная обработка ошибок сервиса
func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.IsBadRequest(err):
		h.httpError(w, errors.NewBadRequest(err.Error(), nil))
	case errors.IsNotFound(err):
		h.httpError(w, errors.NewNotFound(err.Error(), nil))
	case errors.IsInvalidToken(err):
		h.httpError(w, errors.NewInvalidToken(errors.ErrMsgInvalidToken, nil))
	default:
		h.httpError(w, errors.NewInternal(errors.ErrMsgInternal, err))
	}
}

// Обработка HTTP ошибок
func (h *Handler) httpError(w http.ResponseWriter, err *errors.Error) {
	h.logger.Error("Handling error", zap.Error(err))
	var status int
	var errorResponse map[string]string

	if errors.IsNotFound(err) {
		status = http.StatusNotFound
		errorResponse = map[string]string{"error": err.Error()}
	} else if errors.IsBadRequest(err) {
		status = http.StatusBadRequest
		errorResponse = map[string]string{"error": err.Error()}
	} else {
		status = http.StatusInternalServerError
		errorResponse = map[string]string{"error": "internal server error"}
	}

	h.jsonResponse(w, status, errorResponse)
}

// Общая функция для отправки JSON-ответов
func (h *Handler) jsonResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
