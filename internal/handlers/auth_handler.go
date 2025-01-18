package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"net/http"
	"time"
)

// UserIDResponse структура для возврата идентификатора пользователя
type UserIDResponse struct {
	Id int64 `json:"user_id"`
}

// Регистрация нового пользователя
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling  register new user in system request")
	var user models.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.httpError(w, errors.NewBadRequest(errors.ErrMsgInvalidInput, err))
		return
	}

	if err := h.Services.Auth.Register(r.Context(), &user); err != nil {
		h.httpError(w, errors.NewBadRequest("Failed to register user: "+err.Error(), err))
		return
	}
	h.jsonResponse(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

// Авторизация пользователя
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling  login user request")
	var user models.SignIn
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.httpError(w, errors.NewBadRequest(errors.ErrMsgInvalidInput, err))
		return
	}

	token, err := h.Services.Auth.Login(r.Context(), &user)
	if err != nil {
		h.httpError(w, errors.NewBadRequest("Failed to authenticate user", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(time.Hour),
	})

	h.jsonResponse(w, http.StatusOK, map[string]string{"message": "Login successful"})
}

// Получение идентификатора пользователя
func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling  user login an password request")
	var req models.SignIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.httpError(w, errors.NewBadRequest("Invalid request payload", err))
		return
	}

	userId, err := h.Services.Auth.GetUser(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, UserIDResponse{Id: userId})
}
