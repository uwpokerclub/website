package models

import (
	"time"

	"github.com/google/uuid"
)

type Semester struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                  string    `json:"name"`
	Meta                  string    `json:"meta"`
	StartDate             time.Time `json:"startDate" gorm:"not null;default:CURRENT_TIMESTAMP"`
	EndDate               time.Time `json:"endDate" gorm:"not null;default:CURRENT_TIMESTAMP"`
	StartingBudget        float32   `json:"startingBudget" gorm:"not null;default:0"`
	CurrentBudget         float32   `json:"currentBudget" gorm:"not null;default:0"`
	MembershipFee         uint8     `json:"membershipFee" gorm:"not null;default:0"`
	MembershipDiscountFee uint8     `json:"membershipDiscountFee" gorm:"not null;default:0"`
	RebuyFee              uint8     `json:"rebuyFee" gorm:"not null;default:0"`
}

type CreateSemesterRequest struct {
	Name                  string    `json:"name" binding:"required"`
	Meta                  string    `json:"meta"`
	StartDate             time.Time `json:"startDate" binding:"required"`
	EndDate               time.Time `json:"endDate" binding:"required"`
	StartingBudget        float32   `json:"startingBudget" binding:"gte=0,omitempty"`
	MembershipFee         uint8     `json:"membershipFee" binding:"required"`
	MembershipDiscountFee uint8     `json:"membershipDiscountFee" binding:"required"`
	RebuyFee              uint8     `json:"rebuyFee" binding:"required"`
}
