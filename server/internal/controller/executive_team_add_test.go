package controller_test

import (
	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAddExecutiveTeamMemberPostgres(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	router := testutils.NewTestAPIServer(db)

	var semester models.Semester
	var session uuid.UUID
	var member models.User
	reset := func(t *testing.T, questID string) {
		t.Helper()
		require.NoError(t, container.ResetDatabase(ctx))
		semester = models.Semester{ID: uuid.New(), Name: "Fall", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour), MembershipFee: 50, MembershipDiscountFee: 25}
		require.NoError(t, db.Create(&semester).Error)
		member = models.User{ID: 78101, FirstName: "Avery", LastName: "Officer", Email: "avery@example.com", Faculty: models.FacultyMath, QuestID: questID}
		require.NoError(t, db.Create(&member).Error)
		session, err = testutils.CreateTestSession(db, "auth-secretary", "secretary")
		require.NoError(t, err)
	}
	request := func(t *testing.T, semesterID uuid.UUID, questID, role string) *httptest.ResponseRecorder {
		t.Helper()
		req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/semesters/"+semesterID.String()+"/executive-team", map[string]string{"questId": questID, "role": role})
		require.NoError(t, err)
		testutils.SetAuthCookie(req, session)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("new account activates through existing workflow", func(t *testing.T) {
		reset(t, "Quest-New")
		w := request(t, semester.ID, " quest-new ", "tournament_director")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var response models.AddExecutiveTeamMemberResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, member.ID, response.MemberID)
		require.Equal(t, "Quest-New", response.Username)
		require.Equal(t, models.LoginStatusPendingActivation, response.AccountStatus)
		require.NotEmpty(t, response.ActivationToken)
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", response.Username).Error)
		require.Equal(t, "tournament_director", login.Role)
		var activation models.AccountActivation
		require.NoError(t, db.First(&activation, "username = ?", response.Username).Error)
		wantHash := sha256.Sum256([]byte(*response.ActivationToken))
		require.Equal(t, wantHash[:], activation.TokenHash)
		require.WithinDuration(t, time.Now().UTC().Add(14*24*time.Hour), activation.ExpiresAt, 5*time.Second)
		var membership models.Membership
		require.NoError(t, db.First(&membership, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
		require.True(t, membership.Executive)
		require.NotContains(t, w.Body.String(), login.Password)
		verify, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/verify", map[string]string{"token": *response.ActivationToken})
		require.NoError(t, err)
		verified := httptest.NewRecorder()
		router.ServeHTTP(verified, verify)
		require.Equal(t, http.StatusOK, verified.Code, verified.Body.String())
		var verifyResponse models.VerifyActivationResponse
		require.NoError(t, json.Unmarshal(verified.Body.Bytes(), &verifyResponse))
		require.Equal(t, member.FirstName, verifyResponse.FirstName)
		require.Equal(t, member.LastName, verifyResponse.LastName)
		complete, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/complete", map[string]string{"token": *response.ActivationToken, "password": "new-password"})
		require.NoError(t, err)
		completed := httptest.NewRecorder()
		router.ServeHTTP(completed, complete)
		require.Equal(t, http.StatusCreated, completed.Code, completed.Body.String())
		require.NoError(t, db.First(&login, "username = ?", response.Username).Error)
		require.Equal(t, models.LoginStatusActive, login.Status)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("new-password")))
	})

	t.Run("active account keeps password and removes changed-role sessions and stale links", func(t *testing.T) {
		reset(t, "Quest-Active")
		password, err := bcrypt.GenerateFromPassword([]byte("existing-password"), bcrypt.DefaultCost)
		require.NoError(t, err)
		require.NoError(t, db.Create(&models.Login{Username: member.QuestID, Password: string(password), Role: "executive", Status: models.LoginStatusActive}).Error)
		staleToken := createPostgresActivation(t, db, member.QuestID)
		staleSession := uuid.New()
		err = db.Create(&models.Session{ID: staleSession, Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error
		require.NoError(t, err)
		w := request(t, semester.ID, member.QuestID, "tournament_director")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var response models.AddExecutiveTeamMemberResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Nil(t, response.ActivationToken)
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, models.LoginStatusActive, login.Status)
		require.Equal(t, "tournament_director", login.Role)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("existing-password")))
		var sessions int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", staleSession).Count(&sessions).Error)
		require.Zero(t, sessions)
		verify, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/verify", map[string]string{"token": staleToken})
		require.NoError(t, err)
		verified := httptest.NewRecorder()
		router.ServeHTTP(verified, verify)
		require.Equal(t, http.StatusUnauthorized, verified.Code)
	})

	t.Run("disabled account returns active with same password and no activation", func(t *testing.T) {
		reset(t, "Quest-Disabled")
		password, err := bcrypt.GenerateFromPassword([]byte("valid-old-password"), bcrypt.DefaultCost)
		require.NoError(t, err)
		require.NoError(t, db.Create(&models.Login{Username: member.QuestID, Password: string(password), Role: "executive", Status: models.LoginStatusDisabled}).Error)
		staleToken := createPostgresActivation(t, db, member.QuestID)
		staleSession := uuid.New()
		require.NoError(t, db.Create(&models.Session{ID: staleSession, Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
		w := request(t, semester.ID, member.QuestID, "executive")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var response models.AddExecutiveTeamMemberResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Nil(t, response.ActivationToken)
		require.Equal(t, models.LoginStatusActive, response.AccountStatus)
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, models.LoginStatusActive, login.Status)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("valid-old-password")))
		var sessions int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", staleSession).Count(&sessions).Error)
		require.Zero(t, sessions)
		verify, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/verify", map[string]string{"token": staleToken})
		require.NoError(t, err)
		verified := httptest.NewRecorder()
		router.ServeHTTP(verified, verify)
		require.Equal(t, http.StatusUnauthorized, verified.Code)
	})

	t.Run("pending account rotates token without replacing password hash", func(t *testing.T) {
		reset(t, "Quest-Pending")
		password, err := bcrypt.GenerateFromPassword([]byte("pending-password"), bcrypt.DefaultCost)
		require.NoError(t, err)
		require.NoError(t, db.Create(&models.Login{Username: member.QuestID, Password: string(password), Role: "executive", Status: models.LoginStatusPendingActivation}).Error)
		oldToken := createPostgresActivation(t, db, member.QuestID)
		staleSession := uuid.New()
		require.NoError(t, db.Create(&models.Session{ID: staleSession, Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error)
		w := request(t, semester.ID, member.QuestID, "tournament_director")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var response models.AddExecutiveTeamMemberResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.NotEmpty(t, response.ActivationToken)
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("pending-password")))
		var sessions int64
		require.NoError(t, db.Model(&models.Session{}).Where("id = ?", staleSession).Count(&sessions).Error)
		require.Zero(t, sessions)
		verify := func(token string) int {
			req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/activations/verify", map[string]string{"token": token})
			require.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			return w.Code
		}
		require.Equal(t, http.StatusUnauthorized, verify(oldToken))
		require.Equal(t, http.StatusOK, verify(*response.ActivationToken))
	})

	t.Run("protected roles fail before writes", func(t *testing.T) {
		reset(t, "Quest-Protected")
		for _, role := range []string{"president", "secretary", "treasurer", "vice_president", "webmaster", "bot"} {
			w := request(t, semester.ID, member.QuestID, role)
			require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		}
		var count int64
		for _, role := range []string{"president", "secretary", "treasurer", "vice_president", "webmaster", "bot"} {
			reset(t, "Quest-Protected")
			require.NoError(t, db.Create(&models.Login{Username: member.QuestID, Password: "keep", Role: role, Status: models.LoginStatusActive}).Error)
			w := request(t, semester.ID, member.QuestID, "executive")
			require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			require.NoError(t, db.Model(&models.AccountActivation{}).Count(&count).Error)
			require.Zero(t, count)
			require.NoError(t, db.Model(&models.Membership{}).Count(&count).Error)
			require.Zero(t, count)
		}
	})

	t.Run("staged officer-transition account is left untouched", func(t *testing.T) {
		reset(t, "Quest-Staged")
		transition := models.OfficerTransition{ID: uuid.New(), InitiatedBy: "auth-secretary", PresidentUsername: "pres", VicePresidentUsername: "vp", SecretaryUsername: "sec", TreasurerUsername: "treas", Status: models.OfficerTransitionPending}
		require.NoError(t, db.Create(&transition).Error)
		require.NoError(t, db.Create(&models.Login{Username: member.QuestID, Password: "unchanged", Role: "executive", Status: models.LoginStatusPendingActivation, StagedTransitionID: &transition.ID}).Error)
		w := request(t, semester.ID, member.QuestID, "tournament_director")
		require.Equal(t, http.StatusConflict, w.Code, w.Body.String())
		var login models.Login
		require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
		require.Equal(t, "unchanged", login.Password)
		require.Equal(t, "executive", login.Role)
		var count int64
		require.NoError(t, db.Model(&models.Membership{}).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Model(&models.AccountActivation{}).Count(&count).Error)
		require.Zero(t, count)
	})

	t.Run("already executive missing identity and semester return typed errors", func(t *testing.T) {
		reset(t, "Quest-Duplicate")
		require.NoError(t, db.Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: semester.ID, Executive: true}).Error)
		w := request(t, semester.ID, member.QuestID, "executive")
		require.Equal(t, http.StatusConflict, w.Code, w.Body.String())
		var count int64
		require.NoError(t, db.Model(&models.Login{}).Where("username = ?", "Quest-Duplicate").Count(&count).Error)
		require.Zero(t, count)
		w = request(t, semester.ID, "missing-quest", "executive")
		require.Equal(t, http.StatusNotFound, w.Code)
		w = request(t, semester.ID, "  ", "executive")
		require.Equal(t, http.StatusBadRequest, w.Code)
		w = request(t, uuid.New(), member.QuestID, "executive")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("concurrent external account creation is typed conflict or protected-role rejection", func(t *testing.T) {
		reset(t, "Quest-External-Race")
		webmasterSession, err := testutils.CreateTestSession(db, "auth-webmaster", "webmaster")
		require.NoError(t, err)
		addRequest, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/semesters/"+semester.ID.String()+"/executive-team", map[string]string{"questId": member.QuestID, "role": "executive"})
		require.NoError(t, err)
		testutils.SetAuthCookie(addRequest, session)
		loginRequest, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/logins", map[string]string{"username": member.QuestID, "password": "external-password", "role": "president"})
		require.NoError(t, err)
		testutils.SetAuthCookie(loginRequest, webmasterSession)
		var wg sync.WaitGroup
		responses := make(chan struct{ add, login int }, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			router.ServeHTTP(w, addRequest)
			responses <- struct{ add, login int }{add: w.Code}
		}()
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			router.ServeHTTP(w, loginRequest)
			responses <- struct{ add, login int }{login: w.Code}
		}()
		wg.Wait()
		close(responses)
		var addCode, loginCode int
		for result := range responses {
			if result.add != 0 {
				addCode = result.add
			}
			if result.login != 0 {
				loginCode = result.login
			}
		}
		require.Contains(t, []int{http.StatusCreated, http.StatusForbidden, http.StatusConflict}, addCode)
		require.Contains(t, []int{http.StatusCreated, http.StatusConflict}, loginCode)
		require.NotEqual(t, http.StatusInternalServerError, addCode)
		require.NotEqual(t, http.StatusInternalServerError, loginCode)
		if addCode == http.StatusCreated {
			require.Equal(t, http.StatusConflict, loginCode)
			var login models.Login
			require.NoError(t, db.First(&login, "username = ?", "Quest-External-Race").Error)
			require.Equal(t, models.LoginStatusPendingActivation, login.Status)
			require.NotEqual(t, "president", login.Role)
		} else {
			require.Equal(t, http.StatusCreated, loginCode)
			var login models.Login
			require.NoError(t, db.First(&login, "username = ?", member.QuestID).Error)
			require.Equal(t, "president", login.Role)
			var count int64
			require.NoError(t, db.Model(&models.Membership{}).Count(&count).Error)
			require.Zero(t, count)
		}
	})

	t.Run("failed membership, activation, and commit roll back all writes and token response", func(t *testing.T) {
		reset(t, "Quest-Rollback")
		require.NoError(t, db.Exec(`CREATE FUNCTION reject_membership_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced membership failure'; END $$`).Error)
		require.NoError(t, db.Exec(`CREATE TRIGGER reject_membership_insert BEFORE INSERT ON memberships FOR EACH ROW EXECUTE FUNCTION reject_membership_insert()`).Error)
		w := request(t, semester.ID, member.QuestID, "executive")
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.NotContains(t, w.Body.String(), "activationToken")
		var count int64
		require.NoError(t, db.Model(&models.Login{}).Where("username = ?", member.QuestID).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Model(&models.AccountActivation{}).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Exec(`DROP TRIGGER reject_membership_insert ON memberships`).Error)
		require.NoError(t, db.Exec(`DROP FUNCTION reject_membership_insert()`).Error)

		require.NoError(t, db.Exec(`CREATE FUNCTION reject_activation_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced activation failure'; END $$`).Error)
		require.NoError(t, db.Exec(`CREATE TRIGGER reject_activation_insert BEFORE INSERT ON account_activations FOR EACH ROW EXECUTE FUNCTION reject_activation_insert()`).Error)
		w = request(t, semester.ID, member.QuestID, "executive")
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.NotContains(t, w.Body.String(), "activationToken")
		require.NoError(t, db.Model(&models.Login{}).Where("username = ?", member.QuestID).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Model(&models.Membership{}).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Exec(`DROP TRIGGER reject_activation_insert ON account_activations`).Error)
		require.NoError(t, db.Exec(`DROP FUNCTION reject_activation_insert()`).Error)

		require.NoError(t, db.Exec(`CREATE FUNCTION reject_activation_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced commit failure'; END $$`).Error)
		require.NoError(t, db.Exec(`CREATE CONSTRAINT TRIGGER reject_activation_commit AFTER INSERT ON account_activations DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_activation_commit()`).Error)
		w = request(t, semester.ID, member.QuestID, "executive")
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.NotContains(t, w.Body.String(), "activationToken")
		require.NoError(t, db.Model(&models.Login{}).Where("username = ?", member.QuestID).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Model(&models.Membership{}).Count(&count).Error)
		require.Zero(t, count)
		require.NoError(t, db.Exec(`DROP TRIGGER reject_activation_commit ON account_activations`).Error)
		require.NoError(t, db.Exec(`DROP FUNCTION reject_activation_commit()`).Error)
	})

	t.Run("concurrent account creation returns one conflict", func(t *testing.T) {
		reset(t, "Quest-Race")
		requests := make([]*http.Request, 2)
		for i := range requests {
			req, err := testutils.MakeJSONRequest(http.MethodPost, "/api/v2/semesters/"+semester.ID.String()+"/executive-team", map[string]string{"questId": member.QuestID, "role": "executive"})
			require.NoError(t, err)
			testutils.SetAuthCookie(req, session)
			requests[i] = req
		}
		var wg sync.WaitGroup
		codes := make(chan int, 2)
		for _, req := range requests {
			wg.Add(1)
			go func(req *http.Request) {
				defer wg.Done()
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				codes <- w.Code
			}(req)
		}
		wg.Wait()
		close(codes)
		var created, conflict int
		for code := range codes {
			if code == http.StatusCreated {
				created++
			} else if code == http.StatusConflict {
				conflict++
			} else {
				t.Errorf("concurrent add returned %d", code)
			}
		}
		require.Equal(t, 1, created)
		require.Equal(t, 1, conflict)
	})

	t.Run("paid membership promotion refunds existing fee", func(t *testing.T) {
		reset(t, "Quest-Paid")
		semester.CurrentBudget = 50
		require.NoError(t, db.Model(&semester).Update("current_budget", semester.CurrentBudget).Error)
		require.NoError(t, db.Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: semester.ID, Paid: true}).Error)
		w := request(t, semester.ID, member.QuestID, "executive")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		require.NoError(t, db.First(&semester, "id = ?", semester.ID).Error)
		require.Equal(t, float32(0), semester.CurrentBudget)
		var membership models.Membership
		require.NoError(t, db.First(&membership, "user_id = ? AND semester_id = ?", member.ID, semester.ID).Error)
		require.True(t, membership.Executive)
		require.False(t, membership.Paid)
	})
}

func createPostgresActivation(t *testing.T, db *gorm.DB, username string) string {
	t.Helper()
	tx, err := postgres.NewStore(db).BeginTx()
	require.NoError(t, err)
	token, err := services.NewAccountActivationService(tx).Create(username)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	return token
}
