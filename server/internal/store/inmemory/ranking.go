package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type inMemoryRankingRepository struct {
	mu       sync.RWMutex
	rankings map[int64]*models.Ranking
	nextID   int64
}

var _ store.RankingRepository = (*inMemoryRankingRepository)(nil)

func newRankingRepository() *inMemoryRankingRepository {
	return &inMemoryRankingRepository{
		rankings: make(map[int64]*models.Ranking),
	}
}

func NewRankingRepository() store.RankingRepository {
	return newRankingRepository()
}

func (r *inMemoryRankingRepository) clone() *inMemoryRankingRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemoryRankingRepository{
		rankings: make(map[int64]*models.Ranking, len(r.rankings)),
		nextID:   r.nextID,
	}
	for id, rk := range r.rankings {
		rc := *rk
		c.rankings[id] = &rc
	}
	return c
}

// findByMembershipIDLocked must be called with r.mu already held.
func (r *inMemoryRankingRepository) findByMembershipIDLocked(membershipID uuid.UUID) *models.Ranking {
	for _, rk := range r.rankings {
		if rk.MembershipID == membershipID {
			return rk
		}
	}
	return nil
}

func (r *inMemoryRankingRepository) Create(ranking *models.Ranking) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing := r.findByMembershipIDLocked(ranking.MembershipID); existing != nil {
		return fmt.Errorf("ranking for membership %s already exists", ranking.MembershipID)
	}

	if ranking.ID == 0 {
		r.nextID++
		ranking.ID = r.nextID
	} else if _, exists := r.rankings[ranking.ID]; exists {
		return fmt.Errorf("ranking with ID %d already exists", ranking.ID)
	}

	copy := *ranking
	r.rankings[ranking.ID] = &copy

	return nil
}

func (r *inMemoryRankingRepository) FindByMembershipID(membershipID uuid.UUID) (models.Ranking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ranking := r.findByMembershipIDLocked(membershipID)
	if ranking == nil {
		return models.Ranking{}, store.ErrNotFound
	}

	return *ranking, nil
}

func (r *inMemoryRankingRepository) Update(ranking *models.Ranking) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.rankings[ranking.ID]
	if !exists {
		return store.ErrNotFound
	}

	existing.Points = ranking.Points

	return nil
}

func (r *inMemoryRankingRepository) BatchIncrementPoints(updates map[uuid.UUID]int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for membershipID, points := range updates {
		if existing := r.findByMembershipIDLocked(membershipID); existing != nil {
			existing.Points += points
			continue
		}

		r.nextID++
		r.rankings[r.nextID] = &models.Ranking{
			ID:           r.nextID,
			MembershipID: membershipID,
			Points:       points,
		}
	}

	return nil
}
