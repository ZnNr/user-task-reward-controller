package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

// Получение информации о пользователе
func (h *Handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr, exists := vars["user_id"]
	if !exists {
		h.logger.Error("User ID is required")
		h.httpError(w, errors.NewBadRequest("User ID is required", nil))
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid User ID", zap.Error(err))
		h.httpError(w, errors.NewBadRequest("Invalid user id param", err))
		return
	}

	userInfo, err := h.Services.GetUserInfo(userID)
	if err != nil {
		h.logger.Error("Failed to get user info", zap.Int64("UserID", userID), zap.Error(err))
		h.handleServiceError(w, err)
		return
	}

	response := struct {
		User models.User `json:"user"`
	}{
		User: userInfo,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// Топ пользователей на основе статистики
func (h *Handler) UsersLeaderboard(w http.ResponseWriter, r *http.Request) {
	users, err := h.Services.GetUsersLeaderboard()
	if err != nil {
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}
	response := struct {
		Data []models.User `json:"data"`
	}{
		Data: users,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// Обработка реферального кода
func (h *Handler) UserReferrerCode(w http.ResponseWriter, r *http.Request) {
	var referral struct {
		ReferrerCode string `json:"referral_code"`
	}

	userId, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil {
		h.httpError(w, errors.NewBadRequest("Invalid user_id param", err))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&referral); err != nil {
		h.httpError(w, errors.NewBadRequest("Invalid input body", err))
		return
	}

	if err := h.Services.ReferrerCode(int64(userId), referral.ReferrerCode); err != nil {
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}

	response := map[string]interface{}{
		"success": "ok",
	}

	h.jsonResponse(w, http.StatusOK, response)
}

func (h *Handler) GetUserIDbyUsernameOrEmailHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling Get User ID by username or email request")
	// Проверка метода запроса (GET)
	if r.Method != http.MethodGet {
		h.httpError(w, errors.NewBadRequest("Method not allowed", nil))
		return
	}

	// Извлечение параметра username_or_email из пути URL
	vars := mux.Vars(r)
	usernameOrEmail := vars["username_or_email"]
	if strings.TrimSpace(usernameOrEmail) == "" {
		http.Error(w, "Path parameter 'username_or_email' is required", http.StatusBadRequest)
		return
	}

	// Вызов сервиса для получения ID пользователя
	userID, err := h.Services.User.GetUserID(r.Context(), usernameOrEmail)
	if err != nil {
		h.handleServiceError(w, err) // централизованная обработка ошибок
		return
	}
	response := UserIDResponse{Id: userID}

	// Ответ клиенту с информацией о пользователе
	h.jsonResponse(w, http.StatusOK, response)
}
