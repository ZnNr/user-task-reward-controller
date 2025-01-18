package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"github.com/ZnNr/user-task-reward-controller/internal/service/refercode"
)

const (
	// Запрос для создания пользователя
	CreateUserQuery = `
    INSERT INTO users (Username, Password, Refer_code) VALUES ($1, $2, $3) RETURNING ID`

	// Проверка существования пользователя
	CheckUserExistsQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE Username = $1 OR Email = $2)`

	// Получение пользователя по имени пользователя и паролю
	GetUserQuery = `SELECT user_id, username FROM users WHERE Username = $1 AND Password = $2`
)

// PostgresUserRepository реализует репозиторий пользователей для PostgreSQL
type PostgresAuthRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository создает новый экземпляр репозитория пользователей
func NewPostgresAuthRepository(db *sql.DB) repository.AuthRepository {
	return &PostgresAuthRepository{db: db}
}

// CreateUser создает нового пользователя
func (r *PostgresAuthRepository) CreateUser(ctx context.Context, user *models.CreateUser) (int64, error) {
	// Проверка существования пользователя
	exists, err := r.checkUserExists(ctx, user)
	if err != nil {
		return 0, errors.NewValidation("Can't check user existence", err)
	}
	if exists {
		return 0, errors.NewAlreadyExists("User already exists", nil)
	}

	// Генерация реферального кода
	referCode := refercode.RandStringBytes()

	// Подготовка SQL-запроса
	stmt, err := r.db.PrepareContext(ctx, CreateUserQuery)
	if err != nil {
		return 0, errors.NewInternal("Failed to prepare statement", err)
	}
	defer stmt.Close()

	// Выполнение SQL-запроса
	res, err := stmt.ExecContext(ctx, user.Username, user.Password, referCode)
	if err != nil {
		return 0, errors.NewInternal("Failed to execute query to create user", err)
	}

	// Получение ID созданного пользователя
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, errors.NewInternal("Failed to retrieve last insert ID", err)
	}

	return lastID, nil
}

// checkUserExists проверяет, существует ли пользователь с указанным именем пользователя или электронной почтой
func (r *PostgresAuthRepository) checkUserExists(ctx context.Context, user *models.CreateUser) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, CheckUserExistsQuery, user.Username, user.Email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// GetUser возвращает пользователя по имени и паролю
func (r *PostgresAuthRepository) GetUser(ctx context.Context, req *models.SignIn) (models.User, error) {
	row := r.db.QueryRowContext(ctx, GetUserQuery, req.Username, req.Password)

	var user models.User
	err := row.Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.NewNotFound("User not found", err)
		}
		return models.User{}, errors.NewInternal("Error fetching user", err)
	}

	return user, nil
}
