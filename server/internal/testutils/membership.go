package testutils

import (
	"api/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var TEST_MEMBERSHIPS = []models.Membership{
	{
		ID:         uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID:     TEST_USERS[0].ID,
		SemesterID: TEST_SEMESTERS[0].ID,
		Paid:       true,
		Discounted: false,
	},
	{
		ID:         uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		UserID:     TEST_USERS[1].ID,
		SemesterID: TEST_SEMESTERS[0].ID,
		Paid:       true,
		Discounted: true,
	},
	{
		ID:         uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		UserID:     TEST_USERS[2].ID,
		SemesterID: TEST_SEMESTERS[0].ID,
		Paid:       true,
		Discounted: false,
	},
}

func SeedMemberships(db *gorm.DB, seedDependencies bool) error {
	// Ensure semesters are seeded first if requested
	if seedDependencies {
		if err := SeedSemesters(db); err != nil {
			return err
		}

		// Ensure users are seeded first
		if err := SeedUsers(db); err != nil {
			return err
		}
	}

	for _, membership := range TEST_MEMBERSHIPS {
		if err := db.Create(&membership).Error; err != nil {
			return err
		}
	}

	return nil
}

func CreateTestMembership(db *gorm.DB, userId uint64, semesterId uuid.UUID) (*models.Membership, error) {
	membership := models.Membership{
		UserID:     userId,
		SemesterID: semesterId,
		Paid:       true,
		Discounted: false,
	}

	if err := db.Create(&membership).Error; err != nil {
		return nil, err
	}

	return &membership, nil
}

func FindMembershipByID(id uuid.UUID) (*models.Membership, error) {
	for _, membership := range TEST_MEMBERSHIPS {
		if membership.ID == id {
			return &membership, nil
		}
	}
	return nil, fmt.Errorf("membership not found")
}
