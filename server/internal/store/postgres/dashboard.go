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
