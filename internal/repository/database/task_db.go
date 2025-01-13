package database

import (
	"database/sql"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
)

// TaskRepository для работы с задачами
type PostgresTaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository with a given database connection.
func NewPostgresTaskRepository(db *sql.DB) repository.TaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (t *PostgresTaskRepository) CreateTask(task models.TaskCreate) (int, error) {
	var taskId int

	query := fmt.Sprintf("INSERT INTO %s (title, description, price) VALUES ($1, $2, $3) RETURNING id", taskTable)

	result := t.db.QueryRow(query, task.Title, task.Description, task.Price)
	if err := result.Scan(&taskId); err != nil {
		return 0, err
	}

	return taskId, nil
}

func (r *PostgresTaskRepository) CompleteTask(userId, taskId int) error {

	// check if task exists
	var task models.Task
	taskQuery := fmt.Sprintf("SELECT id, price FROM %s WHERE id=$1", taskTable)
	err := r.db.Get(&task, taskQuery, taskId)
	if err != nil {
		return fmt.Errorf("task with id %d not found", taskId)
	}

	// check if user exists
	var user models.User
	userQuery := fmt.Sprintf("SELECT id, balance, refer_from FROM %s WHERE id=$1", usersTable)
	err = r.db.Get(&user, userQuery, userId)
	if err != nil {
		return fmt.Errorf("user with id %d not found", userId)
	}

	// update balance and refer_from balance
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	completeQuery := fmt.Sprintf("INSERT INTO %s (user_id, task_id) VALUES ($1, $2)", completeTable)
	_, err = tx.Exec(completeQuery, userId, taskId)
	if err != nil {
		tx.Rollback()
		return err
	}

	balanceToUserQuery := fmt.Sprintf("UPDATE %s SET balance=balance+$1 WHERE id=$2", usersTable)
	_, err = tx.Exec(balanceToUserQuery, task.Price, userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Если пользователь имеет реф источник - бонус выдать реферу
	if user.ReferFrom != nil {
		err := r.referralReward(user.ReferFrom, task.Price)
		if err != nil {
			logrus.Println(err.Error())
		}
	}

	return tx.Commit()
}

func (r *PostgresTaskRepository) GetAllTasks() ([]models.Task, error) {
	var tasks []models.Task

	query := fmt.Sprintf("SELECT id, title, description, price FROM %s ORDER BY id", taskTable)
	err := r.db.Select(&tasks, query)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *PostgresTaskRepository) referralReward(ref_id *int, price int) error {
	var refId int
	referalQuery := fmt.Sprintf("SELECT id FROM %s WHERE id=$1", usersTable)
	if err := r.db.Get(&refId, referalQuery, ref_id); err != nil {
		return fmt.Errorf("user with id \"%d\" not found", *ref_id)
	}

	rewardCount := utils.Reward(price)
	rewardQuery := fmt.Sprintf("UPDATE %s SET balance=balance+$1 WHERE id=$2", usersTable)
	_, err := r.db.Exec(rewardQuery, rewardCount, *ref_id)
	if err != nil {
		return err
	}

	logrus.Printf("user %d Referral reward: %d", refId, rewardCount)
	return nil
}

func (r *PostgresTaskRepository) ReferrerCode(userId int, refCode string) error {
	var refId int
	referQuery := fmt.Sprintf("SELECT id FROM %s WHERE refer_code=$1", usersTable)
	err := r.db.Get(&refId, referQuery, refCode)
	if err != nil {
		return fmt.Errorf("user with refer_code \"%s\" not found", refCode)
	}

	query := fmt.Sprintf("UPDATE %s SET refer_from=$1 WHERE id=$2", usersTable)
	_, err = r.db.Exec(query, refId, userId)
	if err != nil {
		return err
	}

	return nil
}
