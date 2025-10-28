package testutils

import (
	"api/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var TEST_SEMESTERS = []models.Semester{
	{
		ID:                    uuid.MustParse("f74be353-7855-4b05-875c-1b6eaa3cdcbf"),
		Name:                  "Fall 2023",
		Meta:                  "Earlier semester",
		StartDate:             time.Date(2023, 9, 1, 0, 0, 0, 0, time.Now().Local().Location()),
		EndDate:               time.Date(2023, 12, 15, 0, 0, 0, 0, time.Now().Local().Location()),
		StartingBudget:        50.0,
		CurrentBudget:         50.0,
		MembershipFee:         8,
		MembershipDiscountFee: 4,
		RebuyFee:              1,
	},
	{
		ID:                    uuid.MustParse("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
		Name:                  "Spring 2024",
		Meta:                  "Middle semester",
		StartDate:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.Now().Local().Location()),
		EndDate:               time.Date(2024, 4, 30, 0, 0, 0, 0, time.Now().Local().Location()),
		StartingBudget:        75.0,
		CurrentBudget:         75.0,
		MembershipFee:         12,
		MembershipDiscountFee: 6,
		RebuyFee:              3,
	},
	{
		ID:                    uuid.MustParse("b2c3d4e5-f6a7-8901-bcde-f12345678901"),
		Name:                  "Fall 2024",
		Meta:                  "Latest semester",
		StartDate:             time.Date(2024, 9, 1, 0, 0, 0, 0, time.Now().Local().Location()),
		EndDate:               time.Date(2024, 12, 15, 0, 0, 0, 0, time.Now().Local().Location()),
		StartingBudget:        100.0,
		CurrentBudget:         100.0,
		MembershipFee:         15,
		MembershipDiscountFee: 8,
		RebuyFee:              4,
	},
}

func SeedSemesters(db *gorm.DB) error {
	for _, semester := range TEST_SEMESTERS {
		if err := db.Create(&semester).Error; err != nil {
			return err
		}
	}

	return nil
}

func FindSemesterByID(id string) (*models.Semester, error) {
	for _, semester := range TEST_SEMESTERS {
		if semester.ID.String() == id {
			return &semester, nil
		}
	}
	return nil, fmt.Errorf("semester not found")
}

// CreateTestSemester creates a test semester
func CreateTestSemester(db *gorm.DB, name string) (*models.Semester, error) {
	semester := models.Semester{
		Name:                  name,
		Meta:                  "Test semester",
		StartDate:             time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:               time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
		StartingBudget:        100.0,
		CurrentBudget:         100.0,
		MembershipFee:         10,
		MembershipDiscountFee: 5,
		RebuyFee:              2,
	}

	if err := db.Create(&semester).Error; err != nil {
		return nil, err
	}

	return &semester, nil
}
