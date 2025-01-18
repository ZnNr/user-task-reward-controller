package models

type User struct {
	ID        int64   `json:"user_id" db:"user_id"`
	Username  string  `json:"username" db:"username"`
	Email     string  `json:"email" validate:"required,email"`
	Balance   int     `json:"balance" db:"Balance"`
	ReferCode *string `json:"refer_code" db:"refer_code"`
	ReferFrom *int    `json:"refer_from" db:"refer_from"`
}

// структура для входа в систему
type SignIn struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// структура для создания нового пользователя в системе
type CreateUser struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" validate:"required,email"`
}

// возвращает пользователя по имени и паролю
type SignUp struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
