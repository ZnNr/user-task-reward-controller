package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"go.uber.org/zap"
)

// SQL-запросы
const (
	// Получение таблицы лидеров по балансу
	GetLeaderboardByBalanceQuery = `SELECT user_id, username, balance, refer_code, refer_from FROM users ORDER BY balance DESC`

	// Получение информации о пользователе по ID
	GetUserByIDQuery = `SELECT user_id, username, email, balance, refer_code, refer_from FROM users WHERE user_id = $1`

	// Получить ID пользователя по имени пользователя или email
	GetUserIDQuery = `SELECT user_id FROM users WHERE username = $1 OR email = $2`
)

// PostgresUserRepository реализует репозиторий пользователей для PostgreSQL
type PostgresUserRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresUserRepository создает новый экземпляр репозитория пользователей
func NewPostgresUserRepository(db *sql.DB, logger *zap.Logger) *PostgresUserRepository {
	return &PostgresUserRepository{db: db, logger: logger}
}

// executeQuery выполняет SQL-запрос и возвращает результат
func (r *PostgresUserRepository) executeQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return nil, errors.NewInternal("Failed to execute query", err)
	}
	return rows, nil
}

// executeQueryRow выполняет SQL-запрос и сканирует результат
func (r *PostgresUserRepository) executeQueryRow(ctx context.Context, query string, args ...interface{}) error {
	row := r.db.QueryRowContext(ctx, query, args...)
	return row.Err()
}

// GetUsersLeaderboard возвращает список пользователей, отсортированный по балансу
func (r *PostgresUserRepository) GetUsersLeaderboard(ctx context.Context) ([]models.User, error) {
	rows, err := r.executeQuery(ctx, GetLeaderboardByBalanceQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Username, &user.Balance, &user.ReferCode, &user.ReferFrom)
		if err != nil {
			r.logger.Error("Failed to scan leaderboard row", zap.Error(err))
			return nil, errors.NewInternal("Failed to scan leaderboard row", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating leaderboard rows", zap.Error(err))
		return nil, errors.NewInternal("Error iterating leaderboard rows", err)
	}

	return users, nil
}

// GetUserInfo возвращает информацию о пользователе по ID
func (r *PostgresUserRepository) GetUserInfo(ctx context.Context, userID int64) (models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, GetUserByIDQuery, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Balance, &user.ReferCode, &user.ReferFrom)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info("User not found", zap.Int64("user_id", userID))
			return models.User{}, errors.NewNotFound("User not found", err)
		}
		r.logger.Error("Error fetching user info", zap.Error(err))
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
			r.logger.Info("User ID not found", zap.String("username_or_email", usernameOrEmail))
			return 0, errors.NewNotFound("User ID not found", nil)
		}
		r.logger.Error("Failed to fetch user ID", zap.Error(err))
		return 0, errors.NewInternal("Failed to fetch user ID", err)
	}
	return userID, nil
}

// ReferrerCode сохраняет реферальный код для пользователя
func (r *PostgresUserRepository) ReferrerCode(ctx context.Context, userId int64, referCode string) error {
	var refId int
	err := r.db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE refer_code=$1", referCode).Scan(&refId)
	if err == sql.ErrNoRows {
		r.logger.Info("User with refer_code not found", zap.String("refer_code", referCode))
		return errors.NewNotFound(fmt.Sprintf("user with refer_code \"%s\" not found", referCode), err)
	} else if err != nil {
		r.logger.Error("Error querying for refer_code", zap.String("refer_code", referCode), zap.Error(err))
		return errors.NewInternal("failed to query user by refer_code", err)
	}

	r.logger.Info("Found user with refer_code", zap.String("refer_code", referCode), zap.Int("user_id", refId))

	result, err := r.db.ExecContext(ctx, "UPDATE users SET refer_from=$1 WHERE user_id=$2", refId, userId)
	if err != nil {
		r.logger.Error("Error updating refer_from", zap.Int64("user_id", userId), zap.Error(err))
		return errors.NewInternal("failed to set referrer code", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error getting rows affected", zap.Int64("user_id", userId), zap.Error(err))
		return errors.NewInternal("failed to get rows affected", err)
	}
	if rowsAffected == 0 {
		r.logger.Info("No rows were updated", zap.Int64("user_id", userId))
		return errors.NewNotFound("no rows were updated", nil)
	}

	r.logger.Info("Successfully set refer_from", zap.Int64("user_id", userId), zap.Int("refer_id", refId))
	return nil
}
