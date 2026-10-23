package services

import (
	"api/internal/models"
	"api/internal/store/inmemory"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetExecutiveTeamUnionAndAccountFiltering(t *testing.T) {
	st := inmemory.NewStore()
	semesterID := uuid.New()
	people := []models.User{
		{ID: 1, FirstName: "Alex", LastName: "Same", QuestID: "Quest-A"},
		{ID: 2, FirstName: "Blair", LastName: "Member", QuestID: "quest-b"},
		{ID: 3, FirstName: "Casey", LastName: "Member", QuestID: "quest-c"},
	}
	for i := range people {
		require.NoError(t, st.Members().Create(&people[i]))
	}
	for _, userID := range []uint64{1, 2, 3} {
		require.NoError(t, st.Memberships().Create(&models.Membership{UserID: userID, SemesterID: semesterID, Executive: true}))
	}
	for _, login := range []models.Login{
		{Username: " quest-a ", Password: "not-returned", Role: "president", Status: models.LoginStatusPendingActivation},
		{Username: "QUEST-B", Password: "not-returned", Role: "executive", Status: models.LoginStatusDisabled},
		{Username: "quest-c", Password: "not-returned", Role: "webmaster", Status: models.LoginStatusActive},
		{Username: "unmatched", Password: "not-returned", Role: "secretary", Status: models.LoginStatusActive},
		{Username: "bot-login", Password: "not-returned", Role: "bot", Status: models.LoginStatusActive},
	} {
		copy := login
		require.NoError(t, st.Logins().Create(&copy))
	}

	response, err := GetExecutiveTeam(st, semesterID)
	require.NoError(t, err)
	require.Len(t, response.Members, 4)
	byQuestID := map[string]models.ExecutiveTeamMember{}
	for _, member := range response.Members {
		byQuestID[*member.QuestID] = member
	}

	linked := byQuestID["Quest-A"]
	require.Equal(t, uint64(1), *linked.MemberID)
	require.Equal(t, " quest-a ", *linked.Username)
	require.Equal(t, "president", *linked.Role)
	require.Equal(t, models.LoginStatusPendingActivation, *linked.AccountStatus)
	require.True(t, linked.ActivationOutstanding)
	require.True(t, linked.HasLogin)
	require.True(t, linked.HasExecutiveMembership)

	for _, questID := range []string{"quest-b", "quest-c"} {
		member := byQuestID[questID]
		require.False(t, member.HasLogin, questID)
		require.True(t, member.HasExecutiveMembership, questID)
		require.False(t, member.ActivationOutstanding, questID)
		require.Nil(t, member.Username, questID)
		require.Nil(t, member.Role, questID)
		require.Nil(t, member.AccountStatus, questID)
	}
	unmatched := byQuestID["unmatched"]
	require.Nil(t, unmatched.MemberID)
	require.Nil(t, unmatched.FirstName)
	require.True(t, unmatched.HasLogin)
	require.False(t, unmatched.HasExecutiveMembership)

}

func TestGetExecutiveTeamScopesMembershipsToRequestedSemester(t *testing.T) {
	st := inmemory.NewStore()
	firstSemesterID, secondSemesterID := uuid.New(), uuid.New()
	users := []models.User{
		{ID: 10, FirstName: "First", LastName: "Member", QuestID: "first-member"},
		{ID: 11, FirstName: "Second", LastName: "Member", QuestID: "second-member"},
	}
	for i := range users {
		require.NoError(t, st.Members().Create(&users[i]))
	}
	require.NoError(t, st.Memberships().Create(&models.Membership{UserID: users[0].ID, SemesterID: firstSemesterID, Executive: true}))
	require.NoError(t, st.Memberships().Create(&models.Membership{UserID: users[1].ID, SemesterID: secondSemesterID, Executive: true}))

	first, err := GetExecutiveTeam(st, firstSemesterID)
	require.NoError(t, err)
	require.Len(t, first.Members, 1)
	require.Equal(t, users[0].ID, *first.Members[0].MemberID)
	require.True(t, first.Members[0].HasExecutiveMembership)

	second, err := GetExecutiveTeam(st, secondSemesterID)
	require.NoError(t, err)
	require.Len(t, second.Members, 1)
	require.Equal(t, users[1].ID, *second.Members[0].MemberID)
	require.True(t, second.Members[0].HasExecutiveMembership)
}

func TestGetExecutiveTeamRejectsNormalizedDuplicateEligibleLogins(t *testing.T) {
	st := inmemory.NewStore()
	for _, username := range []string{"Quest-ID", " quest-id "} {
		login := models.Login{Username: username, Password: "secret", Role: "executive", Status: models.LoginStatusActive}
		require.NoError(t, st.Logins().Create(&login))
	}
	_, err := GetExecutiveTeam(st, uuid.New())
	require.ErrorIs(t, err, ErrDuplicateExecutiveTeamLoginIdentity)
}
