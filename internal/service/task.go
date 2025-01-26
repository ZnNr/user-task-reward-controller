package service

import (
	"context"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"go.uber.org/zap"
)

type TaskService struct {
	repo   repository.TaskRepository
	logger *zap.Logger
}

func NewTaskService(repo repository.TaskRepository, logger *zap.Logger) *TaskService {
	return &TaskService{
		repo:   repo,
		logger: logger,
	}
}

// CreateTask создает новую задачу.
func (s *TaskService) CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error) {
	const op = "service.Task.CreateTask"
	logger := s.logger.With(zap.String("op", op))

	logger.Info("Creating new task",
		zap.String("title", req.Title),
		zap.String("description", req.Description),
		zap.Int("price", req.Price))

	// Валидация запроса
	if err := validateTaskRequest(req); err != nil {
		logger.Error("Validation failed", zap.Error(err))
		return 0, err
	}

	taskID, err := s.repo.CreateTask(ctx, req)
	if err != nil {
		logger.Error("Failed to create task", zap.Error(err))
		return 0, err
	}

	logger.Info("Task created successfully", zap.Int64("task_id", taskID), zap.String("title", req.Title))
	return taskID, nil
}

// validateTaskRequest выполняет проверку валидности запроса на создание или обновление задачи.
func validateTaskRequest(req *models.TaskCreate) error {
	if req.Title == "" {
		return errors.NewValidation("task title cannot be empty", nil)
	}
	if req.Price < 1 {
		return errors.NewValidation("minimum value for the Price field is 1", nil)
	}
	return nil
}

// CompleteTask завершает задачу и обновляет баланс пользователя.
func (s *TaskService) CompleteTask(ctx context.Context, userId, taskId int64) error {
	const op = "service.Task.CompleteTask"
	logger := s.logger.With(zap.String("op", op))

	logger.Info("Completing task", zap.Int64("user_id", userId), zap.Int64("task_id", taskId))

	err := s.repo.CompleteTask(ctx, userId, taskId)
	if err != nil {
		logger.Error("Failed to complete task", zap.Error(err))
		return err
	}

	logger.Info("Task completed successfully", zap.Int64("user_id", userId), zap.Int64("task_id", taskId))
	return nil
}

// GetAllTasks возвращает все задачи.
func (s *TaskService) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	const op = "service.Task.GetAllTasks"
	logger := s.logger.With(zap.String("op", op))

	logger.Info("Fetching all tasks")

	tasks, err := s.repo.GetAllTasks(ctx)
	if err != nil {
		logger.Error("Failed to fetch all tasks", zap.Error(err))
		return nil, err
	}

	logger.Info("All tasks fetched successfully", zap.Int("tasks_count", len(tasks)))
	return tasks, nil
}
