package services

import (
	"api/internal/models"
	"api/internal/store/inmemory"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginService_CreateLogin(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	svc := NewLoginService(st)

	require.NoError(t, svc.CreateLogin("alice", "password123", "executive"))

	login, err := st.Logins().FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "executive", login.Role)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("password123")))
}

func TestLoginService_UpdateLogin_NoFields(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	svc := NewLoginService(st)
	require.NoError(t, svc.CreateLogin("alice", "password123", "executive"))

	err := svc.UpdateLogin("alice", nil, nil, nil)
	require.ErrorIs(t, err, ErrUpdateLoginNoFields)
}

func TestLoginService_UpdateLogin_Password(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	svc := NewLoginService(st)
	require.NoError(t, svc.CreateLogin("alice", "password123", "executive"))

	newPassword := "newpassword456"
	require.NoError(t, svc.UpdateLogin("alice", &newPassword, nil, nil))

	login, err := st.Logins().FindByUsername("alice")
	require.NoError(t, err)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte(newPassword)))
	require.Equal(t, "executive", login.Role)
}

func TestLoginService_UpdateLogin_Role(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	svc := NewLoginService(st)
	require.NoError(t, svc.CreateLogin("alice", "password123", "executive"))

	newRole := "treasurer"
	require.NoError(t, svc.UpdateLogin("alice", nil, &newRole, nil))

	login, err := st.Logins().FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, "treasurer", login.Role)
}

func TestLoginService_UpdateLogin_NotFound(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	svc := NewLoginService(st)

	newRole := "treasurer"
	err := svc.UpdateLogin("nobody", nil, &newRole, nil)
	require.ErrorIs(t, err, ErrLoginNotFound)
}

func TestLoginService_UpdateLogin_DisableInvalidatesSessions(t *testing.T) {
	t.Parallel()

	st := inmemory.NewStore()
	svc := NewLoginService(st)
	require.NoError(t, svc.CreateLogin("alice", "password123", "executive"))
	session := &models.Session{
		Username:  "alice",
		Role:      "executive",
		StartedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	require.NoError(t, st.Sessions().Create(session))

	disabled := models.LoginStatusDisabled
	require.NoError(t, svc.UpdateLogin("alice", nil, nil, &disabled))

	login, err := st.Logins().FindByUsername("alice")
	require.NoError(t, err)
	require.Equal(t, models.LoginStatusDisabled, login.Status)
	_, err = st.Sessions().FindByID(session.ID)
	require.Error(t, err)
}
