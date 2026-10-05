package controller

import (
	"api/internal/store"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignupComparisonDailyTotalsUsesCalendarOffsetsAcrossDST(t *testing.T) {
	current := []store.SignupTimelinePoint{
		{Date: "2026-03-07"},
		{Date: "2026-03-08"},
		{Date: "2026-03-09"},
		{Date: "2026-03-10"},
	}
	comparison := []store.SignupTimelinePoint{
		{Date: "2025-03-09", Admin: 2},
		{Date: "2025-03-10"},
		{Date: "2025-03-11", Discord: 1},
		{Date: "2025-03-12", Unknown: 3},
		{Date: "2025-03-13", Admin: 9},
	}

	points, ok := signupComparisonDailyTotals(current, comparison)
	require.True(t, ok)
	require.Equal(t, []SignupTimelineComparisonPoint{
		{ElapsedDay: 0, Total: 2},
		{ElapsedDay: 1, Total: 0},
		{ElapsedDay: 2, Total: 1},
		{ElapsedDay: 3, Total: 3},
	}, points)
}

func TestSignupComparisonDailyTotalsKeepsDifferentStartDatesAndTermCoverage(t *testing.T) {
	current := []store.SignupTimelinePoint{
		{Date: "2026-09-12"},
		{Date: "2026-09-13"},
		{Date: "2026-09-14"},
	}
	comparison := []store.SignupTimelinePoint{
		{Date: "2025-09-02", Admin: 1},
		{Date: "2025-09-03"},
		{Date: "2025-09-04", Discord: 2},
	}

	points, ok := signupComparisonDailyTotals(current, comparison)
	require.True(t, ok)
	require.Equal(t, []SignupTimelineComparisonPoint{
		{ElapsedDay: 0, Total: 1},
		{ElapsedDay: 1, Total: 0},
		{ElapsedDay: 2, Total: 2},
	}, points)
}
