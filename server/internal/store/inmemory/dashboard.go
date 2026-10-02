package inmemory

import (
	"api/internal/store"
	"time"

	"github.com/google/uuid"
)

// inMemoryDashboardRepository deliberately implements none of the dashboard's
// aggregate queries. For these endpoints, the hand-written SQL in the Postgres
// repository *is* the logic being tested - a Go reimplementation here would only
// verify itself, not the shipped query - so every method returns
// store.ErrNotImplemented and these endpoints are tested against real Postgres via
// testcontainers instead.
type inMemoryDashboardRepository struct{}

var _ store.DashboardRepository = (*inMemoryDashboardRepository)(nil)

func newDashboardRepository() *inMemoryDashboardRepository {
	return &inMemoryDashboardRepository{}
}

func NewDashboardRepository() store.DashboardRepository {
	return newDashboardRepository()
}

func (r *inMemoryDashboardRepository) Spotlight(semesterID uuid.UUID, now time.Time) (*store.SpotlightEvent, error) {
	return nil, store.ErrNotImplemented
}

func (r *inMemoryDashboardRepository) MembershipStats(semesterID uuid.UUID) (store.MembershipStats, error) {
	return store.MembershipStats{}, store.ErrNotImplemented
}

func (r *inMemoryDashboardRepository) EngagementStats(semesterID uuid.UUID, asOf time.Time) (store.EngagementStats, error) {
	return store.EngagementStats{}, store.ErrNotImplemented
}

func (r *inMemoryDashboardRepository) EventActivity(semesterID uuid.UUID, asOf time.Time) (store.EventActivityStats, []store.EventSeriesPoint, error) {
	return store.EventActivityStats{}, nil, store.ErrNotImplemented
}

func (r *inMemoryDashboardRepository) SignupTimeline(semesterID uuid.UUID, now time.Time) (store.SignupTimeline, error) {
	return store.SignupTimeline{}, store.ErrNotImplemented
}

func (r *inMemoryDashboardRepository) MembershipTotalAsOf(semesterID uuid.UUID, asOf time.Time) (*int64, error) {
	return nil, store.ErrNotImplemented
}
