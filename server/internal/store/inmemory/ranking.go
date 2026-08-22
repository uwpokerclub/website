package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"strings"
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

func (r *inMemoryRankingRepository) rankingsInSemesterLocked(semesterID uuid.UUID) []*models.Ranking {
	var results []*models.Ranking
	for _, rk := range r.rankings {
		if rk.Membership != nil && rk.Membership.SemesterID == semesterID {
			results = append(results, rk)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Points > results[j].Points
	})
	return results
}

func (r *inMemoryRankingRepository) FindBySemesterAndMembershipID(semesterID uuid.UUID, membershipID uuid.UUID) (models.GetRankingResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ranked := r.rankingsInSemesterLocked(semesterID)
	for i, rk := range ranked {
		if rk.MembershipID == membershipID {
			return models.GetRankingResponse{Points: rk.Points, Position: int32(i + 1)}, nil
		}
	}

	return models.GetRankingResponse{}, store.ErrNotFound
}

func (r *inMemoryRankingRepository) List(filter *models.ListRankingsFilter) ([]models.RankingResponse, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ranked := r.rankingsInSemesterLocked(filter.SemesterID)

	var matched []models.RankingResponse
	for i, rk := range ranked {
		var firstName, lastName string
		var userID uint64
		if rk.Membership.User != nil {
			firstName = rk.Membership.User.FirstName
			lastName = rk.Membership.User.LastName
			userID = rk.Membership.User.ID
		}
		if filter.Search != "" {
			s := strings.ToLower(filter.Search)
			full := strings.ToLower(firstName + " " + lastName)
			if !strings.Contains(strings.ToLower(firstName), s) && !strings.Contains(strings.ToLower(lastName), s) && !strings.Contains(full, s) {
				continue
			}
		}
		matched = append(matched, models.RankingResponse{
			ID: userID, FirstName: firstName, LastName: lastName, Points: rk.Points, Position: int32(i + 1),
		})
	}

	total := int64(len(matched))

	offset := 0
	if filter.Pagination.Offset != nil && *filter.Pagination.Offset > 0 {
		offset = *filter.Pagination.Offset
	}
	if offset >= len(matched) {
		return []models.RankingResponse{}, total, nil
	}
	matched = matched[offset:]
	if filter.Pagination.Limit != nil && *filter.Pagination.Limit > 0 && *filter.Pagination.Limit < len(matched) {
		matched = matched[:*filter.Pagination.Limit]
	}

	return matched, total, nil
}
