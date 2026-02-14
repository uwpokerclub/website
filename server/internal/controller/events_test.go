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
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateEvent(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Seed test date
	semester, err := testutils.CreateTestSemester(db, "Fall 2025")
	require.NoError(t, err)
	structure, err := testutils.CreateTestStructure(db, "Test Structure")
	require.NoError(t, err)

	// Run default tests for authentication and authorization
	unauthorizedRoles := []string{"bot", "executive"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"POST",
		fmt.Sprintf("/api/v2/semesters/%s/events", semester.ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name                 string
		userRole             string
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
		expectedResponse     map[string]any
		useSemesterID        string // Optional: override semester ID in URL path
	}{
		{
			name:     "successful request with all fields",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event 1",
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"semesterId":       semester.ID.String(),
				"startDate":        "2025-09-01T19:00:00Z",
				"structureId":      structure.ID,
				"pointsMultiplier": 1.0,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			expectedResponse: map[string]any{
				"name":             "Event 1",
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"semesterId":       semester.ID.String(),
				"startDate":        "2025-09-01T19:00:00Z",
				"structureId":      float64(structure.ID),
				"state":            float64(models.EventStateStarted),
				"rebuys":           float64(0),
				"pointsMultiplier": 1.0,
			},
		},
		{
			name:     "successful request with minimal data",
			userRole: authorization.ROLE_SECRETARY.ToString(),
			requestBody: map[string]any{
				"name":             "Minimal Event",
				"format":           "Texas Hold'em",
				"semesterId":       semester.ID.String(),
				"startDate":        "2025-09-15T20:00:00Z",
				"structureId":      structure.ID,
				"pointsMultiplier": 1.5,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			expectedResponse: map[string]any{
				"name":             "Minimal Event",
				"format":           "Texas Hold'em",
				"notes":            "",
				"semesterId":       semester.ID.String(),
				"startDate":        "2025-09-15T20:00:00Z",
				"structureId":      float64(structure.ID),
				"state":            float64(models.EventStateStarted),
				"rebuys":           float64(0),
				"pointsMultiplier": 1.5,
			},
		},
		{
			name:     "missing required name field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"semesterId":       semester.ID.String(),
				"startDate":        "2025-09-01T19:00:00Z",
				"structureId":      structure.ID,
				"pointsMultiplier": 1.0,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateEventRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:     "missing required format field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event 1",
				"notes":            "This is a test event",
				"semesterId":       semester.ID.String(),
				"startDate":        "2025-09-01T19:00:00Z",
				"structureId":      structure.ID,
				"pointsMultiplier": 1.0,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateEventRequest.Format' Error:Field validation for 'Format' failed on the 'required' tag",
		},
		{
			name:     "missing required semesterId field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event 1",
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"startDate":        "2025-09-01T19:00:00Z",
				"structureId":      structure.ID,
				"pointsMultiplier": 1.0,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateEventRequest.SemesterID' Error:Field validation for 'SemesterID' failed on the 'required' tag",
		},
		{
			name:     "missing required startDate field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event 1",
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"semesterId":       semester.ID.String(),
				"structureId":      structure.ID,
				"pointsMultiplier": 1.0,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateEventRequest.StartDate' Error:Field validation for 'StartDate' failed on the 'required' tag",
		},
		{
			name:     "missing required structureId field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event 1",
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"semesterId":       semester.ID.String(),
				"startDate":        "2025-09-01T19:00:00Z",
				"pointsMultiplier": 1.0,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateEventRequest.StructureID' Error:Field validation for 'StructureID' failed on the 'required' tag",
		},
		{
			name:     "missing required pointsMultiplier field",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":        "Event 1",
				"format":      "No Limit Hold'em",
				"notes":       "This is a test event",
				"semesterId":  semester.ID.String(),
				"startDate":   "2025-09-01T19:00:00Z",
				"structureId": structure.ID,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateEventRequest.PointsMultiplier' Error:Field validation for 'PointsMultiplier' failed on the 'required' tag",
		},
		{
			name:                 "invalid request empty body",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody:          nil,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Key: 'CreateEventRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'CreateEventRequest.Format' Error:Field validation for 'Format' failed on the 'required' tag\nKey: 'CreateEventRequest.SemesterID' Error:Field validation for 'SemesterID' failed on the 'required' tag\nKey: 'CreateEventRequest.StartDate' Error:Field validation for 'StartDate' failed on the 'required' tag\nKey: 'CreateEventRequest.StructureID' Error:Field validation for 'StructureID' failed on the 'required' tag\nKey: 'CreateEventRequest.PointsMultiplier' Error:Field validation for 'PointsMultiplier' failed on the 'required' tag",
		},
		{
			name:     "invalid start date format",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event 1",
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"semesterId":       semester.ID.String(),
				"startDate":        "invalid-date",
				"structureId":      structure.ID,
				"pointsMultiplier": 1.0,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "parsing time \"invalid-date\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid-date\" as \"2006\"",
		},
		{
			name:     "invalid semester ID in path",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event 1",
				"format":           "No Limit Hold'em",
				"notes":            "This is a test event",
				"semesterId":       "00000000-0000-0000-0000-000000000000",
				"startDate":        "2025-09-01T19:00:00Z",
				"structureId":      structure.ID,
				"pointsMultiplier": 1.0,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Semester ID 'invalid-uuid' is not a valid UUID",
			useSemesterID:        "invalid-uuid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Re-seed test data
			semester, err := testutils.CreateTestSemester(db, "Fall 2025")
			require.NoError(t, err)
			structure, err := testutils.CreateTestStructure(db, "Test Structure")
			require.NoError(t, err)

			// Update semesterId and structureId in request body if they exist
			if tc.requestBody != nil {
				if _, exists := tc.requestBody["semesterId"]; exists {
					tc.requestBody["semesterId"] = semester.ID.String()
				}
				if _, exists := tc.requestBody["structureId"]; exists {
					tc.requestBody["structureId"] = structure.ID
				}
			}

			// Update expected response with actual semester ID for successful cases
			if !tc.expectError && tc.expectedResponse != nil {
				tc.expectedResponse["semesterId"] = semester.ID.String()
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Determine which semester ID to use in the URL path
			urlSemesterID := semester.ID.String()
			if tc.useSemesterID != "" {
				urlSemesterID = tc.useSemesterID
			}

			// Create request
			req, err := testutils.MakeJSONRequest(
				"POST",
				fmt.Sprintf("/api/v2/semesters/%s/events", urlSemesterID),
				tc.requestBody,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
			} else {
				// Parse response to get the auto-generated ID and actual startDate
				var actualResponse map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actualResponse))

				// Set the actual ID and startDate in the expected response (to handle timezone conversion)
				tc.expectedResponse["id"] = actualResponse["id"]
				tc.expectedResponse["startDate"] = actualResponse["startDate"]

				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}

func TestListEvents(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
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
		"/api/v2/semesters/8d5cad26-5c20-4ed1-963b-eb35250ab555/events",
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		expectedEvents   []models.Event
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse map[string]any
	}{
		{
			name:           "successful request no events",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[2].ID.String(),
			expectedEvents: []models.Event{},
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectedResponse: map[string]any{
				"data":  []map[string]any{},
				"total": float64(0),
			},
		},
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       "invalid-uuid",
			expectedEvents:   []models.Event{},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
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
			semesterID       string
			expectedEvents   []models.Event
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
			expectedResponse map[string]any
		}{
			name:       fmt.Sprintf("successful request with events %s", role),
			userRole:   role,
			semesterID: testutils.TEST_SEMESTERS[0].ID.String(),
			expectedEvents: []models.Event{
				testutils.TEST_EVENTS[1],
				testutils.TEST_EVENTS[2],
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		})
	}

	// Preprocess expected responses to convert Event models to ListResponse shape
	for i, tc := range testCases {
		if !tc.expectError && len(tc.expectedEvents) > 0 {
			var eventsData []map[string]any
			b, err := json.Marshal(tc.expectedEvents)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(b, &eventsData))
			testCases[i].expectedResponse = map[string]any{
				"data":  eventsData,
				"total": float64(len(tc.expectedEvents)),
			}
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			testutils.SeedEvents(db, true)

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"GET",
				fmt.Sprintf("/api/v2/semesters/%s/events", tc.semesterID),
				nil,
			)
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

func TestUpdateEvent(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
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
		fmt.Sprintf("/api/v2/semesters/%s/events/1", testutils.TEST_SEMESTERS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name                 string
		userRole             string
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
		expectedResponse     map[string]any
		setupEvent           bool
		useSemesterID        string
		useEventID           string
	}{
		{
			name:     "successful update with all fields",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":      "Updated Event Name",
				"format":    "Updated Format",
				"notes":     "Updated notes",
				"startDate": "2025-09-15T20:00:00Z",
			},
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
			setupEvent:       true,
			useEventID:       "1",
		},
		{
			name:     "successful update with partial fields",
			userRole: authorization.ROLE_SECRETARY.ToString(),
			requestBody: map[string]any{
				"name":   "Partially Updated Event",
				"format": "New Format",
			},
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
			setupEvent:       true,
			useEventID:       "1",
		},
		{
			name:     "successful update with notes set to null",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":  "Event with null notes",
				"notes": nil,
			},
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
			setupEvent:       true,
			useEventID:       "1",
		},
		{
			name:     "name cannot be null",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":   nil,
				"format": "Some Format",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: name cannot be null",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "name cannot be empty string",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":   "",
				"format": "Some Format",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: name must be a non-empty string",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "format cannot be null",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":   "Valid Name",
				"format": nil,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: format cannot be null",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "format cannot be empty string",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":   "Valid Name",
				"format": "",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: format must be a non-empty string",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "startDate cannot be null",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":      "Valid Name",
				"startDate": nil,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: startDate cannot be null",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "invalid startDate format",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":      "Valid Name",
				"startDate": "invalid-date-format",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: startDate must be in RFC3339 format",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "unknown field in request",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":         "Valid Name",
				"unknownField": "some value",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: failed to validate event update request: unknown field: unknownField",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:           "empty body - no fields updated",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody:    nil,
			expectedStatus: http.StatusOK,
			expectError:    false,
			setupEvent:     true,
			useEventID:     "1",
		},
		{
			name:     "invalid semester ID format",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name": "Valid Name",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Semester ID 'invalid-uuid' is not a valid UUID",
			setupEvent:           true,
			useSemesterID:        "invalid-uuid",
			useEventID:           "1",
		},
		{
			name:     "invalid event ID format",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name": "Valid Name",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Event ID 'invalid-id' is not a valid integer",
			setupEvent:           true,
			useEventID:           "invalid-id",
		},
		{
			name:     "non-existent event",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name": "Valid Name",
			},
			expectedStatus:       http.StatusNotFound,
			expectError:          true,
			expectedErrorMessage: "Event '999' not found for semester '550e8400-e29b-41d4-a716-446655440000'",
			setupEvent:           true,
			useSemesterID:        "550e8400-e29b-41d4-a716-446655440000",
			useEventID:           "999",
		},
		{
			name:     "successful update with pointsMultiplier",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event with updated multiplier",
				"pointsMultiplier": 2.5,
			},
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
			setupEvent:       true,
			useEventID:       "1",
		},
		{
			name:     "successful update with pointsMultiplier zero",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Event with zero multiplier",
				"pointsMultiplier": 0.0,
			},
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
			setupEvent:       true,
			useEventID:       "1",
		},
		{
			name:     "pointsMultiplier cannot be null",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Valid Name",
				"pointsMultiplier": nil,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: pointsMultiplier cannot be null",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "pointsMultiplier cannot be negative",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Valid Name",
				"pointsMultiplier": -1.5,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: pointsMultiplier must be a non-negative number",
			setupEvent:           true,
			useEventID:           "1",
		},
		{
			name:     "pointsMultiplier invalid type",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: map[string]any{
				"name":             "Valid Name",
				"pointsMultiplier": "invalid-string",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Error converting request to update map: pointsMultiplier must be a number",
			setupEvent:           true,
			useEventID:           "1",
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
			expectedResponse     map[string]any
			setupEvent           bool
			useSemesterID        string
			useEventID           string
		}{
			name:     fmt.Sprintf("successful update with role %s", role),
			userRole: role,
			requestBody: map[string]any{
				"name":   fmt.Sprintf("Updated by %s", role),
				"format": "Updated Format",
			},
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
			setupEvent:       true,
			useEventID:       "1",
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			var eventID string = "1"
			if tc.setupEvent {
				// Seed test events
				testutils.SeedEvents(db, true)

				// Build expected response dynamically for successful cases
				if !tc.expectError && tc.expectedResponse == nil {
					parsedEventID, err := strconv.ParseInt(eventID, 10, 32)
					require.NoError(t, err)

					// Find the original event
					originalEvent, err := testutils.FindEventById(int32(parsedEventID))
					require.NoError(t, err)

					// Get nested associations
					semesterMap, err := testutils.FindSemesterByIDAsMap(
						originalEvent.SemesterID.String(),
					)
					require.NoError(t, err)

					structureMap, err := testutils.FindStructureByIDAsMap(originalEvent.StructureID)
					require.NoError(t, err)

					// Build base expected response from original event
					tc.expectedResponse = map[string]any{
						"id":         float64(originalEvent.ID),
						"name":       originalEvent.Name,
						"format":     originalEvent.Format,
						"notes":      originalEvent.Notes,
						"semesterId": originalEvent.SemesterID.String(),
						"semester":   semesterMap,
						"startDate": originalEvent.StartDate.Format(
							"2006-01-02T15:04:05Z07:00",
						),
						"state":            float64(originalEvent.State),
						"rebuys":           float64(originalEvent.Rebuys),
						"pointsMultiplier": originalEvent.PointsMultiplier,
						"structureId":      float64(originalEvent.StructureID),
						"structure":        structureMap,
					}

					// Apply updates from request body
					if tc.requestBody != nil {
						if name, ok := tc.requestBody["name"]; ok && name != nil {
							tc.expectedResponse["name"] = name
						}
						if format, ok := tc.requestBody["format"]; ok && format != nil {
							tc.expectedResponse["format"] = format
						}
						if notes, ok := tc.requestBody["notes"]; ok {
							if notes == nil {
								tc.expectedResponse["notes"] = ""
							} else {
								tc.expectedResponse["notes"] = notes
							}
						}
						if startDate, ok := tc.requestBody["startDate"]; ok && startDate != nil {
							tc.expectedResponse["startDate"] = startDate
						}
						if pointsMultiplier, ok := tc.requestBody["pointsMultiplier"]; ok && pointsMultiplier != nil {
							tc.expectedResponse["pointsMultiplier"] = pointsMultiplier
						}
					}
				}
			}

			// Override eventID if specified in test case
			if tc.useEventID != "" {
				eventID = tc.useEventID
			}

			// Determine which semester ID to use in the URL path
			urlSemesterID := testutils.TEST_SEMESTERS[0].ID.String()
			if tc.useSemesterID != "" {
				urlSemesterID = tc.useSemesterID
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"PATCH",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s", urlSemesterID, eventID),
				tc.requestBody,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
			} else {
				// Parse response to handle timezone conversion
				var actualResponse map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actualResponse))

				// Set the actual startDate in the expected response (to handle timezone conversion)
				if tc.expectedResponse != nil {
					tc.expectedResponse["startDate"] = actualResponse["startDate"]
				}

				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}

func TestGetEvent(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	// Only bot role is unauthorized for event.get (requires ROLE_EXECUTIVE)
	unauthorizedRoles := []string{"bot"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"GET",
		"/api/v2/semesters/550e8400-e29b-41d4-a716-446655440000/events/101",
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		setupEvent       bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse map[string]any
	}{
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       "invalid-uuid",
			eventID:          "1",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       "550e8400-e29b-41d4-a716-446655440000",
			eventID:          "invalid-id",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "non-existent semester",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       "550e8400-e29b-41d4-a716-446655440000",
			eventID:          "1",
			setupEvent:       false,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "Event '1' not found for semester '550e8400-e29b-41d4-a716-446655440000'",
		},
		{
			name:             "non-existent event ID",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       "550e8400-e29b-41d4-a716-446655440000",
			eventID:          "999",
			setupEvent:       false,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "Event '999' not found for semester '550e8400-e29b-41d4-a716-446655440000'",
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
			semesterID       string
			eventID          string
			setupEvent       bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
			expectedResponse map[string]any
		}{
			name:             fmt.Sprintf("successful request with valid IDs %s", role),
			userRole:         role,
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          fmt.Sprintf("%d", testutils.TEST_EVENTS[2].ID),
			setupEvent:       true,
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.setupEvent {
				testutils.SeedEvents(db, true)

				// if we expect an event to exist, find it and set the expected response
				if !tc.expectError {
					eventID, err := strconv.ParseInt(tc.eventID, 10, 32)
					require.NoError(t, err)

					event, err := testutils.FindEventById(int32(eventID))
					require.NoError(t, err)

					// Get nested associations
					semesterMap, err := testutils.FindSemesterByIDAsMap(event.SemesterID.String())
					require.NoError(t, err)

					structureMap, err := testutils.FindStructureByIDAsMap(event.StructureID)
					require.NoError(t, err)

					tc.expectedResponse = map[string]any{
						"id":               event.ID,
						"name":             event.Name,
						"format":           event.Format,
						"notes":            event.Notes,
						"semesterId":       event.SemesterID.String(),
						"semester":         semesterMap,
						"startDate":        event.StartDate.Format("2006-01-02T15:04:05Z07:00"),
						"state":            float64(event.State),
						"rebuys":           float64(event.Rebuys),
						"pointsMultiplier": event.PointsMultiplier,
						"structureId":      float64(event.StructureID),
						"structure":        structureMap,
					}
				}
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"GET",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s", tc.semesterID, tc.eventID),
				nil,
			)
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

func TestEndEvent(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	// Only bot, executive, and tournament_director roles are unauthorized for event.end (requires ROLE_SECRETARY or higher)
	unauthorizedRoles := []string{"bot", "executive", "tournament_director"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"POST",
		fmt.Sprintf("/api/v2/semesters/%s/events/1/end", testutils.TEST_SEMESTERS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		setupEvent       bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       "invalid-uuid",
			eventID:          "1",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       "550e8400-e29b-41d4-a716-446655440000",
			eventID:          "invalid-id",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "non-existent event ID",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "999",
			setupEvent:       true,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "record not found",
		},
		{
			name:             "event already ended",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "1",
			setupEvent:       true,
			expectedStatus:   http.StatusForbidden,
			expectError:      true,
			expectedErrorMsg: "This event has already ended, it cannot be ended again.",
		},
	}

	// Add tests for every authorized role (ROLE_SECRETARY and higher)
	authorizedRoles := []string{
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
			eventID          string
			setupEvent       bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
		}{
			name:           fmt.Sprintf("successful end event with role %s", role),
			userRole:       role,
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			setupEvent:     true,
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.setupEvent {
				testutils.SeedEvents(db, true)
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"POST",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/end", tc.semesterID, tc.eventID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				require.Equal(t, tc.expectedStatus, w.Code)
				require.Empty(t, w.Body.String())
			}
		})
	}
}

func TestRestartEvent(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	// Only bot, executive, and tournament_director roles are unauthorized for event.restart (requires ROLE_SECRETARY or higher)
	unauthorizedRoles := []string{"bot", "executive", "tournament_director"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"POST",
		fmt.Sprintf("/api/v2/semesters/%s/events/1/restart", testutils.TEST_SEMESTERS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		setupEvent       bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       "invalid-uuid",
			eventID:          "1",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       "550e8400-e29b-41d4-a716-446655440000",
			eventID:          "invalid-id",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "non-existent event ID",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "999",
			setupEvent:       true,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "record not found",
		},
		{
			name:             "event not ended - cannot restart",
			userRole:         authorization.ROLE_SECRETARY.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "2",
			setupEvent:       true,
			expectedStatus:   http.StatusForbidden,
			expectError:      true,
			expectedErrorMsg: "This event has not been ended",
		},
	}

	// Add tests for every authorized role (ROLE_SECRETARY and higher)
	authorizedRoles := []string{
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
			eventID          string
			setupEvent       bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
		}{
			name:           fmt.Sprintf("successful restart event with role %s", role),
			userRole:       role,
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "1",
			setupEvent:     true,
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.setupEvent {
				testutils.SeedEvents(db, true)
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"POST",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/restart", tc.semesterID, tc.eventID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				require.Equal(t, tc.expectedStatus, w.Code)
				require.Empty(t, w.Body.String())
			}
		})
	}
}

func TestRebuyEvent(t *testing.T) {
	t.Parallel()

	// Setup test database and API server once
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Run default tests for authentication and authorization
	// Only bot and executive roles are unauthorized for event.rebuy (requires ROLE_TOURNAMENT_DIRECTOR or higher)
	unauthorizedRoles := []string{"bot", "executive"}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"POST",
		fmt.Sprintf("/api/v2/semesters/%s/events/1/rebuy", testutils.TEST_SEMESTERS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		setupEvent       bool
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       "invalid-uuid",
			eventID:          "1",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       "550e8400-e29b-41d4-a716-446655440000",
			eventID:          "invalid-id",
			setupEvent:       false,
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "non-existent event ID",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "999",
			setupEvent:       true,
			expectedStatus:   http.StatusNotFound,
			expectError:      true,
			expectedErrorMsg: "record not found",
		},
		{
			name:             "event already ended - cannot rebuy",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "1",
			setupEvent:       true,
			expectedStatus:   http.StatusForbidden,
			expectError:      true,
			expectedErrorMsg: "This event has already ended, it cannot be ended again.",
		},
	}

	// Add tests for every authorized role (ROLE_TOURNAMENT_DIRECTOR and higher)
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
			semesterID       string
			eventID          string
			setupEvent       bool
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
		}{
			name:           fmt.Sprintf("successful rebuy event with role %s", role),
			userRole:       role,
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			setupEvent:     true,
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))

			if tc.setupEvent {
				testutils.SeedEvents(db, true)
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"POST",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/rebuy", tc.semesterID, tc.eventID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert response
			if tc.expectError {
				testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMsg)
			} else {
				require.Equal(t, tc.expectedStatus, w.Code)
				require.Empty(t, w.Body.String())
			}
		})
	}
}
