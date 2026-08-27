package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

// EventRepository is the interface for accessing the events in the data store. It provides methods for creating, reading, updating, and listing events.
type EventRepository interface {
	// Create creates a new event in the data store.
	Create(event *models.Event) error

	// FindByID retrieves an event from the data store by its ID, preloaded with
	// its semester, structure (with blinds), and entries.
	FindByID(id int32) (models.Event, error)

	// FindBySemesterAndID retrieves an event from the data store scoped to a
	// specific semester, preloaded with its semester, structure (with blinds),
	// and entries. It returns store.ErrNotFound if no event with the given ID
	// exists within the given semester.
	FindBySemesterAndID(semesterID uuid.UUID, id int32) (models.Event, error)

	// List retrieves events matching the given filter (semester — or every semester if nil —
	// plus optional name search), ordered by start date descending, along with the total
	// matching count before pagination is applied.
	List(filter *models.ListEventsFilter) ([]models.Event, int64, error)

	// Update applies a partial update to an event using the given column/value
	// map, and writes the applied values back onto event.
	Update(event *models.Event, values map[string]any) error

	// Delete removes an event from the data store by its ID. The event's entries are removed
	// with it by the participants.event_id foreign key (ON DELETE CASCADE). It returns
	// store.ErrNotFound if no event with the given ID exists.
	Delete(id int32) error
}
