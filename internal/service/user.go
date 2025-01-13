package service

import (
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"go.uber.org/zap"
)

// UserService представляет собой службу управления пользователями
type UserService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

// NewUserService создает новый экземпляр UserService
func NewUserService(repo repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}
