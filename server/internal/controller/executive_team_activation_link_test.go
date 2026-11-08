package controller_test

import (
	"api/internal/models"
	"api/internal/testutils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateExecutiveTeamActivationLinkPostgres(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	router := testutils.NewTestAPIServer(db)

	var semester models.Semester
	var otherSemester models.Semester
	var secretarySession uuid.UUID
	reset := func(t *testing.T, role, status string) (models.User, string) {
		t.Helper()
		require.NoError(t, container.ResetDatabase(ctx))
		semester = models.Semester{ID: uuid.New(), Name: "Fall", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour), CurrentBudget: 17}
		otherSemester = models.Semester{ID: uuid.New(), Name: "Winter", StartDate: time.Now(), EndDate: time.Now().Add(48 * time.Hour)}
		require.NoError(t, db.Create(&[]models.Semester{semester, otherSemester}).Error)
		member := models.User{ID: 78101, FirstName: "Avery", LastName: "Officer", Email: "avery@example.com", Faculty: models.FacultyMath, QuestID: "Quest-Target"}
		require.NoError(t, db.Create(&member).Error)
		require.NoError(t, db.Create(&models.Login{Username: member.QuestID, Password: "preserved-password-hash", Role: role, Status: status}).Error)
		require.NoError(t, db.Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: semester.ID, Executive: true, Paid: true, Discounted: true}).Error)
		secretarySession, err = testutils.CreateTestSession(db, "auth-secretary", "secretary")
		require.NoError(t, err)
		return member, "preserved-password-hash"
	}
	request := func(t *testing.T, semesterID uuid.UUID, username string, session *uuid.UUID) *httptest.ResponseRecorder {
		t.Helper()
		req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/semesters/"+semesterID.String()+"/executive-team/"+username+"/activation-link", nil)
		require.NoError(t, err)
		if session != nil {
			testutils.SetAuthCookie(req, *session)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	verify := func(t *testing.T, token string) int {
		t.Helper()
		req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/verify", map[string]string{"token": token})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("replaces the prior token and preserves account and membership state", func(t *testing.T) {
		member, password := reset(t, "executive", models.LoginStatusPendingActivation)
		oldToken := createPostgresActivation(t, db, member.QuestID)
		preservedSessionID := uuid.New()
		require.NoError(t, db.Create(&models.Session{ID: preservedSessionID, Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)

		w := request(t, semester.ID, member.QuestID, &secretarySession)
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var response models.ExecutiveTeamActivationLinkResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.NotEmpty(t, response.ActivationToken)
		tokenHash := sha256.Sum256([]byte(response.ActivationToken))
		require.NotContains(t, w.Body.String(), hex.EncodeToString(tokenHash[:]))
		require.Equal(t, http.StatusUnauthorized, verify(t, oldToken), "the previous token must be revoked")
		require.Equal(t, http.StatusOK, verify(t, response.ActivationToken), "the returned token must be usable")

		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, password, login.Password)
		require.Equal(t, "executive", login.Role)
		require.Equal(t, models.LoginStatusPendingActivation, login.Status)
		var membership models.Membership
		require.NoError(t, db.First(&membership, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
		require.True(t, membership.Executive)
		require.True(t, membership.Paid)
		require.True(t, membership.Discounted)
		var storedSemester models.Semester
		require.NoError(t, db.First(&storedSemester, "id = ?", semester.ID).Error)
		require.Equal(t, semester.CurrentBudget, storedSemester.CurrentBudget)
		require.Equal(t, semester.StartingBudget, storedSemester.StartingBudget)
		var sessionCount int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", preservedSessionID).Count(&sessionCount).Error)
		require.EqualValues(t, 1, sessionCount)
		var persisted models.AccountActivation
		require.NoError(t, db.First(&persisted, "username = ? AND used_at IS NULL", member.QuestID).Error)
		expectedHash := sha256.Sum256([]byte(response.ActivationToken))
		require.Equal(t, expectedHash[:], persisted.TokenHash)
		require.NotContains(t, string(persisted.TokenHash), response.ActivationToken)
	})

	t.Run("requires the secretary-level reissue permission", func(t *testing.T) {
		member, _ := reset(t, "executive", models.LoginStatusPendingActivation)
		executiveSession, err := testutils.CreateTestSession(db, "auth-executive", "executive")
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, request(t, semester.ID, member.QuestID, &executiveSession).Code)
		require.Equal(t, http.StatusUnauthorized, request(t, semester.ID, member.QuestID, nil).Code)
	})

	t.Run("rejects ineligible account, role, membership, and identity targets", func(t *testing.T) {
		cases := []struct {
			name       string
			role       string
			status     string
			mutate     func(*testing.T, models.User)
			pathUser   func(models.User) string
			wantStatus int
		}{
			{name: "active account", role: "executive", status: models.LoginStatusActive, wantStatus: http.StatusConflict},
			{name: "disabled account", role: "executive", status: models.LoginStatusDisabled, wantStatus: http.StatusConflict},
			{name: "protected president", role: "president", status: models.LoginStatusPendingActivation, wantStatus: http.StatusForbidden},
			{name: "protected vice president", role: "vice_president", status: models.LoginStatusPendingActivation, wantStatus: http.StatusForbidden},
			{name: "protected secretary", role: "secretary", status: models.LoginStatusPendingActivation, wantStatus: http.StatusForbidden},
			{name: "protected treasurer", role: "treasurer", status: models.LoginStatusPendingActivation, wantStatus: http.StatusForbidden},
			{name: "protected webmaster", role: "webmaster", status: models.LoginStatusPendingActivation, wantStatus: http.StatusForbidden},
			{name: "protected bot", role: "bot", status: models.LoginStatusPendingActivation, wantStatus: http.StatusForbidden},
			{name: "not executive in selected semester", role: "executive", status: models.LoginStatusPendingActivation, mutate: func(t *testing.T, member models.User) {
				require.NoError(t, db.Model(&models.Membership{}).Where("user_id = ? AND semester_id = ?", member.ID, semester.ID).Update("executive", false).Error)
			}, wantStatus: http.StatusNotFound},
			{name: "executive only in another semester", role: "executive", status: models.LoginStatusPendingActivation, mutate: func(t *testing.T, member models.User) {
				require.NoError(t, db.Model(&models.Membership{}).Where("user_id = ? AND semester_id = ?", member.ID, semester.ID).Update("semester_id", otherSemester.ID).Error)
			}, wantStatus: http.StatusNotFound},
			{name: "login does not exactly link to an executive member", role: "executive", status: models.LoginStatusPendingActivation, mutate: func(t *testing.T, member models.User) {
				require.NoError(t, db.Model(&models.Login{}).Where("username = ?", member.QuestID).Update("username", "unlinked-login").Error)
			}, pathUser: func(models.User) string { return "unlinked-login" }, wantStatus: http.StatusNotFound},
			{name: "staged officer-transition account", role: "executive", status: models.LoginStatusPendingActivation, mutate: func(t *testing.T, member models.User) {
				transition := models.OfficerTransition{ID: uuid.New(), InitiatedBy: "auth-secretary", PresidentUsername: member.QuestID, VicePresidentUsername: "vp", SecretaryUsername: "sec", TreasurerUsername: "treasurer", Status: models.OfficerTransitionPending}
				require.NoError(t, db.Create(&transition).Error)
				require.NoError(t, db.Model(&models.Login{}).Where("username = ?", member.QuestID).Update("staged_transition_id", transition.ID).Error)
			}, wantStatus: http.StatusConflict},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				member, _ := reset(t, tc.role, tc.status)
				if tc.mutate != nil {
					tc.mutate(t, member)
				}
				username := member.QuestID
				if tc.pathUser != nil {
					username = tc.pathUser(member)
				}
				w := request(t, semester.ID, username, &secretarySession)
				require.Equal(t, tc.wantStatus, w.Code, w.Body.String())
				var activations int64
				require.NoError(t, db.Model(&models.AccountActivation{}).Count(&activations).Error)
				require.Zero(t, activations, "rejected requests must not mint or persist a token")
			})
		}
	})

	t.Run("returns no token and restores the previous link when commit fails", func(t *testing.T) {
		member, _ := reset(t, "executive", models.LoginStatusPendingActivation)
		oldToken := createPostgresActivation(t, db, member.QuestID)
		require.NoError(t, db.Exec(`CREATE FUNCTION fail_activation_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'activation commit rejected'; END; $$`).Error)
		require.NoError(t, db.Exec(`CREATE CONSTRAINT TRIGGER fail_activation_commit AFTER INSERT ON account_activations DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION fail_activation_commit()`).Error)
		w := request(t, semester.ID, member.QuestID, &secretarySession)
		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		require.NotContains(t, w.Body.String(), "activationToken")
		require.Equal(t, http.StatusOK, verify(t, oldToken), "the prior token should survive the failed transaction")
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, models.LoginStatusPendingActivation, login.Status)
	})
}
