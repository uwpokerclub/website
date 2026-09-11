package services

import (
	"testing"
	"time"

	"api/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSeasonOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		date   time.Time
		season Season
	}{
		{"Aug 31 is Spring", time.Date(2025, time.August, 31, 0, 0, 0, 0, time.UTC), SeasonSpring},
		{"Sep 1 is Fall", time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC), SeasonFall},
		{"Dec 31 is Fall", time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC), SeasonFall},
		{"Jan 1 is Winter", time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), SeasonWinter},
		{"Apr 30 is Winter", time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC), SeasonWinter},
		{"May 1 is Spring", time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC), SeasonSpring},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.season, SeasonOf(tt.date))
		})
	}
}

func semesterAt(name string, date time.Time) models.Semester {
	return models.Semester{ID: uuid.New(), Name: name, StartDate: date}
}

func TestResolveComparisonSemester(t *testing.T) {
	t.Parallel()

	fall2024 := semesterAt("Fall 2024", time.Date(2024, time.September, 8, 0, 0, 0, 0, time.UTC))
	fall2025 := semesterAt("Fall 2025", time.Date(2025, time.September, 10, 0, 0, 0, 0, time.UTC))
	fall2026 := semesterAt("Fall 2026", time.Date(2026, time.September, 3, 0, 0, 0, 0, time.UTC))
	winter2026 := semesterAt("Winter 2026", time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC))
	spring2026 := semesterAt("Spring 2026", time.Date(2026, time.May, 5, 0, 0, 0, 0, time.UTC))
	spring2025 := semesterAt("Spring 2025", time.Date(2025, time.May, 1, 0, 0, 0, 0, time.UTC))

	t.Run("resolves to the same season one year prior", func(t *testing.T) {
		t.Parallel()
		got := ResolveComparisonSemester(fall2026, []models.Semester{fall2024, fall2025, fall2026, winter2026, spring2026})
		require.NotNil(t, got)
		require.Equal(t, fall2025.ID, got.ID)
	})

	t.Run("falls back to an older same-season semester when the most recent one is missing", func(t *testing.T) {
		t.Parallel()
		got := ResolveComparisonSemester(fall2026, []models.Semester{fall2024, fall2026, winter2026, spring2026})
		require.NotNil(t, got)
		require.Equal(t, fall2024.ID, got.ID)
	})

	t.Run("returns nil when there is no comparable prior season", func(t *testing.T) {
		t.Parallel()
		got := ResolveComparisonSemester(spring2025, []models.Semester{spring2025, fall2024, winter2026})
		require.Nil(t, got)
	})

	t.Run("never resolves to itself", func(t *testing.T) {
		t.Parallel()
		got := ResolveComparisonSemester(fall2025, []models.Semester{fall2025})
		require.Nil(t, got)
	})

	t.Run("never picks an adjacent season", func(t *testing.T) {
		t.Parallel()
		got := ResolveComparisonSemester(fall2026, []models.Semester{winter2026, spring2026})
		require.Nil(t, got)
	})

	t.Run("resolves across a calendar year boundary even under a year apart", func(t *testing.T) {
		t.Parallel()
		// fall2025 starts Sep 10 and fall2026 starts Sep 3: under 365 days apart,
		// but fall2025's UTC year (2025) is strictly earlier than fall2026's (2026).
		got := ResolveComparisonSemester(fall2026, []models.Semester{fall2025, fall2026})
		require.NotNil(t, got)
		require.Equal(t, fall2025.ID, got.ID)
	})

	t.Run("never picks a same-season semester from the same calendar year", func(t *testing.T) {
		t.Parallel()
		fall2026Second := semesterAt("Fall 2026 (Co-op)", time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC))
		got := ResolveComparisonSemester(fall2026, []models.Semester{fall2026Second, spring2025})
		require.Nil(t, got)
	})

	t.Run("picks the most recent among multiple qualifying prior years", func(t *testing.T) {
		t.Parallel()
		got := ResolveComparisonSemester(fall2026, []models.Semester{fall2024, fall2025})
		require.NotNil(t, got)
		require.Equal(t, fall2025.ID, got.ID)
	})

	t.Run("returns a pointer to a copy, not into the caller's slice", func(t *testing.T) {
		t.Parallel()
		semesters := []models.Semester{fall2025}
		got := ResolveComparisonSemester(fall2026, semesters)
		require.NotNil(t, got)
		got.Name = "mutated"
		require.Equal(t, "Fall 2025", semesters[0].Name)
	})
}
