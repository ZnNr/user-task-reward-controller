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

// UserInfo получает информацию о пользователе по его ID
func (h *Handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UserInfo"
	logger := h.logger.With(zap.String("op", op))

	vars := mux.Vars(r)
	userIDStr, exists := vars["user_id"]
	if !exists {
		logger.Error("User ID is required")
		h.httpError(w, errors.NewBadRequest("User ID is required", nil))
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Error("Invalid User ID", zap.Error(err))
		h.httpError(w, errors.NewBadRequest("Invalid user id param", err))
		return
	}

	userInfo, err := h.Services.User.GetUserInfo(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user info", zap.Int64("UserID", userID), zap.Error(err))
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

// UsersLeaderboard получает топ пользователей на основе статистики
func (h *Handler) UsersLeaderboard(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UsersLeaderboard"
	logger := h.logger.With(zap.String("op", op))

	users, err := h.Services.User.GetUsersLeaderboard(r.Context())
	if err != nil {
		logger.Error("Failed to get users leaderboard", zap.Error(err))
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

// UserReferrerCode обрабатывает реферальный код пользователя
func (h *Handler) UserReferrerCode(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.UserReferrerCode"
	logger := h.logger.With(zap.String("op", op))

	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	if userIDStr == "" {
		logger.Info("Missing user_id param in URL", zap.String("url", r.URL.String()))
		h.httpError(w, errors.NewBadRequest("Missing user_id param", nil))
		return
	}

	userId, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.Info("Invalid user_id param", zap.String("user_id", userIDStr), zap.Error(err))
		h.httpError(w, errors.NewBadRequest("Invalid user_id param", err))
		return
	}

	var referral struct {
		ReferrerCode string `json:"refer_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&referral); err != nil {
		logger.Error("Failed to decode request body", zap.Error(err))
		h.httpError(w, errors.NewBadRequest("Invalid input body", err))
		return
	}

	if referral.ReferrerCode == "" {
		logger.Info("Missing refer_code param in request body")
		h.httpError(w, errors.NewBadRequest("Missing refer_code param", nil))
		return
	}

	logger.Info("Received user_id and referral code", zap.Int("user_id", userId), zap.String("referral_code", referral.ReferrerCode))

	if err := h.Services.User.ReferrerCode(r.Context(), int64(userId), referral.ReferrerCode); err != nil {
		logger.Error("Failed to process referrer code", zap.Error(err))
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}

	response := map[string]interface{}{
		"success": "ok",
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// GetUserIDbyUsernameOrEmailHandler получает ID пользователя по имени пользователя или email
func (h *Handler) GetUserIDbyUsernameOrEmailHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetUserIDbyUsernameOrEmailHandler"
	logger := h.logger.With(zap.String("op", op))

	logger.Debug("Handling Get User ID by username or email request")

	if r.Method != http.MethodGet {
		logger.Error("Method not allowed", zap.String("method", r.Method))
		h.httpError(w, errors.NewBadRequest("Method not allowed", nil))
		return
	}

	vars := mux.Vars(r)
	usernameOrEmail := vars["username_or_email"]
	if strings.TrimSpace(usernameOrEmail) == "" {
		logger.Error("Path parameter 'username_or_email' is required", zap.String("path_param", "username_or_email"))
		http.Error(w, "Path parameter 'username_or_email' is required", http.StatusBadRequest)
		return
	}

	userID, err := h.Services.User.GetUserID(r.Context(), usernameOrEmail)
	if err != nil {
		logger.Error("Failed to get user ID", zap.String("username_or_email", usernameOrEmail), zap.Error(err))
		h.handleServiceError(w, err)
		return
	}

	response := UserIDResponse{Id: userID}
	h.jsonResponse(w, http.StatusOK, response)
}
