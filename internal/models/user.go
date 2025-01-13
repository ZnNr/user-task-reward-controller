package models

type User struct {
	ID        int     `json:"id" db:"id"`
	Username  string  `json:"username" db:"username"`
	Email     string  `json:"Email" validate:"required,email"`
	Balance   int     `json:"balance" db:"balance"`
	ReferCode *string `json:"refer_code" db:"refer_code"`
	ReferFrom *int    `json:"refer_from" db:"refer_from"`
}

type SignUpInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignInInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateUser struct {
	ID        int     `json:"id" db:"id"`
	Username  string  `json:"username" db:"username"`
	Password  string  `json:"password" binding:"required"`
	Email     string  `json:"Email" validate:"required,email"`
	Balance   int     `json:"balance" db:"balance"`
	ReferCode *string `json:"refer_code" db:"refer_code"`
	ReferFrom *int    `json:"refer_from" db:"refer_from"`
}

type CheckUser struct {
	Username string `json:"username" db:"username"`
	Email    string `json:"Email" validate:"required,email"`
}
