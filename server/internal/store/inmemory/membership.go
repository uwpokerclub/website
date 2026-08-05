package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type inMemoryMembershipRepository struct {
	mu          sync.RWMutex
	memberships map[uuid.UUID]*models.Membership
}

var _ store.MembershipRepository = (*inMemoryMembershipRepository)(nil)

func newMembershipRepository() *inMemoryMembershipRepository {
	return &inMemoryMembershipRepository{
		memberships: make(map[uuid.UUID]*models.Membership),
	}
}

func NewMembershipRepository() store.MembershipRepository {
	return newMembershipRepository()
}

func (r *inMemoryMembershipRepository) clone() *inMemoryMembershipRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemoryMembershipRepository{
		memberships: make(map[uuid.UUID]*models.Membership, len(r.memberships)),
	}
	for id, m := range r.memberships {
		mc := *m
		c.memberships[id] = &mc
	}
	return c
}

func (r *inMemoryMembershipRepository) Create(membership *models.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if membership.ID == uuid.Nil {
		membership.ID = uuid.New()
	} else if _, exists := r.memberships[membership.ID]; exists {
		return fmt.Errorf("membership with ID %s already exists", membership.ID)
	}

	copy := *membership
	r.memberships[membership.ID] = &copy

	return nil
}

func (r *inMemoryMembershipRepository) FindByID(id uuid.UUID) (models.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	membership, exists := r.memberships[id]
	if !exists {
		return models.Membership{}, store.ErrNotFound
	}

	return *membership, nil
}

func (r *inMemoryMembershipRepository) FindByIDAndSemesterID(id uuid.UUID, semesterID uuid.UUID) (models.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	membership, exists := r.memberships[id]
	if !exists || membership.SemesterID != semesterID {
		return models.Membership{}, store.ErrNotFound
	}

	return *membership, nil
}

func (r *inMemoryMembershipRepository) List(filter *models.ListMembershipsFilter) ([]models.Membership, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var memberships []models.Membership
	for _, membership := range r.memberships {
		if filter.SemesterID != nil && membership.SemesterID != *filter.SemesterID {
			continue
		}
		if filter.UserID != nil && membership.UserID != *filter.UserID {
			continue
		}
		if filter.Paid != nil && membership.Paid != *filter.Paid {
			continue
		}
		if filter.Discounted != nil && membership.Discounted != *filter.Discounted {
			continue
		}
		memberships = append(memberships, *membership)
	}

	sort.Slice(memberships, func(i, j int) bool {
		return memberships[i].UserID < memberships[j].UserID
	})

	total := int64(len(memberships))

	offset := 0
	if filter.Pagination.Offset != nil && *filter.Pagination.Offset > 0 {
		offset = *filter.Pagination.Offset
	}

	if offset >= len(memberships) {
		return []models.Membership{}, total, nil
	}

	memberships = memberships[offset:]

	if filter.Pagination.Limit != nil && *filter.Pagination.Limit > 0 && *filter.Pagination.Limit < len(memberships) {
		memberships = memberships[:*filter.Pagination.Limit]
	}

	return memberships, total, nil
}

func (r *inMemoryMembershipRepository) Update(membership *models.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.memberships[membership.ID]
	if !exists {
		return store.ErrNotFound
	}

	existing.Paid = membership.Paid
	existing.Discounted = membership.Discounted

	return nil
}

func (r *inMemoryMembershipRepository) Delete(id uuid.UUID, semesterID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	membership, exists := r.memberships[id]
	if !exists || membership.SemesterID != semesterID {
		return store.ErrNotFound
	}

	delete(r.memberships, id)

	return nil
}
