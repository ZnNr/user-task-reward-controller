package database

import (
	"database/sql"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	//"github.com/ZnNr/user-task-reward-controller/internal/repository"
)

// SQL-запросы
const (
	// Получение таблицы лидеров по балансу
	GetLeaderboardByBalanceQuery = `SELECT user_id, Username, Balance, Refer_code, Refer_from FROM users ORDER BY Balance DESC`

	// Получение информации о пользователе по ID
	GetUserByIDQuery = `SELECT user_id, Username, Email, Balance, Refer_code, Refer_from FROM users WHERE ID = $1`
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
