package models

import "github.com/google/uuid"

type Transaction struct {
	ID          int32     `json:"id" gorm:"type:integer;primaryKey;autoIncrement"`
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
	ID          int32
	Amount      float32 `json:"amount"`
	Description string  `json:"description"`
}
