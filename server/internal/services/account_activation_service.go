package services

import (
	"api/internal/authentication"
	"api/internal/models"
	"api/internal/store"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrActivationNotFound = errors.New("activation token not found")

type accountActivationService struct{ store store.Store }

func NewAccountActivationService(st store.Store) *accountActivationService {
	return &accountActivationService{store: st}
}

// Create mints a 32-byte base64url token and persists only its SHA-256 hash.
func (svc *accountActivationService) Create(username string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate activation token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	activation := models.AccountActivation{TokenHash: hash[:], Username: username, ExpiresAt: time.Now().UTC().Add(14 * 24 * time.Hour)}
	tx, err := svc.store.BeginTx()
	if err != nil {
		return "", fmt.Errorf("begin activation transaction: %w", err)
	}
	defer tx.Rollback()
	if err := tx.AccountActivations().DeleteUnusedByUsername(username); err != nil {
		return "", fmt.Errorf("revoke prior activations: %w", err)
	}
	if err := tx.AccountActivations().Create(&activation); err != nil {
		return "", fmt.Errorf("create activation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit activation transaction: %w", err)
	}
	return token, nil
}

func (svc *accountActivationService) Verify(token string) (models.AccountActivation, error) {
	activation, err := svc.store.AccountActivations().FindValid(tokenHash(token))
	if errors.Is(err, store.ErrNotFound) {
		return models.AccountActivation{}, ErrActivationNotFound
	}
	if err != nil {
		return models.AccountActivation{}, fmt.Errorf("find activation: %w", err)
	}
	return activation, nil
}

// Complete atomically consumes the token, sets the password and activates the login.
// Transition-bound tokens additionally prove that their exact pending transition
// still exists; only its president completes the handover.
func (svc *accountActivationService) Complete(token, password string) (models.Login, uuid.UUID, error) {
	for attempt := 0; attempt < 2; attempt++ {
		login, sessionToken, err := svc.completeOnce(token, password)
		if !errors.Is(err, store.ErrTransactionConflict) {
			return login, sessionToken, err
		}
	}
	return models.Login{}, uuid.UUID{}, fmt.Errorf("complete activation: %w", store.ErrTransactionConflict)
}

func (svc *accountActivationService) completeOnce(token, password string) (models.Login, uuid.UUID, error) {
	tx, err := svc.store.BeginTx()
	if err != nil {
		return models.Login{}, uuid.UUID{}, fmt.Errorf("begin activation transaction: %w", err)
	}
	defer tx.Rollback()
	activation, err := tx.AccountActivations().Consume(tokenHash(token))
	if errors.Is(err, store.ErrNotFound) {
		return models.Login{}, uuid.UUID{}, ErrActivationNotFound
	}
	if err != nil {
		return models.Login{}, uuid.UUID{}, fmt.Errorf("consume activation: %w", err)
	}
	var transition *models.OfficerTransition
	if activation.TransitionID != nil {
		locked, e := tx.OfficerTransitions().FindByIDForUpdate(*activation.TransitionID)
		if errors.Is(e, store.ErrNotFound) {
			return models.Login{}, uuid.UUID{}, ErrActivationNotFound
		}
		if e != nil {
			return models.Login{}, uuid.UUID{}, fmt.Errorf("lock activation transition: %w", e)
		}
		if locked.Status != models.OfficerTransitionPending || !isTransitionNominee(locked, activation.Username) {
			return models.Login{}, uuid.UUID{}, ErrActivationNotFound
		}
		transition = &locked
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.Login{}, uuid.UUID{}, fmt.Errorf("hash password: %w", err)
	}
	var login models.Login
	if transition != nil && activation.Username == transition.PresidentUsername {
		claimed, e := tx.OfficerTransitions().ClaimCompletion(transition.ID)
		if errors.Is(e, store.ErrNotFound) {
			return models.Login{}, uuid.UUID{}, ErrActivationNotFound
		}
		if e != nil {
			return models.Login{}, uuid.UUID{}, fmt.Errorf("claim officer transition: %w", e)
		}
		login, err = completeOfficerTransition(tx, claimed, string(hash))
	} else {
		login, err = tx.Logins().Activate(activation.Username, string(hash))
	}
	if errors.Is(err, store.ErrNotFound) {
		return models.Login{}, uuid.UUID{}, ErrActivationNotFound
	}
	if err != nil {
		return models.Login{}, uuid.UUID{}, fmt.Errorf("activate login: %w", err)
	}
	sessionToken, err := authentication.NewSessionManager(tx).Create(login.Username, login.Role)
	if err != nil {
		return models.Login{}, uuid.UUID{}, fmt.Errorf("create activation session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return models.Login{}, uuid.UUID{}, fmt.Errorf("commit activation transaction: %w", err)
	}
	return login, sessionToken, nil
}

func isTransitionNominee(transition models.OfficerTransition, username string) bool {
	return username == transition.PresidentUsername ||
		username == transition.VicePresidentUsername ||
		username == transition.SecretaryUsername ||
		username == transition.TreasurerUsername
}

func completeOfficerTransition(tx store.Store, transition models.OfficerTransition, password string) (models.Login, error) {
	president, err := tx.Logins().ActivateWithRole(transition.PresidentUsername, password, "president")
	if err != nil {
		return models.Login{}, fmt.Errorf("activate transition president: %w", err)
	}
	nominees := []string{transition.PresidentUsername, transition.VicePresidentUsername, transition.SecretaryUsername, transition.TreasurerUsername}
	for _, nominee := range []struct{ username, role string }{
		{transition.VicePresidentUsername, "vice_president"},
		{transition.SecretaryUsername, "secretary"},
		{transition.TreasurerUsername, "treasurer"},
	} {
		if err := tx.Logins().Update(nominee.username, map[string]any{"role": nominee.role}); err != nil {
			return models.Login{}, fmt.Errorf("set transition nominee role: %w", err)
		}
	}
	disabled, err := tx.Logins().DisableExecutiveExcept(nominees)
	if err != nil {
		return models.Login{}, fmt.Errorf("disable outgoing executives: %w", err)
	}
	if err := tx.Sessions().DeleteByUsernames(append(nominees, disabled...)); err != nil {
		return models.Login{}, fmt.Errorf("delete transition sessions: %w", err)
	}
	return president, nil
}

func tokenHash(token string) []byte { sum := sha256.Sum256([]byte(token)); return sum[:] }
