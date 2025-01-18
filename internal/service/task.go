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
	s.logger.Info("Creating new task",
		zap.String("title", req.Title),
		zap.String("description", req.Description),
		zap.Int("price", req.Price))

	// Валидация запроса
	if err := validateTaskRequest(req); err != nil {
		s.logger.Error("Validation failed", zap.Error(err))
		return 0, err
	}

	task := &models.TaskCreate{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
	}

	taskID, err := s.repo.CreateTask(ctx, task)
	if err != nil {
		s.logger.Error("Failed to create task", zap.Error(err))
		return 0, err
	}

	s.logger.Info("Task created successfully", zap.String("title", req.Title))
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

func (s *TaskService) CompleteTask(userId, taskId int64) error {
	return s.repo.CompleteTask(userId, taskId)
}

func (s *TaskService) GetAllTasks() ([]models.Task, error) {
	return s.repo.GetAllTasks()
}

func (s *TaskService) ReferrerCode(userId int64, refCode string) error {
	return s.repo.ReferrerCode(userId, refCode)
}
