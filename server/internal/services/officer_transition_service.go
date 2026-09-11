package services

import (
	"api/internal/authorization"
	"api/internal/models"
	"api/internal/store"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrTransitionInvalid    = errors.New("invalid transition nominee")
	ErrTransitionForbidden  = errors.New("protected account cannot be nominated")
	ErrTransitionPending    = errors.New("an officer transition is already pending")
	ErrTransitionNotFound   = errors.New("officer transition not found")
	ErrTransitionResolved   = errors.New("officer transition is already resolved")
	ErrTransitionIneligible = errors.New("transition nominee has no pending activation")
)

type officerTransitionService struct{ store store.Store }

var officerTransitionRoles = []authorization.Role{
	authorization.ROLE_PRESIDENT,
	authorization.ROLE_VICE_PRESIDENT,
	authorization.ROLE_SECRETARY,
	authorization.ROLE_TREASURER,
}

func NewOfficerTransitionService(s store.Store) *officerTransitionService {
	return &officerTransitionService{s}
}
func (s *officerTransitionService) Create(initiatedBy string, r models.CreateOfficerTransitionRequest) (models.OfficerTransition, map[string]string, error) {
	ids := []string{r.PresidentQuestID, r.VicePresidentQuestID, r.SecretaryQuestID, r.TreasurerQuestID}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			return models.OfficerTransition{}, nil, fmt.Errorf("%w: quest IDs must be distinct", ErrTransitionInvalid)
		}
		seen[id] = true
	}
	users := make([]models.User, 4)
	for i, id := range ids {
		found, err := s.store.Members().FindByQuestID(id)
		if err != nil {
			return models.OfficerTransition{}, nil, err
		}
		if len(found) != 1 {
			return models.OfficerTransition{}, nil, fmt.Errorf("%w: %s Quest ID must resolve to exactly one member", ErrTransitionInvalid, []string{"president", "vice president", "secretary", "treasurer"}[i])
		}
		users[i] = found[0]
	}
	tx, err := s.store.BeginTx()
	if err != nil {
		return models.OfficerTransition{}, nil, err
	}
	defer tx.Rollback()
	if _, currentErr := tx.OfficerTransitions().Current(); currentErr == nil {
		return models.OfficerTransition{}, nil, ErrTransitionPending
	} else if !errors.Is(currentErr, store.ErrNotFound) {
		return models.OfficerTransition{}, nil, currentErr
	}
	transition := models.OfficerTransition{ID: uuid.New(), InitiatedBy: initiatedBy, PresidentUsername: users[0].QuestID, VicePresidentUsername: users[1].QuestID, SecretaryUsername: users[2].QuestID, TreasurerUsername: users[3].QuestID, Status: models.OfficerTransitionPending}
	transition.Nominees = map[string]models.OfficerTransitionNominee{}
	for i, role := range officerTransitionRoles {
		transition.Nominees[role.ToString()] = models.OfficerTransitionNominee{FirstName: users[i].FirstName, LastName: users[i].LastName}
	}
	// Tokens are foreign-keyed to the transition, so persist the parent first.
	// This remains atomic: any later nominee/token failure rolls it back.
	if err := tx.OfficerTransitions().Create(&transition); err != nil {
		if errors.Is(err, store.ErrConflict) {
			return models.OfficerTransition{}, nil, ErrTransitionPending
		}
		return models.OfficerTransition{}, nil, err
	}
	tokens := map[string]string{}
	for i, u := range users {
		role := officerTransitionRoles[i]
		login, err := tx.Logins().FindByUsernameForUpdate(u.QuestID)
		mint := role == authorization.ROLE_PRESIDENT
		if errors.Is(err, store.ErrNotFound) {
			hash, e := randomHash()
			if e != nil {
				return models.OfficerTransition{}, nil, e
			}
			login = models.Login{Username: u.QuestID, Password: hash, Role: role.ToString(), Status: models.LoginStatusPendingActivation, StagedTransitionID: &transition.ID}
			if e = tx.Logins().Create(&login); e != nil {
				return models.OfficerTransition{}, nil, e
			}
			mint = true
		} else if err != nil {
			return models.OfficerTransition{}, nil, err
		} else {
			if authorization.ToRole(login.Role) == authorization.ROLE_WEBMASTER || authorization.ToRole(login.Role) == authorization.ROLE_BOT {
				return models.OfficerTransition{}, nil, ErrTransitionForbidden
			}
			if login.Status != models.LoginStatusActive {
				hash, e := randomHash()
				if e != nil {
					return models.OfficerTransition{}, nil, e
				}
				if e = tx.Logins().Update(login.Username, map[string]any{"status": models.LoginStatusPendingActivation, "password": hash}); e != nil {
					return models.OfficerTransition{}, nil, e
				}
				mint = true
			}
		}
		if mint {
			token, e := createTransitionToken(tx, u.QuestID, transition.ID)
			if e != nil {
				return models.OfficerTransition{}, nil, e
			}
			tokens[role.ToString()] = token
		}
	}
	if err := tx.Commit(); err != nil {
		return models.OfficerTransition{}, nil, err
	}
	// The in-memory repository deliberately copies records, whereas GORM fills
	// defaults on the passed struct. Preserve the explicit pending state in the
	// response consistently across both implementations.
	transition.Status = models.OfficerTransitionPending
	return transition, tokens, nil
}

// Cancel marks a pending transition cancelled and removes its activation links
// plus staging-owned pending logins that no other transition names.
func (s *officerTransitionService) Cancel(id uuid.UUID) (models.OfficerTransition, error) {
	for attempt := 0; attempt < 2; attempt++ {
		transition, err := s.cancelOnce(id)
		if !errors.Is(err, store.ErrTransactionConflict) {
			return transition, err
		}
	}
	return models.OfficerTransition{}, store.ErrTransactionConflict
}

func (s *officerTransitionService) cancelOnce(id uuid.UUID) (models.OfficerTransition, error) {
	tx, err := s.store.BeginTx()
	if err != nil {
		return models.OfficerTransition{}, err
	}
	defer tx.Rollback()
	transition, err := tx.OfficerTransitions().FindByIDForUpdate(id)
	if errors.Is(err, store.ErrNotFound) {
		return models.OfficerTransition{}, ErrTransitionNotFound
	}
	if err != nil {
		return models.OfficerTransition{}, err
	}
	if transition.Status != models.OfficerTransitionPending {
		return models.OfficerTransition{}, ErrTransitionResolved
	}
	transition, err = tx.OfficerTransitions().Cancel(id)
	if err != nil {
		return models.OfficerTransition{}, err
	}
	// The transition row is locked before its activation rows and login cleanup,
	// matching activation completion's lock order.
	if err := tx.AccountActivations().DeleteByTransition(id); err != nil {
		return models.OfficerTransition{}, err
	}
	for _, username := range transitionUsernames(transition) {
		login, err := tx.Logins().FindByUsernameForUpdate(username)
		if errors.Is(err, store.ErrNotFound) {
			continue
		}
		if err != nil {
			return models.OfficerTransition{}, err
		}
		if login.Status != models.LoginStatusPendingActivation || login.StagedTransitionID == nil || *login.StagedTransitionID != id {
			continue
		}
		referenced, err := tx.OfficerTransitions().ReferencesUsernameElsewhere(id, username)
		if err != nil {
			return models.OfficerTransition{}, err
		}
		if !referenced {
			if err := tx.Logins().Delete(username); err != nil {
				return models.OfficerTransition{}, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return models.OfficerTransition{}, err
	}
	return transition, nil
}

// Reissue replaces a nominee's unused activation token while the transition is pending.
func (s *officerTransitionService) Reissue(id uuid.UUID, role authorization.Role) (string, error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, err := s.reissueOnce(id, role)
		if !errors.Is(err, store.ErrTransactionConflict) {
			return token, err
		}
	}
	return "", store.ErrTransactionConflict
}

func (s *officerTransitionService) reissueOnce(id uuid.UUID, role authorization.Role) (string, error) {
	tx, err := s.store.BeginTx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	transition, err := tx.OfficerTransitions().FindByIDForUpdate(id)
	if errors.Is(err, store.ErrNotFound) {
		return "", ErrTransitionNotFound
	}
	if err != nil {
		return "", err
	}
	if transition.Status != models.OfficerTransitionPending {
		return "", ErrTransitionResolved
	}
	username, ok := transitionUsernameForRole(transition, role)
	if !ok {
		return "", ErrTransitionInvalid
	}
	login, err := tx.Logins().FindByUsernameForUpdate(username)
	if errors.Is(err, store.ErrNotFound) {
		return "", ErrTransitionIneligible
	}
	if err != nil {
		return "", err
	}
	if login.Status == models.LoginStatusDisabled || (role != authorization.ROLE_PRESIDENT && login.Status != models.LoginStatusPendingActivation) {
		return "", ErrTransitionIneligible
	}
	token, err := createTransitionToken(tx, username, id)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return token, nil
}

func transitionUsernameForRole(t models.OfficerTransition, role authorization.Role) (string, bool) {
	switch role {
	case authorization.ROLE_PRESIDENT:
		return t.PresidentUsername, true
	case authorization.ROLE_VICE_PRESIDENT:
		return t.VicePresidentUsername, true
	case authorization.ROLE_SECRETARY:
		return t.SecretaryUsername, true
	case authorization.ROLE_TREASURER:
		return t.TreasurerUsername, true
	default:
		return "", false
	}
}

func transitionUsernames(t models.OfficerTransition) []string {
	return []string{t.PresidentUsername, t.VicePresidentUsername, t.SecretaryUsername, t.TreasurerUsername}
}
func randomHash() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	h, e := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	return string(h), e
}
func createTransitionToken(tx store.Store, username string, id uuid.UUID) (string, error) {
	raw := make([]byte, 32)
	if _, e := rand.Read(raw); e != nil {
		return "", e
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	h := sha256.Sum256([]byte(token))
	if e := tx.AccountActivations().DeleteUnusedByUsername(username); e != nil {
		return "", e
	}
	return token, tx.AccountActivations().Create(&models.AccountActivation{TokenHash: h[:], Username: username, TransitionID: &id, ExpiresAt: time.Now().UTC().Add(14 * 24 * time.Hour)})
}
