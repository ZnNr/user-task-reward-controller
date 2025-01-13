package models

import "fmt"

type Task struct {
	TaskID      int    `json:"task_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price" db:"price"`
	//AssigneeID  *string    `json:"assignee_id,omitempty"`
}

type TaskCreate struct {
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description" db:"description"`
	Price       int    `json:"price" db:"price"`
}

func (t *TaskCreate) Validate() error {
	if t.Price < 1 {
		return fmt.Errorf("minimum value for the Price field is 1")
	}

	return nil
}
