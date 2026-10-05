package models

import (
	"errors"
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
	FreeTrialLimit        uint8     `json:"freeTrialLimit" gorm:"not null;default:0" example:"4"`
} //@name Semester

const MaxSemesterCalendarDays = 366

var ErrSemesterDateRange = errors.New("semester date range must contain at most 366 inclusive UTC calendar days")

// SemesterCalendarDays counts the UTC calendar dates represented by the
// timestamptz semester bounds. The signup timeline includes both endpoints.
func SemesterCalendarDays(start, end time.Time) int {
	startUTC := start.UTC()
	startDate := time.Date(startUTC.Year(), startUTC.Month(), startUTC.Day(), 0, 0, 0, 0, time.UTC)
	endUTC := end.UTC()
	endDate := time.Date(endUTC.Year(), endUTC.Month(), endUTC.Day(), 0, 0, 0, 0, time.UTC)
	if endDate.Before(startDate) {
		return 0
	}
	return int(endDate.Sub(startDate)/(24*time.Hour)) + 1
}

// ValidateSemesterDateRange ensures signup timeline generation stays bounded.
func ValidateSemesterDateRange(start, end time.Time) error {
	if !end.After(start) || SemesterCalendarDays(start, end) > MaxSemesterCalendarDays {
		return ErrSemesterDateRange
	}
	return nil
}

type CreateSemesterRequest struct {
	Name                  string    `json:"name" binding:"required" example:"Fall 2023"`
	Meta                  string    `json:"meta"`
	StartDate             time.Time `json:"startDate" binding:"required" example:"2023-09-01T00:00:00Z"`
	EndDate               time.Time `json:"endDate" binding:"required,gtfield=StartDate" example:"2023-12-31T23:59:59Z"`
	StartingBudget        float32   `json:"startingBudget" binding:"omitempty,gte=0" example:"100.00"`
	MembershipFee         *uint8    `json:"membershipFee" binding:"required,gte=0" example:"10"`
	MembershipDiscountFee *uint8    `json:"membershipDiscountFee" binding:"required,gte=0" example:"5"`
	RebuyFee              *uint8    `json:"rebuyFee" binding:"required,gte=0" example:"2"`
	FreeTrialLimit        uint8     `json:"freeTrialLimit" binding:"omitempty,gte=0" example:"4"`
} //@name CreateSemesterRequest

func (s Semester) TableName() string {
	return "semesters"
}
