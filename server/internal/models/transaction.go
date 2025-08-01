package models

import "github.com/google/uuid"

type Transaction struct {
	ID          uint      `json:"id" gorm:"type:serial;primaryKey"`
	SemesterID  uuid.UUID `json:"semesterId" gorm:"type:uuid"`
	Semester    Semester  `json:"semester"`
	Amount      float32   `json:"amount" gorm:"not null;default:0"`
	Description string    `json:"description"`
}

type CreateTransactionRequest struct {
	Amount      float32 `json:"amount" binding:"required"`
	Description string  `json:"description" binding:"required"`
}

type UpdateTransactionRequest struct {
	ID          uint
	Amount      float32 `json:"amount"`
	Description string  `json:"description"`
}
