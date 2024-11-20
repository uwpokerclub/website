package authentication

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sessionManager struct {
	db *gorm.DB
}

func NewSessionManager(db *gorm.DB) *sessionManager {
	return &sessionManager{
		db: db,
	}
}

func (svc *sessionManager) Create(username string) (uuid.UUID, error) {
	// Get the current time
	now := time.Now().UTC()
	// Set the expiry time to 8 hours in the future
	expiry := now.Add(time.Hour * 8).UTC()

	// Create the session in the database
	session := models.Session{StartedAt: time.Now(), ExpiresAt: expiry, Username: username}
	res := svc.db.Create(&session)

	if err := res.Error; err != nil {
		return uuid.UUID{}, e.InternalServerError(err.Error())
	}

	return session.ID, nil
}

func (svc *sessionManager) Invalidate(sessionID uuid.UUID) error {
	session := models.Session{ID: sessionID}

	res := svc.db.Delete(&session)
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}

func (svc *sessionManager) Authenticate(sessionID uuid.UUID) error {
	session := models.Session{ID: sessionID}
	res := svc.db.First(&session)

	// Check if session exists
	err := res.Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return e.Unauthorized("Authentication required")
	}

	if err != nil {
		return e.InternalServerError(err.Error())
	}

	// Check if session has expired, if it is delete it from the table and return 401
	if time.Now().UTC().After(session.ExpiresAt) {
		res = svc.db.Delete(&session)
		if err := res.Error; err != nil {
			return e.InternalServerError(err.Error())
		}

		return e.Unauthorized("Session has expired. Please reauthenticate")
	}

	return nil
}

func (svc *sessionManager) Get(sessionID uuid.UUID) (*models.GetSessionResponse, error) {
	session := models.Session{ID: sessionID}
	res := svc.db.First(&session)

	err := res.Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.Unauthorized("Authentication required")
	}

	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	response := models.GetSessionResponse{
		Username: session.Username,
	}

	return &response, nil
}
