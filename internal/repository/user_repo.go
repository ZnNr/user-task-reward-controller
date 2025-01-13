package repository

import (
	"context"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
)

// UserRepository определяет методы для взаимодействия с данными пользователя в базе данных
type UserRepository interface {
	// CreateUser создает нового пользователя в базе данных
	CreateUser(ctx context.Context, user *models.CreateUser) (*models.CreateUser, error)

	// GetUsers возвращает список пользователей, соответствующих заданному фильтру
	GetUsers(ctx context.Context, filter *models.User) (*models.UsersResponse, error)

	// GetUserByID возвращает пользователя по его уникальному идентификатору
	GetUserByID(ctx context.Context, id string) (*models.User, error)

	// UpdateUser обновляет информацию о существующем пользователе
	//UpdateUser(ctx context.Context, user *models.User) (*models.User, error)

	// DeleteUser удаляет пользователя из базы данных
	//DeleteUser(ctx context.Context, id string) error

	GetUserInfo(userId int) (models.User, error)
	GetUsersLeaderboard() ([]models.User, error)
}
