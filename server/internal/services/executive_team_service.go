package services

import (
	"api/internal/authorization"
	"api/internal/models"
	"api/internal/store"
	"api/internal/utils"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrExecutiveTeamInvalidRole      = errors.New("role is not grantable for executive team additions")
	ErrExecutiveTeamInvalidIdentity  = errors.New("quest ID is required")
	ErrExecutiveTeamMemberNotFound   = errors.New("member not found")
	ErrExecutiveTeamMemberAmbiguous  = errors.New("quest ID resolves to multiple members")
	ErrExecutiveTeamIdentityConflict = errors.New("member and login identities do not match")
	ErrExecutiveTeamProtectedLogin   = errors.New("login role cannot be granted through executive team additions")
	ErrExecutiveTeamStagedLogin      = errors.New("staged officer-transition login cannot be changed")
	ErrExecutiveTeamAlreadyMember    = errors.New("member is already on the executive team")
	ErrExecutiveTeamSemesterNotFound = errors.New("semester not found")
)

type executiveTeamService struct{ store store.Store }

func NewExecutiveTeamService(st store.Store) *executiveTeamService {
	return &executiveTeamService{store: st}
}

// AddMember creates or updates the account and ensures its semester executive
// membership in one transaction. A raw activation token is returned only after
// that transaction commits successfully.
func (svc *executiveTeamService) AddMember(semesterID uuid.UUID, questID, role string) (models.AddExecutiveTeamMemberResponse, error) {
	if role != authorization.ROLE_EXECUTIVE.ToString() && role != authorization.ROLE_TOURNAMENT_DIRECTOR.ToString() {
		return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamInvalidRole
	}
	questID = strings.TrimSpace(questID)
	if questID == "" {
		return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamInvalidIdentity
	}

	tx, err := svc.store.BeginTx()
	if err != nil {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("begin executive team add: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Semesters().FindByID(semesterID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamSemesterNotFound
		}
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("find semester: %w", err)
	}

	members, err := tx.Members().FindByQuestID(questID)
	if err != nil {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("find member: %w", err)
	}
	if len(members) == 0 {
		return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamMemberNotFound
	}
	if len(members) != 1 {
		return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamMemberAmbiguous
	}
	member := members[0]
	normalizedQuestID := normalizeExecutiveIdentity(member.QuestID)
	if normalizedQuestID == "" || normalizedQuestID != normalizeExecutiveIdentity(questID) {
		return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamIdentityConflict
	}
	login, loginErr := tx.Logins().FindByUsername(member.QuestID)
	hasLogin := loginErr == nil
	if loginErr != nil && !errors.Is(loginErr, store.ErrNotFound) {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("find login: %w", loginErr)
	}
	if hasLogin {
		if login.StagedTransitionID != nil {
			return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamStagedLogin
		}
		if login.Role != authorization.ROLE_EXECUTIVE.ToString() && login.Role != authorization.ROLE_TOURNAMENT_DIRECTOR.ToString() {
			return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamProtectedLogin
		}
	}

	existingMembership, err := tx.Memberships().FindByUserAndSemester(member.ID, semesterID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("find membership: %w", err)
	}
	if err == nil && existingMembership.Executive {
		return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamAlreadyMember
	}

	username := member.QuestID
	status := models.LoginStatusPendingActivation
	var token *string
	if !hasLogin {
		password, err := unusablePasswordHash()
		if err != nil {
			return models.AddExecutiveTeamMemberResponse{}, err
		}
		login := models.Login{Username: username, Password: password, Role: role, Status: models.LoginStatusPendingActivation}
		if err := tx.Logins().Create(&login); err != nil {
			if errors.Is(err, store.ErrAlreadyExists) || errors.Is(err, store.ErrConflict) {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("%w: login was created concurrently", store.ErrConflict)
			}
			return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("create login: %w", err)
		}
		activationToken, err := NewAccountActivationService(tx).Create(login.Username)
		if err != nil {
			return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("create activation: %w", err)
		}
		token = utils.StringToPtr(activationToken)
	} else {
		username = login.Username
		status = login.Status
		switch login.Status {
		case models.LoginStatusActive:
			updates := map[string]any{"role": role}
			if login.Role != role {
				if err := tx.Sessions().DeleteByUsername(username); err != nil {
					return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("delete changed-role sessions: %w", err)
				}
			}
			if err := tx.Logins().Update(username, updates); err != nil {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("update login role: %w", err)
			}
			if err := tx.AccountActivations().DeleteAllUnusedByUsername(username); err != nil {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("revoke stale activations: %w", err)
			}
		case models.LoginStatusDisabled:
			if err := tx.Sessions().DeleteByUsername(username); err != nil {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("delete disabled-login sessions: %w", err)
			}
			if err := tx.Logins().Update(username, map[string]any{"role": role, "status": models.LoginStatusActive}); err != nil {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("reactivate login: %w", err)
			}
			if err := tx.AccountActivations().DeleteAllUnusedByUsername(username); err != nil {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("revoke stale activations: %w", err)
			}
			status = models.LoginStatusActive
		case models.LoginStatusPendingActivation:
			if login.Role != role {
				if err := tx.Sessions().DeleteByUsername(username); err != nil {
					return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("delete changed-role sessions: %w", err)
				}
			}
			if err := tx.Logins().Update(username, map[string]any{"role": role}); err != nil {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("update pending login role: %w", err)
			}
			activationToken, err := NewAccountActivationService(tx).Create(username)
			if err != nil {
				return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("rotate activation: %w", err)
			}
			token = utils.StringToPtr(activationToken)
		default:
			return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("unsupported login status %q", login.Status)
		}
	}

	// Recheck the exact login identity after account changes and before creating
	// the executive membership.
	confirmed, err := tx.Logins().FindByUsername(member.QuestID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("%w: login changed during add", store.ErrConflict)
		}
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("recheck login identity: %w", err)
	}
	if confirmed.Username != username {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("%w: login identity changed during add", store.ErrConflict)
	}
	if confirmed.StagedTransitionID != nil {
		return models.AddExecutiveTeamMemberResponse{}, ErrExecutiveTeamStagedLogin
	}
	if confirmed.Role != role {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("%w: login role changed during add", store.ErrConflict)
	}
	confirmedMembers, err := tx.Members().FindByQuestID(questID)
	if err != nil {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("recheck member identity: %w", err)
	}
	if len(confirmedMembers) != 1 || confirmedMembers[0].ID != member.ID || normalizeExecutiveIdentity(confirmedMembers[0].QuestID) != normalizedQuestID {
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("%w: member identity changed during add", store.ErrConflict)
	}

	if _, err := NewMembershipService(tx).EnsureExecutiveMembership(member.ID, semesterID, models.MembershipSourceAdmin); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) || errors.Is(err, store.ErrConflict) {
			return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("%w: membership was created concurrently", store.ErrConflict)
		}
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("ensure executive membership: %w", err)
	}

	if err := tx.Commit(); err != nil {
		if errors.Is(err, store.ErrTransactionConflict) {
			return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("%w: concurrent executive team add", store.ErrConflict)
		}
		return models.AddExecutiveTeamMemberResponse{}, fmt.Errorf("commit executive team add: %w", err)
	}
	return models.AddExecutiveTeamMemberResponse{MemberID: member.ID, Username: username, Role: role, AccountStatus: status, ActivationToken: token}, nil
}

func normalizeExecutiveIdentity(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func unusablePasswordHash() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate unusable login password: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(base64.RawURLEncoding.EncodeToString(raw)), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash unusable login password: %w", err)
	}
	return string(hash), nil
}
