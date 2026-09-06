package controller_test

import (
	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAccountActivations(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)
	seed := func(t *testing.T) string {
		t.Helper()
		require.NoError(t, container.ResetDatabase(ctx))
		require.NoError(t, db.Create(&models.User{ID: 1, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com", Faculty: models.FacultyMath, QuestID: "questid"}).Error)
		require.NoError(t, db.Create(&models.Login{Username: "questid", Password: "pending", Role: "executive", Status: models.LoginStatusPendingActivation}).Error)
		token, err := services.NewAccountActivationService(postgres.NewStore(db)).Create("questid")
		require.NoError(t, err)
		return token
	}

	t.Run("verify does not consume a valid token", func(t *testing.T) {
		token := seed(t)
		req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/verify", map[string]string{"token": token})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var response models.VerifyActivationResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, models.VerifyActivationResponse{FirstName: "Ada", LastName: "Lovelace"}, response)
		var activation models.AccountActivation
		require.NoError(t, db.First(&activation).Error)
		require.Nil(t, activation.UsedAt)
	})

	t.Run("complete activates login, consumes token, and creates session", func(t *testing.T) {
		token := seed(t)
		complete := func() *httptest.ResponseRecorder {
			req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/complete", map[string]string{"token": token, "password": "newpassword"})
			require.NoError(t, err)
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)
			return w
		}
		w := complete()
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		require.NotNil(t, findCookie(w, devSessionCookie))
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", "questid").Error)
		require.Equal(t, models.LoginStatusActive, login.Status)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("newpassword")))
		require.Equal(t, http.StatusUnauthorized, complete().Code)
	})

	t.Run("new activation invalidates a previous token", func(t *testing.T) {
		seed(t)
		svc := services.NewAccountActivationService(postgres.NewStore(db))
		first, err := svc.Create("questid")
		require.NoError(t, err)
		second, err := svc.Create("questid")
		require.NoError(t, err)

		verify := func(token string) int {
			req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/verify", map[string]string{"token": token})
			require.NoError(t, err)
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)
			return w.Code
		}
		require.Equal(t, http.StatusUnauthorized, verify(first))
		require.Equal(t, http.StatusOK, verify(second))
	})
}
