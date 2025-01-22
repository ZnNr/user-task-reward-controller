package database

import (
	"context"
	"database/sql"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	//"github.com/ZnNr/user-task-reward-controller/internal/repository"
)

// SQL-запросы
const (
	// Получение таблицы лидеров по балансу
	GetLeaderboardByBalanceQuery = `SELECT user_id, username, balance, refer_code, refer_from FROM users ORDER BY balance DESC`

	// Получение информации о пользователе по ID
	GetUserByIDQuery = `SELECT user_id, username, email, balance, refer_code, refer_from FROM users WHERE user_id = $1`

	// Получить ID пользователя по имени пользователя или email
	GetUserIDQuery = `
    SELECT user_id 
    FROM users 
    WHERE username = $1 OR email = $2`
)

// PostgresUserRepository реализует репозиторий пользователей для PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository создает новый экземпляр репозитория пользователей
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// GetUsersLeaderboard возвращает список пользователей, отсортированный по балансу
func (r *PostgresUserRepository) GetUsersLeaderboard() ([]models.User, error) {
	rows, err := r.db.Query(GetLeaderboardByBalanceQuery)
	if err != nil {
		return nil, errors.NewInternal("Failed to fetch leaderboard", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Username, &user.Balance, &user.ReferCode, &user.ReferFrom)
		if err != nil {
			return nil, errors.NewInternal("Failed to scan leaderboard row", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewInternal("Error iterating leaderboard rows", err)
	}

	return users, nil
}

// GetUserInfo возвращает информацию о пользователе по ID
func (r *PostgresUserRepository) GetUserInfo(userID int64) (models.User, error) {
	var user models.User
	err := r.db.QueryRow(GetUserByIDQuery, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Balance, &user.ReferCode, &user.ReferFrom)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.NewNotFound("User not found", err)
		}
		return models.User{}, errors.NewInternal("Error fetching user info", err)
	}

	return user, nil
}

// GetUserID возвращает user_id пользователя по имени пользователя или email
func (r *PostgresUserRepository) GetUserID(ctx context.Context, usernameOrEmail string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, GetUserIDQuery, usernameOrEmail, usernameOrEmail).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.NewNotFound("User ID not found", nil)
		}
		return 0, errors.NewInternal("Failed to fetch user ID", err)
	}
	return userID, nil
}
