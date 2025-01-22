package service

import (
	"context"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
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

func (u *UserService) GetUserInfo(userId int64) (models.User, error) {
	u.logger.Debug("Service log Fetching user info", zap.Int64("user_id", userId))
	user, err := u.repo.GetUserInfo(userId)
	if err != nil {
		u.logger.Error("Failed to fetch user info", zap.Error(err))
		return models.User{}, err
	}
	return user, nil
}

func (u *UserService) GetUsersLeaderboard() ([]models.User, error) {
	u.logger.Debug("Service log: Fetching users leaderboard")
	users, err := u.repo.GetUsersLeaderboard()
	if err != nil {
		u.logger.Error("Failed to fetch leaderboard", zap.Error(err))
		return nil, err
	}
	return users, nil
}

// GetUserID возвращает ID пользователя по имени пользователя или email
func (u *UserService) GetUserID(ctx context.Context, usernameOrEmail string) (int64, error) {
	u.logger.Debug("Fetching user ID", zap.String("username_or_email", usernameOrEmail))
	userID, err := u.repo.GetUserID(ctx, usernameOrEmail)
	if err != nil {
		if errors.IsNotFound(err) {
			u.logger.Warn("User not found", zap.String("username_or_email", usernameOrEmail))
			return 0, err
		}
		u.logger.Error("Failed to fetch user ID", zap.Error(err))
		return 0, err
	}
	return userID, nil
}
