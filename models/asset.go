package models

import (
	"time"
)

type AssetSchema struct {
	ID        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Account   string    `json:"account"`
	Amount    string    `json:"amount"`
	Date      time.Time `json:"date"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
