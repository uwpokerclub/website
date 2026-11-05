package controller_test

import (
	"api/internal/models"
	"api/internal/testutils"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChangeExecutiveTeamRolePostgres(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	router := testutils.NewTestAPIServer(db)

	semester := models.Semester{ID: uuid.New(), Name: "Fall", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour), StartingBudget: 125, CurrentBudget: 125, MembershipFee: 50, MembershipDiscountFee: 25}
	otherSemester := models.Semester{ID: uuid.New(), Name: "Winter", StartDate: time.Now(), EndDate: time.Now().Add(48 * time.Hour)}
	member := models.User{ID: 79101, FirstName: "Taylor", LastName: "Officer", Email: "taylor@example.com", Faculty: models.FacultyMath, QuestID: "Quest-One"}
	var authSession uuid.UUID
	reset := func(t *testing.T, username, memberQuestID, currentRole, status, callerRole string, executive bool) {
		t.Helper()
		require.NoError(t, container.ResetDatabase(ctx))
		semester.ID = uuid.New()
		otherSemester.ID = uuid.New()
		member.QuestID = memberQuestID
		require.NoError(t, db.Create(&semester).Error)
		require.NoError(t, db.Create(&otherSemester).Error)
		require.NoError(t, db.Create(&member).Error)
		require.NoError(t, db.Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: semester.ID, Paid: true, Discounted: true, Executive: executive}).Error)
		require.NoError(t, db.Create(&models.Login{Username: username, Password: "password-hash", Role: currentRole, Status: status}).Error)
		authSession, err = testutils.CreateTestSession(db, "auth-"+callerRole, callerRole)
		require.NoError(t, err)
	}
	request := func(t *testing.T, semesterID uuid.UUID, username, role string) *httptest.ResponseRecorder {
		t.Helper()
		req, err := testutils.MakeJSONRequest(http.MethodPatch, "/api/v2/semesters/"+semesterID.String()+"/executive-team/"+username, map[string]string{"role": role})
		require.NoError(t, err)
		testutils.SetAuthCookie(req, authSession)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("role changes delete sessions and preserve account and membership fields", func(t *testing.T) {
		cases := []struct {
			name, currentRole, requestedRole, status string
		}{
			{"executive to tournament director active", "executive", "tournament_director", models.LoginStatusActive},
			{"tournament director to executive pending", "tournament_director", "executive", models.LoginStatusPendingActivation},
			{"disabled account remains disabled", "executive", "tournament_director", models.LoginStatusDisabled},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				reset(t, member.QuestID, member.QuestID, tc.currentRole, tc.status, "secretary", true)
				token := createPostgresActivation(t, db, member.QuestID)
				targetSession := uuid.New()
				otherSession := uuid.New()
				require.NoError(t, db.Create(&models.Session{ID: targetSession, Username: member.QuestID, Role: tc.currentRole, StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
				require.NoError(t, db.Create(&models.Login{Username: "other-user", Password: "other-hash", Role: "executive", Status: models.LoginStatusActive}).Error)
				require.NoError(t, db.Create(&models.Session{ID: otherSession, Username: "other-user", Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
				w := request(t, semester.ID, member.QuestID, tc.requestedRole)
				require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

				var login models.Login
				require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
				require.Equal(t, tc.requestedRole, login.Role)
				require.Equal(t, tc.status, login.Status)
				require.Equal(t, "password-hash", login.Password)
				var sessions int64
				require.NoError(t, db.Model(&models.Session{}).Where("id = ?", targetSession).Count(&sessions).Error)
				require.Zero(t, sessions)
				require.NoError(t, db.Model(&models.Session{}).Where("id = ?", otherSession).Count(&sessions).Error)
				require.EqualValues(t, 1, sessions)
				var activationCount int64
				require.NoError(t, db.Model(&models.AccountActivation{}).Where("username = ? AND used_at IS NULL", member.QuestID).Count(&activationCount).Error)
				require.EqualValues(t, 1, activationCount)
				require.NotEmpty(t, token)
				var savedMembership models.Membership
				require.NoError(t, db.First(&savedMembership, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
				require.True(t, savedMembership.Executive)
				require.True(t, savedMembership.Paid)
				require.True(t, savedMembership.Discounted)
				var savedSemester models.Semester
				require.NoError(t, db.First(&savedSemester, "id = ?", semester.ID).Error)
				require.Equal(t, float32(125), savedSemester.CurrentBudget)
			})
		}
	})

	t.Run("same role is a no-op and keeps sessions", func(t *testing.T) {
		reset(t, member.QuestID, member.QuestID, "executive", models.LoginStatusActive, "secretary", true)
		targetSession := uuid.New()
		require.NoError(t, db.Create(&models.Session{ID: targetSession, Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
		w := request(t, semester.ID, member.QuestID, "executive")
		require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
		var sessions int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", targetSession).Count(&sessions).Error)
		require.EqualValues(t, 1, sessions)
	})

	t.Run("protected requested and current roles are forbidden", func(t *testing.T) {
		for _, role := range []string{"secretary", "treasurer", "vice_president", "president", "bot", "webmaster"} {
			t.Run("requested "+role, func(t *testing.T) {
				reset(t, member.QuestID, member.QuestID, "executive", models.LoginStatusActive, "secretary", true)
				w := request(t, semester.ID, member.QuestID, role)
				require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			})
			t.Run("current "+role, func(t *testing.T) {
				reset(t, member.QuestID, member.QuestID, role, models.LoginStatusActive, "secretary", true)
				w := request(t, semester.ID, member.QuestID, "executive")
				require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			})
		}
	})

	t.Run("executive and tournament director callers are denied", func(t *testing.T) {
		for _, callerRole := range []string{"executive", "tournament_director"} {
			t.Run(callerRole, func(t *testing.T) {
				reset(t, member.QuestID, member.QuestID, "executive", models.LoginStatusActive, callerRole, true)
				w := request(t, semester.ID, member.QuestID, "tournament_director")
				require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			})
		}
	})

	t.Run("exact identity and selected semester membership are required", func(t *testing.T) {
		reset(t, "quest-one", "Quest-One", "executive", models.LoginStatusActive, "secretary", true)
		w := request(t, semester.ID, "quest-one", "tournament_director")
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())

		reset(t, member.QuestID, member.QuestID, "executive", models.LoginStatusActive, "secretary", false)
		w = request(t, semester.ID, member.QuestID, "tournament_director")
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())

		reset(t, member.QuestID, member.QuestID, "executive", models.LoginStatusActive, "secretary", true)
		var membership models.Membership
		require.NoError(t, db.First(&membership, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
		require.NoError(t, db.Model(&models.Membership{}).Where("id = ?", membership.ID).Update("semester_id", otherSemester.ID).Error)
		w = request(t, semester.ID, member.QuestID, "tournament_director")
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("invalid and missing semester or target map to request and not found errors", func(t *testing.T) {
		reset(t, member.QuestID, member.QuestID, "executive", models.LoginStatusActive, "secretary", true)
		req, err := testutils.MakeJSONRequest(http.MethodPatch, "/api/v2/semesters/not-a-uuid/executive-team/Quest-One", map[string]string{"role": "tournament_director"})
		require.NoError(t, err)
		testutils.SetAuthCookie(req, authSession)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

		require.Equal(t, http.StatusNotFound, request(t, uuid.New(), member.QuestID, "tournament_director").Code)
		require.Equal(t, http.StatusNotFound, request(t, semester.ID, "missing-user", "tournament_director").Code)
	})

	t.Run("session deletion failure rolls back the role update", func(t *testing.T) {
		reset(t, member.QuestID, member.QuestID, "executive", models.LoginStatusActive, "secretary", true)
		targetSession := uuid.New()
		require.NoError(t, db.Create(&models.Session{ID: targetSession, Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
		require.NoError(t, db.Exec(`CREATE FUNCTION reject_session_delete() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'forced session delete failure'; END; $$ LANGUAGE plpgsql`).Error)
		require.NoError(t, db.Exec(`CREATE TRIGGER reject_session_delete BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION reject_session_delete()`).Error)
		w := request(t, semester.ID, member.QuestID, "tournament_director")
		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, "executive", login.Role)
		var sessions int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", targetSession).Count(&sessions).Error)
		require.EqualValues(t, 1, sessions)
	})
}
