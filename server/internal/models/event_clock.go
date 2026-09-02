package models

import "time"

// EventClock is the server-authoritative state for one event's tournament
// clock. A row is created lazily on first read - see EventClockService in a
// later issue - so its existence implies the clock is either running or
// paused; there is no separate "not started" state.
type EventClock struct {
	EventID     int32      `json:"eventId"     gorm:"type:integer;primaryKey;autoIncrement:false"`
	Event       *Event     `json:"event,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	LevelIndex  int32      `json:"levelIndex"  gorm:"not null"`
	LevelEndsAt time.Time  `json:"levelEndsAt" gorm:"not null"`
	PausedAt    *time.Time `json:"pausedAt"`
	Version     int64      `json:"version"     gorm:"not null"`
	UpdatedAt   time.Time  `json:"updatedAt"   gorm:"not null"`
} //@name EventClock

func (EventClock) TableName() string {
	return "event_clocks"
}

// DerivedClock is the fully rolled-forward, read-only view of an EventClock at
// a moment in time. It is never persisted - see Derive.
type DerivedClock struct {
	LevelIndex  int32
	LevelEndsAt time.Time
	PausedAt    *time.Time
	Remaining   time.Duration
}

// Derive rolls the clock forward to now given the ordered level durations. It
// returns false if levels is empty - a structure with no blinds has no clock.
//
// Derive is a pure function of the receiver, levels, and now: it does not
// mutate the receiver and never touches a store.
//
// Roll-forward accumulates each level's duration onto LevelEndsAt rather than
// resetting to now, so a clock that was never read for an arbitrary gap still
// lands on exactly the right level and exactly the right remaining time on
// the next call - no catch-up job, no reconciliation. While PausedAt is set,
// effectiveNow is frozen there and no roll-forward happens, which is what
// makes pause immune to clock skew between rooms.
func (c EventClock) Derive(levels []time.Duration, now time.Time) (DerivedClock, bool) {
	if len(levels) == 0 {
		return DerivedClock{}, false
	}

	effectiveNow := now
	if c.PausedAt != nil {
		effectiveNow = *c.PausedAt
	}

	levelIndex := c.LevelIndex
	levelEndsAt := c.LevelEndsAt
	for c.PausedAt == nil && !now.Before(levelEndsAt) && int(levelIndex) < len(levels)-1 {
		levelIndex++
		levelEndsAt = levelEndsAt.Add(levels[levelIndex])
	}

	remaining := levelEndsAt.Sub(effectiveNow)
	if remaining < 0 {
		remaining = 0
	}

	return DerivedClock{
		LevelIndex:  levelIndex,
		LevelEndsAt: levelEndsAt,
		PausedAt:    c.PausedAt,
		Remaining:   remaining,
	}, true
}
