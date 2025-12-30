package controller_test

import (
	"api/internal/authorization"
	"api/internal/testutils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListStructures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	unauthorizedRoles := []string{"bot"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"GET",
		"/api/v2/structures",
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		seedStructures   bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:           "successful request no structures",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			seedStructures: false,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful request with structures",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	// Add tests for every authorized role
	authorizedRoles := []string{
		authorization.ROLE_EXECUTIVE.ToString(),
		authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
		authorization.ROLE_SECRETARY.ToString(),
		authorization.ROLE_TREASURER.ToString(),
		authorization.ROLE_VICE_PRESIDENT.ToString(),
		authorization.ROLE_PRESIDENT.ToString(),
		authorization.ROLE_WEBMASTER.ToString(),
	}
	for _, role := range authorizedRoles {
		testCases = append(testCases, struct {
			name             string
			userRole         string
			seedStructures   bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
		}{
			name:           fmt.Sprintf("successful request with role %s", role),
			userRole:       role,
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.seedStructures {
				require.NoError(t, testutils.SeedStructures(db))
			}

			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			req, err := testutils.MakeJSONRequest("GET", "/api/v2/structures", nil)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectError {
				var response []map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

				if tc.seedStructures {
					require.GreaterOrEqual(t, len(response), 1)
				} else {
					require.Equal(t, 0, len(response))
				}
			}
		})
	}
}

func TestCreateStructure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	unauthorizedRoles := []string{"bot", "executive"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"POST",
		"/api/v2/structures",
		unauthorizedRoles,
	)

	testCases := []struct {
		name                 string
		userRole             string
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:     "successful request with valid data",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name": "Test Structure",
				"blinds": []map[string]any{
					{"small": 25, "big": 50, "ante": 0, "time": 15},
					{"small": 50, "big": 100, "ante": 0, "time": 15},
					{"small": 75, "big": 150, "ante": 25, "time": 15},
				},
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:     "missing name field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"blinds": []map[string]any{
					{"small": 25, "big": 50, "ante": 0, "time": 15},
				},
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateStructureRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:     "missing blinds field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name": "Test Structure",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateStructureRequest.Blinds' Error:Field validation for 'Blinds' failed on the 'required' tag",
		},
		{
			name:     "empty blinds array",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":   "Test Structure",
				"blinds": []map[string]any{},
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateStructureRequest.Blinds' Error:Field validation for 'Blinds' failed on the 'min' tag",
		},
		{
			name:                 "empty body",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody:          nil,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateStructureRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'CreateStructureRequest.Blinds' Error:Field validation for 'Blinds' failed on the 'required' tag",
		},
	}

	// Add tests for every authorized role
	authorizedRoles := []string{
		authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
		authorization.ROLE_SECRETARY.ToString(),
		authorization.ROLE_TREASURER.ToString(),
		authorization.ROLE_VICE_PRESIDENT.ToString(),
		authorization.ROLE_PRESIDENT.ToString(),
		authorization.ROLE_WEBMASTER.ToString(),
	}
	for _, role := range authorizedRoles {
		testCases = append(testCases, struct {
			name                 string
			userRole             string
			requestBody          map[string]any
			expectedStatus       int
			expectError          bool
			expectedErrorMessage string
		}{
			name:     fmt.Sprintf("successful request with role %s", role),
			userRole: role,
			requestBody: map[string]any{
				"name": fmt.Sprintf("Structure by %s", role),
				"blinds": []map[string]any{
					{"small": 25, "big": 50, "ante": 0, "time": 15},
				},
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			req, err := testutils.MakeJSONRequest("POST", "/api/v2/structures", tc.requestBody)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
			} else {
				require.Equal(t, tc.expectedStatus, w.Code)

				var response map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
				require.NotNil(t, response["id"])
				require.Equal(t, tc.requestBody["name"], response["name"])
			}
		})
	}
}

func TestGetStructure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	unauthorizedRoles := []string{"bot"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"GET",
		"/api/v2/structures/1",
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		structureID      string
		seedStructures   bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:           "successful request",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			structureID:    "1",
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:             "invalid structure ID format",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			structureID:      "invalid-id",
			seedStructures:   false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Structure ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "negative structure ID",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			structureID:      "-1",
			seedStructures:   false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Structure ID must be a positive integer",
		},
		{
			name:             "non-existent structure",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			structureID:      "999",
			seedStructures:   false,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "record not found",
		},
	}

	// Add tests for every authorized role
	authorizedRoles := []string{
		authorization.ROLE_EXECUTIVE.ToString(),
		authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
		authorization.ROLE_SECRETARY.ToString(),
		authorization.ROLE_TREASURER.ToString(),
		authorization.ROLE_VICE_PRESIDENT.ToString(),
		authorization.ROLE_PRESIDENT.ToString(),
		authorization.ROLE_WEBMASTER.ToString(),
	}
	for _, role := range authorizedRoles {
		testCases = append(testCases, struct {
			name             string
			userRole         string
			structureID      string
			seedStructures   bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
		}{
			name:           fmt.Sprintf("successful request with role %s", role),
			userRole:       role,
			structureID:    "1",
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.seedStructures {
				require.NoError(t, testutils.SeedStructures(db))
			}

			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			req, err := testutils.MakeJSONRequest(
				"GET",
				fmt.Sprintf("/api/v2/structures/%s", tc.structureID),
				nil,
			)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				require.Equal(t, tc.expectedStatus, w.Code)

				var response map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
				require.NotNil(t, response["id"])
				require.NotNil(t, response["name"])
				require.NotNil(t, response["blinds"])
			}
		})
	}
}

func TestUpdateStructure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	unauthorizedRoles := []string{"bot", "executive"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"PATCH",
		"/api/v2/structures/1",
		unauthorizedRoles,
	)

	testCases := []struct {
		name                 string
		userRole             string
		structureID          string
		requestBody          map[string]any
		seedStructures       bool
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:        "successful update with name only",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"name": "Updated Structure Name",
			},
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "successful update with blinds only",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"blinds": []map[string]any{
					{"small": 100, "big": 200, "ante": 25, "time": 20},
					{"small": 200, "big": 400, "ante": 50, "time": 20},
				},
			},
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "successful update with all fields",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"name": "Fully Updated Structure",
				"blinds": []map[string]any{
					{"small": 50, "big": 100, "ante": 10, "time": 15},
				},
			},
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty body - no updates",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID:    "1",
			requestBody:    nil,
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "name cannot be null",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"name": nil,
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "name cannot be null",
		},
		{
			name:        "name cannot be empty string",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"name": "",
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "name must be a non-empty string",
		},
		{
			name:        "blinds cannot be null",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"blinds": nil,
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "blinds cannot be null",
		},
		{
			name:        "blinds cannot be empty array",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"blinds": []map[string]any{},
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "blinds array cannot be empty",
		},
		{
			name:        "unknown field in request",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"unknownField": "some value",
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "unknown field: unknownField",
		},
		{
			name:        "invalid structure ID format",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "invalid-id",
			requestBody: map[string]any{
				"name": "Valid Name",
			},
			seedStructures:       false,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Structure ID 'invalid-id' is not a valid integer",
		},
		{
			name:        "non-existent structure",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "999",
			requestBody: map[string]any{
				"name": "Valid Name",
			},
			seedStructures:       false,
			expectedStatus:       http.StatusNotFound,
			expectError:          true,
			expectedErrorMessage: "Structure not found",
		},
		{
			name:        "blind missing small field",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"blinds": []map[string]any{
					{"big": 50, "ante": 0, "time": 15},
				},
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "blind at index 0 is missing 'small' field",
		},
		{
			name:        "blind missing big field",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"blinds": []map[string]any{
					{"small": 25, "ante": 0, "time": 15},
				},
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "blind at index 0 is missing 'big' field",
		},
		{
			name:        "blind missing time field",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"blinds": []map[string]any{
					{"small": 25, "big": 50, "ante": 0},
				},
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "blind at index 0 is missing 'time' field",
		},
		{
			name:        "blind time out of range",
			userRole:    authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID: "1",
			requestBody: map[string]any{
				"blinds": []map[string]any{
					{"small": 25, "big": 50, "ante": 0, "time": 61},
				},
			},
			seedStructures:       true,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "blind at index 0: 'time' must be between 1 and 60",
		},
	}

	// Add tests for every authorized role
	authorizedRoles := []string{
		authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
		authorization.ROLE_SECRETARY.ToString(),
		authorization.ROLE_TREASURER.ToString(),
		authorization.ROLE_VICE_PRESIDENT.ToString(),
		authorization.ROLE_PRESIDENT.ToString(),
		authorization.ROLE_WEBMASTER.ToString(),
	}
	for _, role := range authorizedRoles {
		testCases = append(testCases, struct {
			name                 string
			userRole             string
			structureID          string
			requestBody          map[string]any
			seedStructures       bool
			expectedStatus       int
			expectError          bool
			expectedErrorMessage string
		}{
			name:        fmt.Sprintf("successful update with role %s", role),
			userRole:    role,
			structureID: "1",
			requestBody: map[string]any{
				"name": fmt.Sprintf("Updated by %s", role),
			},
			seedStructures: true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.seedStructures {
				require.NoError(t, testutils.SeedStructures(db))
			}

			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			req, err := testutils.MakeJSONRequest(
				"PATCH",
				fmt.Sprintf("/api/v2/structures/%s", tc.structureID),
				tc.requestBody,
			)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
			} else {
				require.Equal(t, tc.expectedStatus, w.Code)

				var response map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
				require.NotNil(t, response["id"])
				require.NotNil(t, response["name"])

				// Verify updates were applied
				if tc.requestBody != nil {
					if name, ok := tc.requestBody["name"]; ok && name != nil {
						require.Equal(t, name, response["name"])
					}
					if blinds, ok := tc.requestBody["blinds"]; ok && blinds != nil {
						responseBlinds := response["blinds"].([]any)
						requestBlinds := blinds.([]map[string]any)
						require.Equal(t, len(requestBlinds), len(responseBlinds))
					}
				}
			}
		})
	}
}

func TestDeleteStructure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	unauthorizedRoles := []string{"bot", "executive"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"DELETE",
		"/api/v2/structures/1",
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		structureID      string
		seedStructures   bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:           "successful delete",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID:    "1",
			seedStructures: true,
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:             "invalid structure ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID:      "invalid-id",
			seedStructures:   false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Structure ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "non-existent structure",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			structureID:      "999",
			seedStructures:   false,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "Structure not found",
		},
	}

	// Add tests for every authorized role
	authorizedRoles := []string{
		authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
		authorization.ROLE_SECRETARY.ToString(),
		authorization.ROLE_TREASURER.ToString(),
		authorization.ROLE_VICE_PRESIDENT.ToString(),
		authorization.ROLE_PRESIDENT.ToString(),
		authorization.ROLE_WEBMASTER.ToString(),
	}
	for _, role := range authorizedRoles {
		testCases = append(testCases, struct {
			name             string
			userRole         string
			structureID      string
			seedStructures   bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
		}{
			name:           fmt.Sprintf("successful delete with role %s", role),
			userRole:       role,
			structureID:    "1",
			seedStructures: true,
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.seedStructures {
				require.NoError(t, testutils.SeedStructures(db))
			}

			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			req, err := testutils.MakeJSONRequest(
				"DELETE",
				fmt.Sprintf("/api/v2/structures/%s", tc.structureID),
				nil,
			)
			require.NoError(t, err)
			testutils.SetAuthCookie(req, sessionID)

			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				require.Equal(t, tc.expectedStatus, w.Code)
				require.Empty(t, w.Body.String())
			}
		})
	}
}
