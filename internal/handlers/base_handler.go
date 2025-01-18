package handlers

import (
	"context"
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/service"
	"go.uber.org/zap"
	"net/http"
)

func JWTMiddleware(authService service.Auth) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}
			tokenString := cookie.Value

			userId, err := authService.ParseToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}
			r = r.WithContext(context.WithValue(r.Context(), "userID", userId))
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
