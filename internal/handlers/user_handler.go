package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"net/http"
	"strconv"
	"strings"
)

// Получение информации о пользователе
func (h *Handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		h.httpError(w, errors.NewBadRequest("Invalid user id param", err))
		return
	}

	userInfo, err := h.Services.GetUserInfo(int64(userId))
	if err != nil {
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}

	h.jsonResponse(w, http.StatusOK, struct {
		User models.User `json:"user"`
	}{
		User: userInfo,
	})
}

// Топ пользователей на основе статистики
func (h *Handler) UsersLeaderboard(w http.ResponseWriter, r *http.Request) {
	users, err := h.Services.GetUsersLeaderboard()
	if err != nil {
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}

	h.jsonResponse(w, http.StatusOK, struct {
		Data []models.User `json:"data"`
	}{
		Data: users,
	})
}

// Обработка реферального кода
func (h *Handler) UserReferrerCode(w http.ResponseWriter, r *http.Request) {
	var referral struct {
		ReferrerCode string `json:"referral_code"`
	}

	userId, err := strconv.Atoi(r.URL.Query().Get("id"))
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

	h.jsonResponse(w, http.StatusOK, map[string]string{
		"success": "ok",
	})
}

func (h *Handler) GetUserIDbyUsernameOrEmailHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling Get User ID by username or email request")
	// Проверка метода запроса (GET)
	if r.Method != http.MethodGet {
		h.httpError(w, errors.NewBadRequest("Method not allowed", nil))
		return
	}

	// Получаем параметр имени пользователя или email из строки запроса
	usernameOrEmail := r.URL.Query().Get("username_or_email")
	if strings.TrimSpace(usernameOrEmail) == "" {
		http.Error(w, "Query parameter 'username_or_email' is required", http.StatusBadRequest)
		return
	}

	//// Получаем контекст для передачи в репозиторий
	//ctx := r.Context()

	// Вызов сервиса для получения ID пользователя
	userID, err := h.Services.User.GetUserID(r.Context(), usernameOrEmail)
	if err != nil {
		h.handleServiceError(w, err) // централизованная обработка ошибок
		return
	}

	// Формирование и отправка JSON-ответа
	h.jsonResponse(w, http.StatusOK, UserIDResponse{Id: userID})
}
