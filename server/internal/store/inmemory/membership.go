package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

	for _, m := range r.memberships {
		if m.UserID == membership.UserID && m.SemesterID == membership.SemesterID {
			return fmt.Errorf("membership for user %d in semester %s already exists", membership.UserID, membership.SemesterID)
		}
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

// List retrieves memberships matching filter, with a computed attendance count.
// Unlike the Postgres implementation, the in-memory store has no cross-repository join
// capability, so Attendance is always 0 here — this mirrors the existing accepted gap where
// in-memory repositories don't enforce or compute cross-table relationships. Tests that need a
// non-zero attendance count should assert against the Postgres implementation instead.
func (r *inMemoryMembershipRepository) List(filter *models.ListMembershipsFilter) ([]models.MembershipWithAttendance, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*models.Membership
	for _, m := range r.memberships {
		if filter.SemesterID != nil && m.SemesterID != *filter.SemesterID {
			continue
		}
		if filter.UserID != nil && m.UserID != *filter.UserID {
			continue
		}
		if filter.Paid != nil && m.Paid != *filter.Paid {
			continue
		}
		if filter.Discounted != nil && m.Discounted != *filter.Discounted {
			continue
		}
		if !matchesJoinedFilter(m, filter) {
			continue
		}
		matched = append(matched, m)
	}

	sort.Slice(matched, func(i, j int) bool {
		iName, jName := "", ""
		if matched[i].User != nil {
			iName = matched[i].User.FirstName + matched[i].User.LastName
		}
		if matched[j].User != nil {
			jName = matched[j].User.FirstName + matched[j].User.LastName
		}
		return iName < jName
	})

	total := int64(len(matched))

	offset := 0
	if filter.Pagination.Offset != nil && *filter.Pagination.Offset > 0 {
		offset = *filter.Pagination.Offset
	}
	if offset >= len(matched) {
		return []models.MembershipWithAttendance{}, total, nil
	}
	matched = matched[offset:]
	if filter.Pagination.Limit != nil && *filter.Pagination.Limit > 0 && *filter.Pagination.Limit < len(matched) {
		matched = matched[:*filter.Pagination.Limit]
	}

	results := make([]models.MembershipWithAttendance, len(matched))
	for i, m := range matched {
		results[i] = models.MembershipWithAttendance{Membership: *m, Attendance: 0}
	}

	return results, total, nil
}

func matchesJoinedFilter(m *models.Membership, filter *models.ListMembershipsFilter) bool {
	if filter.Search == "" && filter.Name == nil && filter.Email == nil && filter.Faculty == nil && filter.StudentID == nil {
		return true
	}
	if m.User == nil {
		return false
	}
	full := strings.ToLower(m.User.FirstName + " " + m.User.LastName)
	if filter.Search != "" {
		s := strings.ToLower(filter.Search)
		if !strings.Contains(strings.ToLower(m.User.FirstName), s) && !strings.Contains(strings.ToLower(m.User.LastName), s) &&
			!strings.Contains(strings.ToLower(m.User.Email), s) && !strings.Contains(full, s) {
			return false
		}
	}
	if filter.Name != nil {
		n := strings.ToLower(*filter.Name)
		if !strings.Contains(strings.ToLower(m.User.FirstName), n) && !strings.Contains(strings.ToLower(m.User.LastName), n) && !strings.Contains(full, n) {
			return false
		}
	}
	if filter.Email != nil && !strings.Contains(strings.ToLower(m.User.Email), strings.ToLower(*filter.Email)) {
		return false
	}
	if filter.Faculty != nil && m.User.Faculty != *filter.Faculty {
		return false
	}
	if filter.StudentID != nil && strconv.FormatUint(m.User.ID, 10) != *filter.StudentID {
		return false
	}
	return true
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
