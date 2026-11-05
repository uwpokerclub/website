package services

import (
	"api/internal/models"
	"api/internal/store"
	"api/internal/store/inmemory"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func executiveAddFixture(t *testing.T) (store.Store, uuid.UUID, models.User) {
	t.Helper()
	st := inmemory.NewStore()
	semester := models.Semester{ID: uuid.New(), Name: "Fall", StartDate: time.Now(), EndDate: time.Now().Add(24 * time.Hour), MembershipFee: 50, MembershipDiscountFee: 25}
	require.NoError(t, st.Semesters().Create(&semester))
	member := models.User{ID: 77101, FirstName: "Tess", LastName: "Test", Email: "tess@example.com", Faculty: models.FacultyMath, QuestID: "Quest-Id"}
	require.NoError(t, st.Members().Create(&member))
	return st, semester.ID, member
}

func TestExecutiveTeamService_AddNewMemberCreatesPendingActivationAndMembership(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	response, err := NewExecutiveTeamService(st).AddMember(semesterID, "quest-id", "tournament_director")
	require.NoError(t, err)
	require.Equal(t, models.AddExecutiveTeamMemberResponse{MemberID: member.ID, Username: member.QuestID, Role: "tournament_director", AccountStatus: models.LoginStatusPendingActivation, ActivationToken: response.ActivationToken}, response)
	require.NotNil(t, response.ActivationToken)
	login, err := st.Logins().FindByUsername(member.QuestID)
	require.NoError(t, err)
	require.Equal(t, models.LoginStatusPendingActivation, login.Status)
	require.Error(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("anything-known-to-user")), "new login password must be unusable")
	activation, err := NewAccountActivationService(st).Verify(*response.ActivationToken)
	require.NoError(t, err)
	wantHash := sha256.Sum256([]byte(*response.ActivationToken))
	require.Equal(t, wantHash[:], activation.TokenHash)
	require.WithinDuration(t, time.Now().UTC().Add(14*24*time.Hour), activation.ExpiresAt, 5*time.Second)
	membership, err := st.Memberships().FindByUserAndSemester(member.ID, semesterID)
	require.NoError(t, err)
	require.True(t, membership.Executive)

	tx, err := st.BeginTx()
	require.NoError(t, err)
	completed, _, err := NewAccountActivationService(tx).Complete(*response.ActivationToken, "password-set-through-flow")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, models.LoginStatusActive, completed.Status)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(completed.Password), []byte("password-set-through-flow")))
}

func TestExecutiveTeamService_AddActiveMemberPreservesPasswordAndInvalidatesChangedRoleSession(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	require.NoError(t, st.Logins().Create(&models.Login{Username: member.QuestID, Password: string(hash), Role: "executive", Status: models.LoginStatusActive}))
	session := models.Session{ID: uuid.New(), Username: member.QuestID, Role: "executive", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, st.Sessions().Create(&session))
	response, err := NewExecutiveTeamService(st).AddMember(semesterID, "QUEST-ID", "tournament_director")
	require.NoError(t, err)
	require.Nil(t, response.ActivationToken)
	login, err := st.Logins().FindByUsername(member.QuestID)
	require.NoError(t, err)
	require.Equal(t, models.LoginStatusActive, login.Status)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("current-password")))
	_, err = st.Sessions().FindByID(session.ID)
	require.Error(t, err)
	membership, err := st.Memberships().FindByUserAndSemester(member.ID, semesterID)
	require.NoError(t, err)
	require.True(t, membership.Executive)
}

func TestExecutiveTeamService_AddDisabledMemberReactivatesWithoutChangingPasswordOrKeepingTokens(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("still-valid-password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	require.NoError(t, st.Logins().Create(&models.Login{Username: member.QuestID, Password: string(hash), Role: "executive", Status: models.LoginStatusDisabled}))
	oldToken := createActivationInCallerTx(t, st, member.QuestID)
	response, err := NewExecutiveTeamService(st).AddMember(semesterID, "Quest-Id", "executive")
	require.NoError(t, err)
	require.Nil(t, response.ActivationToken)
	require.Equal(t, models.LoginStatusActive, response.AccountStatus)
	login, err := st.Logins().FindByUsername(member.QuestID)
	require.NoError(t, err)
	require.Equal(t, models.LoginStatusActive, login.Status)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("still-valid-password")))
	_, err = NewAccountActivationService(st).Verify(oldToken)
	require.ErrorIs(t, err, ErrActivationNotFound)
}

func TestExecutiveTeamService_AddPendingMemberRotatesActivationAndKeepsPasswordHash(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("pending-password"), bcrypt.DefaultCost)
	require.NoError(t, err)
	require.NoError(t, st.Logins().Create(&models.Login{Username: member.QuestID, Password: string(hash), Role: "executive", Status: models.LoginStatusPendingActivation}))
	oldToken := createActivationInCallerTx(t, st, member.QuestID)
	response, err := NewExecutiveTeamService(st).AddMember(semesterID, "Quest-Id", "tournament_director")
	require.NoError(t, err)
	require.NotNil(t, response.ActivationToken)
	_, err = NewAccountActivationService(st).Verify(oldToken)
	require.ErrorIs(t, err, ErrActivationNotFound)
	activation, err := NewAccountActivationService(st).Verify(*response.ActivationToken)
	require.NoError(t, err)
	wantHash := sha256.Sum256([]byte(*response.ActivationToken))
	require.Equal(t, wantHash[:], activation.TokenHash)
	login, err := st.Logins().FindByUsername(member.QuestID)
	require.NoError(t, err)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(login.Password), []byte("pending-password")))
}

func TestExecutiveTeamService_RejectsProtectedAndAlreadyExecutiveWithoutSideEffects(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	for _, role := range []string{"president", "webmaster", "bot"} {
		_, err := NewExecutiveTeamService(st).AddMember(semesterID, member.QuestID, role)
		require.ErrorIs(t, err, ErrExecutiveTeamInvalidRole)
	}
	require.NoError(t, st.Memberships().Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: semesterID, Executive: true}))
	_, err := NewExecutiveTeamService(st).AddMember(semesterID, member.QuestID, "executive")
	require.ErrorIs(t, err, ErrExecutiveTeamAlreadyMember)
	_, err = st.Logins().FindByUsername(member.QuestID)
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestExecutiveTeamService_RejectsProtectedExistingAndAmbiguousLogins(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	require.NoError(t, st.Logins().Create(&models.Login{Username: member.QuestID, Password: "protected", Role: "president", Status: models.LoginStatusActive}))
	_, err := NewExecutiveTeamService(st).AddMember(semesterID, member.QuestID, "executive")
	require.ErrorIs(t, err, ErrExecutiveTeamProtectedLogin)
	membership, err := st.Memberships().FindByUserAndSemester(member.ID, semesterID)
	require.Error(t, err)
	_ = membership

	st, semesterID, member = executiveAddFixture(t)
	require.NoError(t, st.Logins().Create(&models.Login{Username: member.QuestID, Password: "hash-a", Role: "executive", Status: models.LoginStatusActive}))
	_, err = NewExecutiveTeamService(st).AddMember(semesterID, member.QuestID, "executive")
	require.NoError(t, err)
}

func TestExecutiveTeamService_RejectsStagedLoginAndDoesNotChangeIt(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	transitionID := uuid.New()
	require.NoError(t, st.Logins().Create(&models.Login{Username: member.QuestID, Password: "unchanged", Role: "executive", Status: models.LoginStatusPendingActivation, StagedTransitionID: &transitionID}))
	_, err := NewExecutiveTeamService(st).AddMember(semesterID, member.QuestID, "tournament_director")
	require.ErrorIs(t, err, ErrExecutiveTeamStagedLogin)
	login, err := st.Logins().FindByUsername(member.QuestID)
	require.NoError(t, err)
	require.Equal(t, "unchanged", login.Password)
	require.Equal(t, "executive", login.Role)
	membership, err := st.Memberships().FindByUserAndSemester(member.ID, semesterID)
	require.Error(t, err)
	_ = membership
}

func TestExecutiveTeamService_PaidMembershipPromotionRefundsBudget(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	semester, err := st.Semesters().FindByID(semesterID)
	require.NoError(t, err)
	semester.CurrentBudget = float32(semester.MembershipFee)
	require.NoError(t, st.Semesters().Update(&semester))
	require.NoError(t, st.Memberships().Create(&models.Membership{ID: uuid.New(), UserID: member.ID, SemesterID: semesterID, Paid: true}))
	_, err = NewExecutiveTeamService(st).AddMember(semesterID, member.QuestID, "executive")
	require.NoError(t, err)
	semester, err = st.Semesters().FindByID(semesterID)
	require.NoError(t, err)
	require.Equal(t, float32(0), semester.CurrentBudget)
	membership, err := st.Memberships().FindByUserAndSemester(member.ID, semesterID)
	require.NoError(t, err)
	require.True(t, membership.Executive)
	require.False(t, membership.Paid)
}

func TestExecutiveTeamService_RejectsBlankIdentityAndMissingSemester(t *testing.T) {
	st, semesterID, member := executiveAddFixture(t)
	_, err := NewExecutiveTeamService(st).AddMember(semesterID, "  ", "executive")
	require.ErrorIs(t, err, ErrExecutiveTeamInvalidIdentity)
	_, err = NewExecutiveTeamService(st).AddMember(uuid.New(), member.QuestID, "executive")
	require.ErrorIs(t, err, ErrExecutiveTeamSemesterNotFound)
	_, err = st.Logins().FindByUsername(member.QuestID)
	require.ErrorIs(t, err, store.ErrNotFound)
}
