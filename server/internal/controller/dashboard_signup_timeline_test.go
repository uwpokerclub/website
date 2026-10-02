package controller

import (
	"api/internal/models"
	"api/internal/store"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSignupComparisonCutoffUsesCalendarDaysAcrossDST(t *testing.T) {
	toronto, err := time.LoadLocation("America/Toronto")
	require.NoError(t, err)

	current := models.Semester{
		StartDate: time.Date(2026, time.March, 7, 5, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, time.June, 30, 3, 59, 59, 0, time.UTC),
	}
	comparison := models.Semester{
		StartDate: time.Date(2025, time.March, 9, 4, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, time.June, 30, 3, 59, 59, 0, time.UTC),
	}
	series := []store.SignupTimelinePoint{
		{Date: "2026-03-07"},
		{Date: "2026-03-08"},
		{Date: "2026-03-09"},
		{Date: "2026-03-10"},
	}

	cutoff, ok := signupComparisonCutoff(current, comparison, series)
	require.True(t, ok)
	require.Equal(t, "2025-03-12", cutoff.In(toronto).Format("2006-01-02"))
	require.Equal(t, 12, cutoff.In(toronto).Day())
}

func TestSignupComparisonCutoffClipsAtComparisonEndDate(t *testing.T) {
	current := models.Semester{
		StartDate: time.Date(2026, time.September, 1, 4, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, time.December, 31, 4, 59, 59, 0, time.UTC),
	}
	comparison := models.Semester{
		StartDate: time.Date(2025, time.September, 3, 4, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, time.September, 5, 4, 0, 0, 0, time.UTC),
	}
	series := []store.SignupTimelinePoint{
		{Date: "2026-09-01"},
		{Date: "2026-09-10"},
	}

	cutoff, ok := signupComparisonCutoff(current, comparison, series)
	require.True(t, ok)
	require.Equal(t, "2025-09-05", cutoff.In(mustToronto(t)).Format("2006-01-02"))
}

func mustToronto(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("America/Toronto")
	require.NoError(t, err)
	return location
}
