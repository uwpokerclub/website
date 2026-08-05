package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

// EntryRepository is the interface for accessing the entries (event participants) in the data store.
// It provides methods for creating, reading, updating, listing, and deleting entries.
type EntryRepository interface {
	// Create creates a new entry in the data store.
	Create(participant *models.Participant) error

	// FindByID retrieves an entry from the data store by its ID.
	FindByID(id int32) (models.Participant, error)

	// FindByMembershipAndEventID retrieves an entry from the data store by its membership ID,
	// scoped to a specific event. It returns store.ErrNotFound if no entry exists for the given
	// membership within the given event.
	FindByMembershipAndEventID(membershipID uuid.UUID, eventID int32) (models.Participant, error)

	// List retrieves entries for a given event, preloaded with their membership (and the
	// membership's user, semester, and ranking), ordered by signed-out time descending
	// (entries that have not yet signed out sort first), along with the total matching count
	// before pagination is applied.
	List(filter *models.ListParticipantsFilter) ([]models.Participant, int64, error)

	// Update applies a partial update to an entry using the given column/value map, and writes
	// the applied values back onto participant.
	Update(participant *models.Participant, values map[string]any) error

	// Delete deletes an entry from the data store by its membership ID, scoped to a specific
	// event. Returns store.ErrNotFound if no matching record exists.
	Delete(membershipID uuid.UUID, eventID int32) error
}
