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

	member, login, hasLogin, err := resolveExecutiveAddTarget(tx, semesterID, questID)
	if err != nil {
		return models.AddExecutiveTeamMemberResponse{}, err
	}
	username, status, token, err := prepareExecutiveLogin(tx, member, login, hasLogin, role)
	if err != nil {
		return models.AddExecutiveTeamMemberResponse{}, err
	}
	if err := recheckExecutiveAddIdentity(tx, member, questID, username, role); err != nil {
		return models.AddExecutiveTeamMemberResponse{}, err
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

// resolveExecutiveAddTarget validates the semester, member, login, and membership
// before any account or membership changes are made.
func resolveExecutiveAddTarget(tx store.Store, semesterID uuid.UUID, questID string) (models.User, models.Login, bool, error) {
	if _, err := tx.Semesters().FindByID(semesterID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return models.User{}, models.Login{}, false, ErrExecutiveTeamSemesterNotFound
		}
		return models.User{}, models.Login{}, false, fmt.Errorf("find semester: %w", err)
	}

	members, err := tx.Members().FindByQuestID(questID)
	if err != nil {
		return models.User{}, models.Login{}, false, fmt.Errorf("find member: %w", err)
	}
	if len(members) == 0 {
		return models.User{}, models.Login{}, false, ErrExecutiveTeamMemberNotFound
	}
	if len(members) != 1 {
		return models.User{}, models.Login{}, false, ErrExecutiveTeamMemberAmbiguous
	}
	member := members[0]
	if normalizedQuestID := normalizeExecutiveIdentity(member.QuestID); normalizedQuestID == "" || normalizedQuestID != normalizeExecutiveIdentity(questID) {
		return models.User{}, models.Login{}, false, ErrExecutiveTeamIdentityConflict
	}

	login, err := tx.Logins().FindByUsername(member.QuestID)
	hasLogin := err == nil
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return models.User{}, models.Login{}, false, fmt.Errorf("find login: %w", err)
	}
	if hasLogin {
		if login.StagedTransitionID != nil {
			return models.User{}, models.Login{}, false, ErrExecutiveTeamStagedLogin
		}
		if login.Role != authorization.ROLE_EXECUTIVE.ToString() && login.Role != authorization.ROLE_TOURNAMENT_DIRECTOR.ToString() {
			return models.User{}, models.Login{}, false, ErrExecutiveTeamProtectedLogin
		}
	}

	membership, err := tx.Memberships().FindByUserAndSemester(member.ID, semesterID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return models.User{}, models.Login{}, false, fmt.Errorf("find membership: %w", err)
	}
	if err == nil && membership.Executive {
		return models.User{}, models.Login{}, false, ErrExecutiveTeamAlreadyMember
	}

	return member, login, hasLogin, nil
}

// prepareExecutiveLogin creates a pending login or applies the requested role to
// an existing executive login, returning its status and any activation token.
func prepareExecutiveLogin(tx store.Store, member models.User, login models.Login, hasLogin bool, role string) (string, string, *string, error) {
	if !hasLogin {
		password, err := unusablePasswordHash()
		if err != nil {
			return "", "", nil, err
		}
		login = models.Login{Username: member.QuestID, Password: password, Role: role, Status: models.LoginStatusPendingActivation}
		if err := tx.Logins().Create(&login); err != nil {
			if errors.Is(err, store.ErrAlreadyExists) || errors.Is(err, store.ErrConflict) {
				return "", "", nil, fmt.Errorf("%w: login was created concurrently", store.ErrConflict)
			}
			return "", "", nil, fmt.Errorf("create login: %w", err)
		}
		token, err := NewAccountActivationService(tx).Create(login.Username)
		if err != nil {
			return "", "", nil, fmt.Errorf("create activation: %w", err)
		}
		return login.Username, login.Status, utils.StringToPtr(token), nil
	}

	status, token, err := updateExecutiveLogin(tx, login, role)
	if err != nil {
		return "", "", nil, err
	}
	return login.Username, status, token, nil
}

// updateExecutiveLogin applies role-specific session and activation changes for
// an existing executive login while preserving its password.
func updateExecutiveLogin(tx store.Store, login models.Login, role string) (string, *string, error) {
	username := login.Username
	updates := map[string]any{"role": role}
	status := login.Status
	switch login.Status {
	case models.LoginStatusActive:
		if login.Role != role {
			if err := tx.Sessions().DeleteByUsername(username); err != nil {
				return "", nil, fmt.Errorf("delete changed-role sessions: %w", err)
			}
		}
		if err := tx.AccountActivations().DeleteAllUnusedByUsername(username); err != nil {
			return "", nil, fmt.Errorf("revoke stale activations: %w", err)
		}
	case models.LoginStatusDisabled:
		if err := tx.Sessions().DeleteByUsername(username); err != nil {
			return "", nil, fmt.Errorf("delete disabled-login sessions: %w", err)
		}
		updates["status"] = models.LoginStatusActive
		status = models.LoginStatusActive
		if err := tx.AccountActivations().DeleteAllUnusedByUsername(username); err != nil {
			return "", nil, fmt.Errorf("revoke stale activations: %w", err)
		}
	case models.LoginStatusPendingActivation:
		if login.Role != role {
			if err := tx.Sessions().DeleteByUsername(username); err != nil {
				return "", nil, fmt.Errorf("delete changed-role sessions: %w", err)
			}
		}
	default:
		return "", nil, fmt.Errorf("unsupported login status %q", login.Status)
	}

	if err := tx.Logins().Update(username, updates); err != nil {
		return "", nil, fmt.Errorf("update login: %w", err)
	}
	if login.Status != models.LoginStatusPendingActivation {
		return status, nil, nil
	}
	token, err := NewAccountActivationService(tx).Create(username)
	if err != nil {
		return "", nil, fmt.Errorf("rotate activation: %w", err)
	}
	return status, utils.StringToPtr(token), nil
}

// recheckExecutiveAddIdentity confirms the login and member still identify the
// same target before the executive membership is ensured.
func recheckExecutiveAddIdentity(tx store.Store, member models.User, questID, username, role string) error {
	confirmed, err := tx.Logins().FindByUsername(member.QuestID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("%w: login changed during add", store.ErrConflict)
		}
		return fmt.Errorf("recheck login identity: %w", err)
	}
	if confirmed.Username != username {
		return fmt.Errorf("%w: login identity changed during add", store.ErrConflict)
	}
	if confirmed.StagedTransitionID != nil {
		return ErrExecutiveTeamStagedLogin
	}
	if confirmed.Role != role {
		return fmt.Errorf("%w: login role changed during add", store.ErrConflict)
	}

	confirmedMembers, err := tx.Members().FindByQuestID(questID)
	if err != nil {
		return fmt.Errorf("recheck member identity: %w", err)
	}
	if len(confirmedMembers) != 1 || confirmedMembers[0].ID != member.ID || normalizeExecutiveIdentity(confirmedMembers[0].QuestID) != normalizeExecutiveIdentity(member.QuestID) {
		return fmt.Errorf("%w: member identity changed during add", store.ErrConflict)
	}
	return nil
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
