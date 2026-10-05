package services

import (
	"api/internal/models"
	"time"
)

var torontoCalendar, _ = time.LoadLocation("America/Toronto")

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

// ComparisonCutoff returns the instant in comparison that corresponds to how far
// now has progressed through target, so a term in progress is compared against the
// same point of the earlier term rather than against its completed total.
//
// Without this, every accumulating figure — memberships sold, distinct players,
// entries — reads as a collapse for most of a term: a month in, the current term is
// being measured against four months of the previous one.
//
// Elapsed time is measured in days from each term's own start rather than as a
// fraction of its length. Signups and attendance cluster by calendar — orientation
// week, the first event — so day 14 is comparable to day 14 even when the two terms
// differ in length.
//
// Once the target's inclusive Toronto end date has passed, or it has run at least as
// long as the comparison term, the comparison is its complete self and the returned
// cutoff clips nothing. It deliberately does not clamp to the comparison term's
// end_date: an event is not guaranteed to fall inside its semester's declared range,
// and clamping would silently drop any that lie outside rather than counting a term
// that has finished.
func ComparisonCutoff(target, comparison models.Semester, now time.Time) time.Time {
	start := comparison.StartDate.UTC()
	if !now.UTC().Before(inclusiveSemesterEndExclusive(target.EndDate)) {
		return noCutoff
	}

	elapsed := now.UTC().Sub(target.StartDate.UTC())
	if elapsed < 0 {
		return start
	}

	if elapsed >= inclusiveSemesterEndExclusive(comparison.EndDate).Sub(start) {
		return noCutoff
	}

	return start.Add(elapsed)
}

// inclusiveSemesterEndExclusive converts the UTC-encoded date-only semester end
// into midnight Toronto time on the following calendar date. The stored UTC
// midnight is a date label, not an instant in Toronto.
func inclusiveSemesterEndExclusive(endDate time.Time) time.Time {
	endDate = endDate.UTC()
	nextDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day()+1, 0, 0, 0, 0, torontoCalendar)
	return nextDay.UTC()
}

// noCutoff is a cutoff far enough in the future to clip nothing, used when the
// comparison term has fully elapsed.
var noCutoff = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)
