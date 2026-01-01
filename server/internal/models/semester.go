package models

import (
	"time"

	"github.com/google/uuid"
)

type Semester struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name                  string    `json:"name" example:"Fall 2023"`
	Meta                  string    `json:"meta"`
	StartDate             time.Time `json:"startDate" gorm:"not null;default:CURRENT_TIMESTAMP" example:"2023-09-01T00:00:00Z"`
	EndDate               time.Time `json:"endDate" gorm:"not null;default:CURRENT_TIMESTAMP" example:"2023-12-31T23:59:59Z"`
	StartingBudget        float32   `json:"startingBudget" gorm:"not null;default:0" example:"100.00"`
	CurrentBudget         float32   `json:"currentBudget" gorm:"not null;default:0" example:"100.00"`
	MembershipFee         uint8     `json:"membershipFee" gorm:"not null;default:0" example:"10"`
	MembershipDiscountFee uint8     `json:"membershipDiscountFee" gorm:"not null;default:0" example:"5"`
	RebuyFee              uint8     `json:"rebuyFee" gorm:"not null;default:0" example:"2"`
} //@name Semester

type CreateSemesterRequest struct {
	Name                  string    `json:"name" binding:"required" example:"Fall 2023"`
	Meta                  string    `json:"meta"`
	StartDate             time.Time `json:"startDate" binding:"required" example:"2023-09-01T00:00:00Z"`
	EndDate               time.Time `json:"endDate" binding:"required,gtfield=StartDate" example:"2023-12-31T23:59:59Z"`
	StartingBudget        float32   `json:"startingBudget" binding:"omitempty,gte=0" example:"100.00"`
	MembershipFee         uint8     `json:"membershipFee" binding:"required,gte=0" example:"10"`
	MembershipDiscountFee uint8     `json:"membershipDiscountFee" binding:"required,gte=0" example:"5"`
	RebuyFee              uint8     `json:"rebuyFee" binding:"required,gte=0" example:"2"`
} //@name CreateSemesterRequest

func (s Semester) TableName() string {
	return "semesters"
}
