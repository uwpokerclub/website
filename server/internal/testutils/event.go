package testutils

import (
	"api/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var TEST_EVENTS = []models.Event{
	{
		ID:               3,
		Name:             "Spring 2024 Event #1",
		Format:           "No Limit Hold'em",
		Notes:            "",
		SemesterID:       TEST_SEMESTERS[1].ID,
		StartDate:        time.Date(2024, 2, 10, 18, 0, 0, 0, time.Now().Local().Location()),
		State:            models.EventStateStarted,
		StructureID:      TEST_STRUCTURES[0].ID,
		Rebuys:           3,
		PointsMultiplier: 1.5,
	},
	{
		ID:               2,
		Name:             "Fall 2023 Event #2",
		Format:           "No Limit Hold'em",
		Notes:            "",
		SemesterID:       TEST_SEMESTERS[0].ID,
		StartDate:        time.Date(2023, 10, 20, 18, 0, 0, 0, time.Now().Local().Location()),
		State:            models.EventStateStarted,
		StructureID:      TEST_STRUCTURES[0].ID,
		Rebuys:           5,
		PointsMultiplier: 1.0,
	},
	{
		ID:               1,
		Name:             "Fall 2023 Event #1",
		Format:           "No Limit Hold'em",
		Notes:            "",
		SemesterID:       TEST_SEMESTERS[0].ID,
		StartDate:        time.Date(2023, 9, 15, 18, 0, 0, 0, time.Now().Local().Location()),
		State:            models.EventStateEnded,
		StructureID:      TEST_STRUCTURES[0].ID,
		Rebuys:           2,
		PointsMultiplier: 1.0,
	},
}

func SeedEvents(db *gorm.DB, seedDependencies bool) error {
	// Seed related structures and semesters first if requested
	if seedDependencies {
		if err := SeedStructures(db); err != nil {
			return err
		}

		if err := SeedSemesters(db); err != nil {
			return err
		}
	}

	for _, event := range TEST_EVENTS {
		if err := db.Create(&event).Error; err != nil {
			return err
		}
	}

	return nil
}

func FindEventById(id int32) (*models.Event, error) {
	for _, event := range TEST_EVENTS {
		if event.ID == id {
			return &event, nil
		}
	}

	return nil, fmt.Errorf("event not found")
}

// CreateTestEvent creates a test event with the given parameters
func CreateTestEvent(
	db *gorm.DB,
	semesterID uuid.UUID,
	structureID int32,
	name string,
) (*models.Event, error) {
	event := models.Event{
		Name:             name,
		Format:           "No Limit Hold'em",
		Notes:            "Test event",
		SemesterID:       semesterID,
		StartDate:        time.Now().UTC(),
		State:            models.EventStateStarted,
		StructureID:      structureID,
		Rebuys:           0,
		PointsMultiplier: 1.0,
	}

	if err := db.Create(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}
