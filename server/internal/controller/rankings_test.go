package controller_test

import (
	"api/internal/authorization"
	"api/internal/models"
	"api/internal/testutils"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListRankings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Authorization: BOT+ can access (semester.rankings.list)
	// All authenticated users can list rankings, so only test unauthenticated access
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"GET",
		fmt.Sprintf("/api/v2/semesters/%s/rankings", testutils.TEST_SEMESTERS[0].ID),
		[]string{}, // Empty slice means only test unauthenticated
	)

	testCases := []struct {
		name                 string
		userRole             string
		semesterID           string
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
		minResults           int
	}{
		{
			name:                 "successful request with rankings - BOT role",
			userRole:             authorization.ROLE_BOT.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus:       http.StatusOK,
			expectError:          false,
			expectedErrorMessage: "",
			minResults:           3, // TEST_RANKINGS has 3 rankings for semester 0
		},
		{
			name:                 "successful request with rankings - EXECUTIVE role",
			userRole:             authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus:       http.StatusOK,
			expectError:          false,
			expectedErrorMessage: "",
			minResults:           3,
		},
		{
			name:                 "successful request with rankings - TOURNAMENT_DIRECTOR role",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus:       http.StatusOK,
			expectError:          false,
			expectedErrorMessage: "",
			minResults:           3,
		},
		{
			name:                 "semester with no rankings",
			userRole:             authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[2].ID.String(),
			expectedStatus:       http.StatusOK,
			expectError:          false,
			expectedErrorMessage: "",
			minResults:           0,
		},
		{
			name:                 "invalid semester ID format",
			userRole:             authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:           "invalid-uuid",
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Semester ID 'invalid-uuid' is not a valid UUID",
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
				fmt.Sprintf("/api/v2/semesters/%s/rankings", tc.semesterID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if tc.expectError {
				if tc.expectedErrorMessage != "" {
					require.Contains(t, w.Body.String(), tc.expectedErrorMessage, "Response: %s", w.Body.String())
				}
			} else {
				var rankings []models.RankingResponse
				err = json.Unmarshal(w.Body.Bytes(), &rankings)
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(rankings), tc.minResults)

				// Validate structure of rankings if we have results
				if len(rankings) > 0 {
					for _, ranking := range rankings {
						require.NotZero(t, ranking.ID)
						require.NotEmpty(t, ranking.FirstName)
						require.NotEmpty(t, ranking.LastName)
						require.GreaterOrEqual(t, ranking.Points, int32(0))
						require.GreaterOrEqual(t, ranking.Position, int32(1))
					}

					// Verify rankings are sorted by points descending
					for i := 1; i < len(rankings); i++ {
						require.GreaterOrEqual(t, rankings[i-1].Points, rankings[i].Points,
							"Rankings should be sorted by points descending")
					}
				}
			}
		})
	}
}

func TestGetRanking(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Authorization: BOT+ can access (semester.rankings.get)
	// All authenticated users can get rankings, so only test unauthenticated access
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"GET",
		fmt.Sprintf("/api/v2/semesters/%s/rankings/%s",
			testutils.TEST_SEMESTERS[0].ID,
			testutils.TEST_MEMBERSHIPS[0].ID),
		[]string{}, // Empty slice means only test unauthenticated access
	)

	testCases := []struct {
		name                 string
		userRole             string
		semesterID           string
		membershipID         string
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:                 "successful retrieval - BOT role",
			userRole:             authorization.ROLE_BOT.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:         testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus:       http.StatusOK,
			expectError:          false,
			expectedErrorMessage: "",
		},
		{
			name:                 "successful retrieval - EXECUTIVE role",
			userRole:             authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:         testutils.TEST_MEMBERSHIPS[1].ID.String(),
			expectedStatus:       http.StatusOK,
			expectError:          false,
			expectedErrorMessage: "",
		},
		{
			name:                 "successful retrieval - TOURNAMENT_DIRECTOR role",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:         testutils.TEST_MEMBERSHIPS[2].ID.String(),
			expectedStatus:       http.StatusOK,
			expectError:          false,
			expectedErrorMessage: "",
		},
		{
			name:                 "ranking not found - membership from different semester",
			userRole:             authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[1].ID.String(),
			membershipID:         testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus:       http.StatusNotFound,
			expectError:          true,
			expectedErrorMessage: "record not found",
		},
		{
			name:                 "ranking not found - non-existent membership ID",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:         "00000000-0000-0000-0000-000000000000",
			expectedStatus:       http.StatusNotFound,
			expectError:          true,
			expectedErrorMessage: "record not found",
		},
		{
			name:                 "invalid membership ID format",
			userRole:             authorization.ROLE_BOT.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:         "invalid-uuid",
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "membershipId 'invalid-uuid' is not a valid UUID",
		},
		{
			name:                 "invalid semester ID format",
			userRole:             authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:           "invalid-uuid",
			membershipID:         testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Semester ID 'invalid-uuid' is not a valid UUID",
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
				fmt.Sprintf("/api/v2/semesters/%s/rankings/%s", tc.semesterID, tc.membershipID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if tc.expectError {
				if tc.expectedErrorMessage != "" {
					require.Contains(t, w.Body.String(), tc.expectedErrorMessage, "Response: %s", w.Body.String())
				}
			} else {
				var ranking models.GetRankingResponse
				err := json.Unmarshal(w.Body.Bytes(), &ranking)
				require.NoError(t, err)
				require.GreaterOrEqual(t, ranking.Points, int32(0))
				require.GreaterOrEqual(t, ranking.Position, int32(1))
			}
		})
	}
}

func TestExportRankings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized access - SECRETARY+ can access
	unauthorizedRoles := []string{
		authorization.ROLE_BOT.ToString(),
		authorization.ROLE_EXECUTIVE.ToString(),
		authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
	}
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"GET",
		fmt.Sprintf("/api/v2/semesters/%s/rankings/export", testutils.TEST_SEMESTERS[0].ID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name           string
		userRole       string
		semesterID     string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful export - SECRETARY role",
			userRole:       authorization.ROLE_SECRETARY.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful export - TREASURER role",
			userRole:       authorization.ROLE_TREASURER.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful export - VICE_PRESIDENT role",
			userRole:       authorization.ROLE_VICE_PRESIDENT.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful export - PRESIDENT role",
			userRole:       authorization.ROLE_PRESIDENT.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful export - WEBMASTER role",
			userRole:       authorization.ROLE_WEBMASTER.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "export with empty rankings",
			userRole:       authorization.ROLE_SECRETARY.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[2].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid semester ID format",
			userRole:       authorization.ROLE_SECRETARY.ToString(),
			semesterID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "non-existent semester - returns empty CSV",
			userRole:       authorization.ROLE_SECRETARY.ToString(),
			semesterID:     "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusOK,
			expectError:    false,
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
				fmt.Sprintf("/api/v2/semesters/%s/rankings/export", tc.semesterID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if !tc.expectError {
				// Verify response headers for file download
				require.NotEmpty(t, w.Header().Get("Content-Disposition"))
				require.Contains(t, w.Header().Get("Content-Disposition"), "attachment")

				// Parse and validate CSV content
				csvReader := csv.NewReader(bytes.NewReader(w.Body.Bytes()))
				records, err := csvReader.ReadAll()
				require.NoError(t, err)
				require.NotEmpty(t, records, "CSV should have at least a header row")

				// Validate header row (now includes position column)
				require.Equal(t, []string{"position", "id", "first_name", "last_name", "points"}, records[0],
					"CSV header should match expected format")

				// For semesters with rankings, validate data rows
				if tc.semesterID == testutils.TEST_SEMESTERS[0].ID.String() {
					// Should have 3 rankings for semester 0
					require.GreaterOrEqual(t, len(records), 4, "Should have header + 3 data rows")

					// Validate each data row has correct number of columns
					for i := 1; i < len(records); i++ {
						require.Len(t, records[i], 5, "Each CSV row should have 5 columns")
						require.NotEmpty(t, records[i][0], "Position should not be empty")
						require.NotEmpty(t, records[i][1], "ID should not be empty")
						require.NotEmpty(t, records[i][2], "First name should not be empty")
						require.NotEmpty(t, records[i][3], "Last name should not be empty")
						require.NotEmpty(t, records[i][4], "Points should not be empty")
					}
				} else {
					// For semesters without rankings, should only have header
					require.Len(t, records, 1, "Empty semester should only have CSV header")
				}
			} else {
				// Verify error response
				var errResp map[string]any
				err = json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.NotEmpty(t, errResp["message"])
			}
		})
	}
}
