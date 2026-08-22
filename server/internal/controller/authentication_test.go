package controller_test

import (
	"api/internal/authorization"
	"api/internal/models"
	"api/internal/testutils"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const devSessionCookie = "uwpsc-dev-session-id"

// seedLoginWithPassword creates a login whose password column holds a real
// bcrypt hash. testutils.CreateTestSession deliberately does not — it stores
// the literal string "hashed_password" — so it cannot be used to test login.
func seedLoginWithPassword(t *testing.T, db *gorm.DB, username, password, role string) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	require.NoError(t, db.Create(&models.Login{
		Username: username,
		Password: string(hash),
		Role:     role,
	}).Error)
}

// findCookie returns the named cookie from the response, or nil.
func findCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	testCases := []struct {
		name           string
		body           any
		expectedStatus int
		expectSession  bool
	}{
		{
			name:           "valid credentials",
			body:           models.NewSessionRequest{Username: "testuser", Password: "password"},
			expectedStatus: http.StatusCreated,
			expectSession:  true,
		},
		{
			name:           "incorrect password",
			body:           models.NewSessionRequest{Username: "testuser", Password: "wrongpassword"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unknown username",
			body:           models.NewSessionRequest{Username: "nobody", Password: "password"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing password",
			body:           map[string]string{"username": "testuser"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))
			seedLoginWithPassword(t, db, "testuser", "password", authorization.ROLE_EXECUTIVE.ToString())

			req, err := testutils.MakeJSONRequest("POST", "/api/v2/session", tc.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			cookie := findCookie(w, devSessionCookie)
			if !tc.expectSession {
				require.Nil(t, cookie, "no session cookie should be set on a failed login")
				return
			}

			require.NotNil(t, cookie, "expected a %s cookie", devSessionCookie)
			require.NoError(t, uuid.Validate(cookie.Value))
			require.True(t, cookie.HttpOnly)

			// The session must actually be persisted, which proves the handler
			// went through store.Sessions().Create.
			var session models.Session
			require.NoError(t, db.Where("id = ?", cookie.Value).First(&session).Error)
			require.Equal(t, "testuser", session.Username)
			require.Equal(t, authorization.ROLE_EXECUTIVE.ToString(), session.Role)
		})
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	t.Run("existing session", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))

		sessionID, err := testutils.CreateTestSession(db, "testuser", authorization.ROLE_EXECUTIVE.ToString())
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("POST", "/api/v2/session/logout", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code, "Response: %s", w.Body.String())

		var count int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", sessionID).Count(&count).Error)
		require.Zero(t, count, "session should have been deleted")
	})

	// Pins the store.ErrNotFound swallow in sessionManager.Invalidate: the
	// repository reports a missing row as an error, but logging out an
	// already-gone session is still a success.
	t.Run("unknown session id", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))

		req, err := testutils.MakeJSONRequest("POST", "/api/v2/session/logout", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, uuid.New())

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code, "Response: %s", w.Body.String())
	})

	t.Run("no cookie", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))

		req, err := testutils.MakeJSONRequest("POST", "/api/v2/session/logout", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code, "Response: %s", w.Body.String())
	})

	t.Run("malformed cookie", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))

		req, err := testutils.MakeJSONRequest("POST", "/api/v2/session/logout", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: devSessionCookie, Value: "not-a-uuid"})

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code, "Response: %s", w.Body.String())
	})
}

func TestGetSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// GET /session is authenticated but not authorized against any action, so
	// only the unauthenticated case applies here.
	t.Run("unauthenticated", func(t *testing.T) {
		testutils.TestUnauthenticatedEndpoint(t, container, apiServer, "GET", "/api/v2/session")
	})

	t.Run("authenticated", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))

		role := authorization.ROLE_EXECUTIVE.ToString()
		sessionID, err := testutils.CreateTestSession(db, "testuser", role)
		require.NoError(t, err)

		req, err := testutils.MakeJSONRequest("GET", "/api/v2/session", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, sessionID)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Response: %s", w.Body.String())

		var resp models.GetSessionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "testuser", resp.Username)
		require.Equal(t, role, resp.Role)

		// Compare against the same source the handler uses rather than
		// hardcoding the permission matrix, which changes with the authorizers.
		authSvc := authorization.NewAuthorizationService(role, authorization.DefaultAuthorizerMap)
		expected, err := json.Marshal(authSvc.GetPermissions())
		require.NoError(t, err)
		actual, err := json.Marshal(resp.Permissions)
		require.NoError(t, err)
		require.JSONEq(t, string(expected), string(actual))
	})

	// A session past its expiry must be rejected and cleaned up.
	t.Run("expired session", func(t *testing.T) {
		require.NoError(t, container.ResetDatabase(ctx))

		require.NoError(t, db.Create(&models.Login{
			Username: "expireduser",
			Password: "hashed_password",
			Role:     authorization.ROLE_EXECUTIVE.ToString(),
		}).Error)

		expiredID := uuid.New()
		require.NoError(t, db.Create(&models.Session{
			ID:        expiredID,
			StartedAt: time.Now().Add(-9 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Username:  "expireduser",
			Role:      authorization.ROLE_EXECUTIVE.ToString(),
		}).Error)

		req, err := testutils.MakeJSONRequest("GET", "/api/v2/session", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, expiredID)

		w := httptest.NewRecorder()
		apiServer.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code, "Response: %s", w.Body.String())

		var count int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", expiredID).Count(&count).Error)
		require.Zero(t, count, "expired session should have been deleted")
	})
}
