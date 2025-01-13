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

// SQL-запросы
const (
	// Запрос для создания пользователя
	CreateUserQuery = `
	INSERT INTO users (
		Username,
		Password,
		Email,
		Balance,
		Refer_code,
		Refer_from
	) VALUES ($1, $2, $3, $4, $5, $6)`

	// Проверка существования пользователя
	CheckUserExistsQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE Username = $1 OR Email = $2)`

	// Получение пользователя по имени пользователя и паролю
	GetUserQuery = `SELECT id, username FROM users WHERE Username = $1 AND Password = $2`

	// Получение таблицы лидеров по балансу
	GetLeaderboardByBalanceQuery = `SELECT ID, Username, Balance, Refer_code, Refer_from FROM users ORDER BY Balance DESC`

	// Получение информации о пользователе по ID
	GetUserByIDQuery = `
	SELECT 
		ID, Username, Email, Balance, Refer_code, Refer_from, TasksCompleted, CreatedAt, UpdatedAt 
	FROM 
		users 
	WHERE 
		ID = $1`
)

// PostgresUserRepository реализует репозиторий пользователей для PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository создает новый экземпляр репозитория пользователей
func NewPostgresUserRepository(db *sql.DB) repository.UserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser создает нового пользователя
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *models.CreateUser, referFrom string) (int64, error) {
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
	res, err := stmt.ExecContext(ctx, user.Username, user.Password, user.Email, user.Balance, referCode, referFrom)
	if err != nil {
		return 0, errors.NewInternal("Failed to execute query to create user", err)
	}

	// Получение ID созданного пользователя
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, errors.NewInternal("Failed to get last insert ID", err)
	}

	return lastID, nil
}

// checkUserExists проверяет, существует ли пользователь с указанным именем пользователя или электронной почтой
func (r *PostgresUserRepository) checkUserExists(ctx context.Context, user *models.CreateUser) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, CheckUserExistsQuery, user.Username, user.Email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// GetUserByID возвращает пользователя по имени и паролю
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, username, password string) (models.User, error) {
	row := r.db.QueryRowContext(ctx, GetUserQuery, username, password)

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

// GetUsersLeaderboard возвращает список пользователей, отсортированный по балансу
func (r *PostgresUserRepository) GetUsersLeaderboard(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx, GetLeaderboardByBalanceQuery)
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
func (r *PostgresUserRepository) GetUserInfo(ctx context.Context, userID int64) (models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, GetUserByIDQuery, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Balance,
		&user.ReferCode,
		&user.ReferFrom,
		//&user.TasksCompleted,
		//&user.CreatedAt,
		//&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.NewNotFound("User not found", err)
		}
		return models.User{}, errors.NewInternal("Error fetching user info", err)
	}

	return user, nil
}
