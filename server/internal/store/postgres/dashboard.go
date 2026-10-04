package postgres

import (
	"api/internal/models"
	"api/internal/store"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgresDashboardRepository struct {
	db *gorm.DB
}

var _ store.DashboardRepository = (*postgresDashboardRepository)(nil)

func NewDashboardRepository(db *gorm.DB) store.DashboardRepository {
	return &postgresDashboardRepository{db: db}
}

// spotlightQuery builds the shared event-plus-entry-count query for Spotlight,
// scoped to semesterID and any additional where clauses.
func (r *postgresDashboardRepository) spotlightQuery(semesterID uuid.UUID) *gorm.DB {
	return r.db.Table("events e").
		Select("e.id, e.name, e.format, e.start_date, e.state, e.rebuys, COUNT(p.id) AS entries").
		Joins("LEFT JOIN participants p ON p.event_id = e.id").
		Where("e.semester_id = ?", semesterID).
		Group("e.id")
}

func (r *postgresDashboardRepository) Spotlight(semesterID uuid.UUID, now time.Time) (*store.SpotlightEvent, error) {
	var event store.SpotlightEvent

	err := r.spotlightQuery(semesterID).
		Where("e.state = ? AND e.start_date <= ?", models.EventStateStarted, now).
		Order("e.start_date DESC").
		Limit(1).
		Scan(&event).Error
	if err != nil {
		return nil, err
	}
	if event.ID != 0 {
		return &event, nil
	}

	err = r.spotlightQuery(semesterID).
		Where("e.state = ? AND e.start_date > ?", models.EventStateStarted, now).
		Order("e.start_date ASC").
		Limit(1).
		Scan(&event).Error
	if err != nil {
		return nil, err
	}
	if event.ID != 0 {
		return &event, nil
	}

	return nil, nil
}

// MembershipStats implements store.DashboardRepository.
func (r *postgresDashboardRepository) MembershipStats(semesterID uuid.UUID) (store.MembershipStats, error) {
	var stats store.MembershipStats

	err := r.db.Raw(`
		WITH target AS (
			SELECT start_date FROM semesters WHERE id = ?
		)
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE m.executive) AS executive,
			COUNT(*) FILTER (WHERE NOT m.executive AND NOT m.paid) AS unpaid,
			COUNT(*) FILTER (WHERE NOT m.executive AND m.paid AND m.discounted) AS discounted,
			COUNT(*) FILTER (WHERE NOT m.executive AND m.paid AND NOT m.discounted) AS paid,
			COUNT(*) FILTER (WHERE NOT EXISTS (
				SELECT 1 FROM memberships pm
				JOIN semesters ps ON ps.id = pm.semester_id
				WHERE pm.user_id = m.user_id AND ps.start_date < target.start_date
			)) AS new
		FROM memberships m, target
		WHERE m.semester_id = ?
	`, semesterID, semesterID).Scan(&stats).Error
	if err != nil {
		return store.MembershipStats{}, err
	}

	stats.Returning = stats.Total - stats.New

	return stats, nil
}

// EngagementStats implements store.DashboardRepository.
func (r *postgresDashboardRepository) EngagementStats(semesterID uuid.UUID, asOf time.Time) (store.EngagementStats, error) {
	var stats store.EngagementStats

	err := r.db.Raw(`
		WITH per_member AS (
			SELECT m.user_id, COUNT(DISTINCT p.event_id) AS events
			FROM participants p
			JOIN events e ON e.id = p.event_id
			JOIN memberships m ON m.id = p.membership_id AND m.semester_id = e.semester_id
			WHERE e.semester_id = ? AND e.start_date <= ?
			GROUP BY m.user_id
		)
		SELECT
			COUNT(*) AS players,
			COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY events), 0) AS median_events_attended,
			COUNT(*) FILTER (WHERE events = 1) AS played_once_count,
			COUNT(*) FILTER (WHERE events >= 10) AS ten_plus_count
		FROM per_member
	`, semesterID, asOf).Scan(&stats).Error
	if err != nil {
		return store.EngagementStats{}, err
	}

	if stats.Players > 0 {
		stats.PlayedOnceShare = float64(stats.PlayedOnceCount) / float64(stats.Players)
	}

	return stats, nil
}

// TrialConversionStats implements store.DashboardRepository.
func (r *postgresDashboardRepository) TrialConversionStats(semesterID uuid.UUID, freeTrialLimit uint8) (store.TrialConversionStats, error) {
	var stats store.TrialConversionStats

	err := r.db.Raw(`
		WITH per_member AS (
			SELECT m.user_id, COUNT(DISTINCT p.event_id) AS entries, m.paid, m.executive
			FROM participants p
			JOIN events e ON e.id = p.event_id
			JOIN memberships m ON m.id = p.membership_id AND m.semester_id = ?
			WHERE e.semester_id = ?
			GROUP BY m.user_id, m.paid, m.executive
		)
		SELECT
			COUNT(*) AS players,
			COUNT(*) FILTER (WHERE executive) AS executive,
			COUNT(*) FILTER (WHERE NOT executive AND paid) AS paid,
			COUNT(*) FILTER (WHERE NOT executive AND NOT paid AND entries >= ?) AS trial_spent,
			COUNT(*) FILTER (WHERE NOT executive AND NOT paid AND entries < ?) AS trial_open
		FROM per_member
	`, semesterID, semesterID, freeTrialLimit, freeTrialLimit).Scan(&stats).Error
	if err != nil {
		return store.TrialConversionStats{}, err
	}

	return stats, nil
}

// TrialConversionCohortStats implements store.DashboardRepository.
func (r *postgresDashboardRepository) TrialConversionCohortStats(semesterID uuid.UUID) (store.TrialConversionCohortStats, error) {
	var stats store.TrialConversionCohortStats
	err := r.db.Raw(`
		SELECT
			COUNT(*) FILTER (WHERE trial_started_at IS NOT NULL AND converted_at IS NOT NULL) AS numerator,
			COUNT(*) FILTER (WHERE trial_started_at IS NOT NULL) AS denominator,
			(
				SELECT COUNT(DISTINCT m.id)
				FROM memberships m
				WHERE m.semester_id = ?
				  AND m.trial_started_at IS NULL
				  AND NOT m.executive
				  AND EXISTS (
					SELECT 1
					FROM participants p
					JOIN events e ON e.id = p.event_id
					WHERE p.membership_id = m.id AND e.semester_id = m.semester_id
				  )
			) AS untracked_entrants
		FROM memberships
		WHERE semester_id = ?
	`, semesterID, semesterID).Scan(&stats).Error
	if err != nil {
		return store.TrialConversionCohortStats{}, err
	}
	if stats.Denominator > 0 {
		rate := float64(stats.Numerator) / float64(stats.Denominator)
		stats.Rate = &rate
	}
	return stats, nil
}

// EventActivity implements store.DashboardRepository.
func (r *postgresDashboardRepository) EventActivity(semesterID uuid.UUID, asOf time.Time) (store.EventActivityStats, []store.EventSeriesPoint, error) {
	series := []store.EventSeriesPoint{}

	err := r.db.Table("events e").
		Select("e.id, e.name, e.start_date, COUNT(p.id) AS entries").
		Joins("LEFT JOIN participants p ON p.event_id = e.id").
		Where("e.semester_id = ? AND e.state = ? AND e.start_date <= ?", semesterID, models.EventStateEnded, asOf).
		Group("e.id").
		Order("e.start_date ASC, e.id ASC").
		Scan(&series).Error
	if err != nil {
		return store.EventActivityStats{}, nil, err
	}

	stats := store.EventActivityStats{EventsRun: int64(len(series))}
	for _, point := range series {
		stats.TotalEntries += point.Entries
	}
	if stats.EventsRun > 0 {
		stats.AverageFieldSize = float64(stats.TotalEntries) / float64(stats.EventsRun)
	}

	// Deliberately not clipped to asOf: a scheduled event is one that has not ended,
	// which for the current term usually means it is future-dated. Applying the cutoff
	// here would filter out precisely the events this count exists to report.
	err = r.db.Model(&models.Event{}).
		Where("semester_id = ? AND state = ?", semesterID, models.EventStateStarted).
		Count(&stats.EventsScheduled).Error
	if err != nil {
		return store.EventActivityStats{}, nil, err
	}

	return stats, series, nil
}

func (r *postgresDashboardRepository) AverageFieldSize(semesterID uuid.UUID, asOf time.Time) (float64, error) {
	var average float64
	err := r.db.Raw(`
		SELECT COALESCE(AVG(event_entries.entries), 0)
		FROM (
			SELECT e.id, COUNT(p.id) AS entries
			FROM events e
			LEFT JOIN participants p ON p.event_id = e.id
			WHERE e.semester_id = ? AND e.state = ? AND e.start_date <= ?
			GROUP BY e.id
		) AS event_entries
	`, semesterID, models.EventStateEnded, asOf).Scan(&average).Error
	return average, err
}

// SignupTimeline implements store.DashboardRepository. Membership creation times
// are UTC wall-clock values in a timestamp column, event starts are timestamptz
// instants, and semester bounds encode date-only values at UTC midnight. Convert
// each explicitly so output is independent of the PostgreSQL session timezone.
func (r *postgresDashboardRepository) SignupTimeline(semesterID uuid.UUID, now time.Time) (store.SignupTimeline, error) {
	timeline := store.SignupTimeline{Series: []store.SignupTimelinePoint{}, EventDates: []string{}}
	var bounds struct {
		StartDate *time.Time
		EndDate   *time.Time
	}
	err := r.db.Raw(`
		SELECT (start_date AT TIME ZONE 'UTC')::date AS start_date,
			LEAST((end_date AT TIME ZONE 'UTC')::date, (?::timestamptz AT TIME ZONE 'America/Toronto')::date) AS end_date
		FROM semesters
		WHERE id = ?
	`, now, semesterID).Scan(&bounds).Error
	if err != nil {
		return store.SignupTimeline{}, err
	}
	if bounds.StartDate == nil || bounds.EndDate == nil {
		return timeline, nil
	}
	if models.SemesterCalendarDays(*bounds.StartDate, *bounds.EndDate) > models.MaxSemesterCalendarDays {
		return store.SignupTimeline{}, store.ErrSignupTimelineRange
	}

	err = r.db.Raw(`
		WITH daily AS (
			SELECT
				(m.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Toronto')::date AS date,
				COUNT(*) FILTER (WHERE m.source = 'admin') AS admin,
				COUNT(*) FILTER (WHERE m.source = 'discord') AS discord,
				COUNT(*) FILTER (WHERE m.source IS DISTINCT FROM 'admin' AND m.source IS DISTINCT FROM 'discord') AS unknown
			FROM memberships m
			WHERE m.semester_id = ?
				AND m.created_at IS NOT NULL
				AND (m.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Toronto')::date BETWEEN ?::date AND ?::date
			GROUP BY (m.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Toronto')::date
		)
		SELECT
			TO_CHAR(days.date, 'YYYY-MM-DD') AS date,
			COALESCE(d.admin, 0) AS admin,
			COALESCE(d.discord, 0) AS discord,
			COALESCE(d.unknown, 0) AS unknown
		FROM generate_series(?::date, ?::date, INTERVAL '1 day') AS days(date)
		LEFT JOIN daily d ON d.date = days.date
		ORDER BY days.date
	`, semesterID, bounds.StartDate, bounds.EndDate, bounds.StartDate, bounds.EndDate).Scan(&timeline.Series).Error
	if err != nil {
		return store.SignupTimeline{}, err
	}

	for _, point := range timeline.Series {
		dayTotal := point.Admin + point.Discord + point.Unknown
		timeline.Total += dayTotal
		if dayTotal > 0 && timeline.DataStartsAt == nil {
			date := point.Date
			timeline.DataStartsAt = &date
		}
	}

	err = r.db.Raw(`
		SELECT TO_CHAR((start_date AT TIME ZONE 'America/Toronto')::date, 'YYYY-MM-DD')
		FROM events
		WHERE semester_id = ?
		ORDER BY start_date ASC, id ASC
	`, semesterID).Scan(&timeline.EventDates).Error
	if err != nil {
		return store.SignupTimeline{}, err
	}

	return timeline, nil
}

// minDatedShare is the proportion of a semester's memberships that must carry a
// creation date before a point-in-time count means anything.
//
// The figure is a baseline the current term is judged against, so a baseline drawn
// from a minority of the term understates it and flatters the present. The term that
// straddled the created_at migration is the motivating case: 352 memberships, one of
// them dated. Counting that one row would have reported a near-zero baseline and made
// any current term look like a runaway success.
//
// Terms are either almost entirely dated (96-100% once backfilled) or not dated at
// all, so this sits in the empty space between those two populations rather than
// trying to draw a fine line.
const minDatedShare = 0.8

// MembershipTotalAsOf returns the dated membership count at or before asOf, or nil
// when that partial count cannot be known reliably.
//
// Memberships predating the created_at migration carry NULL. The backfill recovers
// most of them from first participation, but members who never entered an event keep
// a NULL date and are invisible to this count — so it is reported only when the dated
// rows are a large enough majority to stand in for the term. Completed comparisons
// bypass this method and use the exact total from MembershipStats, including undated
// rows. Callers must treat nil as "unknowable", never as zero.
func (r *postgresDashboardRepository) MembershipTotalAsOf(semesterID uuid.UUID, asOf time.Time) (*int64, error) {
	var counts struct {
		Dated int64
		Total int64
	}
	err := r.db.Model(&models.Membership{}).
		Select("COUNT(*) FILTER (WHERE created_at IS NOT NULL) AS dated, COUNT(*) AS total").
		Where("semester_id = ?", semesterID).
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	if counts.Total == 0 || float64(counts.Dated)/float64(counts.Total) < minDatedShare {
		return nil, nil
	}

	var total int64
	err = r.db.Model(&models.Membership{}).
		Where("semester_id = ? AND created_at IS NOT NULL AND created_at <= (?::timestamptz AT TIME ZONE 'UTC')", semesterID, asOf).
		Count(&total).Error
	if err != nil {
		return nil, err
	}

	return &total, nil
}
