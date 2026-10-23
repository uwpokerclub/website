package services

import (
	"api/internal/models"
	"api/internal/store"
	"api/internal/utils"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

var ErrDuplicateExecutiveTeamLoginIdentity = errors.New("multiple logins resolve to the same executive-team member")

func normalizedQuestID(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

// GetExecutiveTeam returns executive memberships for one semester, enriched
// with matching login details when an account exists.
func GetExecutiveTeam(st store.Store, semesterID uuid.UUID) (models.ExecutiveTeamResponse, error) {
	response := models.ExecutiveTeamResponse{Members: make([]models.ExecutiveTeamMember, 0)}
	memberships, _, err := st.Memberships().List(&models.ListMembershipsFilter{
		SemesterID: &semesterID,
		Executive:  utils.BoolToPtr(true),
		Pagination: models.Pagination{},
	})
	if err != nil {
		return response, err
	}
	if len(memberships) == 0 {
		return response, nil
	}

	memberIDs := make([]uint64, 0, len(memberships))
	for _, membership := range memberships {
		memberIDs = append(memberIDs, membership.UserID)
	}
	users, err := st.Members().FindByIDs(memberIDs)
	if err != nil {
		return response, err
	}
	usersByID := make(map[uint64]models.User, len(users))
	for _, user := range users {
		usersByID[user.ID] = user
	}

	rowByMemberID := make(map[uint64]int, len(memberships))
	rowByQuestID := make(map[string]int, len(memberships))
	questIDs := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		user, exists := usersByID[membership.UserID]
		if !exists {
			return response, fmt.Errorf("executive membership references missing member %d", membership.UserID)
		}
		if _, exists := rowByMemberID[user.ID]; exists {
			return response, fmt.Errorf("duplicate executive membership for member %d", user.ID)
		}
		row := models.ExecutiveTeamMember{
			MemberID:               utils.Uint64ToPtr(user.ID),
			FirstName:              utils.StringToPtr(user.FirstName),
			LastName:               utils.StringToPtr(user.LastName),
			QuestID:                utils.StringToPtr(user.QuestID),
			HasExecutiveMembership: true,
		}
		rowByMemberID[user.ID] = len(response.Members)
		if identity := normalizedQuestID(user.QuestID); identity != "" {
			if _, exists := rowByQuestID[identity]; exists {
				return response, fmt.Errorf("%w: duplicate normalized member Quest ID %q", ErrDuplicateExecutiveTeamLoginIdentity, identity)
			}
			rowByQuestID[identity] = len(response.Members)
			questIDs = append(questIDs, user.QuestID)
		}
		response.Members = append(response.Members, row)
	}

	logins, err := st.Logins().FindByNormalizedUsernames(questIDs)
	if err != nil {
		return response, err
	}
	for _, login := range logins {
		index, matchesMember := rowByQuestID[normalizedQuestID(login.Username)]
		if !matchesMember {
			continue
		}
		row := &response.Members[index]
		if row.HasLogin {
			return response, fmt.Errorf("%w: normalized username %q", ErrDuplicateExecutiveTeamLoginIdentity, normalizedQuestID(login.Username))
		}
		row.Username = utils.StringToPtr(login.Username)
		row.Role = utils.StringToPtr(login.Role)
		row.AccountStatus = utils.StringToPtr(login.AccountStatus)
		row.ActivationOutstanding = login.AccountStatus == models.LoginStatusPendingActivation
		row.HasLogin = true
	}

	sort.Slice(response.Members, func(i, j int) bool {
		a, b := response.Members[i], response.Members[j]
		if utils.StringPtrValue(a.LastName) != utils.StringPtrValue(b.LastName) {
			return utils.StringPtrValue(a.LastName) < utils.StringPtrValue(b.LastName)
		}
		if utils.StringPtrValue(a.FirstName) != utils.StringPtrValue(b.FirstName) {
			return utils.StringPtrValue(a.FirstName) < utils.StringPtrValue(b.FirstName)
		}
		if utils.StringPtrValue(a.QuestID) != utils.StringPtrValue(b.QuestID) {
			return utils.StringPtrValue(a.QuestID) < utils.StringPtrValue(b.QuestID)
		}
		if utils.Uint64PtrValue(a.MemberID) != utils.Uint64PtrValue(b.MemberID) {
			return utils.Uint64PtrValue(a.MemberID) < utils.Uint64PtrValue(b.MemberID)
		}
		return utils.StringPtrValue(a.Username) < utils.StringPtrValue(b.Username)
	})
	return response, nil
}
