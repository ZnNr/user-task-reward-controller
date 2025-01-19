package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	//"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"github.com/ZnNr/user-task-reward-controller/internal/service/refercode"
	"log"
)

// SQL Queries
const (
	addTaskQuery            = `INSERT INTO tasks (title, description, price) VALUES ($1, $2, $3) RETURNING task_id`
	checkTaskDuplicateQuery = `SELECT COUNT(*) FROM tasks WHERE title = $1 AND description = $2 AND task_id <> $3`
	completeTaskQuery       = `SELECT task_id, price FROM tasks WHERE id=$1`
	userQuery               = `SELECT task_id, balance, refer_from FROM users WHERE task_id=$1`
	completeQuery           = `INSERT INTO task_complete(user_id, task_id) VALUES ($1, $2)`
	balanceToUserQuery      = `UPDATE users SET balance=balance+$1 WHERE id=$2`
)

// TaskRepository для работы с задачами
type PostgresTaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository with a given database connection.
func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (t *PostgresTaskRepository) CreateTask(ctx context.Context, task *models.TaskCreate) (int64, error) {
	// Проверяем на дубликаты
	if isDuplicate, err := t.checkForDuplicateTask(ctx, task, 0); err != nil {
		return 0, err
	} else if isDuplicate {
		return 0, errors.NewAlreadyExists("task with the same title and description already exists", nil)
	}

	stmt, err := t.db.PrepareContext(ctx, addTaskQuery)
	if err != nil {
		return 0, errors.NewInternal("Cannot prepare statement", err)
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Printf("Error closing statement: %v", closeErr)
		}
	}()

	// Передаем аргументы без указателей
	res, err := stmt.ExecContext(ctx, task.Title, task.Description, task.Price)
	if err != nil {
		return 0, errors.NewInternal("Cannot create task", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, errors.NewInternal("Cannot retrieve last insert ID", err)
	}

	return lastID, nil
}

// checkForDuplicateTask проверяет наличие дубликатов задач по заголовку и описанию.
func (r *PostgresTaskRepository) checkForDuplicateTask(ctx context.Context, task *models.TaskCreate, excludeID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		checkTaskDuplicateQuery,
		task.Title,
		task.Description,
		excludeID).Scan(&count)
	if err != nil {
		return false, errors.NewInternal("failed to check for duplicate task", err)
	}
	return count > 0, nil
}

func (r *PostgresTaskRepository) CompleteTask(userId, taskId int64) error {
	// Проверяем, существует ли задача
	var task models.Task

	err := r.db.QueryRow(completeTaskQuery, taskId).Scan(&task.TaskID, &task.Price)
	if err != nil {
		return errors.NewNotFound(fmt.Sprintf("task with id %d not found", taskId), err)
	}

	// Проверяем, существует ли пользователь
	var user models.User

	err = r.db.QueryRow(userQuery, userId).Scan(&user.ID, &user.Balance, &user.ReferFrom)
	if err != nil {
		return errors.NewNotFound(fmt.Sprintf("user with id %d not found", userId), err)
	}

	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return errors.NewInternal("failed to begin transaction", err)
	}

	// Выполняем запись о завершении задачи
	if _, err = tx.Exec(completeQuery, userId, taskId); err != nil {
		tx.Rollback()
		return errors.NewInternal("failed to complete task", err)
	}

	// Обновляем баланс пользователя
	if _, err = tx.Exec(balanceToUserQuery, task.Price, userId); err != nil {
		tx.Rollback()
		return errors.NewInternal("failed to update user balance", err)
	}

	// Если у пользователя есть реферал, выплачиваем бонус
	if user.ReferFrom != nil {
		if err := r.referralReward(user.ReferFrom, task.Price); err != nil {
			log.Println(err.Error())
		}
	}

	// Подтверждаем транзакцию
	return tx.Commit()
}

func (r *PostgresTaskRepository) GetAllTasks() ([]models.Task, error) {
	// Подготовка SQL-запроса
	rows, err := r.db.Query("SELECT id, title, description FROM tasks")
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task

	// Сканирование результатов
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.TaskID, &task.Title, &task.Description); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	// Проверка на наличие ошибок после завершения итерации
	if err := rows.Err(); err != nil {
		log.Println("Error encountered during rows iteration:", err)
		return nil, err
	}

	return tasks, nil
}

func (r *PostgresTaskRepository) referralReward(ref_id *int, price int) error {
	var refId int
	if err := r.db.QueryRow("SELECT id FROM users WHERE id=$1", *ref_id).Scan(&refId); err != nil {
		return errors.NewNotFound(fmt.Sprintf("user with id \"%d\" not found", *ref_id), err)
	}

	rewardCount := refercode.Reward(price)
	_, err := r.db.Exec("UPDATE users SET balance=balance+$1 WHERE id=$2", rewardCount, *ref_id)
	if err != nil {
		return errors.NewInternal("failed to update referrer balance", err)
	}

	log.Printf("user %d Referral reward: %d", refId, rewardCount)
	return nil
}

func (r *PostgresTaskRepository) ReferrerCode(userId int64, refCode string) error {
	var refId int
	err := r.db.QueryRow("SELECT id FROM users WHERE refer_code=$1", refCode).Scan(&refId)
	if err != nil {
		return errors.NewNotFound(fmt.Sprintf("user with refer_code \"%s\" not found", refCode), err)
	}

	if _, err = r.db.Exec("UPDATE users SET refer_from=$1 WHERE id=$2", refId, userId); err != nil {
		return errors.NewInternal("failed to set referrer code", err)
	}

	return nil
}
