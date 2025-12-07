package controller_test

import (
	"api/internal/authorization"
	"api/internal/models"
	"api/internal/testutils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	unauthorizedRoles := []string{authorization.ROLE_BOT.ToString(), authorization.ROLE_EXECUTIVE.ToString()}
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "POST", "/api/v2/members", unauthorizedRoles)

	testCases := []struct {
		name                 string
		userRole             string
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:     "successful creation",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"id":        uint64(12345678),
				"firstName": "John",
				"lastName":  "Doe",
				"email":     "john.doe@uwaterloo.ca",
				"faculty":   "Math",
				"questId":   "jdoe",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:     "missing required field - email",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"id":        uint64(12345679),
				"firstName": "Jane",
				"lastName":  "Smith",
				"faculty":   "Engineering",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateUserRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag",
		},
		{
			name:     "invalid faculty",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"id":        uint64(12345680),
				"firstName": "Bob",
				"lastName":  "Johnson",
				"email":     "bob@uwaterloo.ca",
				"faculty":   "InvalidFaculty",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateUserRequest.Faculty' Error:Field validation for 'Faculty' failed on the 'oneof' tag",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data
			require.NoError(t, testutils.SeedAll(db))

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"POST",
				"/api/v2/members",
				tc.requestBody,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert status code
			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			// Assert response body
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
			} else {
				var user models.User
				err := json.Unmarshal(w.Body.Bytes(), &user)
				require.NoError(t, err)
				require.Equal(t, tc.requestBody["email"], user.Email)
			}
		})
	}
}

func TestListMembers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// No unauthorized roles - all roles have access (bot+)

	testCases := []struct {
		name           string
		userRole       string
		queryParams    string
		expectedStatus int
		expectEmpty    bool
	}{
		{
			name:           "list all members",
			userRole:       authorization.ROLE_BOT.ToString(),
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectEmpty:    false,
		},
		{
			name:           "filter by email",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			queryParams:    "?email=john.doe@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter by name",
			userRole:       authorization.ROLE_BOT.ToString(),
			queryParams:    "?name=John",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter by faculty",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			queryParams:    "?faculty=Math",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter by id",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			queryParams:    "?id=20780648",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data
			require.NoError(t, testutils.SeedAll(db))

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"GET",
				"/api/v2/members"+tc.queryParams,
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			var users []models.User
			err = json.Unmarshal(w.Body.Bytes(), &users)
			require.NoError(t, err)

			if tc.expectEmpty {
				require.Empty(t, users)
			}
		})
	}
}

func TestGetMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden cases
	unauthorizedRoles := []string{authorization.ROLE_BOT.ToString()}
	testUserID := testutils.TEST_USERS[0].ID
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"GET",
		fmt.Sprintf("/api/v2/members/%d", testUserID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name           string
		userRole       string
		memberID       string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful retrieval",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			memberID:       fmt.Sprint(testUserID),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "member not found",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			memberID:       "99999998",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "invalid ID format",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			memberID:       "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "negative ID",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			memberID:       "-1",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data
			require.NoError(t, testutils.SeedAll(db))

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"GET",
				"/api/v2/members/"+tc.memberID,
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if !tc.expectError {
				var user models.User
				err := json.Unmarshal(w.Body.Bytes(), &user)
				require.NoError(t, err)
				require.NotEmpty(t, user.ID)
			}
		})
	}
}

func TestUpdateMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden cases
	unauthorizedRoles := []string{authorization.ROLE_BOT.ToString(), authorization.ROLE_EXECUTIVE.ToString()}
	testUserID := testutils.TEST_USERS[0].ID
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"PATCH",
		fmt.Sprintf("/api/v2/members/%d", testUserID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name           string
		userRole       string
		memberID       string
		requestBody    map[string]any
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "successful update",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			memberID: fmt.Sprint(testUserID),
			requestBody: map[string]any{
				"firstName": "UpdatedName",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:     "update email",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			memberID: fmt.Sprint(testUserID),
			requestBody: map[string]any{
				"email": "newemail@uwaterloo.ca",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "member not found",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			memberID:       "99999998",
			requestBody:    map[string]any{"firstName": "Test"},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:     "invalid faculty",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			memberID: fmt.Sprint(testUserID),
			requestBody: map[string]any{
				"faculty": "InvalidFaculty",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data
			require.NoError(t, testutils.SeedAll(db))

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"PATCH",
				"/api/v2/members/"+tc.memberID,
				tc.requestBody,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if !tc.expectError {
				var user models.User
				err := json.Unmarshal(w.Body.Bytes(), &user)
				require.NoError(t, err)
				if firstName, ok := tc.requestBody["firstName"]; ok {
					require.Equal(t, firstName, user.FirstName)
				}
			}
		})
	}
}

func TestDeleteMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden cases
	unauthorizedRoles := []string{authorization.ROLE_BOT.ToString(), authorization.ROLE_EXECUTIVE.ToString()}
	testUserID := testutils.TEST_USERS[0].ID
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"DELETE",
		fmt.Sprintf("/api/v2/members/%d", testUserID),
		unauthorizedRoles,
	)

	// Use TEST_USERS[3] (Alice) who has no memberships
	testDeleteUserID := testutils.TEST_USERS[3].ID

	testCases := []struct {
		name           string
		userRole       string
		memberID       string
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			memberID:       fmt.Sprint(testDeleteUserID),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "member not found",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			memberID:       "99999998",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid ID format",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			memberID:       "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data
			require.NoError(t, testutils.SeedAll(db))

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"DELETE",
				"/api/v2/members/"+tc.memberID,
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			// Verify deletion if successful
			if tc.expectedStatus == http.StatusNoContent {
				// Create request to verify deletion
				getReq, err := testutils.MakeJSONRequest(
					"GET",
					"/api/v2/members/"+tc.memberID,
					nil,
				)
				require.NoError(t, err)

				testutils.SetAuthCookie(getReq, sessionID)

				getW := httptest.NewRecorder()
				apiServer.ServeHTTP(getW, getReq)
				require.Equal(t, http.StatusNotFound, getW.Code)
			}
		})
	}
}
