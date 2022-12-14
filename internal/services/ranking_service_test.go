package services

import (
	"api/internal/database"
	"api/internal/models"
	"api/internal/testhelpers"
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestRankingService_UpdateRanking_NoPreviousRanking(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	set, err := testhelpers.SetupSemester(db, "Fall 2022")
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	svc := NewRankingService(db)
	err = svc.UpdateRanking(set.Memberships[0].ID, 32)
	if err != nil {
		t.Errorf("UpdateRanking() error = %v", err)
		return
	}

	// Retrieve ranking to ensure it was created
	ranking := models.Ranking{MembershipID: set.Memberships[0].ID}
	res := db.First(&ranking)

	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("UpdateRanking() did not create a new ranking record: %v", err)
		return
	}

	if err := res.Error; err != nil {
		t.Errorf("Failed to retrieve ranking: %v", err)
		return
	}

	if ranking.Points != 32 {
		t.Errorf("Points = %v, expected = %v", ranking.Points, 32)
	}
}

func TestRankingService_UpdateRanking_PreviousRanking(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	set, err := testhelpers.SetupSemester(db, "Fall 2022")
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	ranking := models.Ranking{
		MembershipID: set.Memberships[0].ID,
		Points:       32,
	}
	res := db.Create(&ranking)
	if res.Error != nil {
		t.Fatalf("Failed to create ranking: %v", err)
	}

	svc := NewRankingService(db)
	err = svc.UpdateRanking(set.Memberships[0].ID, 16)

	if err != nil {
		t.Errorf("UpdateRanking() error = %v", err)
		return
	}

	// Retrieve ranking to ensure it was created
	found := models.Ranking{MembershipID: set.Memberships[0].ID}
	res = db.First(&found)

	if err := res.Error; err != nil {
		t.Fatalf("Failed to retrieve ranking: %v", err)
		return
	}

	if found.Points != 48 {
		t.Errorf("Points = %v, expected = %v", ranking.Points, 48)
	}
}
