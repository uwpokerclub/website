package models

import "github.com/google/uuid"

type Transaction struct {
	ID          uint32    `json:"id"`
	SemesterID  uuid.UUID `json:"semesterId" gorm:"type:uuid"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
}

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	Description string  `json:"description" binding:"required"`
}

type UpdateTransactionRequest struct {
	ID          uint32
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}
