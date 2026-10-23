package controller_test

import (
	"api/internal/models"
	"api/internal/testutils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetExecutiveTeamRoster(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	router := testutils.NewTestAPIServer(db)

	semesterID := uuid.New()
	for _, semester := range []models.Semester{
		{ID: semesterID, Name: "Current", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour)},
	} {
		require.NoError(t, db.Create(&semester).Error)
	}
	users := []models.User{
		{ID: 71001, FirstName: "Alex", LastName: "Able", QuestID: "Quest-A"},
		{ID: 71002, FirstName: "Blair", LastName: "Baker", QuestID: "quest-b"},
		{ID: 71003, FirstName: "Casey", LastName: "Clark", QuestID: "quest-c"},
		{ID: 71004, FirstName: "Drew", LastName: "Dale", QuestID: "quest-d"},
		{ID: 71005, FirstName: "Evan", LastName: "Pending", QuestID: "quest-e"},
		{ID: 71006, FirstName: "Fran", LastName: "NoLogin", QuestID: "quest-f"},
		{ID: 71007, FirstName: "Noah", LastName: "Nonexec", QuestID: "quest-not-exec"},
	}
	for i := range users {
		require.NoError(t, db.Create(&users[i]).Error)
	}
	memberships := []models.Membership{
		{UserID: users[0].ID, SemesterID: semesterID, Executive: true},
		{UserID: users[1].ID, SemesterID: semesterID, Executive: true},
		{UserID: users[2].ID, SemesterID: semesterID, Executive: true},
		{UserID: users[3].ID, SemesterID: semesterID, Executive: true},
		{UserID: users[4].ID, SemesterID: semesterID, Executive: true},
		{UserID: users[5].ID, SemesterID: semesterID, Executive: true},
		{UserID: users[6].ID, SemesterID: semesterID, Executive: false},
	}
	require.NoError(t, db.Create(&memberships).Error)
	logins := []models.Login{
		{Username: " quest-a ", Password: "private-hash", Role: "president", Status: models.LoginStatusActive},
		{Username: "quest-b", Password: "private-hash", Role: "secretary", Status: models.LoginStatusDisabled},
		{Username: "quest-c", Password: "private-hash", Role: "webmaster", Status: models.LoginStatusActive},
		{Username: "quest-d", Password: "private-hash", Role: "bot", Status: models.LoginStatusActive},
		{Username: "quest-e", Password: "private-hash", Role: "treasurer", Status: models.LoginStatusPendingActivation},
		{Username: "login-only", Password: "private-hash", Role: "executive", Status: models.LoginStatusActive},
		{Username: "unmatched-pending", Password: "private-hash", Role: "treasurer", Status: models.LoginStatusPendingActivation},
		{Username: "not-exec", Password: "private-hash", Role: "tournament_director", Status: models.LoginStatusActive},
	}
	require.NoError(t, db.Create(&logins).Error)

	// More than the generic list endpoint's default 100 rows verifies this roster is unpaginated.
	extraUsers := make([]models.User, 105)
	extraMemberships := make([]models.Membership, 105)
	for i := range extraUsers {
		extraUsers[i] = models.User{ID: uint64(72000 + i), FirstName: fmt.Sprintf("Extra%03d", i), LastName: "Member", QuestID: fmt.Sprintf("extra-%03d", i)}
		extraMemberships[i] = models.Membership{UserID: extraUsers[i].ID, SemesterID: semesterID, Executive: true}
	}
	require.NoError(t, db.Create(&extraUsers).Error)
	require.NoError(t, db.Create(&extraMemberships).Error)

	sessionID, err := testutils.CreateTestSession(db, "auth-webmaster", "webmaster")
	require.NoError(t, err)
	req, err := testutils.MakeJSONRequest("GET", "/api/v2/semesters/"+semesterID.String()+"/executive-team", nil)
	require.NoError(t, err)
	testutils.SetAuthCookie(req, sessionID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var response models.ExecutiveTeamResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Len(t, response.Members, 111)
	byQuestID := map[string]models.ExecutiveTeamMember{}
	for i, member := range response.Members {
		byQuestID[member.QuestID] = member
		if i > 0 {
			previous := response.Members[i-1]
			previousKey := fmt.Sprintf("%s\x00%s\x00%s\x00%020d", previous.LastName, previous.FirstName, previous.QuestID, previous.MemberID)
			currentKey := fmt.Sprintf("%s\x00%s\x00%s\x00%020d", member.LastName, member.FirstName, member.QuestID, member.MemberID)
			require.LessOrEqual(t, previousKey, currentKey, "SQL roster ordering must be deterministic")
		}
	}
	require.Equal(t, users[0].ID, byQuestID["Quest-A"].MemberID, "normalized username must link to the user and dedupe membership")
	require.True(t, byQuestID["Quest-A"].HasExecutiveMembership)
	require.True(t, byQuestID["Quest-A"].HasLogin)
	require.Equal(t, " quest-a ", *byQuestID["Quest-A"].Username)
	require.True(t, byQuestID["quest-b"].HasExecutiveMembership)
	require.True(t, byQuestID["quest-b"].HasLogin)
	require.Equal(t, models.LoginStatusDisabled, *byQuestID["quest-b"].AccountStatus)
	require.Equal(t, "secretary", *byQuestID["quest-b"].Role)
	require.True(t, byQuestID["quest-c"].HasExecutiveMembership)
	require.True(t, byQuestID["quest-c"].HasLogin)
	require.Equal(t, "webmaster", *byQuestID["quest-c"].Role)
	require.True(t, byQuestID["quest-d"].HasExecutiveMembership)
	require.True(t, byQuestID["quest-d"].HasLogin)
	require.Equal(t, "bot", *byQuestID["quest-d"].Role)
	require.True(t, byQuestID["quest-e"].HasExecutiveMembership)
	require.True(t, byQuestID["quest-e"].ActivationOutstanding)
	require.True(t, byQuestID["quest-e"].HasLogin)
	require.True(t, byQuestID["quest-f"].HasExecutiveMembership)
	require.False(t, byQuestID["quest-f"].HasLogin)
	require.Nil(t, byQuestID["quest-f"].Username)
	require.Nil(t, byQuestID["quest-f"].Role)
	require.Nil(t, byQuestID["quest-f"].AccountStatus)
	require.False(t, byQuestID["quest-f"].ActivationOutstanding)
	require.NotContains(t, byQuestID, "login-only")
	require.NotContains(t, byQuestID, "unmatched-pending")
	require.NotContains(t, w.Body.String(), "private-hash")
	require.NotContains(t, byQuestID, "not-exec")
}

func TestGetExecutiveTeamEmptyAndErrors(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	router := testutils.NewTestAPIServer(db)
	semester := models.Semester{ID: uuid.New(), Name: "Empty", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour)}
	require.NoError(t, db.Create(&semester).Error)
	webmasterSession, err := testutils.CreateTestSession(db, "auth-webmaster", "webmaster")
	require.NoError(t, err)
	botSession, err := testutils.CreateTestSession(db, "auth-bot", "bot")
	require.NoError(t, err)
	request := func(path string, session *uuid.UUID) *httptest.ResponseRecorder {
		req, reqErr := testutils.MakeJSONRequest("GET", path, nil)
		require.NoError(t, reqErr)
		if session != nil {
			testutils.SetAuthCookie(req, *session)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		return recorder
	}

	empty := request("/api/v2/semesters/"+semester.ID.String()+"/executive-team", &webmasterSession)
	require.Equal(t, http.StatusOK, empty.Code, empty.Body.String())
	require.JSONEq(t, `{"members":[]}`, empty.Body.String())
	require.Equal(t, http.StatusBadRequest, request("/api/v2/semesters/not-a-uuid/executive-team", &webmasterSession).Code)
	require.Equal(t, http.StatusNotFound, request("/api/v2/semesters/"+uuid.NewString()+"/executive-team", &webmasterSession).Code)
	require.Equal(t, http.StatusUnauthorized, request("/api/v2/semesters/"+semester.ID.String()+"/executive-team", nil).Code)
	require.Equal(t, http.StatusForbidden, request("/api/v2/semesters/"+semester.ID.String()+"/executive-team", &botSession).Code)
}
