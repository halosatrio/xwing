package db

import (
	"time"
)

type UserSchema struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransactionSchema struct {
	ID        int       `json:"id"`
	UserId    string    `json:"user_id"`
	Type      string    `json:"type"`
	Amount    string    `json:"amount"`
	Category  string    `json:"category"`
	Date      time.Time `json:"date"`
	Notes     string    `json:"notes"`
	IsActive  string    `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
