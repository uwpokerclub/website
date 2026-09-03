package store

import "api/internal/models"

// EventClockRepository is the interface for accessing event clocks in the data
// store. A row exists only once an event's clock has been read or acted on at
// least once — see models.EventClock and services.EventClockService.
type EventClockRepository interface {
	// FindByEventID retrieves an event's clock by its event ID. It returns
	// ErrNotFound if no clock exists yet for the event.
	FindByEventID(eventID int32) (models.EventClock, error)

	// FindByEventIDForUpdate retrieves an event's clock by its event ID,
	// locking the row for the remainder of the enclosing transaction. It must
	// only be called within a transaction started by Store.BeginTx. It
	// returns ErrNotFound if no clock exists yet for the event.
	FindByEventIDForUpdate(eventID int32) (models.EventClock, error)

	// Create creates a new clock in the data store. It returns
	// ErrAlreadyExists, without creating anything, if a clock already exists
	// for the event.
	Create(clock *models.EventClock) error

	// Update writes every mutable field of an existing clock back to the data
	// store.
	Update(clock *models.EventClock) error

	// DeleteByEventID deletes an event's clock by its event ID. It returns
	// ErrNotFound if no clock exists for the event.
	DeleteByEventID(eventID int32) error
}
