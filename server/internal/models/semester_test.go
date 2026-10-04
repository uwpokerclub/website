package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSemesterCalendarDaysUsesInclusiveUTCDates(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  int
	}{
		{
			name:  "leap year boundary is 366 rows",
			start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC),
			want:  366,
		},
		{
			name:  "one row beyond leap year limit is 367",
			start: time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
			want:  367,
		},
		{
			name:  "spring DST transition does not change UTC calendar row count",
			start: time.Date(2026, time.March, 7, 23, 30, 0, 0, mustTorontoLocation(t)),
			end:   time.Date(2026, time.March, 9, 0, 30, 0, 0, mustTorontoLocation(t)),
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, SemesterCalendarDays(tt.start, tt.end))
		})
	}
}

func TestValidateSemesterDateRangeAllows366AndRejects367InclusiveRows(t *testing.T) {
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	end366 := time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC)
	require.NoError(t, ValidateSemesterDateRange(start, end366))

	start = time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC)
	end367 := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)
	require.ErrorIs(t, ValidateSemesterDateRange(start, end367), ErrSemesterDateRange)
}

func mustTorontoLocation(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("America/Toronto")
	require.NoError(t, err)
	return location
}
