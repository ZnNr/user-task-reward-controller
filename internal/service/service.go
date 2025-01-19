package service

import (
	"context"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"go.uber.org/zap"
	"time"
)

type Auth interface {
	Login(ctx context.Context, login *models.SignIn) (string, error)
	Register(ctx context.Context, signUp *models.CreateUser) error
	ParseToken(token string) (int64, error)
	GetUser(ctx context.Context, up *models.SignIn) (int64, error)
}
type User interface {
	GetUserInfo(userId int64) (models.User, error)
	GetUsersLeaderboard() ([]models.User, error)
}

type Task interface {
	CreateTask(ctx context.Context, req *models.TaskCreate) (int64, error)
	CompleteTask(userId, taskId int64) error
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
	Salt     string
}

func NewService(deps ServicesDependencies) *Service {
	return &Service{
		Auth: NewAuthService(AuthDependencies{
			authRepo: deps.Repos.AuthRepository,
			logger:   deps.Logger,
			signKey:  deps.SignKey,
			tokenTTL: deps.TokenTTL,
			salt:     deps.Salt,
		}),
		Task: NewTaskService(deps.Repos.TaskRepository, deps.Logger),
		User: NewUserService(deps.Repos.UserRepository, deps.Logger),
	}
}
