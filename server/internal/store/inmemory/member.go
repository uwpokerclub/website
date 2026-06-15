package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type inMemoryMemberRepository struct {
	mu      sync.RWMutex
	members map[uint64]*models.User
}

var _ store.MemberRepository = (*inMemoryMemberRepository)(nil)

func newMemberRepository() *inMemoryMemberRepository {
	return &inMemoryMemberRepository{
		members: make(map[uint64]*models.User),
	}
}

func NewMemberRepository() store.MemberRepository {
	return newMemberRepository()
}

func (r *inMemoryMemberRepository) clone() *inMemoryMemberRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemoryMemberRepository{
		members: make(map[uint64]*models.User, len(r.members)),
	}
	for id, u := range r.members {
		uc := *u
		c.members[id] = &uc
	}
	return c
}

func (r *inMemoryMemberRepository) Create(member *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.members[member.ID]; exists {
		return fmt.Errorf("member with ID %d already exists", member.ID)
	}

	copy := *member
	r.members[copy.ID] = &copy

	return nil
}

func (r *inMemoryMemberRepository) FindByID(id uint64) (models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	member, exists := r.members[id]
	if !exists {
		return models.User{}, store.ErrNotFound
	}

	return *member, nil
}

func (r *inMemoryMemberRepository) List(filter *models.ListUsersFilter, pagination *models.Pagination) ([]models.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var members []models.User
	for _, member := range r.members {
		if filter.ID != nil && member.ID != *filter.ID {
			continue
		}
		if filter.Name != nil {
			full := strings.ToLower(member.FirstName + " " + member.LastName)
			if !strings.Contains(full, strings.ToLower(*filter.Name)) {
				continue
			}
		}
		if filter.Email != nil && !strings.Contains(strings.ToLower(member.Email), strings.ToLower(*filter.Email)) {
			continue
		}
		if filter.Faculty != nil && member.Faculty != *filter.Faculty {
			continue
		}
		members = append(members, *member)
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].CreatedAt.After(members[j].CreatedAt)
	})

	total := int64(len(members))

	offset := 0
	if pagination.Offset != nil && *pagination.Offset > 0 {
		offset = *pagination.Offset
	}

	if offset >= len(members) {
		return []models.User{}, total, nil
	}

	members = members[offset:]

	if pagination.Limit != nil && *pagination.Limit > 0 && *pagination.Limit < len(members) {
		members = members[:*pagination.Limit]
	}

	return members, total, nil
}

func (r *inMemoryMemberRepository) Update(member *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.members[member.ID]; !exists {
		return store.ErrNotFound
	}

	copy := *member
	r.members[copy.ID] = &copy

	return nil
}

func (r *inMemoryMemberRepository) Delete(id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.members[id]; !exists {
		return store.ErrNotFound
	}

	delete(r.members, id)

	return nil
}
