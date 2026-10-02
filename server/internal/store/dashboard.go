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

// EventActivityStats is the Event Activity card's scalar figures for a semester.
// TotalEntries and AverageFieldSize cover ended events only. AverageFieldSize is
// calculated in Go so a semester with no ended events never divides by zero.
type EventActivityStats struct {
	EventsRun        int64   `json:"eventsRun"`
	EventsScheduled  int64   `json:"eventsScheduled"`
	TotalEntries     int64   `json:"totalEntries"`
	AverageFieldSize float64 `json:"averageFieldSize"`
} //@name EventActivityStats

// EventSeriesPoint is one ended event's attendance for the Event Activity chart.
type EventSeriesPoint struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate"`
	Entries   int64     `json:"entries"`
} //@name EventSeriesPoint

// EngagementStats is the Engagement & Retention card's figures for a single semester:
// distinct players, median events attended, the played-once cohort, and the 10+
// cohort.
//
// The population is distinct members (by memberships.user_id, not membership_id - a
// member who entered this semester's events through two different memberships still
// counts once) with at least one participant row in an event whose semester_id is
// the requested semester. PlayedOnceShare is PlayedOnceCount / Players, computed in
// Go so a semester with zero players never divides by zero.
type EngagementStats struct {
	Players              int64   `json:"players"`
	MedianEventsAttended float64 `json:"medianEventsAttended"`
	PlayedOnceCount      int64   `json:"playedOnceCount"`
	PlayedOnceShare      float64 `json:"playedOnceShare"`
	TenPlusCount         int64   `json:"tenPlusCount"`
} //@name EngagementStats

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

	// EngagementStats returns the Engagement & Retention figures for the given
	// semester, counting only events that started at or before asOf. A semester with
	// no qualifying participant rows returns a zero-valued EngagementStats, not an
	// error.
	//
	// asOf exists so a term in progress is compared against the same point of the
	// earlier term rather than its completed total; see services.ComparisonCutoff.
	EngagementStats(semesterID uuid.UUID, asOf time.Time) (EngagementStats, error)

	// EventActivity returns Event Activity scalar stats and an ascending start-date
	// series for ended events, counting only events that started at or before asOf. A
	// semester with no ended events returns zero stats and an empty, non-nil series.
	EventActivity(semesterID uuid.UUID, asOf time.Time) (EventActivityStats, []EventSeriesPoint, error)

	// MembershipTotalAsOf returns how many of the semester's memberships had been
	// created at or before asOf, or nil when the semester has no dated memberships at
	// all — every row predating the created_at migration carries NULL and cannot be
	// reliably recovered. Callers must treat nil as "unknowable", never as zero.
	MembershipTotalAsOf(semesterID uuid.UUID, asOf time.Time) (*int64, error)
}
