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
// zero when there are no ended events; EventsRun distinguishes that state from a
// genuine zero-entry average.
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

// SignupTimelinePoint is one calendar day in a semester's signup timeline. Date
// is deliberately a YYYY-MM-DD string rather than a timestamp: the dashboard
// renders it as a calendar date and must not shift it across time zones.
type SignupTimelinePoint struct {
	Date    string `json:"date"`
	Admin   int64  `json:"admin"`
	Discord int64  `json:"discord"`
	Unknown int64  `json:"unknown"`
} //@name SignupTimelinePoint

// SignupTimeline is the dashboard's daily membership-creation series for a
// semester. EventDates and DataStartsAt are YYYY-MM-DD strings for the same
// calendar-date reason as SignupTimelinePoint.Date.
type SignupTimeline struct {
	Series       []SignupTimelinePoint `json:"series"`
	EventDates   []string              `json:"eventDates"`
	DataStartsAt *string               `json:"dataStartsAt"`
	Total        int64                 `json:"total"`
} //@name SignupTimeline

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

// TrialConversionStats is the trial-status breakdown for distinct players in a
// semester. The buckets form an exclusive partition of Players.
type TrialConversionStats struct {
	Players    int64 `json:"players"`
	Paid       int64 `json:"paid"`
	TrialSpent int64 `json:"trialSpent"`
	TrialOpen  int64 `json:"trialOpen"`
	Executive  int64 `json:"executive"`
} //@name TrialConversionStats

// TrialConversionCohortStats reports conversions among memberships whose trial
// start was observed. Rate is a nullable fraction in [0,1]; it is nil when no
// starters are observed.
type TrialConversionCohortStats struct {
	Numerator   int64    `json:"numerator"`
	Denominator int64    `json:"denominator"`
	Rate        *float64 `json:"rate"`
} //@name TrialConversionCohortStats

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

	// TrialConversionStats returns the current paid, executive, and free-trial
	// status of distinct players with entries in the semester. Only entries whose
	// memberships also belong to semesterID qualify; this excludes malformed
	// cross-semester participant rows. FreeTrialLimit is supplied by the caller
	// from the resolved semester, and a zero limit deliberately places every
	// unpaid, non-executive player in TrialSpent.
	TrialConversionStats(semesterID uuid.UUID, freeTrialLimit uint8) (TrialConversionStats, error)

	// TrialConversionCohortStats returns first-ever paid transitions over observed
	// trial starters, independently of current participant rows and membership status.
	TrialConversionCohortStats(semesterID uuid.UUID) (TrialConversionCohortStats, error)

	// EventActivity returns Event Activity scalar stats and an ascending start-date
	// series for ended events, counting only events that started at or before asOf. A
	// semester with no ended events returns zero stats and an empty, non-nil series.
	//
	// EventsRun and EventsScheduled are disjoint - ended and not-yet-ended - so a
	// term's total is their sum, not EventsScheduled alone. EventsScheduled ignores
	// asOf, since the events it counts are the future-dated ones.
	EventActivity(semesterID uuid.UUID, asOf time.Time) (EventActivityStats, []EventSeriesPoint, error)

	// AverageFieldSize returns the ended-event mean including events with zero entries,
	// or nil when no events qualify. It is a narrow comparison query and does not compute
	// the current card's other figures.
	AverageFieldSize(semesterID uuid.UUID, asOf time.Time) (*float64, error)

	// SignupTimeline returns a zero-filled daily calendar-date series from the
	// semester start through the earlier of its end date and now's America/Toronto
	// calendar date. The inclusive range is limited to 366 rows; larger stored
	// ranges return ErrSignupTimelineRange without truncation. Memberships without
	// a creation date are excluded; every other source besides admin and discord is
	// reported as unknown. Total and DataStartsAt are derived from the daily series.
	SignupTimeline(semesterID uuid.UUID, now time.Time) (SignupTimeline, error)

	// MembershipTotalAsOf returns how many of the semester's dated memberships had
	// been created at or before asOf, or nil when the all-term dated share is below
	// the repository's reliability threshold. A returned count is the exact observed
	// dated count and can omit undated memberships; it is never extrapolated. Callers must
	// treat nil as "unknowable", never as zero. Completed comparisons use the exact
	// MembershipStats.Total instead.
	MembershipTotalAsOf(semesterID uuid.UUID, asOf time.Time) (*int64, error)
}
