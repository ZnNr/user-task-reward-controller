package service

import (
	"context"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"go.uber.org/zap"
	"time"
)

// Auth интерфейс для аутентификации
type Auth interface {
	Login(ctx context.Context, credentials *models.SignIn) (string, error)
	Register(ctx context.Context, userInfo *models.CreateUser) (int64, error)
	ParseToken(token string) (int64, error)
	GetUser(ctx context.Context, up *models.SignIn) (*models.User, error)
}

// User интерфейс для работы с пользователями
type User interface {
	GetUserInfo(ctx context.Context, userId int64) (models.User, error)
	GetUsersLeaderboard(ctx context.Context) ([]models.User, error)
	GetUserID(ctx context.Context, usernameOrEmail string) (int64, error)
	ReferrerCode(ctx context.Context, userId int64, refCode string) error
}

// Task интерфейс для работы с задачами
type Task interface {
	CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error)
	CompleteTask(tx context.Context, userId, taskId int64) error
	GetAllTasks(ctx context.Context) ([]models.Task, error)
}

// Service структура для объединения всех сервисов
type Service struct {
	Auth
	User
	Task
}

// ServicesDependencies зависимости для создания Service
type ServicesDependencies struct {
	Repos    *repository.Repository
	Logger   *zap.Logger
	SignKey  string
	TokenTTL time.Duration
}

// NewService создает новый экземпляр Service
func NewService(deps ServicesDependencies) *Service {
	return &Service{
		Auth: NewAuthService(AuthDependencies{
			authRepo: deps.Repos.AuthRepository,
			logger:   deps.Logger,
			signKey:  deps.SignKey,
			tokenTTL: deps.TokenTTL,
		}),
		Task: NewTaskService(deps.Repos.TaskRepository, deps.Logger),
		User: NewUserService(deps.Repos.UserRepository, deps.Logger),
	}
}
