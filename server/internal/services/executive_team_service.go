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

type executiveTeamPerson struct {
	member     *models.User
	login      *models.ExecutiveTeamLogin
	membership bool
}

func normalizedQuestID(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

// GetExecutiveTeam unions current eligible login accounts with executive
// memberships for the selected semester.
func GetExecutiveTeam(st store.Store, semesterID uuid.UUID) (models.ExecutiveTeamResponse, error) {
	response := models.ExecutiveTeamResponse{Members: make([]models.ExecutiveTeamMember, 0)}
	loginCandidates, err := st.Logins().ListExecutiveTeamCandidates()
	if err != nil {
		return response, err
	}
	membershipFilter := &models.ListMembershipsFilter{
		SemesterID: &semesterID,
		Executive:  utils.BoolToPtr(true),
		Pagination: models.Pagination{},
	}
	memberships, _, err := st.Memberships().List(membershipFilter)
	if err != nil {
		return response, err
	}

	questIDs := make([]string, 0, len(loginCandidates))
	seenLoginIdentity := make(map[string]struct{}, len(loginCandidates))
	for _, login := range loginCandidates {
		identity := normalizedQuestID(login.Username)
		if identity == "" {
			continue
		}
		if _, exists := seenLoginIdentity[identity]; exists {
			return response, fmt.Errorf("%w: normalized username %q", ErrDuplicateExecutiveTeamLoginIdentity, identity)
		}
		seenLoginIdentity[identity] = struct{}{}
		questIDs = append(questIDs, login.Username)
	}
	usersByQuestID := make(map[string]models.User)
	users, err := st.Members().FindByQuestIDs(questIDs)
	if err != nil {
		return response, err
	}
	for _, user := range users {
		if key := normalizedQuestID(user.QuestID); key != "" {
			if _, exists := usersByQuestID[key]; exists {
				return response, fmt.Errorf("%w: duplicate normalized Quest ID %q", ErrDuplicateExecutiveTeamLoginIdentity, key)
			}
			usersByQuestID[key] = user
		}
	}

	memberIDs := make([]uint64, 0, len(memberships))
	for _, membership := range memberships {
		memberIDs = append(memberIDs, membership.UserID)
	}
	usersByID := make(map[uint64]models.User)
	users, err = st.Members().FindByIDs(memberIDs)
	if err != nil {
		return response, err
	}
	for _, user := range users {
		usersByID[user.ID] = user
	}

	people := make(map[string]*executiveTeamPerson, len(loginCandidates)+len(memberships))
	for i := range loginCandidates {
		login := loginCandidates[i]
		key := "login:" + login.Username
		var member *models.User
		if normalized := normalizedQuestID(login.Username); normalized != "" {
			if found, ok := usersByQuestID[normalized]; ok {
				copy := found
				member = &copy
				key = fmt.Sprintf("member:%d", found.ID)
			}
		}
		people[key] = &executiveTeamPerson{member: member, login: &login}
	}
	for _, membership := range memberships {
		member, ok := usersByID[membership.UserID]
		if !ok {
			return response, fmt.Errorf("executive membership references missing member %d", membership.UserID)
		}
		key := fmt.Sprintf("member:%d", member.ID)
		person, exists := people[key]
		if !exists {
			copy := member
			person = &executiveTeamPerson{member: &copy}
			people[key] = person
		}
		person.membership = true
	}

	for _, person := range people {
		row := models.ExecutiveTeamMember{HasLogin: person.login != nil, HasExecutiveMembership: person.membership}
		if person.member != nil {
			row.MemberID = utils.Uint64ToPtr(person.member.ID)
			row.FirstName = utils.StringToPtr(person.member.FirstName)
			row.LastName = utils.StringToPtr(person.member.LastName)
			row.QuestID = utils.StringToPtr(person.member.QuestID)
		} else if person.login != nil {
			row.QuestID = utils.StringToPtr(person.login.Username)
		}
		if person.login != nil {
			row.Username = utils.StringToPtr(person.login.Username)
			row.Role = utils.StringToPtr(person.login.Role)
			row.AccountStatus = utils.StringToPtr(person.login.AccountStatus)
			row.ActivationOutstanding = person.login.AccountStatus == models.LoginStatusPendingActivation
		}
		response.Members = append(response.Members, row)
	}
	sort.Slice(response.Members, func(i, j int) bool {
		a, b := response.Members[i], response.Members[j]
		if stringValue(a.LastName) != stringValue(b.LastName) {
			return stringValue(a.LastName) < stringValue(b.LastName)
		}
		if stringValue(a.FirstName) != stringValue(b.FirstName) {
			return stringValue(a.FirstName) < stringValue(b.FirstName)
		}
		if stringValue(a.QuestID) != stringValue(b.QuestID) {
			return stringValue(a.QuestID) < stringValue(b.QuestID)
		}
		if uint64Value(a.MemberID) != uint64Value(b.MemberID) {
			return uint64Value(a.MemberID) < uint64Value(b.MemberID)
		}
		return stringValue(a.Username) < stringValue(b.Username)
	})
	return response, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func uint64Value(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}
