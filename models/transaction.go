package models

import (
	"time"
)

type TransactionSchema struct {
	ID        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Type      string    `json:"type"`
	Amount    string    `json:"amount"`
	Category  string    `json:"category"`
	Date      time.Time `json:"date"`
	Notes     string    `json:"notes"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
