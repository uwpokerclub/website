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

	"github.com/stretchr/testify/require"
)

// All non-webmaster roles - logins endpoints are webmaster-only
var loginsUnauthorizedRoles = []string{
	authorization.ROLE_BOT.ToString(),
	authorization.ROLE_EXECUTIVE.ToString(),
	authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
	authorization.ROLE_SECRETARY.ToString(),
	authorization.ROLE_TREASURER.ToString(),
	authorization.ROLE_VICE_PRESIDENT.ToString(),
	authorization.ROLE_PRESIDENT.ToString(),
}

func TestListLogins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "GET", "/api/v2/logins", loginsUnauthorizedRoles)

	testCases := []struct {
		name           string
		setupLogins    func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "empty list",
			setupLogins:    func() {},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "list multiple logins",
			setupLogins: func() {
				db.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"})
				db.Create(&models.Login{Username: "bob", Password: "hash2", Role: "president"})
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "list logins with linked member",
			setupLogins: func() {
				// Create user with QuestID matching login username
				db.Create(&models.User{
					ID:        12345678,
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@example.com",
					Faculty:   "Math",
					QuestID:   "jdoe",
				})
				db.Create(&models.Login{Username: "jdoe", Password: "hash1", Role: "executive"})
				db.Create(&models.Login{Username: "webmaster", Password: "hash2", Role: "webmaster"})
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Setup test data
			tc.setupLogins()

			// Setup authentication as webmaster
			sessionID, err := testutils.CreateTestSession(db, "testwebmaster", authorization.ROLE_WEBMASTER.ToString())
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("GET", "/api/v2/logins", nil)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			var resp models.ListResponse[models.LoginWithMember]
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			// Account for the webmaster session login created for auth (+1 if expectedCount > 0 for setup logins)
			// Actually the session creates its own login, so we need to account for that
			// The test webmaster login is separate, so expected count should be setup count + 1 (webmaster session)
			require.Len(t, resp.Data, tc.expectedCount+1) // +1 for testwebmaster session login
		})
	}
}

func TestGetLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "GET", "/api/v2/logins/testuser", loginsUnauthorizedRoles)

	testCases := []struct {
		name           string
		username       string
		setupLogin     func()
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "successful retrieval",
			username: "alice",
			setupLogin: func() {
				db.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"})
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:     "with linked member",
			username: "jdoe",
			setupLogin: func() {
				db.Create(&models.User{
					ID:        12345678,
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@example.com",
					Faculty:   "Math",
					QuestID:   "jdoe",
				})
				db.Create(&models.Login{Username: "jdoe", Password: "hash1", Role: "executive"})
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			username:       "nonexistent",
			setupLogin:     func() {},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Setup test data
			tc.setupLogin()

			// Setup authentication as webmaster
			sessionID, err := testutils.CreateTestSession(db, "testwebmaster", authorization.ROLE_WEBMASTER.ToString())
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("GET", "/api/v2/logins/"+tc.username, nil)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if !tc.expectError {
				var login models.LoginWithMember
				err := json.Unmarshal(w.Body.Bytes(), &login)
				require.NoError(t, err)
				require.Equal(t, tc.username, login.Username)
			}
		})
	}
}

func TestCreateLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "POST", "/api/v2/logins", loginsUnauthorizedRoles, map[string]any{
		"username": "newuser",
		"password": "password123",
		"role":     "executive",
	})

	testCases := []struct {
		name                 string
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name: "successful creation",
			requestBody: map[string]any{
				"username": "newuser",
				"password": "password123",
				"role":     "executive",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "missing username",
			requestBody: map[string]any{
				"password": "password123",
				"role":     "executive",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateLoginRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag",
		},
		{
			name: "missing password",
			requestBody: map[string]any{
				"username": "newuser",
				"role":     "executive",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateLoginRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag",
		},
		{
			name: "password too short",
			requestBody: map[string]any{
				"username": "newuser",
				"password": "short",
				"role":     "executive",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateLoginRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag",
		},
		{
			name: "missing role",
			requestBody: map[string]any{
				"username": "newuser",
				"password": "password123",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateLoginRequest.Role' Error:Field validation for 'Role' failed on the 'required' tag",
		},
		{
			name: "invalid role",
			requestBody: map[string]any{
				"username": "newuser",
				"password": "password123",
				"role":     "invalid_role",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateLoginRequest.Role' Error:Field validation for 'Role' failed on the 'oneof' tag",
		},
		{
			name: "create webmaster role",
			requestBody: map[string]any{
				"username": "newwebmaster",
				"password": "password123",
				"role":     "webmaster",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Setup authentication as webmaster
			sessionID, err := testutils.CreateTestSession(db, "testwebmaster", authorization.ROLE_WEBMASTER.ToString())
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("POST", "/api/v2/logins", tc.requestBody)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
			} else {
				var login models.LoginWithMember
				err := json.Unmarshal(w.Body.Bytes(), &login)
				require.NoError(t, err)
				require.Equal(t, tc.requestBody["username"], login.Username)
				require.Equal(t, tc.requestBody["role"], login.Role)
			}
		})
	}
}

func TestDeleteLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "DELETE", "/api/v2/logins/testuser", loginsUnauthorizedRoles)

	testCases := []struct {
		name           string
		username       string
		setupLogin     func()
		expectedStatus int
	}{
		{
			name:     "successful deletion",
			username: "alice",
			setupLogin: func() {
				db.Create(&models.Login{Username: "alice", Password: "hash1", Role: "executive"})
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "not found",
			username:       "nonexistent",
			setupLogin:     func() {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Setup test data
			tc.setupLogin()

			// Setup authentication as webmaster
			sessionID, err := testutils.CreateTestSession(db, "testwebmaster", authorization.ROLE_WEBMASTER.ToString())
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("DELETE", "/api/v2/logins/"+tc.username, nil)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			// Verify deletion if successful
			if tc.expectedStatus == http.StatusNoContent {
				// Verify login is deleted
				getReq, err := testutils.MakeJSONRequest("GET", "/api/v2/logins/"+tc.username, nil)
				require.NoError(t, err)
				testutils.SetAuthCookie(getReq, sessionID)

				getW := httptest.NewRecorder()
				apiServer.ServeHTTP(getW, getReq)
				require.Equal(t, http.StatusNotFound, getW.Code)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "PATCH", "/api/v2/logins/testuser/password", loginsUnauthorizedRoles, map[string]any{
		"newPassword": "newpassword123",
	})

	testCases := []struct {
		name                 string
		username             string
		setupLogin           func()
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:     "successful password change",
			username: "alice",
			setupLogin: func() {
				db.Create(&models.Login{Username: "alice", Password: "oldhash", Role: "executive"})
			},
			requestBody: map[string]any{
				"newPassword": "newpassword123",
			},
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:       "login not found",
			username:   "nonexistent",
			setupLogin: func() {},
			requestBody: map[string]any{
				"newPassword": "newpassword123",
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:     "missing password",
			username: "alice",
			setupLogin: func() {
				db.Create(&models.Login{Username: "alice", Password: "oldhash", Role: "executive"})
			},
			requestBody:          map[string]any{},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'ChangePasswordRequest.NewPassword' Error:Field validation for 'NewPassword' failed on the 'required' tag",
		},
		{
			name:     "password too short",
			username: "alice",
			setupLogin: func() {
				db.Create(&models.Login{Username: "alice", Password: "oldhash", Role: "executive"})
			},
			requestBody: map[string]any{
				"newPassword": "short",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'ChangePasswordRequest.NewPassword' Error:Field validation for 'NewPassword' failed on the 'min' tag",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Setup test data
			tc.setupLogin()

			// Setup authentication as webmaster
			sessionID, err := testutils.CreateTestSession(db, "testwebmaster", authorization.ROLE_WEBMASTER.ToString())
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("PATCH", "/api/v2/logins/"+tc.username+"/password", tc.requestBody)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if tc.expectError && tc.expectedErrorMessage != "" {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
			}
		})
	}
}
