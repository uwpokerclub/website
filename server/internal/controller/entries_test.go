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

func TestCreateEntry(t *testing.T) {
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
		"POST",
		fmt.Sprintf("/api/v2/semesters/%s/events/1/entries", testutils.TEST_SEMESTERS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name                 string
		userRole             string
		requestBody          any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
		expectedResponse     []map[string]any
		useSemesterID        string
		useEventID           string
	}{
		{
			name:     "successful request single entry",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: []string{
				testutils.TEST_MEMBERSHIPS[1].ID.String(),
			},
			expectedStatus: http.StatusMultiStatus,
			expectError:    false,
			expectedResponse: []map[string]any{
				{
					"membershipId": testutils.TEST_MEMBERSHIPS[1].ID.String(),
					"status":       "created",
					"participant": map[string]any{
						"membershipId": testutils.TEST_MEMBERSHIPS[1].ID.String(),
						"eventId":      float64(2),
						"placement":    float64(0),
						"signedOutAt":  nil,
					},
				},
			},
		},
		{
			name:     "partial success - one duplicate",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: []string{
				testutils.TEST_MEMBERSHIPS[0].ID.String(), // Already in event 2
				testutils.TEST_MEMBERSHIPS[1].ID.String(),
			},
			expectedStatus: http.StatusMultiStatus,
			expectError:    false,
			// expectedResponse will be validated separately for mixed results
		},
		{
			name:                 "empty array",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody:          []string{},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Array of membership IDs cannot be empty",
		},
		{
			name:     "invalid UUID in array",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: []string{
				"invalid-uuid",
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Invalid UUID format: 'invalid-uuid'",
		},
		{
			name:                 "invalid request - null body",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody:          nil,
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "EOF",
		},
		{
			name:     "invalid semester ID in path",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: []string{
				testutils.TEST_MEMBERSHIPS[0].ID.String(),
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Semester ID 'invalid-uuid' is not a valid UUID",
			useSemesterID:        "invalid-uuid",
		},
		{
			name:     "invalid event ID in path",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: []string{
				testutils.TEST_MEMBERSHIPS[0].ID.String(),
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Event ID 'invalid-id' is not a valid integer",
			useEventID:           "invalid-id",
		},
		{
			name:     "event already ended - cannot create entry",
			userRole: authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			requestBody: []string{
				testutils.TEST_MEMBERSHIPS[1].ID.String(),
			},
			expectedStatus: http.StatusMultiStatus,
			expectError:    false,
			useEventID:     "1",
			// Will validate error in results
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
			requestBody          any
			expectedStatus       int
			expectError          bool
			expectedErrorMessage string
			expectedResponse     []map[string]any
			useSemesterID        string
			useEventID           string
		}{
			name:     fmt.Sprintf("successful request with role %s", role),
			userRole: role,
			requestBody: []string{
				testutils.TEST_MEMBERSHIPS[1].ID.String(),
			},
			expectedStatus: http.StatusMultiStatus,
			expectError:    false,
			expectedResponse: []map[string]any{
				{
					"membershipId": testutils.TEST_MEMBERSHIPS[1].ID.String(),
					"status":       "created",
					"participant": map[string]any{
						"membershipId": testutils.TEST_MEMBERSHIPS[1].ID.String(),
						"eventId":      float64(2),
						"placement":    float64(0),
						"signedOutAt":  nil,
					},
				},
			},
		})
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

			// Determine which semester ID and event ID to use in the URL path
			urlSemesterID := testutils.TEST_SEMESTERS[0].ID.String()
			if tc.useSemesterID != "" {
				urlSemesterID = tc.useSemesterID
			}

			urlEventID := "2"
			if tc.useEventID != "" {
				urlEventID = tc.useEventID
			}

			// Create request
			req, err := testutils.MakeJSONRequest(
				"POST",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/entries", urlSemesterID, urlEventID),
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
				// Parse response to get the results
				var actualResponse []map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actualResponse))

				// Verify status code
				require.Equal(t, tc.expectedStatus, w.Code)

				// Special handling for specific test cases
				switch tc.name {
				case "partial success - one duplicate":
					// Validate mixed results
					require.Len(t, actualResponse, 2)
					// First should be error (duplicate)
					require.Equal(t, testutils.TEST_MEMBERSHIPS[0].ID.String(), actualResponse[0]["membershipId"])
					require.Equal(t, "error", actualResponse[0]["status"])
					require.NotEmpty(t, actualResponse[0]["error"])
					// Second should be success
					require.Equal(t, testutils.TEST_MEMBERSHIPS[1].ID.String(), actualResponse[1]["membershipId"])
					require.Equal(t, "created", actualResponse[1]["status"])
					require.NotNil(t, actualResponse[1]["participant"])

				case "event already ended - cannot create entry":
					// All should be errors
					require.Len(t, actualResponse, 1)
					require.Equal(t, "error", actualResponse[0]["status"])
					require.Contains(t, actualResponse[0]["error"], "Modification of a completed event is forbidden")

				default:
					// For standard success cases, set the auto-generated IDs
					require.Len(t, actualResponse, len(tc.expectedResponse))
					for i := range actualResponse {
						if participant, ok := actualResponse[i]["participant"].(map[string]any); ok {
							if expectedParticipant, ok := tc.expectedResponse[i]["participant"].(map[string]any); ok {
								expectedParticipant["id"] = participant["id"]
							}
						}
					}
					testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
				}
			}
		})
	}
}

func TestListEntries(t *testing.T) {
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
		fmt.Sprintf("/api/v2/semesters/%s/events/1/entries", testutils.TEST_SEMESTERS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse map[string]any
	}{
		{
			name:           "successful request no entries",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "3",
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
			eventID:          "1",
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "invalid-id",
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
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
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
			expectedResponse map[string]any
		}{
			name:             fmt.Sprintf("successful request with role %s", role),
			userRole:         role,
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "2",
			expectedStatus:   http.StatusOK,
			expectError:      false,
			expectedResponse: nil, // Will be set dynamically
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data
			require.NoError(t, testutils.SeedAll(db))

			// Build expected response for tests that expect entries
			if !tc.expectError && tc.expectedResponse == nil {
				// For event 2, we expect 2 participants with nested membership and user
				user0, err := testutils.FindUserByID(testutils.TEST_USERS[0].ID)
				require.NoError(t, err)
				user2, err := testutils.FindUserByID(testutils.TEST_USERS[2].ID)
				require.NoError(t, err)

				tc.expectedResponse = map[string]any{
					"total": float64(2),
					"data": []map[string]any{
					{
						// Participant fields
						"membershipId": testutils.TEST_MEMBERSHIPS[2].ID.String(),
						"eventId":      float64(2),
						"placement":    float64(0),
						"signedOutAt":  nil,
						// Nested membership with nested user
						"membership": map[string]any{
							"id":         testutils.TEST_MEMBERSHIPS[2].ID.String(),
							"userId":     float64(testutils.TEST_USERS[2].ID),
							"semesterId": testutils.TEST_SEMESTERS[0].ID.String(),
							"paid":       testutils.TEST_MEMBERSHIPS[2].Paid,
							"discounted": testutils.TEST_MEMBERSHIPS[2].Discounted,
							"user": map[string]any{
								"id":        float64(user2.ID),
								"firstName": user2.FirstName,
								"lastName":  user2.LastName,
								"email":     user2.Email,
								"faculty":   user2.Faculty,
								"questId":   user2.QuestID,
								"createdAt": user2.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
							},
							"semester": map[string]any{
								"id":                    testutils.TEST_SEMESTERS[0].ID.String(),
								"name":                  testutils.TEST_SEMESTERS[0].Name,
								"meta":                  testutils.TEST_SEMESTERS[0].Meta,
								"startDate":             testutils.TEST_SEMESTERS[0].StartDate.Format("2006-01-02T15:04:05.999999999Z07:00"),
								"endDate":               testutils.TEST_SEMESTERS[0].EndDate.Format("2006-01-02T15:04:05.999999999Z07:00"),
								"startingBudget":        float64(testutils.TEST_SEMESTERS[0].StartingBudget),
								"currentBudget":         float64(testutils.TEST_SEMESTERS[0].CurrentBudget),
								"membershipFee":         float64(testutils.TEST_SEMESTERS[0].MembershipFee),
								"membershipDiscountFee": float64(testutils.TEST_SEMESTERS[0].MembershipDiscountFee),
								"rebuyFee":              float64(testutils.TEST_SEMESTERS[0].RebuyFee),
							},
							"ranking": map[string]any{
								"id":           float64(3),
								"membershipId": testutils.TEST_MEMBERSHIPS[2].ID.String(),
								"points":       float64(0),
								"attendance":   float64(0),
							},
						},
					},
					{
						// Participant fields
						"membershipId": testutils.TEST_MEMBERSHIPS[0].ID.String(),
						"eventId":      float64(2),
						"placement":    float64(0),
						"signedOutAt":  "2023-10-20T20:00:00-04:00",
						// Nested membership with nested user
						"membership": map[string]any{
							"id":         testutils.TEST_MEMBERSHIPS[0].ID.String(),
							"userId":     float64(testutils.TEST_USERS[0].ID),
							"semesterId": testutils.TEST_SEMESTERS[0].ID.String(),
							"paid":       testutils.TEST_MEMBERSHIPS[0].Paid,
							"discounted": testutils.TEST_MEMBERSHIPS[0].Discounted,
							"user": map[string]any{
								"id":        float64(user0.ID),
								"firstName": user0.FirstName,
								"lastName":  user0.LastName,
								"email":     user0.Email,
								"faculty":   user0.Faculty,
								"questId":   user0.QuestID,
								"createdAt": user0.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
							},
							"semester": map[string]any{
								"id":                    testutils.TEST_SEMESTERS[0].ID.String(),
								"name":                  testutils.TEST_SEMESTERS[0].Name,
								"meta":                  testutils.TEST_SEMESTERS[0].Meta,
								"startDate":             testutils.TEST_SEMESTERS[0].StartDate.Format("2006-01-02T15:04:05.999999999Z07:00"),
								"endDate":               testutils.TEST_SEMESTERS[0].EndDate.Format("2006-01-02T15:04:05.999999999Z07:00"),
								"startingBudget":        float64(testutils.TEST_SEMESTERS[0].StartingBudget),
								"currentBudget":         float64(testutils.TEST_SEMESTERS[0].CurrentBudget),
								"membershipFee":         float64(testutils.TEST_SEMESTERS[0].MembershipFee),
								"membershipDiscountFee": float64(testutils.TEST_SEMESTERS[0].MembershipDiscountFee),
								"rebuyFee":              float64(testutils.TEST_SEMESTERS[0].RebuyFee),
							},
							"ranking": map[string]any{
								"id":           float64(1),
								"membershipId": testutils.TEST_MEMBERSHIPS[0].ID.String(),
								"points":       float64(0),
								"attendance":   float64(0),
							},
						},
					},
				},
				}
			}

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"GET",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/entries", tc.semesterID, tc.eventID),
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
				// Parse response to get the actual participant IDs (auto-generated)
				var fullResponse map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &fullResponse))

				actualData, ok := fullResponse["data"].([]any)
				require.True(t, ok)

				// Build expected data from tc.expectedResponse
				expectedData, _ := tc.expectedResponse["data"].([]map[string]any)

				// Set the auto-generated fields in expected response
				for i := range actualData {
					actual, ok := actualData[i].(map[string]any)
					require.True(t, ok)

					if i < len(expectedData) {
						// Set auto-generated participant ID
						expectedData[i]["id"] = actual["id"]

						// Set actual createdAt timestamp for user
						if membership, ok := actual["membership"].(map[string]any); ok {
							if user, ok := membership["user"].(map[string]any); ok {
								if expectedMembership, ok := expectedData[i]["membership"].(map[string]any); ok {
									if expectedUser, ok := expectedMembership["user"].(map[string]any); ok {
										expectedUser["createdAt"] = user["createdAt"]
									}
								}
							}

							// Set actual ranking membershipId (may be zero-value UUID)
							if ranking, ok := membership["ranking"].(map[string]any); ok {
								if expectedMembership, ok := expectedData[i]["membership"].(map[string]any); ok {
									if expectedRanking, ok := expectedMembership["ranking"].(map[string]any); ok {
										expectedRanking["membershipId"] = ranking["membershipId"]
									}
								}
							}
						}
					}
				}

				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}

func TestSignOutEntry(t *testing.T) {
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
		"POST",
		fmt.Sprintf("/api/v2/semesters/%s/events/2/entries/%s/sign-out", testutils.TEST_SEMESTERS[0].ID, testutils.TEST_MEMBERSHIPS[2].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		entryID          string
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse map[string]any
	}{
		{
			name:           "successful sign out",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			entryID:        testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectedResponse: map[string]any{
				"membershipId": testutils.TEST_MEMBERSHIPS[2].ID.String(),
				"eventId":      float64(2),
				"placement":    float64(0),
				// signedOutAt will be set dynamically
			},
		},
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       "invalid-uuid",
			eventID:          "2",
			entryID:          testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "invalid-id",
			entryID:          testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "invalid entry ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "2",
			entryID:          "invalid-uuid",
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Entry ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "event already ended - cannot sign out",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "1",
			entryID:          testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus:   http.StatusForbidden,
			expectError:      true,
			expectedErrorMsg: "Modification of a completed event is forbidden",
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
			semesterID       string
			eventID          string
			entryID          string
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
			expectedResponse map[string]any
		}{
			name:           fmt.Sprintf("successful sign out with role %s", role),
			userRole:       role,
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			entryID:        testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectedResponse: map[string]any{
				"membershipId": testutils.TEST_MEMBERSHIPS[2].ID.String(),
				"eventId":      float64(2),
				"placement":    float64(0),
				// signedOutAt will be set dynamically
			},
		})
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
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/entries/%s/sign-out", tc.semesterID, tc.eventID, tc.entryID),
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
				// Parse response to get the actual signedOutAt time
				var actualResponse map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actualResponse))

				// Set the actual signedOutAt and id in the expected response
				tc.expectedResponse["signedOutAt"] = actualResponse["signedOutAt"]
				tc.expectedResponse["id"] = actualResponse["id"]

				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}

func TestSignInEntry(t *testing.T) {
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
		"POST",
		fmt.Sprintf("/api/v2/semesters/%s/events/1/entries/%s/sign-in", testutils.TEST_SEMESTERS[0].ID, testutils.TEST_MEMBERSHIPS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		entryID          string
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		expectedResponse map[string]any
	}{
		{
			name:           "successful sign in",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			entryID:        testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectedResponse: map[string]any{
				"membershipId": testutils.TEST_MEMBERSHIPS[0].ID.String(),
				"eventId":      float64(2),
				"placement":    float64(0),
				"signedOutAt":  nil,
			},
		},
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       "invalid-uuid",
			eventID:          "1",
			entryID:          testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "invalid-id",
			entryID:          testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "invalid entry ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "1",
			entryID:          "invalid-uuid",
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Entry ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "event already ended - cannot sign in",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "1",
			entryID:          testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus:   http.StatusForbidden,
			expectError:      true,
			expectedErrorMsg: "Modification of a completed event is forbidden",
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
			semesterID       string
			eventID          string
			entryID          string
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
			expectedResponse map[string]any
		}{
			name:           fmt.Sprintf("successful sign in with role %s", role),
			userRole:       role,
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			entryID:        testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectedResponse: map[string]any{
				"membershipId": testutils.TEST_MEMBERSHIPS[0].ID.String(),
				"eventId":      float64(2),
				"placement":    float64(0),
				"signedOutAt":  nil,
			},
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data - participant should already be signed out
			require.NoError(t, testutils.SeedAll(db))

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			// Create request
			req, err := testutils.MakeJSONRequest(
				"POST",
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/entries/%s/sign-in", tc.semesterID, tc.eventID, tc.entryID),
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
				// Parse response to get the actual id
				var actualResponse map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actualResponse))

				// Set the actual id in the expected response
				tc.expectedResponse["id"] = actualResponse["id"]

				testutils.AssertSuccessResponse(t, w, tc.expectedStatus, tc.expectedResponse)
			}
		})
	}
}

func TestDeleteEntry(t *testing.T) {
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
		"DELETE",
		fmt.Sprintf("/api/v2/semesters/%s/events/2/entries/%s", testutils.TEST_SEMESTERS[0].ID, testutils.TEST_MEMBERSHIPS[2].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name             string
		userRole         string
		semesterID       string
		eventID          string
		entryID          string
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:           "successful delete",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			entryID:        testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:             "invalid semester ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       "invalid-uuid",
			eventID:          "2",
			entryID:          testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Semester ID 'invalid-uuid' is not a valid UUID",
		},
		{
			name:             "invalid event ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "invalid-id",
			entryID:          testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Event ID 'invalid-id' is not a valid integer",
		},
		{
			name:             "invalid entry ID format",
			userRole:         authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:       testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:          "2",
			entryID:          "invalid-uuid",
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Entry ID 'invalid-uuid' is not a valid UUID",
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
			semesterID       string
			eventID          string
			entryID          string
			expectedStatus   int
			expectError      bool
			expectedErrorMsg string
		}{
			name:           fmt.Sprintf("successful delete with role %s", role),
			userRole:       role,
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			eventID:        "2",
			entryID:        testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		})
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
				fmt.Sprintf("/api/v2/semesters/%s/events/%s/entries/%s", tc.semesterID, tc.eventID, tc.entryID),
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
