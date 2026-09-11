package services

import (
	"api/internal/models"
	"time"
)

// Season is a semester's season, derived from its start date's month rather than
// its free-text name.
type Season string

const (
	SeasonFall   Season = "Fall"
	SeasonWinter Season = "Winter"
	SeasonSpring Season = "Spring"
)

// SeasonOf derives a semester's season from the month of startDate.
func SeasonOf(startDate time.Time) Season {
	// Explicit UTC: pgx can hand back times in the local zone.
	switch startDate.UTC().Month() {
	case time.September, time.October, time.November, time.December:
		return SeasonFall
	case time.January, time.February, time.March, time.April:
		return SeasonWinter
	default:
		return SeasonSpring
	}
}

// ResolveComparisonSemester returns the most recent semester in semesters that shares
// target's season and started in a strictly earlier calendar year (UTC) than target,
// or nil if no such semester exists.
func ResolveComparisonSemester(target models.Semester, semesters []models.Semester) *models.Semester {
	season := SeasonOf(target.StartDate)
	targetYear := target.StartDate.UTC().Year()

	var best *models.Semester
	for _, candidate := range semesters {
		if candidate.ID == target.ID {
			continue
		}
		if SeasonOf(candidate.StartDate) != season {
			continue
		}
		if candidate.StartDate.UTC().Year() >= targetYear {
			continue
		}
		if best == nil || candidate.StartDate.After(best.StartDate) {
			c := candidate
			best = &c
		}
	}

	return best
}
