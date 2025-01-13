package service

import (
	"context"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"

	"go.uber.org/zap"
	"time"
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
func (s *TaskService) CreateTask(ctx context.Context, req *models.TaskCreate) (*models.Task, error) {
	s.logger.Info("Creating new task",
		zap.String("title", req.Title),
		zap.String("description", req.Description))

	if err := validateTaskRequest(req); err != nil {
		return nil, err
	}

	task := &models.Task{
		TaskID:      generateTaskID(), // Генерация уникального ID задачи
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   time.Now(), // Установка текущего времени в качестве времени создания
		//DueDate:     req.DueDate,
		// Установка статуса на основе проверки
		Status: validateAndSetStatus(req.Status),
		//AssigneeID: req.AssigneeID,
	}

	return s.repo.CreateTask(ctx, task)
}
