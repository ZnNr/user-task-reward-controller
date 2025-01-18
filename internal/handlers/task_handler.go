package handlers

import (
	"encoding/json"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"net/http"
	"strconv"
)

// Создание задачи
func (h *Handler) TaskCreate(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling  creating new task request")
	var task models.TaskCreate
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.httpError(w, errors.NewBadRequest("Invalid input body", err))
		return
	}

	//if err := task.Validate(); err != nil {
	//	h.httpError(w, errors.NewInternal(err.Error(), nil))
	//	return
	//}
	// Извлечение контекста
	ctx := r.Context()
	newTaskId, err := h.Services.CreateTask(ctx, &task)
	if err != nil {
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"new_task_id": newTaskId,
	})
}

// Завершение задачи
func (h *Handler) TaskComplete(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling  complete task request")
	var taskComplete struct {
		TaskId int `json:"task_id"`
	}

	userId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		h.httpError(w, errors.NewBadRequest("Invalid user_id param", err))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&taskComplete); err != nil {
		h.httpError(w, errors.NewBadRequest("Invalid input body", err))
		return
	}

	if err := h.Services.CompleteTask(int64(userId), int64(taskComplete.TaskId)); err != nil {
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success": "ok",
	})
}

// Получение всех задач
func (h *Handler) TaskGetAll(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling  get all tasks from repo request")
	tasks, err := h.Services.GetAllTasks()
	if err != nil {
		h.httpError(w, errors.NewInternal(err.Error(), nil))
		return
	}

	h.jsonResponse(w, http.StatusOK, struct {
		Data []models.Task `json:"tasks"`
	}{
		Data: tasks,
	})
}
