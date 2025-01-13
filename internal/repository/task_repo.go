package repository

import "github.com/ZnNr/user-task-reward-controller/internal/models"

type TaskRepository interface {
	CreateTask(models.TaskCreate) (int, error)
	CompleteTask(userId, taskId int) error
	GetAllTasks() ([]models.Task, error)
	ReferrerCode(userId int, refCode string) error
}
