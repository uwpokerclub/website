package store

import (
	"time"

	"github.com/google/uuid"
)

// SpotlightEvent is the event featured on the dashboard's Event Spotlight card.
//
// This lives in the store package rather than internal/models because
// atlas-provider-gorm scans every type under internal/models to diff the schema; a
// plain response struct placed there would be picked up as a new table.
type SpotlightEvent struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Format    string    `json:"format"`
	StartDate time.Time `json:"startDate"`
	State     uint8     `json:"state"`
	Entries   int64     `json:"entries"`
	Rebuys    uint8     `json:"rebuys"`
} //@name SpotlightEvent

// DashboardRepository is the interface for the dashboard's read-only aggregate
// queries. These span memberships, participants, and events, so they are kept here
// rather than smeared as Stats() methods across those repositories.
type DashboardRepository interface {
	// Spotlight returns the event to feature on the dashboard for the given semester,
	// evaluated against now:
	//
	//  1. The live event - state EventStateStarted with a start date at or before now.
	//  2. Otherwise, the soonest event with a start date after now.
	//  3. Otherwise, nil.
	//
	// The returned event's Entries and Rebuys reflect its current entry count and
	// rebuy count. Returns nil, nil if no qualifying event exists.
	Spotlight(semesterID uuid.UUID, now time.Time) (*SpotlightEvent, error)
}
