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

func TestCreateSemester(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	unauthorizedRoles := []string{"bot", "executive", "tournament_director", "secretary", "treasurer"}
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "POST", "/api/v2/semesters", unauthorizedRoles)

	testCases := []struct {
		name             string
		userRole         string
		requestBody      map[string]interface{}
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse map[string]interface{}
	}{
		{
			name:     "successful request",
			userRole: authorization.ROLE_PRESIDENT.ToString(),
			requestBody: map[string]interface{}{
				"name":                  "Fall 2024",
				"meta":                  "Test semester",
				"startDate":             "2024-08-31T20:00:00-04:00",
				"endDate":               "2024-12-14T19:00:00-05:00",
				"startingBudget":        100.0,
				"membershipFee":         10,
				"membershipDiscountFee": 5,
				"rebuyFee":              2,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			expectedResponse: map[string]interface{}{
				"name":                  "Fall 2024",
				"meta":                  "Test semester",
				"startDate":             "2024-08-31T20:00:00-04:00",
				"endDate":               "2024-12-14T19:00:00-05:00",
				"startingBudget":        float64(100.0),
				"currentBudget":         float64(100.0),
				"membershipFee":         float64(10),
				"membershipDiscountFee": float64(5),
				"rebuyFee":              float64(2),
				"id":                    nil, // Just check that ID exists
			},
		},
		{
			name:     "successful request with minimal data",
			userRole: authorization.ROLE_PRESIDENT.ToString(),
			requestBody: map[string]interface{}{
				"name":                  "Spring 2025",
				"startDate":             "2024-12-31T19:00:00-05:00",
				"endDate":               "2025-04-29T20:00:00-04:00",
				"membershipFee":         15,
				"membershipDiscountFee": 10,
				"rebuyFee":              3,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			expectedResponse: map[string]interface{}{
				"name":                  "Spring 2025",
				"meta":                  "",
				"startDate":             "2024-12-31T19:00:00-05:00",
				"endDate":               "2025-04-29T20:00:00-04:00",
				"startingBudget":        float64(0),
				"currentBudget":         float64(0),
				"membershipFee":         float64(15),
				"membershipDiscountFee": float64(10),
				"rebuyFee":              float64(3),
				"id":                    nil, // Just check that ID exists
			},
		},
		{
			name:     "invalid request missing name",
			userRole: authorization.ROLE_PRESIDENT.ToString(),
			requestBody: map[string]interface{}{
				"startDate":             "2024-09-01T00:00:00Z",
				"endDate":               "2024-12-15T00:00:00Z",
				"membershipFee":         10,
				"membershipDiscountFee": 5,
				"rebuyFee":              2,
			},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Key: 'CreateSemesterRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:     "invalid request missing start date",
			userRole: authorization.ROLE_PRESIDENT.ToString(),
			requestBody: map[string]interface{}{
				"name":                  "Fall 2024",
				"endDate":               "2024-12-15T00:00:00Z",
				"membershipFee":         10,
				"membershipDiscountFee": 5,
				"rebuyFee":              2,
			},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Key: 'CreateSemesterRequest.StartDate' Error:Field validation for 'StartDate' failed on the 'required' tag",
		},
		{
			name:     "invalid request missing end date",
			userRole: authorization.ROLE_PRESIDENT.ToString(),
			requestBody: map[string]interface{}{
				"name":                  "Fall 2024",
				"startDate":             "2024-09-01T00:00:00Z",
				"membershipFee":         10,
				"membershipDiscountFee": 5,
				"rebuyFee":              2,
			},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Key: 'CreateSemesterRequest.EndDate' Error:Field validation for 'EndDate' failed on the 'required' tag",
		},
		{
			name:     "invalid request end date before start date",
			userRole: authorization.ROLE_PRESIDENT.ToString(),
			requestBody: map[string]interface{}{
				"name":                  "Fall 2024",
				"startDate":             "2024-12-15T00:00:00Z",
				"endDate":               "2024-09-01T00:00:00Z",
				"membershipFee":         10,
				"membershipDiscountFee": 5,
				"rebuyFee":              2,
			},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Key: 'CreateSemesterRequest.EndDate' Error:Field validation for 'EndDate' failed on the 'gtfield' tag",
		},
		{
			name:     "invalid request negative starting budget",
			userRole: authorization.ROLE_PRESIDENT.ToString(),
			requestBody: map[string]interface{}{
				"name":                  "Fall 2024",
				"startDate":             "2024-09-01T00:00:00Z",
				"endDate":               "2024-12-15T00:00:00Z",
				"startingBudget":        -50.0,
				"membershipFee":         10,
				"membershipDiscountFee": 5,
				"rebuyFee":              2,
			},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Key: 'CreateSemesterRequest.StartingBudget' Error:Field validation for 'StartingBudget' failed on the 'gte' tag",
		},
		{
			name:             "invalid request empty body",
			userRole:         authorization.ROLE_PRESIDENT.ToString(),
			requestBody:      nil,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "EOF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("POST", "/api/v2/semesters", tc.requestBody)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				// Parse response to get the auto-generated ID
				var actualResponse map[string]interface{}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actualResponse))

				// Set the actual ID in the expected response
				tc.expectedResponse["id"] = actualResponse["id"]

				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}

func TestListSemesters(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	// All authenticated users can list semesters (ROLE_BOT and above have access)
	unauthorizedRoles := []string{}
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "GET", "/api/v2/semesters", unauthorizedRoles)

	testCases := []struct {
		name             string
		userRole         string
		setupSemesters   []models.Semester
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse []map[string]interface{}
	}{
		{
			name:             "successful request no semesters",
			userRole:         authorization.ROLE_BOT.ToString(),
			setupSemesters:   []models.Semester{}, // Empty slice, no semesters to create
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: []map[string]interface{}{},
		},
	}

	// Add tests for every role
	authorizedRoles := []string{
		authorization.ROLE_BOT.ToString(),
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
			setupSemesters   []models.Semester
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
			expectedResponse []map[string]interface{}
		}{
			name:             "successful request " + role,
			userRole:         role,
			setupSemesters:   testutils.TEST_SEMESTERS,
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: []map[string]interface{}{},
		})
	}

	// For test cases with semesters, we need to order them by start_date DESC
	// to match the API response ordering
	for i, tc := range testCases {
		if tc.expectedStatus == http.StatusOK && len(tc.setupSemesters) > 0 {
			// Create a copy of setupSemesters ordered by start_date DESC
			orderedSemesters := make([]models.Semester, len(tc.setupSemesters))
			copy(orderedSemesters, tc.setupSemesters)

			// Sort by start_date DESC (most recent first)
			// testutils.TEST_SEMESTERS should be: Fall 2024, Spring 2024, Fall 2023
			expectedOrder := []models.Semester{
				tc.setupSemesters[2], // Fall 2024 (2024-09-01)
				tc.setupSemesters[1], // Spring 2024 (2024-01-15)
				tc.setupSemesters[0], // Fall 2023 (2023-09-01)
			}

			// Convert to []map[string]interface{} format
			b, err := json.Marshal(expectedOrder)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(b, &testCases[i].expectedResponse))
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed semesters if not empty
			if len(tc.setupSemesters) > 0 {
				testutils.SeedSemesters(db)
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("GET", "/api/v2/semesters", nil)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				// Use the updated AssertSuccessResponse which can handle arrays
				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}

func TestGetSemester(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	// Only roles below ROLE_EXECUTIVE are unauthorized (based on semester_authorizer.go: get requires ROLE_EXECUTIVE)
	unauthorizedRoles := []string{"bot"}
	testutils.TestInvalidAuthForEndpoint(t, container, apiServer, "GET", fmt.Sprintf("/api/v2/semesters/%s", "550e8400-e29b-41d4-a716-446655440000"), unauthorizedRoles)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		setupSemesters   bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse map[string]interface{}
	}{
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       "invalid-uuid",
			setupSemesters:   false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Invalid UUID for semester ID",
		},
		{
			name:             "non-existent semester",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       "550e8400-e29b-41d4-a716-446655440000",
			setupSemesters:   false,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "record not found",
		},
	}

	// Add tests for every role
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
			semesterID       string
			setupSemesters   bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
			expectedResponse map[string]interface{}
		}{
			name:             fmt.Sprintf("successful request with valid ID %s", role),
			userRole:         role,
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(), // Fall 2023
			setupSemesters:   true,
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.setupSemesters {
				// Seed test semesters
				testutils.SeedSemesters(db)

				// If we expect a semester to exist, find it and set the expected response accordingly
				if tc.semesterID != "" && tc.expectedResponse == nil {
					semester, err := testutils.FindSemesterByID(tc.semesterID)
					require.NoError(t, err)
					tc.expectedResponse = map[string]interface{}{
						"id":                    semester.ID.String(),
						"name":                  semester.Name,
						"meta":                  semester.Meta,
						"startDate":             semester.StartDate,
						"endDate":               semester.EndDate,
						"startingBudget":        semester.StartingBudget,
						"currentBudget":         semester.CurrentBudget,
						"membershipFee":         semester.MembershipFee,
						"membershipDiscountFee": semester.MembershipDiscountFee,
						"rebuyFee":              semester.RebuyFee,
					}
				}
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest("GET", fmt.Sprintf("/api/v2/semesters/%s", tc.semesterID), nil)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}
