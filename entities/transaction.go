package entities

import "gorm.io/gorm"

type Transaction struct {
	gorm.Model
	UserID uint   `json:"user_id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
}