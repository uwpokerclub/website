package controller_test

import (
	"api/internal/testutils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHealthCheck runs against a nil database on purpose: the health controller
// takes no store, and postgres.NewStore only hands the handle to its
// repositories rather than dereferencing it, so this stays the one test in the
// package that needs no container.
//
// The body is asserted exactly because the e2e workflow's readiness gate greps
// for it; see .github/workflows/e2e.yml.
func TestHealthCheck(t *testing.T) {
	t.Parallel()

	apiServer := testutils.NewTestAPIServer(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v2/health", nil)

	apiServer.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}
