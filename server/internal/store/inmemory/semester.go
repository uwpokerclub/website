package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type inMemorySemesterRepository struct {
	mu        sync.RWMutex
	semesters map[uuid.UUID]*models.Semester
}

var _ store.SemesterRepository = (*inMemorySemesterRepository)(nil)

func NewSemesterRepository() store.SemesterRepository {
	return &inMemorySemesterRepository{
		semesters: make(map[uuid.UUID]*models.Semester),
	}
}

func (r *inMemorySemesterRepository) Create(semester *models.Semester) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if semester.ID == uuid.Nil {
		semester.ID = uuid.New()
	}
	copy := *semester
	r.semesters[semester.ID] = &copy
	return nil
}

func (r *inMemorySemesterRepository) FindByID(id uuid.UUID) (*models.Semester, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	semester, exists := r.semesters[id]
	if !exists {
		return nil, store.ErrNotFound
	}
	return semester, nil
}

func (r *inMemorySemesterRepository) List(pagination *models.Pagination) ([]*models.Semester, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var semesters []*models.Semester
	for _, semester := range r.semesters {
		semesters = append(semesters, semester)
	}
	sort.Slice(semesters, func(i, j int) bool {
		return semesters[i].StartDate.After(semesters[j].StartDate)
	})
	total := int64(len(semesters))

	offset := 0
	if pagination.Offset != nil && *pagination.Offset > 0 {
		offset = *pagination.Offset
	}
	if offset >= len(semesters) {
		return []*models.Semester{}, total, nil
	}
	semesters = semesters[offset:]

	if pagination.Limit != nil && *pagination.Limit > 0 && *pagination.Limit < len(semesters) {
		semesters = semesters[:*pagination.Limit]
	}

	return semesters, total, nil
}

func (r *inMemorySemesterRepository) Update(semester *models.Semester) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.semesters[semester.ID]
	if !exists {
		return store.ErrNotFound
	}
	r.semesters[semester.ID] = semester
	return nil
}
