package models

type User struct {
	ID        int     `json:"id" db:"id"`
	Username  string  `json:"username" db:"username"`
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
