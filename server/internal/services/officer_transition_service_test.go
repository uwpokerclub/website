package services

import (
	"api/internal/models"
	"api/internal/store"
	"api/internal/store/inmemory"
	"testing"

	"github.com/google/uuid"
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

func TestOfficerTransitionCancelRemovesOnlyOwnedStagingLogins(t *testing.T) {
	st := transitionStore(t)
	require.NoError(t, st.Logins().Create(&models.Login{Username: "pres", Password: "old", Role: "vice_president", Status: models.LoginStatusActive}))
	transition, tokens, err := NewOfficerTransitionService(st).Create("outgoing", transitionRequest())
	require.NoError(t, err)
	// A historical completed transition may still name a login this staging
	// transition created; cancellation must preserve that account.
	require.NoError(t, st.OfficerTransitions().Create(&models.OfficerTransition{ID: uuid.New(), InitiatedBy: "earlier", PresidentUsername: "vp", VicePresidentUsername: "old-vp", SecretaryUsername: "old-sec", TreasurerUsername: "old-treas", Status: models.OfficerTransitionCompleted}))
	_, err = NewOfficerTransitionService(st).Cancel(transition.ID)
	require.NoError(t, err)
	_, err = st.Logins().FindByUsername("sec")
	require.ErrorIs(t, err, store.ErrNotFound)
	_, err = st.Logins().FindByUsername("vp")
	require.NoError(t, err)
	require.Equal(t, models.LoginStatusActive, mustLogin(t, st, "pres").Status)
	_, err = NewAccountActivationService(st).Verify(tokens["president"])
	require.ErrorIs(t, err, ErrActivationNotFound)
}

func TestOfficerTransitionCancelPreservesPreexistingPendingLogin(t *testing.T) {
	st := transitionStore(t)
	require.NoError(t, st.Logins().Create(&models.Login{Username: "vp", Password: "old", Role: "executive", Status: models.LoginStatusPendingActivation}))
	transition, _, err := NewOfficerTransitionService(st).Create("outgoing", transitionRequest())
	require.NoError(t, err)
	require.Nil(t, mustLogin(t, st, "vp").StagedTransitionID)
	_, err = NewOfficerTransitionService(st).Cancel(transition.ID)
	require.NoError(t, err)
	login := mustLogin(t, st, "vp")
	require.Equal(t, models.LoginStatusPendingActivation, login.Status)
	require.Nil(t, login.StagedTransitionID)
}

func TestOfficerTransitionReissueReplacesTokenAndChecksEligibility(t *testing.T) {
	st := transitionStore(t)
	require.NoError(t, st.Logins().Create(&models.Login{Username: "pres", Password: "old", Role: "vice_president", Status: models.LoginStatusActive}))
	require.NoError(t, st.Logins().Create(&models.Login{Username: "vp", Password: "old", Role: "executive", Status: models.LoginStatusActive}))
	transition, tokens, err := NewOfficerTransitionService(st).Create("outgoing", transitionRequest())
	require.NoError(t, err)
	fresh, err := NewOfficerTransitionService(st).Reissue(transition.ID, "president")
	require.NoError(t, err)
	require.NotEqual(t, tokens["president"], fresh)
	_, err = NewAccountActivationService(st).Verify(tokens["president"])
	require.ErrorIs(t, err, ErrActivationNotFound)
	_, err = NewAccountActivationService(st).Verify(fresh)
	require.NoError(t, err)
	_, err = NewOfficerTransitionService(st).Reissue(transition.ID, "vice_president")
	require.ErrorIs(t, err, ErrTransitionIneligible)
	_, err = NewOfficerTransitionService(st).Reissue(transition.ID, "invalid")
	require.ErrorIs(t, err, ErrTransitionInvalid)
	_, err = NewOfficerTransitionService(st).Cancel(transition.ID)
	require.NoError(t, err)
	_, err = NewOfficerTransitionService(st).Reissue(transition.ID, "president")
	require.ErrorIs(t, err, ErrTransitionResolved)
}

func TestOfficerTransitionReissueRejectsDisabledOrMissingPresident(t *testing.T) {
	st := transitionStore(t)
	require.NoError(t, st.Logins().Create(&models.Login{Username: "pres", Password: "old", Role: "vice_president", Status: models.LoginStatusActive}))
	transition, _, err := NewOfficerTransitionService(st).Create("outgoing", transitionRequest())
	require.NoError(t, err)
	require.NoError(t, st.Logins().Update("pres", map[string]any{"status": models.LoginStatusDisabled}))
	_, err = NewOfficerTransitionService(st).Reissue(transition.ID, "president")
	require.ErrorIs(t, err, ErrTransitionIneligible)
	require.NoError(t, st.Logins().Delete("pres"))
	_, err = NewOfficerTransitionService(st).Reissue(transition.ID, "president")
	require.ErrorIs(t, err, ErrTransitionIneligible)
}

func TestOfficerTransitionCancelRejectsResolvedStates(t *testing.T) {
	st := transitionStore(t)
	for _, status := range []string{models.OfficerTransitionCompleted, models.OfficerTransitionCancelled} {
		transition := models.OfficerTransition{ID: uuid.New(), InitiatedBy: "outgoing", PresidentUsername: "pres", VicePresidentUsername: "vp", SecretaryUsername: "sec", TreasurerUsername: "treas", Status: status}
		require.NoError(t, st.OfficerTransitions().Create(&transition))
		_, err := NewOfficerTransitionService(st).Cancel(transition.ID)
		require.ErrorIs(t, err, ErrTransitionResolved)
	}
}

func transitionStore(t *testing.T) store.Store {
	t.Helper()
	st := inmemory.NewStore()
	for i, id := range []string{"pres", "vp", "sec", "treas"} {
		require.NoError(t, st.Members().Create(&models.User{ID: uint64(i + 1), QuestID: id, FirstName: id, LastName: id}))
	}
	return st
}

func transitionRequest() models.CreateOfficerTransitionRequest {
	return models.CreateOfficerTransitionRequest{PresidentQuestID: "pres", VicePresidentQuestID: "vp", SecretaryQuestID: "sec", TreasurerQuestID: "treas"}
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
