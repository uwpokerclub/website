package authentication

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/store"
	"errors"
	"time"

	"github.com/google/uuid"
)

type sessionManager struct {
	store store.Store
}

func NewSessionManager(st store.Store) *sessionManager {
	return &sessionManager{
		store: st,
	}
}

func (svc *sessionManager) Create(username string, role string) (uuid.UUID, error) {
	// Get the current time
	now := time.Now().UTC()
	// Set the expiry time to 8 hours in the future
	expiry := now.Add(time.Hour * 8).UTC()

	// Create the session in the database
	session := models.Session{StartedAt: now, ExpiresAt: expiry, Username: username, Role: role}
	if err := svc.store.Sessions().Create(&session); err != nil {
		return uuid.UUID{}, e.InternalServerError(err.Error())
	}

	return session.ID, nil
}

func (svc *sessionManager) Invalidate(sessionID uuid.UUID) error {
	// Deleting a session that no longer exists is not an error: the caller's
	// intent (this session must not be usable) is already satisfied. The
	// previous GORM implementation reported RowsAffected == 0 as success, and
	// logout depends on that.
	if err := svc.store.Sessions().Delete(sessionID); err != nil && !errors.Is(err, store.ErrNotFound) {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (svc *sessionManager) Authenticate(sessionID uuid.UUID) (*models.Session, error) {
	session, err := svc.store.Sessions().FindByID(sessionID)
	if err != nil {
		// Check if session exists
		if errors.Is(err, store.ErrNotFound) {
			return nil, e.Unauthorized("Authentication required")
		}

		return nil, e.InternalServerError(err.Error())
	}

	// Check if session has expired, if it is delete it from the table and return 401
	if time.Now().UTC().After(session.ExpiresAt) {
		// A concurrent logout may have already removed it; that is not a failure.
		if err := svc.store.Sessions().Delete(session.ID); err != nil && !errors.Is(err, store.ErrNotFound) {
			return nil, e.InternalServerError(err.Error())
		}

		return nil, e.Unauthorized("Session has expired. Please reauthenticate")
	}

	return &session, nil
}
