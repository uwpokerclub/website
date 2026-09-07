package services

import (
	"api/internal/models"
	"api/internal/store"
	"api/internal/store/inmemory"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOfficerTransitionStagingDoesNotChangeActiveRoles(t *testing.T) {
	st := inmemory.NewStore()
	for i, id := range []string{"pres", "vp", "sec", "treas"} {
		require.NoError(t, st.Members().Create(&models.User{ID: uint64(i + 1), QuestID: id, FirstName: "First" + id, LastName: "Last" + id}))
	}
	require.NoError(t, st.Logins().Create(&models.Login{Username: "pres", Password: "old", Role: "vice_president", Status: models.LoginStatusActive}))

	transition, tokens, err := NewOfficerTransitionService(st).Create("outgoing", models.CreateOfficerTransitionRequest{
		PresidentQuestID: "pres", VicePresidentQuestID: "vp", SecretaryQuestID: "sec", TreasurerQuestID: "treas",
	})
	require.NoError(t, err)
	require.Equal(t, models.OfficerTransitionPending, transition.Status)
	require.NotEmpty(t, tokens["president"])
	require.NotEmpty(t, tokens["vice_president"])
	require.Equal(t, models.OfficerTransitionNominee{FirstName: "Firstpres", LastName: "Lastpres"}, transition.Nominees["president"])
	require.Equal(t, "vice_president", mustLogin(t, st, "pres").Role)
	require.Equal(t, models.LoginStatusActive, mustLogin(t, st, "pres").Status)
	require.Equal(t, models.LoginStatusPendingActivation, mustLogin(t, st, "vp").Status)
}

func TestOfficerTransitionRejectsDuplicateQuestIDs(t *testing.T) {
	st := inmemory.NewStore()
	_, _, err := NewOfficerTransitionService(st).Create("outgoing", models.CreateOfficerTransitionRequest{PresidentQuestID: "same", VicePresidentQuestID: "same", SecretaryQuestID: "s", TreasurerQuestID: "t"})
	require.ErrorIs(t, err, ErrTransitionInvalid)
}

func mustLogin(t *testing.T, st store.Store, username string) models.Login {
	t.Helper()
	login, err := st.Logins().FindByUsername(username)
	require.NoError(t, err)
	return login
}
