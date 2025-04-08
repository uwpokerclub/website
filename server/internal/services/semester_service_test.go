package services

import (
	"api/internal/database"
	"api/internal/models"
	"api/internal/testhelpers"
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSemesterService(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "CreateSemester",
			test: CreateSemesterTest(),
		},
		{
			name: "GetSemester",
			test: GetSemesterTest(),
		},
		{
			name: "ListSemesters",
			test: ListSemesterTest(),
		},
		{
			name: "GetRankings",
			test: GetRankingsTest(),
		},
		{
			name: "UpdateBudget_Positive",
			test: UpdateBudget_Positive(),
		},
		{
			name: "UpdateBudget_Negative",
			test: UpdateBudget_Negative(),
		},
		{
			name: "ExportRankings",
			test: ExportRankingsTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func CreateSemesterTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		ss := NewSemesterService(db)

		req := &models.CreateSemesterRequest{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			MembershipFee:         10,
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}

		res, err := ss.CreateSemester(req)
		if err != nil {
			t.Errorf("SemesterService.CreateSemester() error = %v", err)
			return
		}

		if _, err := uuid.Parse(res.ID.String()); err != nil {
			t.Errorf("SemesterService.CreateSemester().ID = \"\", expected valid UUID")
			return
		}

		if res.Name != "Spring 2022" {
			t.Errorf("SemesterService.CreateSemester().Name = %v, wanted = \"Spring 2022\"", res.Name)
			return
		}

		if res.Meta != "" {
			t.Errorf("SemesterService.CreateSemester().Meta = %v, wanted = \"\"", res.Meta)
			return
		}

		if res.StartDate != time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC) {
			t.Errorf("SemesterService.CreateSemester().StartDate = %v, wanted = %v", res.StartDate, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))
			return
		}

		if res.EndDate != time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC) {
			t.Errorf("SemesterService.CreateSemester().EndDate = %v, wanted = %v", res.StartDate, time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC))
			return
		}

		if res.StartingBudget != 105.57 {
			t.Errorf("SemesterService.CreateSemester().StartingBudget = %v, wanted = %v", res.Meta, 105.57)
			return
		}

		if res.CurrentBudget != 105.57 {
			t.Errorf("SemesterService.CreateSemester().CurrentBudget = %v, wanted = %v", res.Meta, 105.57)
			return
		}

		if res.MembershipFee != 10 {
			t.Errorf("SemesterService.CreateSemester().MembershipFee = %v, wanted = %v", res.Meta, 10)
			return
		}

		if res.MembershipDiscountFee != 5 {
			t.Errorf("SemesterService.CreateSemester().MembershipDiscountFee = %v, wanted = %v", res.Meta, 5)
			return
		}

		if res.RebuyFee != 2 {
			t.Errorf("SemesterService.CreateSemester().RebuyFee = %v, wanted = %v", res.Meta, 2)
			return
		}
	}
}

func GetSemesterTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			MembershipFee:         10,
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		ss := NewSemesterService(db)

		acc, err := ss.GetSemester(semester1.ID)
		if err != nil {
			t.Errorf("SemesterService.GetSemester() error = %v", err)
			return
		}

		if !reflect.DeepEqual(*acc, semester1) {
			t.Errorf("SemesterService.GetSemester() = %v, wanted = %v", acc, semester1)
			return
		}
	}
}

func ListSemesterTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

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
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}
		semester2 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			MembershipFee:         10,
			MembershipDiscountFee: 7,
			RebuyFee:              2,
		}
		res = db.Create(&semester2)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		semester3 := models.Semester{
			Name:                  "Fall 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        140,
			MembershipFee:         10,
			MembershipDiscountFee: 7,
			RebuyFee:              1,
		}
		res = db.Create(&semester3)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		ss := NewSemesterService(db)

		semesters, err := ss.ListSemesters()
		if err != nil {
			t.Errorf("SemesterService.ListSemesters() error = %v", err)
			return
		}

		exp := []models.Semester{semester3, semester2, semester1}
		if !reflect.DeepEqual(semesters, exp) {
			t.Errorf("SemesterService.ListSemesters() = %v, wanted = %v", semesters, exp)
			return
		}
	}
}

func GetRankingsTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
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
		res = db.Create(&membership1)
		if res.Error != nil {
			t.Fatalf("Error when creating memberships: %v", res.Error)
		}
		res = db.Create(&membership2)
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
		}
		res = db.Create(&r)
		if res.Error != nil {
			t.Fatalf("Failed to create rankings: %v", res.Error)
		}

		ss := NewSemesterService(db)

		rankings, err := ss.GetRankings(semester1.ID)
		if err != nil {
			t.Errorf("SemesterService.GetRankings() error = %v", err)
			return
		}

		if len(rankings) != 2 {
			t.Errorf("SemesterService.GetRankings() length = %d, expected = %d", len(rankings), 2)
			return
		}

		if rankings[0].ID != user1.ID {
			t.Errorf("SemesterService.GetRankings() order not correct, got = %v, wanted = %v", rankings[0], user1)
			return
		}

		if rankings[1].ID != user2.ID {
			t.Errorf("SemesterService.GetRankings() order not correct, got = %v, wanted = %v", rankings[1], user2)
			return
		}
	}
}

func UpdateBudget_Positive() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			CurrentBudget:         105.57,
			MembershipFee:         10,
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		ss := NewSemesterService(db)

		err = ss.UpdateBudget(semester1.ID, 10.30)
		if err != nil {
			t.Errorf("SemesterService.UpdateBudget() error = %v", err)
			return
		}

		newSem := models.Semester{ID: semester1.ID}
		res = db.First(&newSem)
		if res.Error != nil {
			t.Fatalf("Error when retrieving semester: %v", res.Error)
			return
		}

		if !almostEqual(newSem.CurrentBudget, 115.87) {
			t.Errorf("SemesterService.UpdateBudget() = %v, expected = %v", newSem.CurrentBudget, 115.87)
			return
		}
	}
}

func UpdateBudget_Negative() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			CurrentBudget:         105.57,
			MembershipFee:         10,
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		ss := NewSemesterService(db)

		err = ss.UpdateBudget(semester1.ID, -10.30)
		if err != nil {
			t.Errorf("SemesterService.UpdateBudget() error = %v", err)
			return
		}

		newSem := models.Semester{ID: semester1.ID}
		res = db.First(&newSem)
		if res.Error != nil {
			t.Fatalf("Error when retrieving semester: %v", res.Error)
			return
		}

		if !almostEqual(newSem.CurrentBudget, 95.27) {
			t.Errorf("SemesterService.UpdateBudget() = %v, expected = %v", newSem.CurrentBudget, 95.27)
			return
		}
	}
}

func ExportRankingsTest(t *testing.T) {
	db, err := database.OpenTestConnection()
	if !assert.NoError(t, err, "Failed to initialize the database") {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	semester, err := testhelpers.SetupSemester(db, "Fall 2022")
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to seed a semester: %v", err)
	}

	// Seed rankings
	r := []models.Ranking{
		{
			MembershipID: semester.Memberships[0].ID,
			Points:       10,
		},
		{
			MembershipID: semester.Memberships[1].ID,
			Points:       5,
		},
		{
			MembershipID: semester.Memberships[2].ID,
			Points:       2,
		},
	}
	res := db.Create(&r)
	if !assert.NoError(t, res.Error) {
		t.Fatalf("Failed to seed rankings: %v", res.Error)
	}

	svc := NewSemesterService(db)

	// Ensure filepath was returned
	fp, err := svc.ExportRankings(semester.Semester.ID)
	assert.NoError(t, err, "ExportRankings should not return an error")

	expectedPath, err := filepath.Abs("rankings.csv")
	assert.NoError(t, err, "Should not error creating the expected file path")
	assert.Equal(t, expectedPath, fp)

	// Check that the file exists
	assert.FileExists(t, fp, "The rankings CSV file should exist")

	// Read the file
	file, err := os.Open(fp)
	assert.NoError(t, err, "Should not error when openning a file")

	// Initialize reader
	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	assert.NoError(t, err, "Should not error when reading the CSV file")

	// Check records were inserted correctly
	expectedRecords := [][]string{
		{"id", "first_name", "last_name", "points"},
		{"20780648", "Adam", "Mahood", "10"},
		{"36459367", "Deep", "Kalra", "5"},
		{"13274944", "Jane", "Doe", "2"},
	}

	assert.ElementsMatch(t, expectedRecords, records)

	// Remove file
	os.Remove(fp)
}
