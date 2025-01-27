package tests

import (
	"context"
	service2 "github.com/ZnNr/user-task-reward-controller/internal/service"
	"testing"

	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockRepository реализует интерфейс repository.TaskRepository для тестирования.
type MockRepository struct {
	createTaskFunc   func(ctx context.Context, req *models.TaskCreate) (int64, error)
	completeTaskFunc func(ctx context.Context, userId, taskId int64) error
	getAllTasksFunc  func(ctx context.Context) ([]models.Task, error)
}

func (m *MockRepository) CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error) {
	return m.createTaskFunc(ctx, req)
}

func (m *MockRepository) CompleteTask(ctx context.Context, userId, taskId int64) error {
	return m.completeTaskFunc(ctx, userId, taskId)
}

func (m *MockRepository) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	return m.getAllTasksFunc(ctx)
}

func TestCreateTask(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	tests := []struct {
		name          string
		repo          *MockRepository
		req           *models.TaskCreate
		expectedID    int64
		expectedError error
	}{
		{
			name: "success",
			repo: &MockRepository{
				createTaskFunc: func(ctx context.Context, req *models.TaskCreate) (int64, error) {
					assert.Equal(t, "valid title", req.Title)
					assert.Equal(t, 10, req.Price)
					return 1, nil
				},
			},
			req:           &models.TaskCreate{Title: "valid title", Description: "desc", Price: 10},
			expectedID:    1,
			expectedError: nil,
		},
		{
			name:          "invalid request",
			repo:          &MockRepository{},
			req:           &models.TaskCreate{Title: "", Description: "desc", Price: 10},
			expectedID:    0,
			expectedError: errors.NewValidation("task title cannot be empty", nil),
		},
		{
			name: "repository error",
			repo: &MockRepository{
				createTaskFunc: func(ctx context.Context, req *models.TaskCreate) (int64, error) {
					return 0, errors.NewInternal("repo error", nil)
				},
			},
			req:           &models.TaskCreate{Title: "valid title", Description: "desc", Price: 10},
			expectedID:    0,
			expectedError: errors.NewInternal("repo error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := service2.NewTaskService(tt.repo, logger)
			id, err := service.CreateTask(ctx, tt.req)
			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCompleteTask(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	tests := []struct {
		name          string
		repo          *MockRepository
		userId        int64
		taskId        int64
		expectedError error
	}{
		{
			name: "success",
			repo: &MockRepository{
				completeTaskFunc: func(ctx context.Context, userId, taskId int64) error {
					assert.Equal(t, int64(1), userId)
					assert.Equal(t, int64(1), taskId)
					return nil
				},
			},
			userId:        1,
			taskId:        1,
			expectedError: nil,
		},
		{
			name: "repository error",
			repo: &MockRepository{
				completeTaskFunc: func(ctx context.Context, userId, taskId int64) error {
					return errors.NewInternal("repo error", nil)
				},
			},
			userId:        1,
			taskId:        1,
			expectedError: errors.NewInternal("repo error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := service2.NewTaskService(tt.repo, logger)
			err := service.CompleteTask(ctx, tt.userId, tt.taskId)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestGetAllTasks(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	tests := []struct {
		name          string
		repo          *MockRepository
		expectedTasks []models.Task
		expectedError error
	}{
		{
			name: "success",
			repo: &MockRepository{
				getAllTasksFunc: func(ctx context.Context) ([]models.Task, error) {
					return []models.Task{{TaskID: 1}}, nil
				},
			},
			expectedTasks: []models.Task{{TaskID: 1}},
			expectedError: nil,
		},
		{
			name: "repository error",
			repo: &MockRepository{
				getAllTasksFunc: func(ctx context.Context) ([]models.Task, error) {
					return nil, errors.NewInternal("repo error", nil)
				},
			},
			expectedTasks: nil,
			expectedError: errors.NewInternal("repo error", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := service2.NewTaskService(tt.repo, logger)
			tasks, err := service.GetAllTasks(ctx)
			assert.Equal(t, tt.expectedTasks, tasks)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
