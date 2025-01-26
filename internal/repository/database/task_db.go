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

// SQL Queries
const (
	addTaskQuery            = `INSERT INTO tasks (title, description, price) VALUES ($1, $2, $3) RETURNING task_id`
	checkTaskDuplicateQuery = `SELECT COUNT(*) FROM tasks WHERE title = $1 AND description = $2 AND task_id <> $3`
	completeTaskQuery       = `SELECT task_id, price FROM tasks WHERE task_id=$1`
	userQuery               = `SELECT user_id, balance, refer_from FROM users WHERE user_id=$1`
	completeQuery           = `INSERT INTO task_complete(user_id, task_id) VALUES ($1, $2)`
	balanceToUserQuery      = `UPDATE users SET balance=balance+$1 WHERE user_id=$2`
)

// TaskRepository для работы с задачами
type PostgresTaskRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresTaskRepository создает новый экземпляр репозитория задач
func NewPostgresTaskRepository(db *sql.DB, logger *zap.Logger) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db, logger: logger}
}

// executeQuery выполняет SQL-запрос и возвращает результат
func (r *PostgresTaskRepository) executeQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return nil, errors.NewInternal("Failed to execute query", err)
	}
	return rows, nil
}

// executeQueryRow выполняет SQL-запрос и сканирует результат
func (r *PostgresTaskRepository) executeQueryRow(ctx context.Context, query string, args ...interface{}) error {
	row := r.db.QueryRowContext(ctx, query, args...)
	return row.Err()
}

// executeExec выполняет SQL-запрос на изменение данных и возвращает количество затронутых строк
func (r *PostgresTaskRepository) executeExec(ctx context.Context, query string, args ...interface{}) (int64, error) {
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

// CreateTask создает новую задачу
func (r *PostgresTaskRepository) CreateTask(ctx context.Context, task *models.TaskCreate) (int64, error) {
	// Проверяем на дубликаты
	if isDuplicate, err := r.checkForDuplicateTask(ctx, task, 0); err != nil {
		return 0, err
	} else if isDuplicate {
		return 0, errors.NewAlreadyExists("task with the same title and description already exists", nil)
	}
	var lastID int64
	err := r.db.QueryRowContext(ctx, addTaskQuery, task.Title, task.Description, task.Price).Scan(&lastID)
	if err != nil {
		r.logger.Error("Cannot create task", zap.Error(err))
		return 0, errors.NewInternal("Cannot create task", err)
	}
	return lastID, nil
}

// checkForDuplicateTask проверяет наличие дубликатов задач по заголовку и описанию.
func (r *PostgresTaskRepository) checkForDuplicateTask(ctx context.Context, task *models.TaskCreate, excludeID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, checkTaskDuplicateQuery, task.Title, task.Description, excludeID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to check for duplicate task", zap.Error(err))
		return false, errors.NewInternal("failed to check for duplicate task", err)
	}
	return count > 0, nil
}

// CompleteTask завершает задачу и обновляет баланс пользователя
func (r *PostgresTaskRepository) CompleteTask(ctx context.Context, userId, taskId int64) error {
	// Проверяем, существует ли задача
	var task models.Task
	err := r.db.QueryRowContext(ctx, completeTaskQuery, taskId).Scan(&task.TaskID, &task.Price)
	if err != nil {
		r.logger.Info("task not found", zap.Int64("task_id", taskId), zap.Error(err))
		return errors.NewNotFound(fmt.Sprintf("task with id %d not found", taskId), err)
	}

	// Проверяем, существует ли пользователь
	var user models.User
	err = r.db.QueryRowContext(ctx, userQuery, userId).Scan(&user.ID, &user.Balance, &user.ReferFrom)
	if err != nil {
		r.logger.Info("user not found", zap.Int64("user_id", userId), zap.Error(err))
		return errors.NewNotFound(fmt.Sprintf("user with id %d not found", userId), err)
	}

	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return errors.NewInternal("failed to begin transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Выполняем запись о завершении задачи
	rowsAffected, err := r.executeExec(ctx, completeQuery, userId, taskId)
	if err != nil || rowsAffected == 0 {
		r.logger.Error("failed to complete task", zap.Int64("user_id", userId), zap.Int64("task_id", taskId), zap.Error(err))
		return errors.NewInternal("failed to complete task", err)
	}

	// Обновляем баланс пользователя
	rowsAffected, err = r.executeExec(ctx, balanceToUserQuery, task.Price, userId)
	if err != nil || rowsAffected == 0 {
		r.logger.Error("failed to update user balance", zap.Int64("user_id", userId), zap.Int("price", int(task.Price)), zap.Error(err))
		return errors.NewInternal("failed to update user balance", err)
	}

	// Если у пользователя есть реферал, выплачиваем бонус
	if user.ReferFrom != nil {
		if err := r.referralReward(user.ReferFrom, int(task.Price)); err != nil {
			r.logger.Error("failed to process referral reward", zap.Int("refer_id", *user.ReferFrom), zap.Error(err))
		}
	}
	return nil
}

// GetAllTasks возвращает все задачи
func (r *PostgresTaskRepository) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	rows, err := r.executeQuery(ctx, "SELECT task_id, title, description FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.TaskID, &task.Title, &task.Description); err != nil {
			r.logger.Error("Error scanning row", zap.Error(err))
			return nil, errors.NewInternal("Error scanning row", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error encountered during rows iteration", zap.Error(err))
		return nil, errors.NewInternal("Error encountered during rows iteration", err)
	}
	return tasks, nil
}

// referralReward выплачивает бонус за реферальную программу
func (r *PostgresTaskRepository) referralReward(refer_id *int, price int) error {
	var refId int
	if err := r.db.QueryRow("SELECT user_id FROM users WHERE user_id=$1", *refer_id).Scan(&refId); err != nil {
		r.logger.Error("user not found for referral reward", zap.Int("refer_id", *refer_id), zap.Error(err))
		return errors.NewNotFound(fmt.Sprintf("user with id \"%d\" not found", *refer_id), err)
	}
	rewardCount := refercode.Reward(price)
	_, err := r.db.Exec("UPDATE users SET balance=balance+$1 WHERE user_id=$2", rewardCount, *refer_id)
	if err != nil {
		r.logger.Error("failed to update referrer balance", zap.Int("refer_id", refId), zap.Error(err))
		return errors.NewInternal("failed to update referrer balance", err)
	}
	r.logger.Info("Referral reward processed", zap.Int("refer_id", refId), zap.Int("reward", rewardCount))
	return nil
}
