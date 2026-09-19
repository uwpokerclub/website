package controller_test

import (
	"api/internal/authorization"
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
)

func TestOfficerTransitionRecoveryEndpoints(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	api := testutils.NewTestAPIServer(db)
	seed := func(t *testing.T) models.OfficerTransition {
		t.Helper()
		require.NoError(t, container.ResetDatabase(ctx))
		for i, username := range []string{"pres", "vp", "sec", "treas"} {
			require.NoError(t, db.Create(&models.User{ID: uint64(i + 1), FirstName: username, LastName: username, Email: username + "@example.com", Faculty: models.FacultyMath, QuestID: username}).Error)
		}
		transition, _, err := services.NewOfficerTransitionService(postgres.NewStore(db)).Create("outgoing", models.CreateOfficerTransitionRequest{PresidentQuestID: "pres", VicePresidentQuestID: "vp", SecretaryQuestID: "sec", TreasurerQuestID: "treas"})
		require.NoError(t, err)
		return transition
	}
	auth := func(t *testing.T, method, path, role string, body any) *httptest.ResponseRecorder {
		t.Helper()
		session, err := testutils.CreateTestSession(db, "actor-"+role, role)
		require.NoError(t, err)
		req, err := testutils.MakeJSONRequest(method, path, body)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, session)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		return w
	}

	t.Run("president and webmaster can cancel, lower roles cannot", func(t *testing.T) {
		transition := seed(t)
		w := auth(t, http.MethodPost, "/api/v2/officer-transitions/"+transition.ID.String()+"/cancel", authorization.ROLE_VICE_PRESIDENT.ToString(), nil)
		require.Equal(t, http.StatusForbidden, w.Code)
		w = auth(t, http.MethodPost, "/api/v2/officer-transitions/"+transition.ID.String()+"/cancel", authorization.ROLE_WEBMASTER.ToString(), nil)
		require.Equal(t, http.StatusNoContent, w.Code)
		require.Empty(t, w.Body.String())
	})
	t.Run("reissue validates ID and role", func(t *testing.T) {
		seed(t)
		w := auth(t, http.MethodPost, "/api/v2/officer-transitions/not-a-uuid/reissue", authorization.ROLE_PRESIDENT.ToString(), map[string]string{"role": "president"})
		require.Equal(t, http.StatusBadRequest, w.Code)
		transition := seed(t)
		w = auth(t, http.MethodPost, "/api/v2/officer-transitions/"+transition.ID.String()+"/reissue", authorization.ROLE_PRESIDENT.ToString(), map[string]string{"role": "nope"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("president can reissue and receives the activation token", func(t *testing.T) {
		transition := seed(t)
		w := auth(t, http.MethodPost, "/api/v2/officer-transitions/"+transition.ID.String()+"/reissue", authorization.ROLE_PRESIDENT.ToString(), map[string]string{"role": "president"})
		require.Equal(t, http.StatusCreated, w.Code)
		var response struct {
			ActivationToken string `json:"activationToken"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.NotEmpty(t, response.ActivationToken)
	})
	t.Run("current transition exposes nominee activation progress to the executive ladder", func(t *testing.T) {
		transition := seed(t)
		require.NoError(t, db.Model(&models.Login{}).Where("username = ?", "vp").Update("status", models.LoginStatusActive).Error)
		w := auth(t, http.MethodGet, "/api/v2/officer-transitions/current", authorization.ROLE_EXECUTIVE.ToString(), nil)
		require.Equal(t, http.StatusOK, w.Code)
		var response struct {
			ID        string          `json:"id"`
			Activated map[string]bool `json:"activated"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, transition.ID.String(), response.ID)
		require.False(t, response.Activated["president"])
		require.True(t, response.Activated["vice_president"])
	})
	t.Run("completed transition is available only to its incoming president", func(t *testing.T) {
		transition := seed(t)
		_, err := postgres.NewStore(db).OfficerTransitions().ClaimCompletion(transition.ID)
		require.NoError(t, err)
		require.NoError(t, db.Where("username = ?", "pres").Delete(&models.Login{}).Error)
		session, err := testutils.CreateTestSession(db, "pres", authorization.ROLE_PRESIDENT.ToString())
		require.NoError(t, err)
		req, err := testutils.MakeJSONRequest(http.MethodGet, "/api/v2/officer-transitions/completed", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, session)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		var response struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, transition.ID.String(), response.ID)
	})
}
