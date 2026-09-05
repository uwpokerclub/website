package models

import "time"

// EventClock is the server-authoritative state for one event's tournament
// clock. A row is created lazily on first read, so its existence implies the
// clock is either running or paused; there is no separate "not started" state.
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
// a moment in time. It is never persisted.
type DerivedClock struct {
	LevelIndex  int32
	LevelEndsAt time.Time
	PausedAt    *time.Time
	Remaining   time.Duration
	// Version is carried through from the underlying EventClock unchanged:
	// derivation never bumps it, only a persisted action does.
	Version int64
}

// ClockState is the JSON shape returned by every clock endpoint: the fully
// derived state plus the server's own clock, so a client with a badly-set
// system clock can apply a one-line offset.
type ClockState struct {
	LevelIndex  int32      `json:"levelIndex"`
	LevelEndsAt time.Time  `json:"levelEndsAt"`
	PausedAt    *time.Time `json:"pausedAt"`
	Version     int64      `json:"version"`
	ServerTime  time.Time  `json:"serverTime"`
} //@name ClockState

// NewClockState builds the response shape for a derived clock, stamping it
// with the server's current time.
func NewClockState(derived DerivedClock, serverTime time.Time) ClockState {
	return ClockState{
		LevelIndex:  derived.LevelIndex,
		LevelEndsAt: derived.LevelEndsAt,
		PausedAt:    derived.PausedAt,
		Version:     derived.Version,
		ServerTime:  serverTime,
	}
}

// AdjustClockRequest is the request body for POST .../clock/adjust.
type AdjustClockRequest struct {
	DeltaSeconds int `json:"deltaSeconds" binding:"min=-3600,max=3600"`
} //@name AdjustClockRequest

// SetClockLevelRequest is the request body for POST .../clock/level. Index is
// a pointer so binding's "required" tag can tell an omitted field (nil) apart
// from an explicit index 0, which would otherwise silently reset the clock.
type SetClockLevelRequest struct {
	Index *int32 `json:"index" binding:"required,min=0"`
} //@name SetClockLevelRequest

// Derive rolls the clock forward to now given the ordered level durations. It
// returns false if levels is empty. Derive is pure: it does not mutate the
// receiver and never touches a store.
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
		Version:     c.Version,
	}, true
}
