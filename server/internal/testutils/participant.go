package testutils

import (
	"api/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var TEST_PARTICIPANTS = []models.Participant{
	{
		MembershipID: TEST_MEMBERSHIPS[0].ID,
		EventID:      1, // TEST_EVENTS[2] - Fall 2023 Event #1 (Ended)
		Placement:    1,
		SignedOutAt:  ptrToTime(time.Date(2023, 9, 15, 20, 30, 0, 0, time.Now().Local().Location())),
	},
	{
		MembershipID: TEST_MEMBERSHIPS[1].ID,
		EventID:      1, // TEST_EVENTS[2] - Fall 2023 Event #1 (Ended)
		Placement:    2,
		SignedOutAt:  ptrToTime(time.Date(2023, 9, 15, 20, 15, 0, 0, time.Now().Local().Location())),
	},
	{
		MembershipID: TEST_MEMBERSHIPS[2].ID,
		EventID:      2, // TEST_EVENTS[1] - Fall 2023 Event #2 (Started)
		Placement:    0,
		SignedOutAt:  nil,
	},
	{
		MembershipID: TEST_MEMBERSHIPS[0].ID,
		EventID:      2, // TEST_EVENTS[1] - Fall 2023 Event #2 (Started) - signed out, for testing sign-in
		Placement:    0,
		SignedOutAt:  ptrToTime(time.Date(2023, 10, 20, 20, 0, 0, 0, time.Now().Local().Location())),
	},
}

func ptrToTime(t time.Time) *time.Time {
	return &t
}

func SeedParticipants(db *gorm.DB, seedDependencies bool) error {
	// Ensure events are seeded first (which seeds semesters and structures) if requested
	if seedDependencies {
		if err := SeedEvents(db, true); err != nil {
			return err
		}

		// Ensure memberships are seeded first
		if err := SeedMemberships(db, true); err != nil {
			return err
		}
	}

	for _, participant := range TEST_PARTICIPANTS {
		if err := db.Create(&participant).Error; err != nil {
			return err
		}
	}

	return nil
}

func CreateTestParticipant(db *gorm.DB, membershipId uuid.UUID, eventId int32) (*models.Participant, error) {
	participant := models.Participant{
		MembershipID: membershipId,
		EventID:      eventId,
		Placement:    0,
		SignedOutAt:  nil,
	}

	if err := db.Create(&participant).Error; err != nil {
		return nil, err
	}

	return &participant, nil
}

func FindParticipantById(membershipId uuid.UUID, eventId int32) (*models.Participant, error) {
	for _, participant := range TEST_PARTICIPANTS {
		if participant.MembershipID == membershipId && participant.EventID == eventId {
			return &participant, nil
		}
	}

	return nil, fmt.Errorf("participant not found")
}

func FindParticipantAsMap(membershipId uuid.UUID, eventId int32) (map[string]any, error) {
	participant, err := FindParticipantById(membershipId, eventId)
	if err != nil {
		return nil, err
	}

	// Convert to map for JSON comparison
	b, err := json.Marshal(participant)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return result, nil
}
