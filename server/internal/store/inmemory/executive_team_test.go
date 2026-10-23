package inmemory_test

import (
	"api/internal/models"
	"api/internal/store"
	"api/internal/store/inmemory"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListExecutiveTeamUsesMembershipsAndEnrichesAccounts(t *testing.T) {
	st := inmemory.NewStore()
	semesterID := uuid.New()
	people := []models.User{
		{ID: 1, FirstName: "Alex", LastName: "Able", QuestID: "Quest-A"},
		{ID: 2, FirstName: "Blair", LastName: "Baker", QuestID: "quest-b"},
		{ID: 3, FirstName: "Casey", LastName: "Clark", QuestID: "quest-c"},
		{ID: 4, FirstName: "Drew", LastName: "Dale", QuestID: "quest-d"},
		{ID: 5, FirstName: "Evan", LastName: "Empty", QuestID: "quest-e"},
	}
	for i := range people {
		require.NoError(t, st.Members().Create(&people[i]))
		require.NoError(t, st.Memberships().Create(&models.Membership{UserID: people[i].ID, SemesterID: semesterID, Executive: true}))
	}
	for _, login := range []models.Login{
		{Username: " quest-a ", Password: "not-returned", Role: "president", Status: models.LoginStatusActive},
		{Username: "QUEST-B", Password: "not-returned", Role: "executive", Status: models.LoginStatusDisabled},
		{Username: "quest-c", Password: "not-returned", Role: "webmaster", Status: models.LoginStatusActive},
		{Username: "quest-d", Password: "not-returned", Role: "bot", Status: models.LoginStatusPendingActivation},
		{Username: "login-only", Password: "not-returned", Role: "secretary", Status: models.LoginStatusActive},
		{Username: "unmatched", Password: "not-returned", Role: "executive", Status: models.LoginStatusPendingActivation},
	} {
		copy := login
		require.NoError(t, st.Logins().Create(&copy))
	}

	response, err := st.Memberships().ListExecutiveTeam(semesterID)
	require.NoError(t, err)
	require.Len(t, response, 5)
	byQuestID := map[string]models.ExecutiveTeamMember{}
	for _, member := range response {
		byQuestID[member.QuestID] = member
		require.True(t, member.HasExecutiveMembership)
	}

	active := byQuestID["Quest-A"]
	require.Equal(t, uint64(1), active.MemberID)
	require.Equal(t, " quest-a ", *active.Username)
	require.Equal(t, "president", *active.Role)
	require.Equal(t, models.LoginStatusActive, *active.AccountStatus)
	require.True(t, active.HasLogin)
	require.False(t, active.ActivationOutstanding)

	disabled := byQuestID["quest-b"]
	require.True(t, disabled.HasLogin)
	require.Equal(t, "executive", *disabled.Role)
	require.Equal(t, models.LoginStatusDisabled, *disabled.AccountStatus)
	require.False(t, disabled.ActivationOutstanding)

	for questID, role := range map[string]string{"quest-c": "webmaster", "quest-d": "bot"} {
		member := byQuestID[questID]
		require.True(t, member.HasLogin, questID)
		require.Equal(t, role, *member.Role, questID)
	}
	pending := byQuestID["quest-d"]
	require.Equal(t, models.LoginStatusPendingActivation, *pending.AccountStatus)
	require.True(t, pending.ActivationOutstanding)

	withoutAccount := byQuestID["quest-e"]
	require.False(t, withoutAccount.HasLogin)
	require.Nil(t, withoutAccount.Username)
	require.Nil(t, withoutAccount.Role)
	require.Nil(t, withoutAccount.AccountStatus)
	require.False(t, withoutAccount.ActivationOutstanding)
	require.NotContains(t, byQuestID, "login-only")
	require.NotContains(t, byQuestID, "unmatched")
}

func TestListExecutiveTeamScopesMembershipsToRequestedSemester(t *testing.T) {
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

	first, err := st.Memberships().ListExecutiveTeam(firstSemesterID)
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, users[0].ID, first[0].MemberID)
	require.True(t, first[0].HasExecutiveMembership)

	second, err := st.Memberships().ListExecutiveTeam(secondSemesterID)
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.Equal(t, users[1].ID, second[0].MemberID)
	require.True(t, second[0].HasExecutiveMembership)
}

func TestListExecutiveTeamRejectsMultipleNormalizedLogins(t *testing.T) {
	st := inmemory.NewStore()
	semesterID := uuid.New()
	user := models.User{ID: 20, FirstName: "Quest", LastName: "Member", QuestID: "quest-id"}
	require.NoError(t, st.Members().Create(&user))
	require.NoError(t, st.Memberships().Create(&models.Membership{UserID: user.ID, SemesterID: semesterID, Executive: true}))
	for _, login := range []models.Login{
		{Username: "Quest-ID", Password: "secret", Role: "executive", Status: models.LoginStatusActive},
		{Username: " quest-id ", Password: "secret", Role: "webmaster", Status: models.LoginStatusDisabled},
	} {
		copy := login
		require.NoError(t, st.Logins().Create(&copy))
	}
	_, err := st.Memberships().ListExecutiveTeam(semesterID)
	require.ErrorIs(t, err, store.ErrDuplicateExecutiveTeamLoginIdentity)
}

func TestListExecutiveTeamDoesNotReturnLoginOnlyRows(t *testing.T) {
	st := inmemory.NewStore()
	login := models.Login{Username: "login-only", Password: "secret", Role: "executive", Status: models.LoginStatusActive}
	require.NoError(t, st.Logins().Create(&login))

	response, err := st.Memberships().ListExecutiveTeam(uuid.New())
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Empty(t, response)
}
