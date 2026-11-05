package services

import (
	"api/internal/models"
	"api/internal/store"
	"api/internal/store/inmemory"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createActivationInCallerTx(t *testing.T, st store.Store, username string) string {
	t.Helper()
	tx, err := st.BeginTx()
	require.NoError(t, err)
	token, err := NewAccountActivationService(tx).Create(username)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	return token
}

func TestAccountActivationService_CreateAndVerify(t *testing.T) {
	st := inmemory.NewStore()
	require.NoError(t, st.Logins().Create(&models.Login{
		Username: "questid",
		Password: "unused",
		Role:     "executive",
		Status:   models.LoginStatusPendingActivation,
	}))

	token := createActivationInCallerTx(t, st, "questid")

	activation, err := NewAccountActivationService(st).Verify(token)
	require.NoError(t, err)
	require.Equal(t, "questid", activation.Username)
	require.Nil(t, activation.UsedAt, "verification must not consume the token")
}

func TestAccountActivationService_RejectsExpiredToken(t *testing.T) {
	st := inmemory.NewStore()
	token := "expired-token"
	require.NoError(t, st.AccountActivations().Create(&models.AccountActivation{
		TokenHash: tokenHash(token), Username: "questid", ExpiresAt: time.Now().UTC().Add(-time.Second),
	}))

	_, err := NewAccountActivationService(st).Verify(token)
	require.ErrorIs(t, err, ErrActivationNotFound)
}

func TestAccountActivationService_ConcurrentCompleteAllowsExactlyOne(t *testing.T) {
	st := inmemory.NewStore()
	require.NoError(t, st.Logins().Create(&models.Login{Username: "questid", Password: "unused", Role: "executive", Status: models.LoginStatusPendingActivation}))
	svc := NewAccountActivationService(st)
	token := createActivationInCallerTx(t, st, "questid")

	errs := make(chan error, 2)
	for range 2 {
		go func() { _, _, err := svc.Complete(token, "newpassword"); errs <- err }()
	}
	var successes int
	for range 2 {
		if <-errs == nil {
			successes++
		}
	}
	require.Equal(t, 1, successes)
}

func TestAccountActivationService_CreateRevokesPreviousUnusedToken(t *testing.T) {
	st := inmemory.NewStore()
	require.NoError(t, st.Logins().Create(&models.Login{Username: "questid", Password: "unused", Role: "executive", Status: models.LoginStatusPendingActivation}))
	svc := NewAccountActivationService(st)
	first := createActivationInCallerTx(t, st, "questid")
	second := createActivationInCallerTx(t, st, "questid")

	_, err := svc.Verify(first)
	require.ErrorIs(t, err, ErrActivationNotFound)
	_, err = svc.Verify(second)
	require.NoError(t, err)
}

func TestAccountActivationService_CompleteDoesNotReactivateDisabledLogin(t *testing.T) {
	st := inmemory.NewStore()
	require.NoError(t, st.Logins().Create(&models.Login{Username: "questid", Password: "unused", Role: "president", Status: models.LoginStatusPendingActivation}))
	svc := NewAccountActivationService(st)
	token := createActivationInCallerTx(t, st, "questid")
	require.NoError(t, st.Logins().Update("questid", map[string]any{"status": models.LoginStatusDisabled}))

	_, _, err := svc.Complete(token, "newpassword")
	require.ErrorIs(t, err, ErrActivationNotFound)
	login, err := st.Logins().FindByUsername("questid")
	require.NoError(t, err)
	require.Equal(t, models.LoginStatusDisabled, login.Status)
}

func TestAccountActivationService_CreateUsesCallerTransaction(t *testing.T) {
	st := inmemory.NewStore()
	require.NoError(t, st.Logins().Create(&models.Login{Username: "questid", Password: "unused", Role: "executive", Status: models.LoginStatusPendingActivation}))
	tx, err := st.BeginTx()
	require.NoError(t, err)
	token, err := NewAccountActivationService(tx).Create("questid")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	_, err = NewAccountActivationService(st).Verify(token)
	require.ErrorIs(t, err, ErrActivationNotFound, "Create must not commit its caller's transaction")
}
