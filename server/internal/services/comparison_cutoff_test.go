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

	t.Run("uses the full comparison when a shorter target term has completed", func(t *testing.T) {
		shortTarget := semesterAt("2026-01-01T00:00:00Z", "2026-02-15T00:00:00Z")
		longComparison := semesterAt("2025-01-03T00:00:00Z", "2025-05-01T00:00:00Z")
		loc, err := time.LoadLocation("America/Toronto")
		assert.NoError(t, err)
		now := time.Date(2026, 2, 16, 0, 0, 0, 0, loc)

		cutoff := services.ComparisonCutoff(shortTarget, longComparison, now)

		assert.True(t, cutoff.After(longComparison.EndDate))
	})

	t.Run("keeps the inclusive Toronto end date partial until the following midnight", func(t *testing.T) {
		shortTarget := semesterAt("2026-01-01T00:00:00Z", "2026-02-15T00:00:00Z")
		longComparison := semesterAt("2025-01-03T00:00:00Z", "2025-05-01T00:00:00Z")
		loc, err := time.LoadLocation("America/Toronto")
		assert.NoError(t, err)
		priorEvening := time.Date(2026, 2, 14, 19, 30, 0, 0, loc)
		lastDay := time.Date(2026, 2, 15, 23, 59, 59, 0, loc)
		exclusiveNextDay := time.Date(2026, 2, 16, 0, 0, 0, 0, loc)

		assert.Equal(t, longComparison.StartDate.Add(priorEvening.UTC().Sub(shortTarget.StartDate.UTC())), services.ComparisonCutoff(shortTarget, longComparison, priorEvening))
		assert.Equal(t, longComparison.StartDate.Add(lastDay.UTC().Sub(shortTarget.StartDate.UTC())), services.ComparisonCutoff(shortTarget, longComparison, lastDay))
		assert.True(t, services.ComparisonCutoff(shortTarget, longComparison, exclusiveNextDay).After(longComparison.EndDate))
	})

	t.Run("derives inclusive end boundaries across DST from UTC date components", func(t *testing.T) {
		loc, err := time.LoadLocation("America/Toronto")
		assert.NoError(t, err)
		targetSpring := semesterAt("2026-03-01T00:00:00Z", "2026-03-08T00:00:00Z")
		comparisonSpring := semesterAt("2025-03-01T00:00:00Z", "2025-04-30T00:00:00Z")
		lastDay := time.Date(2026, 3, 8, 23, 59, 0, 0, loc)
		nextDay := time.Date(2026, 3, 9, 0, 0, 0, 0, loc)
		assert.False(t, services.ComparisonCutoff(targetSpring, comparisonSpring, lastDay).After(comparisonSpring.EndDate))
		assert.True(t, services.ComparisonCutoff(targetSpring, comparisonSpring, nextDay).After(comparisonSpring.EndDate))
	})

	t.Run("measures a comparison term through its inclusive Toronto end date", func(t *testing.T) {
		loc, err := time.LoadLocation("America/Toronto")
		assert.NoError(t, err)
		targetLong := semesterAt("2026-01-01T00:00:00Z", "2026-05-31T00:00:00Z")
		comparisonShort := semesterAt("2025-01-01T00:00:00Z", "2025-03-09T00:00:00Z")
		comparisonExclusiveEnd := time.Date(2025, 3, 10, 0, 0, 0, 0, loc)
		targetEquivalentEnd := time.Date(2026, 3, 10, 0, 0, 0, 0, loc)

		beforeEnd := services.ComparisonCutoff(targetLong, comparisonShort, targetEquivalentEnd.Add(-time.Nanosecond))
		atEnd := services.ComparisonCutoff(targetLong, comparisonShort, targetEquivalentEnd)

		assert.Equal(t, comparisonShort.StartDate.Add(targetEquivalentEnd.Add(-time.Nanosecond).UTC().Sub(targetLong.StartDate.UTC())), beforeEnd)
		assert.True(t, atEnd.After(comparisonExclusiveEnd.AddDate(100, 0, 0)))
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
