package service

import (
	"context"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"go.uber.org/zap"
	"time"
)

type Auth interface {
	Login(ctx context.Context, credentials *models.SignIn) (string, error)
	Register(ctx context.Context, userInfo *models.CreateUser) (int64, error)
	ParseToken(token string) (int64, error)
	GetUser(ctx context.Context, up *models.SignIn) (*models.User, error)
}
type User interface {
	GetUserInfo(userId int64) (models.User, error)
	GetUsersLeaderboard() ([]models.User, error)
	GetUserID(ctx context.Context, usernameOrEmail string) (int64, error)
}

type Task interface {
	CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error)
	CompleteTask(tx context.Context, userId, taskId int64) error
	GetAllTasks() ([]models.Task, error)
	ReferrerCode(userId int64, refCode string) error
}

type Service struct {
	Auth
	User
	Task
}

type ServicesDependencies struct {
	Repos    *repository.Repository
	Logger   *zap.Logger
	SignKey  string
	TokenTTL time.Duration
}

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
