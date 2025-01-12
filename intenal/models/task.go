package models

import "fmt"

type Task struct {
	ID          int    `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Price       int    `json:"price" db:"price"`
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
