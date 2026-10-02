package services_test

import (
	"api/internal/models"
	"api/internal/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func semesterAt(start, end string) models.Semester {
	s, _ := time.Parse(time.RFC3339, start)
	e, _ := time.Parse(time.RFC3339, end)
	return models.Semester{StartDate: s, EndDate: e}
}

func TestComparisonCutoff(t *testing.T) {
	target := semesterAt("2026-09-01T00:00:00Z", "2026-12-15T00:00:00Z")
	comparison := semesterAt("2025-09-03T00:00:00Z", "2025-12-12T00:00:00Z")

	t.Run("maps elapsed days from each term's own start", func(t *testing.T) {
		now, _ := time.Parse(time.RFC3339, "2026-09-15T00:00:00Z") // day 14

		cutoff := services.ComparisonCutoff(target, comparison, now)

		// Day 14 of the comparison term, not day 14 of the calendar year.
		assert.Equal(t, "2025-09-17T00:00:00Z", cutoff.Format(time.RFC3339))
	})

	t.Run("clips nothing once the current term has outrun the comparison term's length", func(t *testing.T) {
		now, _ := time.Parse(time.RFC3339, "2026-12-14T00:00:00Z") // past Fall 2025's 100 days

		cutoff := services.ComparisonCutoff(target, comparison, now)

		// Not clamped to end_date: an event dated outside its semester's declared
		// range must still count toward a term that has already finished.
		assert.True(t, cutoff.After(comparison.EndDate.UTC().AddDate(100, 0, 0)))
	})

	t.Run("never runs before the comparison term's start", func(t *testing.T) {
		now, _ := time.Parse(time.RFC3339, "2026-08-20T00:00:00Z") // before term start

		cutoff := services.ComparisonCutoff(target, comparison, now)

		assert.Equal(t, comparison.StartDate.UTC(), cutoff)
	})
}
