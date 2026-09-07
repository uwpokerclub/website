package services

import (
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
	ErrTransitionInvalid   = errors.New("invalid transition nominee")
	ErrTransitionForbidden = errors.New("protected account cannot be nominated")
	ErrTransitionPending   = errors.New("an officer transition is already pending")
)

type officerTransitionService struct{ store store.Store }

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
	for i, role := range []string{"president", "vice_president", "secretary", "treasurer"} {
		transition.Nominees[role] = models.OfficerTransitionNominee{FirstName: users[i].FirstName, LastName: users[i].LastName}
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
	roles := []string{"president", "vice_president", "secretary", "treasurer"}
	for i, u := range users {
		login, err := tx.Logins().FindByUsernameForUpdate(u.QuestID)
		mint := i == 0
		if errors.Is(err, store.ErrNotFound) {
			hash, e := randomHash()
			if e != nil {
				return models.OfficerTransition{}, nil, e
			}
			login = models.Login{Username: u.QuestID, Password: hash, Role: roles[i], Status: models.LoginStatusPendingActivation}
			if e = tx.Logins().Create(&login); e != nil {
				return models.OfficerTransition{}, nil, e
			}
			mint = true
		} else if err != nil {
			return models.OfficerTransition{}, nil, err
		} else {
			if login.Role == "webmaster" || login.Role == "bot" {
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
			tokens[roles[i]] = token
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
