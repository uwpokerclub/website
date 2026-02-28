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

func TestCreateMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	// Note: BOT can create memberships according to membership_authorizer.go:19
	// So we test with a role that doesn't have access (none in this case)
	// Skipping unauthorized test since BOT and TOURNAMENT_DIRECTOR+ all have access

	testCases := []struct {
		name                 string
		userRole             string
		semesterID           string
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:       "successful creation - paid and not discounted",
			userRole:   authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID: testutils.TEST_SEMESTERS[0].ID.String(),
			requestBody: map[string]any{
				"userId":     testutils.TEST_USERS[3].ID, // Alice has no membership
				"paid":       true,
				"discounted": false,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:       "successful creation - paid and discounted",
			userRole:   authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID: testutils.TEST_SEMESTERS[1].ID.String(),
			requestBody: map[string]any{
				"userId":     testutils.TEST_USERS[3].ID,
				"paid":       true,
				"discounted": true,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:       "successful creation - not paid and not discounted",
			userRole:   authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID: testutils.TEST_SEMESTERS[2].ID.String(),
			requestBody: map[string]any{
				"userId":     testutils.TEST_USERS[3].ID,
				"paid":       false,
				"discounted": false,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:       "invalid state - not paid but discounted",
			userRole:   authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID: testutils.TEST_SEMESTERS[0].ID.String(),
			requestBody: map[string]any{
				"userId":     testutils.TEST_USERS[3].ID,
				"paid":       false,
				"discounted": true,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "cannot create membership that is not paid and discounted",
		},
		{
			name:       "missing required field - userId",
			userRole:   authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID: testutils.TEST_SEMESTERS[0].ID.String(),
			requestBody: map[string]any{
				"paid":       true,
				"discounted": false,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "UserID",
		},
		{
			name:       "invalid semester ID",
			userRole:   authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID: "invalid-uuid",
			requestBody: map[string]any{
				"userId":     testutils.TEST_USERS[3].ID,
				"paid":       true,
				"discounted": false,
			},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "Semester ID",
		},
		{
			name:       "non-existent semester",
			userRole:   authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID: "00000000-0000-0000-0000-000000000000",
			requestBody: map[string]any{
				"userId":     testutils.TEST_USERS[3].ID,
				"paid":       true,
				"discounted": false,
			},
			expectedStatus:       http.StatusInternalServerError,
			expectError:          true,
			expectedErrorMessage: "", // Don't check message - it's a DB constraint error
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
				fmt.Sprintf("/api/v2/semesters/%s/memberships", tc.semesterID),
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
				if tc.expectedErrorMessage != "" {
					require.Contains(t, w.Body.String(), tc.expectedErrorMessage, "Response: %s", w.Body.String())
				}
			} else {
				var membership models.Membership
				err := json.Unmarshal(w.Body.Bytes(), &membership)
				require.NoError(t, err)
				require.Equal(t, tc.requestBody["userId"], membership.UserID)
				require.Equal(t, tc.requestBody["paid"], membership.Paid)
				require.Equal(t, tc.requestBody["discounted"], membership.Discounted)
			}
		})
	}
}

func TestListMemberships(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	// Note: BOT+ can list memberships, so no unauthorized roles to test
	// Skipping unauthorized test since all authenticated roles have access

	testCases := []struct {
		name           string
		userRole       string
		semesterID     string
		queryParams    string
		expectedStatus int
		expectError    bool
		minResults     int
		expectedTotal  int64
	}{
		{
			name:           "list all memberships for semester",
			userRole:       authorization.ROLE_BOT.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectError:    false,
			minResults:     3, // We have 3 test memberships for semester 0
			expectedTotal:  3,
		},
		{
			name:           "list memberships with limit",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			queryParams:    "?limit=2",
			expectedStatus: http.StatusOK,
			expectError:    false,
			minResults:     2,
			expectedTotal:  3,
		},
		{
			name:           "list memberships with offset",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			queryParams:    "?offset=1",
			expectedStatus: http.StatusOK,
			expectError:    false,
			minResults:     2,
			expectedTotal:  3,
		},
		{
			name:           "list memberships with limit and offset",
			userRole:       authorization.ROLE_BOT.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			queryParams:    "?limit=1&offset=1",
			expectedStatus: http.StatusOK,
			expectError:    false,
			minResults:     1,
			expectedTotal:  3,
		},
		{
			name:           "semester with no memberships",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[2].ID.String(),
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectError:    false,
			minResults:     0,
			expectedTotal:  0,
		},
		{
			name:           "invalid semester ID",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:     "invalid-uuid",
			queryParams:    "",
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
				fmt.Sprintf("/api/v2/semesters/%s/memberships%s", tc.semesterID, tc.queryParams),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if tc.expectError {
				var errResp map[string]any
				err = json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
			} else {
				var resp models.ListResponse[models.MembershipWithAttendance]
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(resp.Data), tc.minResults)
				require.Equal(t, tc.expectedTotal, resp.Total)
			}
		})
	}
}

func TestGetMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	// Note: BOT+ can get memberships, so no unauthorized roles to test
	// Skipping unauthorized test since all authenticated roles have access

	testCases := []struct {
		name           string
		userRole       string
		semesterID     string
		membershipID   string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful retrieval",
			userRole:       authorization.ROLE_BOT.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "membership not found - wrong semester",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[1].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "membership not found - non-existent ID",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "invalid membership ID format",
			userRole:       authorization.ROLE_BOT.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid semester ID format",
			userRole:       authorization.ROLE_EXECUTIVE.ToString(),
			semesterID:     "invalid-uuid",
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
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
				fmt.Sprintf("/api/v2/semesters/%s/memberships/%s", tc.semesterID, tc.membershipID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if !tc.expectError {
				var membership models.Membership
				err := json.Unmarshal(w.Body.Bytes(), &membership)
				require.NoError(t, err)
				require.NotEmpty(t, membership.ID)
				require.NotNil(t, membership.User)
				require.NotNil(t, membership.Semester)
			}
		})
	}
}

func TestUpdateMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	unauthorizedRoles := []string{authorization.ROLE_BOT.ToString(), authorization.ROLE_EXECUTIVE.ToString()}
	testSemesterID := testutils.TEST_SEMESTERS[0].ID.String()
	testMembershipID := testutils.TEST_MEMBERSHIPS[0].ID.String()
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"PATCH",
		fmt.Sprintf("/api/v2/semesters/%s/memberships/%s", testSemesterID, testMembershipID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name                 string
		userRole             string
		semesterID           string
		membershipID         string
		requestBody          map[string]any
		expectedStatus       int
		expectError          bool
		expectedErrorMessage string
		validateResponse     func(*testing.T, *models.Membership)
	}{
		{
			name:           "successful update - mark as paid",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[2].ID.String(),
			requestBody:    map[string]any{"paid": true},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, m *models.Membership) {
				require.True(t, m.Paid)
			},
		},
		{
			name:           "successful update - mark as discounted",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			requestBody:    map[string]any{"discounted": true},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, m *models.Membership) {
				require.True(t, m.Discounted)
			},
		},
		{
			name:           "successful update - paid and discounted together",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			requestBody:    map[string]any{"paid": true, "discounted": true},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, m *models.Membership) {
				require.True(t, m.Paid)
				require.True(t, m.Discounted)
			},
		},
		{
			name:           "successful update - remove discount",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[1].ID.String(),
			requestBody:    map[string]any{"discounted": false},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, m *models.Membership) {
				require.False(t, m.Discounted)
			},
		},
		{
			name:                 "invalid state - not paid but discounted",
			userRole:             authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:           testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:         testutils.TEST_MEMBERSHIPS[0].ID.String(),
			requestBody:          map[string]any{"paid": false, "discounted": true},
			expectedStatus:       http.StatusBadRequest,
			expectError:          true,
			expectedErrorMessage: "cannot set membership to not paid and discounted",
		},
		{
			name:           "membership not found - wrong semester",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[1].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			requestBody:    map[string]any{"paid": true},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "membership not found - non-existent ID",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   "00000000-0000-0000-0000-000000000000",
			requestBody:    map[string]any{"paid": true},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "invalid membership ID format",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   "invalid-uuid",
			requestBody:    map[string]any{"paid": true},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid semester ID format",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     "invalid-uuid",
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			requestBody:    map[string]any{"paid": true},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "empty request body",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			requestBody:    map[string]any{},
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
				"PATCH",
				fmt.Sprintf("/api/v2/semesters/%s/memberships/%s", tc.semesterID, tc.membershipID),
				tc.requestBody,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())

			if tc.expectError {
				if tc.expectedErrorMessage != "" {
					testutils.AssertErrorResponse(t, w, tc.expectedStatus, tc.expectedErrorMessage)
				}
			} else {
				var membership models.Membership
				err := json.Unmarshal(w.Body.Bytes(), &membership)
				require.NoError(t, err)
				require.NotEmpty(t, membership.ID)

				if tc.validateResponse != nil {
					tc.validateResponse(t, &membership)
				}
			}
		})
	}
}

func TestUpdateMembershipBudgetUpdates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	testCases := []struct {
		name                  string
		initialPaid           bool
		initialDiscounted     bool
		updatePaid            *bool
		updateDiscounted      *bool
		expectedBudgetChange  float32
		semesterMembershipFee int
		semesterDiscountFee   int
	}{
		{
			name:                  "mark unpaid membership as paid (no discount)",
			initialPaid:           false,
			initialDiscounted:     false,
			updatePaid:            boolPtr(true),
			updateDiscounted:      nil,
			expectedBudgetChange:  8.0, // TEST_SEMESTERS[0].MembershipFee
			semesterMembershipFee: 8,
			semesterDiscountFee:   4,
		},
		{
			name:                  "mark unpaid membership as paid with discount",
			initialPaid:           false,
			initialDiscounted:     false,
			updatePaid:            boolPtr(true),
			updateDiscounted:      boolPtr(true),
			expectedBudgetChange:  4.0, // TEST_SEMESTERS[0].MembershipDiscountFee
			semesterMembershipFee: 8,
			semesterDiscountFee:   4,
		},
		{
			name:                  "add discount to paid membership",
			initialPaid:           true,
			initialDiscounted:     false,
			updatePaid:            nil,
			updateDiscounted:      boolPtr(true),
			expectedBudgetChange:  -4.0, // -(MembershipFee - MembershipDiscountFee)
			semesterMembershipFee: 8,
			semesterDiscountFee:   4,
		},
		{
			name:                  "remove discount from paid membership",
			initialPaid:           true,
			initialDiscounted:     true,
			updatePaid:            nil,
			updateDiscounted:      boolPtr(false),
			expectedBudgetChange:  4.0, // (MembershipFee - MembershipDiscountFee)
			semesterMembershipFee: 8,
			semesterDiscountFee:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for clean state
			require.NoError(t, container.ResetDatabase(ctx))

			// Seed test data
			require.NoError(t, testutils.SeedAll(db))

			// Create a test membership with initial state
			membership, err := testutils.CreateTestMembership(
				db,
				testutils.TEST_USERS[3].ID,
				testutils.TEST_SEMESTERS[0].ID,
			)
			require.NoError(t, err)

			// Update to initial state
			membership.Paid = tc.initialPaid
			membership.Discounted = tc.initialDiscounted
			require.NoError(t, db.Save(membership).Error)

			// Get initial budget
			var semester models.Semester
			require.NoError(t, db.First(&semester, testutils.TEST_SEMESTERS[0].ID).Error)
			initialBudget := semester.CurrentBudget

			// Setup authentication
			sessionID, err := testutils.CreateTestSession(db, "testuser", authorization.ROLE_TOURNAMENT_DIRECTOR.ToString())
			require.NoError(t, err)

			// Build request body
			requestBody := make(map[string]any)
			if tc.updatePaid != nil {
				requestBody["paid"] = *tc.updatePaid
			}
			if tc.updateDiscounted != nil {
				requestBody["discounted"] = *tc.updateDiscounted
			}

			// Create request
			req, err := testutils.MakeJSONRequest(
				"PATCH",
				fmt.Sprintf("/api/v2/semesters/%s/memberships/%s",
					testutils.TEST_SEMESTERS[0].ID.String(),
					membership.ID.String()),
				requestBody,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code, "Response: %s", w.Body.String())

			// Verify budget was updated correctly
			require.NoError(t, db.First(&semester, testutils.TEST_SEMESTERS[0].ID).Error)
			expectedBudget := initialBudget + tc.expectedBudgetChange
			require.InDelta(t, expectedBudget, semester.CurrentBudget, 0.01,
				"Budget should be %f but got %f", expectedBudget, semester.CurrentBudget)
		})
	}
}

func TestDeleteMembership(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetDB()
	apiServer := testutils.NewTestAPIServer(db)

	// Test unauthorized/forbidden access
	unauthorizedRoles := []string{authorization.ROLE_BOT.ToString(), authorization.ROLE_EXECUTIVE.ToString()}
	testSemesterID := testutils.TEST_SEMESTERS[0].ID.String()
	testMembershipID := testutils.TEST_MEMBERSHIPS[0].ID.String()
	testutils.TestInvalidAuthForEndpoint(
		t,
		container,
		apiServer,
		"DELETE",
		fmt.Sprintf("/api/v2/semesters/%s/memberships/%s", testSemesterID, testMembershipID),
		unauthorizedRoles,
	)

	testCases := []struct {
		name           string
		userRole       string
		semesterID     string
		membershipID   string
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "membership not found - wrong semester",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[1].ID.String(),
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "membership not found - non-existent ID",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid membership ID format",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     testutils.TEST_SEMESTERS[0].ID.String(),
			membershipID:   "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid semester ID format",
			userRole:       authorization.ROLE_TOURNAMENT_DIRECTOR.ToString(),
			semesterID:     "invalid-uuid",
			membershipID:   testutils.TEST_MEMBERSHIPS[0].ID.String(),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, container.ResetDatabase(ctx))
			require.NoError(t, testutils.SeedAll(db))

			sessionID, err := testutils.CreateTestSession(db, "testuser", tc.userRole)
			require.NoError(t, err)

			req, err := testutils.MakeJSONRequest(
				"DELETE",
				fmt.Sprintf("/api/v2/semesters/%s/memberships/%s", tc.semesterID, tc.membershipID),
				nil,
			)
			require.NoError(t, err)

			testutils.SetAuthCookie(req, sessionID)

			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			require.Equal(t, tc.expectedStatus, w.Code, "Response: %s", w.Body.String())
		})
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}
