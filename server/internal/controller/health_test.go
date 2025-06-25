package controller_test

import (
	"api/internal/server"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	apiServer := server.NewAPIServer(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)

	apiServer.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}
