package controller_test

import (
	"api/internal/testutils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	apiServer := testutils.NewTestAPIServer(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)

	apiServer.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}
