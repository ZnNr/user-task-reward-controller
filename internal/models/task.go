package models

type Task struct {
	TaskID      int64  `json:"task_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price" db:"price"`
}

type TaskCreate struct {
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description" db:"description"`
	Price       int    `json:"price" db:"price"`
}
