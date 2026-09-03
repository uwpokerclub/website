package models_test

import (
	"api/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestEventClock_Derive is written to port cleanly to the equivalent Jest
// suite for deriveClock.ts (#452): every field is a plain int64 (seconds
// since an arbitrary epoch) or a *int64, and levels are plain second counts.
func TestEventClock_Derive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		levelIndex  int32
		levelEndsAt int64
		pausedAt    *int64

		levels []int64
		now    int64

		wantOK          bool
		wantLevelIndex  int32
		wantLevelEndsAt int64
		wantRemaining   int64
		wantPausedAt    *int64
	}{
		{
			name:        "pause freezes remaining regardless of elapsed wall time",
			levelIndex:  0,
			levelEndsAt: 1000,
			pausedAt:    ptr(int64(800)),
			levels:      []int64{600},
			now:         5000,

			wantOK:          true,
			wantLevelIndex:  0,
			wantLevelEndsAt: 1000,
			wantRemaining:   200,
			wantPausedAt:    ptr(int64(800)),
		},
		{
			name:        "resume at the moment of unpause restores the exact frozen remaining",
			levelIndex:  0,
			levelEndsAt: 1000,
			pausedAt:    nil,
			levels:      []int64{600},
			now:         800,

			wantOK:          true,
			wantLevelIndex:  0,
			wantLevelEndsAt: 1000,
			wantRemaining:   200,
			wantPausedAt:    nil,
		},
		{
			name:        "multi-level roll-forward across a long gap lands on the right level and remaining time",
			levelIndex:  0,
			levelEndsAt: 600,
			pausedAt:    nil,
			levels:      []int64{600, 900, 1200, 1500},
			now:         3200,

			wantOK:          true,
			wantLevelIndex:  3,
			wantLevelEndsAt: 4200,
			wantRemaining:   1000,
			wantPausedAt:    nil,
		},
		{
			name:        "negative adjust crossing a level boundary carries the overflow",
			levelIndex:  0,
			levelEndsAt: 500,
			pausedAt:    nil,
			levels:      []int64{600, 900},
			now:         520,

			wantOK:          true,
			wantLevelIndex:  1,
			wantLevelEndsAt: 1400,
			wantRemaining:   880,
			wantPausedAt:    nil,
		},
		{
			name:        "last level expired clamps at 0:00",
			levelIndex:  1,
			levelEndsAt: 1500,
			pausedAt:    nil,
			levels:      []int64{600, 900},
			now:         5000,

			wantOK:          true,
			wantLevelIndex:  1,
			wantLevelEndsAt: 1500,
			wantRemaining:   0,
			wantPausedAt:    nil,
		},
		{
			name:        "empty structure returns no clock",
			levelIndex:  0,
			levelEndsAt: 0,
			pausedAt:    nil,
			levels:      []int64{},
			now:         0,

			wantOK: false,
		},
		{
			name:        "a single level far shorter than the gap rolls to the end and clamps",
			levelIndex:  0,
			levelEndsAt: 600,
			pausedAt:    nil,
			levels:      []int64{600},
			now:         100000,

			wantOK:          true,
			wantLevelIndex:  0,
			wantLevelEndsAt: 600,
			wantRemaining:   0,
			wantPausedAt:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := models.EventClock{
				LevelIndex:  tc.levelIndex,
				LevelEndsAt: epoch(tc.levelEndsAt),
				PausedAt:    epochPtr(tc.pausedAt),
			}

			levels := make([]time.Duration, len(tc.levels))
			for i, s := range tc.levels {
				levels[i] = time.Duration(s) * time.Second
			}

			got, ok := state.Derive(levels, epoch(tc.now))
			require.Equal(t, tc.wantOK, ok)
			if !tc.wantOK {
				return
			}

			require.Equal(t, tc.wantLevelIndex, got.LevelIndex)
			require.Equal(t, epoch(tc.wantLevelEndsAt), got.LevelEndsAt)
			require.Equal(t, time.Duration(tc.wantRemaining)*time.Second, got.Remaining)
			require.Equal(t, epochPtr(tc.wantPausedAt), got.PausedAt)
		})
	}
}

func TestEventClock_Derive_PassesVersionThroughUnchanged(t *testing.T) {
	t.Parallel()

	state := models.EventClock{
		LevelIndex:  0,
		LevelEndsAt: epoch(600),
		Version:     17,
	}

	got, ok := state.Derive([]time.Duration{600 * time.Second, 900 * time.Second}, epoch(3000))
	require.True(t, ok)
	require.Equal(t, int64(17), got.Version, "Derive must carry Version through unchanged; only actions bump it")
}

func TestNewClockState(t *testing.T) {
	t.Parallel()

	pausedAt := epoch(800)
	derived := models.DerivedClock{
		LevelIndex:  2,
		LevelEndsAt: epoch(1000),
		PausedAt:    &pausedAt,
		Version:     5,
	}
	serverTime := epoch(1234)

	state := models.NewClockState(derived, serverTime)

	require.Equal(t, int32(2), state.LevelIndex)
	require.Equal(t, epoch(1000), state.LevelEndsAt)
	require.Equal(t, &pausedAt, state.PausedAt)
	require.Equal(t, int64(5), state.Version)
	require.Equal(t, serverTime, state.ServerTime)
}

func TestEventClock_Derive_DoesNotMutateReceiver(t *testing.T) {
	t.Parallel()

	state := models.EventClock{
		LevelIndex:  0,
		LevelEndsAt: epoch(600),
	}

	_, ok := state.Derive([]time.Duration{600 * time.Second, 900 * time.Second}, epoch(3000))
	require.True(t, ok)

	require.Equal(t, int32(0), state.LevelIndex, "Derive must not mutate the receiver")
	require.Equal(t, epoch(600), state.LevelEndsAt, "Derive must not mutate the receiver")
}

func epoch(seconds int64) time.Time {
	return time.Unix(seconds, 0).UTC()
}

func epochPtr(seconds *int64) *time.Time {
	if seconds == nil {
		return nil
	}
	t := epoch(*seconds)
	return &t
}

func ptr[T any](v T) *T {
	return &v
}
