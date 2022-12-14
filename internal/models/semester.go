package models

import (
	"time"

	"github.com/google/uuid"
)

type Semester struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	Name                  string    `json:"name"`
	Meta                  string    `json:"meta"`
	StartDate             time.Time `json:"startDate"`
	EndDate               time.Time `json:"endDate"`
	StartingBudget        float64   `json:"startingBudget"`
	CurrentBudget         float64   `json:"currentBudget"`
	MembershipFee         int8      `json:"membershipFee"`
	MembershipFeeDiscount int8      `json:"membershipFeeDiscount"`
	RebuyFee              int8      `json:"rebuyFee"`
}

type CreateSemesterRequest struct {
	Name                  string    `json:"name" binding:"required"`
	Meta                  string    `json:"meta"`
	StartDate             time.Time `json:"startDate" binding:"required"`
	EndDate               time.Time `json:"endDate" binding:"required"`
	StartingBudget        float64   `json:"startingBudget" binding:"gte=0,omitempty"`
	MembershipFee         int8      `json:"membershipFee" binding:"required"`
	MembershipDiscountFee int8      `json:"membershipDiscountFee" binding:"required"`
	RebuyFee              int8      `json:"rebuyFee" binding:"required"`
}
