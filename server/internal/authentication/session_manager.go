package authentication

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrSessionNotFound is returned when no session exists for the given ID.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionExpired is returned when a session exists but is past its expiry.
// The session is deleted before this error is returned.
var ErrSessionExpired = errors.New("session has expired")

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
		return uuid.UUID{}, fmt.Errorf("create session: %w", err)
	}

	return session.ID, nil
}

func (svc *sessionManager) Invalidate(sessionID uuid.UUID) error {
	// Deleting a session that no longer exists is not an error: the caller's
	// intent (this session must not be usable) is already satisfied. The
	// previous GORM implementation reported RowsAffected == 0 as success, and
	// logout depends on that.
	if err := svc.store.Sessions().Delete(sessionID); err != nil && !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("invalidate session: %w", err)
	}

	return nil
}

func (svc *sessionManager) Authenticate(sessionID uuid.UUID) (*models.Session, error) {
	session, err := svc.store.Sessions().FindByID(sessionID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf("find session: %w", err)
	}

	// Check if session has expired, if it is delete it from the table and return 401
	if time.Now().UTC().After(session.ExpiresAt) {
		// A concurrent logout may have already removed it; that is not a failure.
		if err := svc.store.Sessions().Delete(session.ID); err != nil && !errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("delete expired session: %w", err)
		}

		return nil, ErrSessionExpired
	}

	return &session, nil
}
