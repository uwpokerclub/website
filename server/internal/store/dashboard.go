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

// SemesterRef identifies a semester by id and name, without the rest of its fields.
// Dashboard endpoints that carry a resolved comparison semester use this instead of
// models.Semester to keep the response focused on what the UI needs to label it.
type SemesterRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
} //@name SemesterRef

// MembershipStats is the membership bucket breakdown for a single semester: the
// Term at a Glance card's figures.
//
// Paid, Discounted and Executive are independent booleans on memberships, so the
// buckets are assigned by an exclusive partition rather than by counting each flag
// independently - otherwise an unpaid executive would double-count as unpaid. First
// match wins: Executive, then Unpaid, then Discounted, then Paid. Total, and New +
// Returning, each sum across the four buckets / two buckets respectively.
type MembershipStats struct {
	Total      int64 `json:"total"`
	Paid       int64 `json:"paid"`
	Unpaid     int64 `json:"unpaid"`
	Discounted int64 `json:"discounted"`
	Executive  int64 `json:"executive"`
	New        int64 `json:"new"`
	Returning  int64 `json:"returning"`
} //@name MembershipStats

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

	// MembershipStats returns the membership bucket breakdown for the given semester.
	// A semester with no memberships returns a zero-valued MembershipStats, not an
	// error. New vs returning is measured against this semester's own start_date: a
	// membership is New when the user has no membership in any semester whose
	// start_date is strictly earlier, otherwise Returning.
	MembershipStats(semesterID uuid.UUID) (MembershipStats, error)
}
