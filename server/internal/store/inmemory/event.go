package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type inMemoryEventRepository struct {
	mu     sync.RWMutex
	events map[int32]*models.Event
	nextID int32
}

var _ store.EventRepository = (*inMemoryEventRepository)(nil)

func newEventRepository() *inMemoryEventRepository {
	return &inMemoryEventRepository{
		events: make(map[int32]*models.Event),
	}
}

func NewEventRepository() store.EventRepository {
	return newEventRepository()
}

func (r *inMemoryEventRepository) clone() *inMemoryEventRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemoryEventRepository{
		events: make(map[int32]*models.Event, len(r.events)),
		nextID: r.nextID,
	}
	for id, e := range r.events {
		ec := *e
		c.events[id] = &ec
	}
	return c
}

func (r *inMemoryEventRepository) Create(event *models.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event.ID == 0 {
		r.nextID++
		event.ID = r.nextID
	} else if _, exists := r.events[event.ID]; exists {
		return fmt.Errorf("event with ID %d already exists", event.ID)
	}

	copy := *event
	r.events[event.ID] = &copy

	return nil
}

func (r *inMemoryEventRepository) FindByID(id int32) (models.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, exists := r.events[id]
	if !exists {
		return models.Event{}, store.ErrNotFound
	}

	return *event, nil
}

func (r *inMemoryEventRepository) FindBySemesterAndID(semesterID uuid.UUID, id int32) (models.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, exists := r.events[id]
	if !exists || event.SemesterID != semesterID {
		return models.Event{}, store.ErrNotFound
	}

	return *event, nil
}

func (r *inMemoryEventRepository) List(filter *models.ListEventsFilter) ([]models.Event, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var events []models.Event
	for _, event := range r.events {
		if filter.SemesterID != nil && event.SemesterID != *filter.SemesterID {
			continue
		}
		if filter.Search != "" &&
			!strings.Contains(strings.ToLower(event.Name), strings.ToLower(filter.Search)) {
			continue
		}
		events = append(events, *event)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartDate.After(events[j].StartDate)
	})

	total := int64(len(events))

	offset := 0
	if filter.Pagination.Offset != nil && *filter.Pagination.Offset > 0 {
		offset = *filter.Pagination.Offset
	}

	if offset >= len(events) {
		return []models.Event{}, total, nil
	}

	events = events[offset:]

	if filter.Pagination.Limit != nil && *filter.Pagination.Limit > 0 &&
		*filter.Pagination.Limit < len(events) {
		events = events[:*filter.Pagination.Limit]
	}

	return events, total, nil
}

func (r *inMemoryEventRepository) Update(event *models.Event, values map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.events[event.ID]
	if !exists {
		return store.ErrNotFound
	}

	for key, value := range values {
		switch key {
		case "name":
			existing.Name = value.(string)
		case "format":
			existing.Format = value.(string)
		case "notes":
			if value == nil {
				existing.Notes = ""
			} else {
				existing.Notes = value.(string)
			}
		case "start_date":
			existing.StartDate = value.(time.Time)
		case "points_multiplier":
			existing.PointsMultiplier = value.(float32)
		case "state":
			existing.State = uint8(value.(int))
		case "rebuys":
			existing.Rebuys = value.(uint8)
		}
	}

	*event = *existing

	return nil
}
