package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type inMemoryEntryRepository struct {
	mu           sync.RWMutex
	participants map[int32]*models.Participant
	nextID       int32
}

var _ store.EntryRepository = (*inMemoryEntryRepository)(nil)

func newEntryRepository() *inMemoryEntryRepository {
	return &inMemoryEntryRepository{
		participants: make(map[int32]*models.Participant),
	}
}

func NewEntryRepository() store.EntryRepository {
	return newEntryRepository()
}

func (r *inMemoryEntryRepository) clone() *inMemoryEntryRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemoryEntryRepository{
		participants: make(map[int32]*models.Participant, len(r.participants)),
		nextID:       r.nextID,
	}
	for id, p := range r.participants {
		pc := *p
		c.participants[id] = &pc
	}
	return c
}

func (r *inMemoryEntryRepository) Create(participant *models.Participant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if participant.ID == 0 {
		r.nextID++
		participant.ID = r.nextID
	} else if _, exists := r.participants[participant.ID]; exists {
		return fmt.Errorf("entry with ID %d already exists", participant.ID)
	}

	if participant.MembershipID != nil {
		for _, p := range r.participants {
			if p.MembershipID != nil && *p.MembershipID == *participant.MembershipID && p.EventID == participant.EventID {
				return fmt.Errorf(
					"entry for membership %s in event %d already exists",
					*participant.MembershipID, participant.EventID,
				)
			}
		}
	}

	copy := *participant
	r.participants[participant.ID] = &copy

	return nil
}

func (r *inMemoryEntryRepository) FindByID(id int32) (models.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participant, exists := r.participants[id]
	if !exists {
		return models.Participant{}, store.ErrNotFound
	}

	return *participant, nil
}

func (r *inMemoryEntryRepository) FindByMembershipAndEventID(membershipID uuid.UUID, eventID int32) (models.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.participants {
		if p.MembershipID != nil && *p.MembershipID == membershipID && p.EventID == eventID {
			return *p, nil
		}
	}

	return models.Participant{}, store.ErrNotFound
}

func (r *inMemoryEntryRepository) List(filter *models.ListParticipantsFilter) ([]models.Participant, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var participants []models.Participant
	for _, p := range r.participants {
		if p.EventID != filter.EventID {
			continue
		}
		participants = append(participants, *p)
	}

	// Matches "ORDER BY signed_out_at DESC" under Postgres' default NULLS FIRST-on-DESC
	// behavior: not-yet-signed-out entries (nil) sort first, then signed-out entries by
	// most recent first.
	sort.Slice(participants, func(i, j int) bool {
		a, b := participants[i].SignedOutAt, participants[j].SignedOutAt
		switch {
		case a == nil && b == nil:
			return false
		case a == nil:
			return true
		case b == nil:
			return false
		default:
			return a.After(*b)
		}
	})

	total := int64(len(participants))

	offset := 0
	if filter.Pagination.Offset != nil && *filter.Pagination.Offset > 0 {
		offset = *filter.Pagination.Offset
	}

	if offset >= len(participants) {
		return []models.Participant{}, total, nil
	}

	participants = participants[offset:]

	if filter.Pagination.Limit != nil && *filter.Pagination.Limit > 0 &&
		*filter.Pagination.Limit < len(participants) {
		participants = participants[:*filter.Pagination.Limit]
	}

	return participants, total, nil
}

func (r *inMemoryEntryRepository) Update(participant *models.Participant, values map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.participants[participant.ID]
	if !exists {
		return store.ErrNotFound
	}

	if v, ok := values["signed_out_at"]; ok {
		if v == nil {
			existing.SignedOutAt = nil
		} else {
			existing.SignedOutAt = v.(*time.Time)
		}
	}

	*participant = *existing

	return nil
}

func (r *inMemoryEntryRepository) Delete(membershipID uuid.UUID, eventID int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, p := range r.participants {
		if p.MembershipID != nil && *p.MembershipID == membershipID && p.EventID == eventID {
			delete(r.participants, id)
			return nil
		}
	}

	return store.ErrNotFound
}
