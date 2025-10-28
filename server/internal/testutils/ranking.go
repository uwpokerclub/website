package testutils

import (
	"api/internal/models"

	"gorm.io/gorm"
)

var TEST_RANKINGS = []models.Ranking{
	{
		MembershipID: TEST_MEMBERSHIPS[0].ID,
		Points:       0,
		Attendance:   0,
	},
	{
		MembershipID: TEST_MEMBERSHIPS[1].ID,
		Points:       0,
		Attendance:   0,
	},
	{
		MembershipID: TEST_MEMBERSHIPS[2].ID,
		Points:       0,
		Attendance:   0,
	},
}

func SeedRankings(db *gorm.DB, seedDependencies bool) error {
	// Ensure memberships are seeded first if requested
	if seedDependencies {
		if err := SeedMemberships(db, true); err != nil {
			return err
		}
	}

	for _, ranking := range TEST_RANKINGS {
		if err := db.Create(&ranking).Error; err != nil {
			return err
		}
	}

	return nil
}
