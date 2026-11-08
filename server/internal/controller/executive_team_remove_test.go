package controller_test

import (
	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRemoveExecutiveTeamMemberPostgres(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	router := testutils.NewTestAPIServer(db)

	var semester, otherSemester models.Semester
	var member models.User
	var authSession uuid.UUID
	reset := func(t *testing.T, role, status, callerRole string) {
		t.Helper()
		require.NoError(t, container.ResetDatabase(ctx))
		semester = models.Semester{ID: uuid.New(), Name: "Fall", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour), FreeTrialLimit: 2, StartingBudget: 125, CurrentBudget: 125, MembershipFee: 50}
		otherSemester = models.Semester{ID: uuid.New(), Name: "Winter", StartDate: time.Now(), EndDate: time.Now().Add(48 * time.Hour), FreeTrialLimit: 2}
		member = models.User{ID: 79101, FirstName: "Taylor", LastName: "Officer", Email: "taylor@example.com", Faculty: models.FacultyMath, QuestID: "Quest-One"}
		require.NoError(t, db.Create(&semester).Error)
		require.NoError(t, db.Create(&otherSemester).Error)
		require.NoError(t, db.Create(&member).Error)
		require.NoError(t, db.Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: semester.ID, Executive: true, FreeTrialAvailable: true}).Error)
		require.NoError(t, db.Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: otherSemester.ID, Executive: true}).Error)
		require.NoError(t, db.Create(&models.Login{Username: member.QuestID, Password: "password-hash", Role: role, Status: status}).Error)
		authSession, err = testutils.CreateTestSession(db, "auth-"+callerRole, callerRole)
		require.NoError(t, err)
	}
	request := func(t *testing.T, semesterID uuid.UUID, username string, session *uuid.UUID) *httptest.ResponseRecorder {
		t.Helper()
		req, err := testutils.MakeJSONRequest(http.MethodDelete, "/api/v2/semesters/"+semesterID.String()+"/executive-team/"+username, nil)
		require.NoError(t, err)
		if session != nil {
			testutils.SetAuthCookie(req, *session)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("executive and tournament director are removed atomically", func(t *testing.T) {
		for _, role := range []string{"executive", "tournament_director"} {
			t.Run(role, func(t *testing.T) {
				reset(t, role, models.LoginStatusActive, "secretary")
				activationToken := createPostgresActivation(t, db, member.QuestID)
				var activationBefore models.AccountActivation
				require.NoError(t, db.First(&activationBefore, "username = ? AND used_at IS NULL", member.QuestID).Error)
				targetSession := uuid.New()
				require.NoError(t, db.Create(&models.Session{ID: targetSession, Username: member.QuestID, Role: role, StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
				require.NoError(t, db.Create(&models.Login{Username: "unrelated-user", Password: "hash", Role: "executive", Status: models.LoginStatusActive}).Error)
				otherSession := uuid.New()
				require.NoError(t, db.Create(&models.Session{ID: otherSession, Username: "unrelated-user", Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
				w := request(t, semester.ID, member.QuestID, &authSession)
				require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

				var login models.Login
				require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
				require.Equal(t, models.LoginStatusDisabled, login.Status)
				require.Equal(t, role, login.Role)
				require.Equal(t, "password-hash", login.Password)
				var sessions int64
				require.NoError(t, db.Model(&models.Session{}).Where("username = ?", member.QuestID).Count(&sessions).Error)
				require.Zero(t, sessions)
				require.NoError(t, db.Model(&models.Session{}).Where("id = ?", otherSession).Count(&sessions).Error)
				require.EqualValues(t, 1, sessions)
				require.NoError(t, db.Model(&models.Session{}).Where("id = ?", targetSession).Count(&sessions).Error)
				require.Zero(t, sessions)
				var activationCount int64
				require.NoError(t, db.Model(&models.AccountActivation{}).Where("username = ? AND used_at IS NULL", member.QuestID).Count(&activationCount).Error)
				require.EqualValues(t, 1, activationCount)
				var activationAfter models.AccountActivation
				require.NoError(t, db.First(&activationAfter, "username = ? AND used_at IS NULL", member.QuestID).Error)
				require.Equal(t, activationBefore, activationAfter)
				require.NotEmpty(t, activationToken)
				var selected, other models.Membership
				require.NoError(t, db.First(&selected, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
				require.False(t, selected.Executive)
				require.NoError(t, db.First(&other, "user_id = ? AND semester_id = ?", member.ID, otherSemester.ID).Error)
				require.True(t, other.Executive)
				var savedSemester models.Semester
				require.NoError(t, db.First(&savedSemester, "id = ?", semester.ID).Error)
				require.Equal(t, float32(125), savedSemester.CurrentBudget)
			})
		}
	})

	t.Run("protected roles are rejected", func(t *testing.T) {
		for _, role := range []string{"secretary", "treasurer", "vice_president", "president", "webmaster", "bot"} {
			t.Run(role, func(t *testing.T) {
				reset(t, role, models.LoginStatusActive, "secretary")
				w := request(t, semester.ID, member.QuestID, &authSession)
				require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
				var login models.Login
				require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
				require.Equal(t, role, login.Role)
				require.Equal(t, models.LoginStatusActive, login.Status)
			})
		}
	})

	t.Run("exact identity and selected executive membership are required", func(t *testing.T) {
		reset(t, "executive", models.LoginStatusActive, "secretary")
		w := request(t, semester.ID, "quest-one", &authSession)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		require.NoError(t, db.Model(&models.Membership{}).Where("user_id = ? AND semester_id = ?", member.ID, semester.ID).Update("executive", false).Error)
		w = request(t, semester.ID, member.QuestID, &authSession)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		w = request(t, uuid.New(), member.QuestID, &authSession)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("staged transition is rejected without changing account", func(t *testing.T) {
		reset(t, "executive", models.LoginStatusPendingActivation, "secretary")
		transitionID := uuid.New()
		transition := models.OfficerTransition{ID: transitionID, InitiatedBy: "auth-secretary", PresidentUsername: "pres", VicePresidentUsername: "vp", SecretaryUsername: "sec", TreasurerUsername: "treas", Status: models.OfficerTransitionPending}
		require.NoError(t, db.Create(&transition).Error)
		require.NoError(t, db.Model(&models.Login{}).Where("username = ?", member.QuestID).Update("staged_transition_id", transitionID).Error)
		w := request(t, semester.ID, member.QuestID, &authSession)
		require.Equal(t, http.StatusConflict, w.Code, w.Body.String())
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, models.LoginStatusPendingActivation, login.Status)
		require.Equal(t, "executive", login.Role)
		require.NotNil(t, login.StagedTransitionID)
	})

	t.Run("caller authorization and request errors", func(t *testing.T) {
		for _, callerRole := range []string{"executive", "tournament_director", "bot"} {
			t.Run("denied "+callerRole, func(t *testing.T) {
				reset(t, "executive", models.LoginStatusActive, callerRole)
				w := request(t, semester.ID, member.QuestID, &authSession)
				require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			})
		}
		reset(t, "executive", models.LoginStatusActive, "secretary")
		require.Equal(t, http.StatusUnauthorized, request(t, semester.ID, member.QuestID, nil).Code)
		req, err := testutils.MakeJSONRequest(http.MethodDelete, "/api/v2/semesters/not-a-uuid/executive-team/Quest-One", nil)
		require.NoError(t, err)
		testutils.SetAuthCookie(req, authSession)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})

	t.Run("session deletion failure rolls back all changes", func(t *testing.T) {
		reset(t, "executive", models.LoginStatusActive, "secretary")
		require.NoError(t, db.Create(&models.Session{ID: uuid.New(), Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
		var savedSession models.Session
		require.NoError(t, db.First(&savedSession, "username = ?", member.QuestID).Error)
		require.NoError(t, db.Exec(`CREATE FUNCTION reject_session_delete() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'forced session delete failure'; END; $$ LANGUAGE plpgsql`).Error)
		require.NoError(t, db.Exec(`CREATE TRIGGER reject_session_delete BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION reject_session_delete()`).Error)
		w := request(t, semester.ID, member.QuestID, &authSession)
		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, models.LoginStatusActive, login.Status)
		var membership models.Membership
		require.NoError(t, db.First(&membership, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
		require.True(t, membership.Executive)
		var retainedSessionCount int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", savedSession.ID).Count(&retainedSessionCount).Error)
		require.EqualValues(t, 1, retainedSessionCount)
	})

	t.Run("free trial cache matches event sign-in after removal", func(t *testing.T) {
		for _, tc := range []struct {
			name              string
			priorEntries      int
			wantAvailable     bool
			wantSignInSuccess bool
		}{{"under limit", 1, true, true}, {"at limit", 2, false, false}, {"over limit", 3, false, false}} {
			t.Run(tc.name, func(t *testing.T) {
				reset(t, "executive", models.LoginStatusActive, "secretary")
				structure, err := testutils.CreateTestStructure(db, "Test")
				require.NoError(t, err)
				events := make([]models.Event, tc.priorEntries+1)
				for i := range events {
					event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Event")
					require.NoError(t, err)
					events[i] = *event
				}
				var membership models.Membership
				require.NoError(t, db.First(&membership, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
				for i := 0; i < tc.priorEntries; i++ {
					require.NoError(t, db.Create(&models.Participant{MembershipID: &membership.ID, EventID: events[i].ID}).Error)
				}
				w := request(t, semester.ID, member.QuestID, &authSession)
				require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
				var updated models.Membership
				require.NoError(t, db.First(&updated, "id = ?", membership.ID).Error)
				require.Equal(t, tc.wantAvailable, updated.FreeTrialAvailable)
				_, err = services.NewParticipantsService(postgres.NewStore(db)).CreateParticipant(&models.CreateParticipantRequest{MembershipID: membership.ID, EventID: events[len(events)-1].ID})
				require.Equal(t, tc.wantSignInSuccess, err == nil, "cached trial status and live event sign-in should agree; error: %v", err)
			})
		}
	})
}
