package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/service/refercode"
	"go.uber.org/zap"
)

const (
	// Запрос для создания пользователя
	CreateUserQuery = `
    INSERT INTO users (username, password, refer_code) VALUES ($1, $2, $3) RETURNING user_id`
	// Проверка существования пользователя
	CheckUserExistsQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)`
	// Получение пользователя по имени пользователя и паролю
	GetUserQuery = `SELECT user_id, username, password, email FROM users WHERE username = $1 AND password = $2`
	// Получение пользователя по имени пользователя
	GetUserByUsernameQuery = `SELECT user_id, username, password, email FROM users WHERE username = $1`
)

// PostgresAuthRepository реализует репозиторий пользователей для PostgresSQL
type PostgresAuthRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresAuthRepository создает новый экземпляр репозитория пользователей
func NewPostgresAuthRepository(db *sql.DB, logger *zap.Logger) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db, logger: logger}
}

// executeQuery выполняет SQL-запрос и возвращает результат
func (r *PostgresAuthRepository) executeQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return nil, errors.NewInternal("Failed to execute query", err)
	}
	return rows, nil
}

// executeQueryRow выполняет SQL-запрос и сканирует результат
func (r *PostgresAuthRepository) executeQueryRow(ctx context.Context, query string, args ...interface{}) error {
	row := r.db.QueryRowContext(ctx, query, args...)
	return row.Err()
}

// executeExec выполняет SQL-запрос на изменение данных и возвращает количество затронутых строк
func (r *PostgresAuthRepository) executeExec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to execute exec query", zap.String("query", query), zap.Error(err))
		return 0, errors.NewInternal("Failed to execute exec query", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.String("query", query), zap.Error(err))
		return 0, errors.NewInternal("Failed to get rows affected", err)
	}
	return rowsAffected, nil
}

// CreateUser создает нового пользователя
func (r *PostgresAuthRepository) CreateUser(ctx context.Context, user *models.CreateUser) (int64, error) {
	// Проверка существования пользователя
	exists, err := r.checkUserExists(ctx, user)
	if err != nil {
		r.logger.Error("Can't check user existence", zap.Error(err))
		return 0, errors.NewValidation("Can't check user existence", err)
	}
	if exists {
		r.logger.Info("User already exists", zap.String("username", user.Username), zap.String("email", user.Email))
		return 0, errors.NewAlreadyExists("User already exists", nil)
	}

	// Генерация реферального кода
	referCode := refercode.RandStringBytes()

	// Подготовка SQL-запроса
	var lastID int64
	err = r.db.QueryRowContext(ctx, CreateUserQuery, user.Username, user.Password, referCode).Scan(&lastID)
	if err != nil {
		r.logger.Error("Failed to execute query to create user", zap.Error(err))
		return 0, errors.NewInternal("Failed to execute query to create user", err)
	}
	r.logger.Info("User created successfully", zap.Int64("user_id", lastID))
	return lastID, nil
}

// checkUserExists проверяет, существует ли пользователь с указанным именем пользователя или электронной почтой
func (r *PostgresAuthRepository) checkUserExists(ctx context.Context, user *models.CreateUser) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, CheckUserExistsQuery, user.Username, user.Email).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check user existence", zap.Error(err))
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// GetUser возвращает пользователя по имени и паролю
func (r *PostgresAuthRepository) GetUser(ctx context.Context, req *models.SignIn) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, GetUserQuery, req.Username, req.Password).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info("User not found", zap.String("username", req.Username))
			return nil, errors.NewNotFound("User not found", err)
		}
		r.logger.Error("Error fetching user", zap.Error(err))
		return nil, errors.NewInternal("Error fetching user", err)
	}
	r.logger.Info("User fetched successfully", zap.Int64("user_id", user.ID))
	return &user, nil
}

// GetUserByUsername возвращает пользователя по имени пользователя
func (r *PostgresAuthRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, GetUserByUsernameQuery, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info("User not found by username", zap.String("username", username))
			return nil, errors.NewNotFound("User not found", err)
		}
		r.logger.Error("Error fetching user by username", zap.Error(err))
		return nil, errors.NewInternal("Error fetching user", err)
	}
	r.logger.Info("User fetched successfully by username", zap.Int64("user_id", user.ID))
	return &user, nil
}
