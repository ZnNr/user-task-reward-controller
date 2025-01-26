package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// TaskCreate создает новую задачу
func (h *Handler) TaskCreate(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.TaskCreate"
	logger := h.logger.With(zap.String("op", op))

	logger.Debug("Handling creating new task request")

	var task models.TaskCreate
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		logger.Error("Failed to decode JSON body", zap.Error(err))
		h.httpError(w, errors.NewBadRequest("Invalid input body", err))
		return
	}

	ctx := r.Context()
	newTaskId, err := h.Services.Task.CreateTask(ctx, &task)
	if err != nil {
		logger.Error("Failed to create task", zap.Error(err))
		h.handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"new_task_id": newTaskId,
	}
	h.jsonResponse(w, http.StatusOK, response)
}

// TaskComplete завершает задачу
func (h *Handler) TaskComplete(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.TaskComplete"
	logger := h.logger.With(zap.String("op", op))

	logger.Debug("Handling complete task request")

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

	var req struct {
		TaskID int64 `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode JSON body", zap.Error(err))
		h.httpError(w, errors.NewBadRequest("Invalid request payload", err))
		return
	}

	err = h.Services.Task.CompleteTask(r.Context(), userID, req.TaskID)
	if err != nil {
		logger.Error("Failed to complete task", zap.Error(err))
		h.handleServiceError(w, err)
		return
	}

	response := map[string]string{
		"message": "Task completed successfully",
	}
	h.jsonResponse(w, http.StatusOK, response)
}

// TaskGetAll получает все задачи
func (h *Handler) TaskGetAll(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.TaskGetAll"
	logger := h.logger.With(zap.String("op", op))

	logger.Debug("Handling get all tasks from repo request")

	tasks, err := h.Services.Task.GetAllTasks(r.Context())
	if err != nil {
		logger.Error("Failed to get all tasks", zap.Error(err))
		h.httpError(w, errors.NewInternal("Failed to get all tasks", err))
		return
	}

	response := struct {
		Data []models.Task `json:"tasks"`
	}{
		Data: tasks,
	}
	h.jsonResponse(w, http.StatusOK, response)
}
