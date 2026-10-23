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

var ErrDuplicateExecutiveTeamLoginIdentity = errors.New("multiple eligible logins resolve to the same executive-team identity")

func normalizedQuestID(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

// GetExecutiveTeam unions current eligible login accounts with executive
// memberships for the selected semester.
func GetExecutiveTeam(st store.Store, semesterID uuid.UUID) (models.ExecutiveTeamResponse, error) {
	response := models.ExecutiveTeamResponse{Members: make([]models.ExecutiveTeamMember, 0)}
	logins, err := st.Logins().ListExecutiveTeamCandidates()
	if err != nil {
		return response, err
	}
	memberships, _, err := st.Memberships().List(&models.ListMembershipsFilter{
		SemesterID: &semesterID,
		Executive:  utils.BoolToPtr(true),
		Pagination: models.Pagination{},
	})
	if err != nil {
		return response, err
	}

	questIDs := make([]string, 0, len(logins))
	wantedQuestIDs := make(map[string]struct{}, len(logins))
	for _, login := range logins {
		identity := normalizedQuestID(login.Username)
		if identity == "" {
			continue
		}
		if _, exists := wantedQuestIDs[identity]; exists {
			return response, fmt.Errorf("%w: normalized username %q", ErrDuplicateExecutiveTeamLoginIdentity, identity)
		}
		wantedQuestIDs[identity] = struct{}{}
		questIDs = append(questIDs, login.Username)
	}

	memberIDs := make([]uint64, 0, len(memberships))
	for _, membership := range memberships {
		memberIDs = append(memberIDs, membership.UserID)
	}
	users, err := st.Members().FindByIDsOrQuestIDs(memberIDs, questIDs)
	if err != nil {
		return response, err
	}
	usersByID := make(map[uint64]models.User, len(users))
	usersByQuestID := make(map[string]models.User, len(questIDs))
	for _, user := range users {
		usersByID[user.ID] = user
		if identity := normalizedQuestID(user.QuestID); identity != "" {
			if _, wanted := wantedQuestIDs[identity]; wanted {
				if _, exists := usersByQuestID[identity]; exists {
					return response, fmt.Errorf("%w: duplicate normalized Quest ID %q", ErrDuplicateExecutiveTeamLoginIdentity, identity)
				}
				usersByQuestID[identity] = user
			}
		}
	}

	rowByMemberID := make(map[uint64]int, len(memberships))
	for _, login := range logins {
		row := models.ExecutiveTeamMember{
			Username:              utils.StringToPtr(login.Username),
			Role:                  utils.StringToPtr(login.Role),
			AccountStatus:         utils.StringToPtr(login.AccountStatus),
			ActivationOutstanding: login.AccountStatus == models.LoginStatusPendingActivation,
			HasLogin:              true,
		}
		if user, linked := usersByQuestID[normalizedQuestID(login.Username)]; linked {
			row.MemberID = utils.Uint64ToPtr(user.ID)
			row.FirstName = utils.StringToPtr(user.FirstName)
			row.LastName = utils.StringToPtr(user.LastName)
			row.QuestID = utils.StringToPtr(user.QuestID)
			rowByMemberID[user.ID] = len(response.Members)
		} else {
			row.QuestID = utils.StringToPtr(login.Username)
		}
		response.Members = append(response.Members, row)
	}

	for _, membership := range memberships {
		user, exists := usersByID[membership.UserID]
		if !exists {
			return response, fmt.Errorf("executive membership references missing member %d", membership.UserID)
		}
		if index, exists := rowByMemberID[user.ID]; exists {
			response.Members[index].HasExecutiveMembership = true
			continue
		}
		rowByMemberID[user.ID] = len(response.Members)
		response.Members = append(response.Members, models.ExecutiveTeamMember{
			MemberID:               utils.Uint64ToPtr(user.ID),
			FirstName:              utils.StringToPtr(user.FirstName),
			LastName:               utils.StringToPtr(user.LastName),
			QuestID:                utils.StringToPtr(user.QuestID),
			HasExecutiveMembership: true,
		})
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
