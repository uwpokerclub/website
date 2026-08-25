package server_test

import (
	"api/internal/testutils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveSPAFixture makes c.File("./public/index.html") resolve during the test by
// chdir'ing into a temp directory holding a stand-in bundle. Without it every
// unmatched route 404s simply because the file is missing, which would hide the
// difference between the API fallback and the SPA fallback.
func serveSPAFixture(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "public"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "public", "index.html"),
		[]byte("<!doctype html><title>UW Poker Studies Club</title>"),
		0o644,
	))
	t.Chdir(dir)
}

// TestNoRouteDoesNotServeSPAForAPIPaths covers the routes deleted with the v1
// API: they must report a JSON 404 rather than falling through to the single
// page app, which would answer an uptime probe with a 200 and an HTML body.
func TestNoRouteDoesNotServeSPAForAPIPaths(t *testing.T) {
	serveSPAFixture(t)

	router := testutils.NewTestAPIServer(nil)

	paths := []string{"/api", "/api/health", "/api/session", "/api/v2/does-not-exist"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var body map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body),
				"expected a JSON error, got %q", w.Body.String())
			assert.Equal(t, float64(http.StatusNotFound), body["code"])
			assert.Equal(t, "NOT_FOUND", body["type"])
		})
	}
}

// TestNoRouteServesSPAForNonAPIPaths guards the other half: React Router owns
// every non-/api path, so deep links must still reach the bundle. "/apiary"
// is here on purpose - the guard matches the "/api/" prefix, not "api", and
// must not swallow a client route that merely starts with those letters.
func TestNoRouteServesSPAForNonAPIPaths(t *testing.T) {
	serveSPAFixture(t)

	router := testutils.NewTestAPIServer(nil)

	paths := []string{"/", "/admin/events", "/events/123", "/rankings/fall-2023", "/apiary"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "UW Poker Studies Club")
		})
	}
}
