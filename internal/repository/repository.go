package repository

import (
	"context"
	"database/sql"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository/database"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *models.CreateUser) (int64, error)
	GetUser(ctx context.Context, req *models.SignIn) (models.User, error)
}

type UserRepository interface {
	GetUserInfo(userID int64) (models.User, error)
	GetUsersLeaderboard() ([]models.User, error)
}

type TaskRepository interface {
	CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error)
	CompleteTask(userId, taskId int64) error
	GetAllTasks() ([]models.Task, error)
	ReferrerCode(userId int64, refCode string) error
}

type Repository struct {
	AuthRepository
	UserRepository
	TaskRepository
}

func NewRepositories(db *sql.DB) *Repository {
	return &Repository{
		AuthRepository: database.NewPostgresAuthRepository(db),
		UserRepository: database.NewPostgresUserRepository(db),
		TaskRepository: database.NewPostgresTaskRepository(db),
	}
}
