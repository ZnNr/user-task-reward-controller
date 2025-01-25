package repository

import (
	"context"
	"database/sql"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository/database"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *models.CreateUser) (int64, error)
	GetUser(ctx context.Context, req *models.SignIn) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type UserRepository interface {
	GetUserInfo(userID int64) (models.User, error)
	GetUsersLeaderboard() ([]models.User, error)
	GetUserID(ctx context.Context, usernameOrEmail string) (int64, error)
}

type TaskRepository interface {
	CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error)
	CompleteTask(ctx context.Context, userId, taskId int64) error
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
