package services

import (
	"api/internal/database"
	"api/internal/models"
	"api/internal/testhelpers"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRankingService_UpdateRanking_NoPreviousRanking(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
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
		t.Fatal(err.Error())
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

func TestRankingService_GetRanking(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	// Create existing users
	user1 := models.User{
		ID:        1,
		FirstName: "adam",
		LastName:  "mahood",
		Email:     "adam@gmail.com",
		Faculty:   "Math",
		QuestID:   "asmahood",
	}
	res := db.Create(&user1)
	if res.Error != nil {
		t.Fatalf("Error when creating existing users: %v", res.Error)
	}

	user2 := models.User{
		ID:        2,
		FirstName: "john",
		LastName:  "doe",
		Email:     "john@gmail.com",
		Faculty:   "Science",
		QuestID:   "jdoe",
	}
	res = db.Create(&user2)
	if res.Error != nil {
		t.Fatalf("Error when creating existing users: %v", res.Error)
	}

	user3, err := testhelpers.CreateUser(db, 3, "jane", "doe", "jdoe@gmail.com", "Arts", "jdoe234")
	assert.NoError(t, err)

	// Create semester
	semester1 := models.Semester{
		Name:                  "Winter 2022",
		Meta:                  "",
		StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		StartingBudget:        100.54,
		MembershipFee:         10,
		MembershipDiscountFee: 5,
		RebuyFee:              2,
	}
	res = db.Create(&semester1)
	if res.Error != nil {
		t.Fatalf("Error when creating existing semester: %v", res.Error)
	}

	// Create semester
	semester2 := models.Semester{
		Name:                  "Fall 2022",
		Meta:                  "",
		StartDate:             time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:               time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC),
		StartingBudget:        100.54,
		MembershipFee:         10,
		MembershipDiscountFee: 5,
		RebuyFee:              2,
	}
	res = db.Create(&semester2)
	if res.Error != nil {
		t.Fatalf("Error when creating existing semester: %v", res.Error)
	}

	// Create memberships for this semester
	membership1 := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	membership2 := models.Membership{
		UserID:     user2.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: true,
	}
	membership3 := models.Membership{
		UserID:     user3.ID,
		SemesterID: semester2.ID,
		Paid:       true,
		Discounted: true,
	}
	res = db.Create(&membership1)
	if res.Error != nil {
		t.Fatalf("Error when creating memberships: %v", res.Error)
	}
	res = db.Create(&membership2)
	if res.Error != nil {
		t.Fatalf("Error when creating memberships: %v", res.Error)
	}
	res = db.Create(&membership3)
	if res.Error != nil {
		t.Fatalf("Error when creating memberships: %v", res.Error)
	}

	// Add rankings
	r := []models.Ranking{
		{
			MembershipID: membership1.ID,
			Points:       10,
		},
		{
			MembershipID: membership2.ID,
			Points:       5,
		},
		{
			MembershipID: membership3.ID,
			Points:       11,
		},
	}
	res = db.Create(&r)
	if res.Error != nil {
		t.Fatalf("Failed to create rankings: %v", res.Error)
	}

	expected1 := models.GetRankingResponse{
		Points:   10,
		Position: 1,
	}
	expected2 := models.GetRankingResponse{
		Points:   5,
		Position: 2,
	}

	svc := NewRankingService(db)

	ranking1, err := svc.GetRanking(semester1.ID, membership1.ID)
	if assert.NoError(t, err, "GetRanking should not error on member 1") {
		assert.Equal(t, expected1, *ranking1)
	}

	ranking2, err := svc.GetRanking(semester1.ID, membership2.ID)
	if assert.NoError(t, err, "GetRanking should not error on member 2") {
		assert.Equal(t, expected2, *ranking2)
	}
}

func TestRankingService_GetRanking_TiedRankings(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer database.WipeDB(db)

	// Create users
	user1, err := testhelpers.CreateUser(db, 1, "alice", "smith", "alice@gmail.com", "Math", "asmith")
	assert.NoError(t, err)
	user2, err := testhelpers.CreateUser(db, 2, "bob", "jones", "bob@gmail.com", "Science", "bjones")
	assert.NoError(t, err)
	user3, err := testhelpers.CreateUser(db, 3, "carol", "white", "carol@gmail.com", "Arts", "cwhite")
	assert.NoError(t, err)

	// Create semester
	semester := models.Semester{
		Name:                  "Winter 2023",
		Meta:                  "",
		StartDate:             time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:               time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
		StartingBudget:        100.00,
		MembershipFee:         10,
		MembershipDiscountFee: 5,
		RebuyFee:              2,
	}
	res := db.Create(&semester)
	if res.Error != nil {
		t.Fatalf("Error when creating semester: %v", res.Error)
	}

	// Create memberships
	membership1 := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester.ID,
		Paid:       true,
		Discounted: false,
	}
	membership2 := models.Membership{
		UserID:     user2.ID,
		SemesterID: semester.ID,
		Paid:       true,
		Discounted: false,
	}
	membership3 := models.Membership{
		UserID:     user3.ID,
		SemesterID: semester.ID,
		Paid:       true,
		Discounted: false,
	}
	res = db.Create(&membership1)
	assert.NoError(t, res.Error)
	res = db.Create(&membership2)
	assert.NoError(t, res.Error)
	res = db.Create(&membership3)
	assert.NoError(t, res.Error)

	// Add rankings with tied scores: 100, 100, 50
	rankings := []models.Ranking{
		{MembershipID: membership1.ID, Points: 100},
		{MembershipID: membership2.ID, Points: 100},
		{MembershipID: membership3.ID, Points: 50},
	}
	res = db.Create(&rankings)
	assert.NoError(t, res.Error)

	svc := NewRankingService(db)

	// Both 100-point members should have position 1
	ranking1, err := svc.GetRanking(semester.ID, membership1.ID)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), ranking1.Points)
	assert.Equal(t, int32(1), ranking1.Position, "First 100-point member should be rank 1")

	ranking2, err := svc.GetRanking(semester.ID, membership2.ID)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), ranking2.Points)
	assert.Equal(t, int32(1), ranking2.Position, "Second 100-point member should be rank 1")

	// 50-point member should have position 3 (not 2)
	ranking3, err := svc.GetRanking(semester.ID, membership3.ID)
	assert.NoError(t, err)
	assert.Equal(t, int32(50), ranking3.Points)
	assert.Equal(t, int32(3), ranking3.Position, "50-point member should be rank 3 (skipping rank 2)")
}
