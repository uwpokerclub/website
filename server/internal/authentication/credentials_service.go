package authentication

import (
	"api/internal/models"
	"api/internal/store"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type credentialsService struct {
	store store.Store
}

func NewCredentialService(st store.Store) *credentialsService {
	return &credentialsService{
		store: st,
	}
}

func (svc *credentialsService) Validate(username string, password string) (bool, string, error) {
	return svc.validate(username, password, false)
}

func (svc *credentialsService) ValidateForSession(username string, password string) (bool, string, error) {
	return svc.validate(username, password, true)
}

func (svc *credentialsService) validate(username string, password string, forSession bool) (bool, string, error) {
	// Find the login with the specified username
	var login models.Login
	var err error
	if forSession {
		login, err = svc.store.Logins().FindByUsernameForUpdate(username)
	} else {
		login, err = svc.store.Logins().FindByUsername(username)
	}
	if err != nil {
		// A missing login is not an error to the caller, just a failed validation
		if errors.Is(err, store.ErrNotFound) {
			return false, "", nil
		}

		// Any other error is a server error
		return false, "", fmt.Errorf("find login: %w", err)
	}

	// Keep inactive accounts indistinguishable from incorrect credentials.
	if login.Status != models.LoginStatusActive {
		return false, "", nil
	}

	// Compare the hashed password and the plaintext password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(login.Password), []byte(password))
	if err != nil {
		return false, "", nil
	}

	return true, login.Role, nil
}
