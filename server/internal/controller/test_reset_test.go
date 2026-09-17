//go:build e2e

package controller_test

import (
	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTestResetControllerClearsStagedOfficerTransition(t *testing.T) {
	t.Setenv("ENVIRONMENT", "test")

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	api := testutils.NewTestAPIServer(container.GetDB())
	reset := func() {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, "/api/v2/test/reset", nil)
		require.NoError(t, err)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	}
	stage := func() {
		t.Helper()
		_, _, err := services.NewOfficerTransitionService(postgres.NewStore(container.GetDB())).Create(
			"test_president",
			models.CreateOfficerTransitionRequest{
				PresidentQuestID:     "hdrust0",
				VicePresidentQuestID: "dhousegoe1",
				SecretaryQuestID:     "eaucock2",
				TreasurerQuestID:     "kduckham3",
			},
		)
		require.NoError(t, err)
	}

	reset()
	stage()
	reset()
	stage()
}
