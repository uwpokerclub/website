package models

import "github.com/google/uuid"

type Membership struct {
	ID         string `gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID     uint64
	SemesterID uuid.UUID `gorm:"type:uuid"`
	Paid       bool
	Discounted bool
}
