package repository

import (
	"context"
	"database/sql"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository/database"
	"go.uber.org/zap"
)

// AuthRepository интерфейс для работы с аутентификацией
type AuthRepository interface {
	CreateUser(ctx context.Context, user *models.CreateUser) (int64, error)
	GetUser(ctx context.Context, req *models.SignIn) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	GetUserInfo(ctx context.Context, userID int64) (models.User, error)
	GetUsersLeaderboard(ctx context.Context) ([]models.User, error)
	GetUserID(ctx context.Context, usernameOrEmail string) (int64, error)
	ReferrerCode(ctx context.Context, userId int64, refCode string) error
}

// TaskRepository интерфейс для работы с задачами
type TaskRepository interface {
	CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error)
	CompleteTask(ctx context.Context, userId, taskId int64) error
	GetAllTasks(ctx context.Context) ([]models.Task, error)
}

// Repository структура для объединения всех репозиториев
type Repository struct {
	AuthRepository
	UserRepository
	TaskRepository
}

// NewRepositories создает новый экземпляр Repository с логированием
func NewRepositories(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		AuthRepository: database.NewPostgresAuthRepository(db, logger),
		UserRepository: database.NewPostgresUserRepository(db, logger),
		TaskRepository: database.NewPostgresTaskRepository(db, logger),
	}
}
