package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// RegisterHandler Регистрация нового пользователя
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.RegisterHandler"
	logger := h.logger.With(zap.String("op", op))

	logger.Debug("Handling register new user in system request")

	var user models.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Error("Failed to decode JSON body", zap.Error(err))
		h.httpError(w, errors.NewInvalidArgument("Invalid input data", err))
		return
	}

	// Вызов метода регистрации с получением идентификатора пользователя
	userId, err := h.Services.Auth.Register(r.Context(), &user)
	if err != nil {
		logger.Error("Failed to register user", zap.Error(err))
		h.handleServiceError(w, err)
		return
	}

	// Ответ клиенту с сообщением об успешной регистрации и идентификатором пользователя
	response := map[string]interface{}{
		"message": "User registered successfully",
		"user_id": userId,
	}
	h.jsonResponse(w, http.StatusCreated, response)
}

// LoginHandler Авторизация пользователя
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.LoginHandler"
	logger := h.logger.With(zap.String("op", op))

	logger.Debug("Handling login user request")

	var user models.SignIn
	// Декодирование JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Error("Failed to decode JSON body", zap.Error(err))
		h.httpError(w, errors.NewInvalidArgument("Invalid input data", err))
		return
	}

	// Вызов метода авторизации
	token, err := h.Services.Auth.Login(r.Context(), &user)
	if err != nil {
		logger.Error("Failed to authenticate user", zap.Error(err))
		h.handleServiceError(w, err)
		return
	}

	// Установка cookie с токеном
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(time.Hour),
		Path:     "/",
		Secure:   true, // используется HTTPS
		HttpOnly: true, // Защита от XSS-атак
	})

	// Ответ клиенту с сообщением об успешном входе
	response := map[string]interface{}{
		"message": "User Login successful",
		"token":   token,
	}
	h.jsonResponse(w, http.StatusOK, response)
}

// GetUserHandler Получение информации о пользователе по имени пользователя и паролю
func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetUserHandler"
	logger := h.logger.With(zap.String("op", op))

	logger.Debug("Handling user login and password request")

	var req models.SignIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode JSON body", zap.Error(err))
		h.httpError(w, errors.NewInvalidArgument(errors.ErrorMessage[errors.BadRequest], err))
		return
	}

	user, err := h.Services.Auth.GetUser(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to get user ID", zap.Error(err))
		h.handleServiceError(w, err)
		return
	}

	response := UserIDResponse{Id: user.ID}
	h.jsonResponse(w, http.StatusOK, response)
}
