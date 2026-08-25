package models

import "github.com/google/uuid"

// Transaction is retained deliberately even though the transactions feature was
// removed in issue #358. Atlas generates migrations by diffing every GORM model
// under ./internal/models against the live schema (see atlas/atlas.hcl, which
// loads the package by path). Deleting this struct would make the next
// `make generate-migration` emit a DROP TABLE transactions and destroy the
// existing rows, which are being kept for a future money-management feature.
//
// Do not delete this file without first writing an intentional Atlas migration.
type Transaction struct {
	ID          int32     `json:"id" gorm:"type:integer;primaryKey;autoIncrement"`
	SemesterID  uuid.UUID `json:"semesterId" gorm:"type:uuid"`
	Semester    Semester  `json:"semester"`
	Amount      float32   `json:"amount" gorm:"not null;default:0"`
	Description string    `json:"description"`
}
