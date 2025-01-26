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

// GetUserInfo возвращает информацию о пользователе по ID
func (u *UserService) GetUserInfo(ctx context.Context, userId int64) (models.User, error) {
	const op = "service.User.GetUserInfo"
	logger := u.logger.With(zap.String("op", op))

	logger.Debug("Fetching user info", zap.Int64("user_id", userId))
	user, err := u.repo.GetUserInfo(ctx, userId)
	if err != nil {
		logger.Error("Failed to fetch user info", zap.Error(err))
		return models.User{}, err
	}
	logger.Info("User info fetched successfully", zap.Int64("user_id", userId))
	return user, nil
}

// GetUsersLeaderboard возвращает список пользователей, отсортированный по балансу
func (u *UserService) GetUsersLeaderboard(ctx context.Context) ([]models.User, error) {
	const op = "service.User.GetUsersLeaderboard"
	logger := u.logger.With(zap.String("op", op))

	logger.Debug("Fetching users leaderboard")
	users, err := u.repo.GetUsersLeaderboard(ctx)
	if err != nil {
		logger.Error("Failed to fetch leaderboard", zap.Error(err))
		return nil, err
	}
	logger.Info("Leaderboard fetched successfully", zap.Int("users_count", len(users)))
	return users, nil
}

// GetUserID возвращает ID пользователя по имени пользователя или email
func (u *UserService) GetUserID(ctx context.Context, usernameOrEmail string) (int64, error) {
	const op = "service.User.GetUserID"
	logger := u.logger.With(zap.String("op", op))

	logger.Debug("Fetching user ID", zap.String("username_or_email", usernameOrEmail))
	userID, err := u.repo.GetUserID(ctx, usernameOrEmail)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Warn("User not found", zap.String("username_or_email", usernameOrEmail))
			return 0, err
		}
		logger.Error("Failed to fetch user ID", zap.Error(err))
		return 0, err
	}
	logger.Info("User ID fetched successfully", zap.String("username_or_email", usernameOrEmail), zap.Int64("user_id", userID))
	return userID, nil
}

// ReferrerCode сохраняет реферальный код для пользователя
func (u *UserService) ReferrerCode(ctx context.Context, userId int64, refCode string) error {
	const op = "service.User.ReferrerCode"
	logger := u.logger.With(zap.String("op", op))

	logger.Debug("Saving referrer code", zap.Int64("user_id", userId), zap.String("ref_code", refCode))
	err := u.repo.ReferrerCode(ctx, userId, refCode)
	if err != nil {
		logger.Error("Failed to save referrer code", zap.Error(err))
		return err
	}
	logger.Info("Referrer code saved successfully", zap.Int64("user_id", userId), zap.String("ref_code", refCode))
	return nil
}
