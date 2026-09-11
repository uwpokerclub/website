package services_test

import (
	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTransitionPresidentActivationCompletesAndInvalidatesSessions(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	transition, tokens := seedActivationTransition(t, db)

	login, sessionID, err := services.NewAccountActivationService(postgres.NewStore(db)).Complete(tokens["president"], "newpassword")
	require.NoError(t, err)
	require.Equal(t, "president", login.Role)

	var session models.Session
	require.NoError(t, db.First(&session, "id = ?", sessionID).Error)
	require.Equal(t, "president", session.Role)
	var completed models.OfficerTransition
	require.NoError(t, db.First(&completed, "id = ?", transition.ID).Error)
	require.Equal(t, "completed", completed.Status)
	require.NotNil(t, completed.ResolvedAt)

	assertLogin := func(username, role, status string) {
		var actual models.Login
		require.NoError(t, db.First(&actual, "username = ?", username).Error)
		require.Equal(t, role, actual.Role, username)
		require.Equal(t, status, actual.Status, username)
	}
	assertLogin("president", "president", models.LoginStatusActive)
	assertLogin("vp", "vice_president", models.LoginStatusPendingActivation)
	assertLogin("secretary", "secretary", models.LoginStatusActive)
	assertLogin("treasurer", "treasurer", models.LoginStatusPendingActivation)
	assertLogin("outgoing", "president", models.LoginStatusDisabled)
	assertLogin("director", "tournament_director", models.LoginStatusDisabled)
	assertLogin("webmaster", "webmaster", models.LoginStatusActive)
	assertLogin("bot", "bot", models.LoginStatusActive)

	var stale int64
	require.NoError(t, db.Model(&models.Session{}).Where("username IN ? AND id <> ?", []string{"president", "vp", "secretary", "treasurer", "outgoing", "director"}, sessionID).Count(&stale).Error)
	require.Zero(t, stale)
	var protected int64
	require.NoError(t, db.Model(&models.Session{}).Where("username IN ?", []string{"webmaster", "bot"}).Count(&protected).Error)
	require.EqualValues(t, 2, protected)
}

func TestTransitionBoundNonPresidentActivatesWithoutCompleting(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	transition, tokens := seedActivationTransition(t, db)

	login, _, err := services.NewAccountActivationService(postgres.NewStore(db)).Complete(tokens["vp"], "newpassword")
	require.NoError(t, err)
	require.Equal(t, "executive", login.Role)
	var actual models.Login
	require.NoError(t, db.First(&actual, "username = ?", "vp").Error)
	require.Equal(t, models.LoginStatusActive, actual.Status)
	var pending models.OfficerTransition
	require.NoError(t, db.First(&pending, "id = ?", transition.ID).Error)
	require.Equal(t, models.OfficerTransitionPending, pending.Status)
}

func TestTransitionBoundTokenRejectsCancelledOrMismatchedTransition(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	transition, tokens := seedActivationTransition(t, db)
	require.NoError(t, db.Model(&models.OfficerTransition{}).Where("id = ?", transition.ID).Update("status", "cancelled").Error)
	_, _, err = services.NewAccountActivationService(postgres.NewStore(db)).Complete(tokens["president"], "newpassword")
	require.ErrorIs(t, err, services.ErrActivationNotFound)
	assertUnusedTransitionToken(t, db, tokens["president"])

	transition, _ = seedActivationTransition(t, db)
	require.NoError(t, db.Create(&models.Login{Username: "outsider", Password: "old", Role: "executive", Status: models.LoginStatusPendingActivation}).Error)
	outsiderToken := createBoundToken(t, db, "outsider", transition.ID)
	_, _, err = services.NewAccountActivationService(postgres.NewStore(db)).Complete(outsiderToken, "newpassword")
	require.ErrorIs(t, err, services.ErrActivationNotFound)
	assertUnusedTransitionToken(t, db, outsiderToken)
}

func TestTransitionPresidentActivationDoesNotReactivateDisabledLogin(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	transition, tokens := seedActivationTransition(t, db)
	require.NoError(t, db.Model(&models.Login{}).Where("username = ?", "president").Update("status", models.LoginStatusDisabled).Error)

	_, _, err = services.NewAccountActivationService(postgres.NewStore(db)).Complete(tokens["president"], "newpassword")
	require.ErrorIs(t, err, services.ErrActivationNotFound)
	assertUnusedTransitionToken(t, db, tokens["president"])
	var pending models.OfficerTransition
	require.NoError(t, db.First(&pending, "id = ?", transition.ID).Error)
	require.Equal(t, models.OfficerTransitionPending, pending.Status)
	var president models.Login
	require.NoError(t, db.First(&president, "username = ?", "president").Error)
	require.Equal(t, models.LoginStatusDisabled, president.Status)
	require.Equal(t, "vice_president", president.Role)
}

func TestTransitionCompletionRollsBackAfterClaimFailure(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	transition, tokens := seedActivationTransition(t, db)
	require.NoError(t, db.Exec(`CREATE FUNCTION fail_transition_role_update() RETURNS trigger AS $$
	BEGIN
	  IF NEW.username = 'vp' THEN RAISE EXCEPTION 'forced role update failure'; END IF;
	  RETURN NEW;
	END;
	$$ LANGUAGE plpgsql`).Error)
	require.NoError(t, db.Exec(`CREATE TRIGGER fail_transition_role_update BEFORE UPDATE ON logins
	FOR EACH ROW EXECUTE FUNCTION fail_transition_role_update()`).Error)

	_, _, err = services.NewAccountActivationService(postgres.NewStore(db)).Complete(tokens["president"], "newpassword")
	require.Error(t, err)
	require.False(t, errors.Is(err, services.ErrActivationNotFound))
	assertUnusedTransitionToken(t, db, tokens["president"])
	var pending models.OfficerTransition
	require.NoError(t, db.First(&pending, "id = ?", transition.ID).Error)
	require.Equal(t, models.OfficerTransitionPending, pending.Status)
	var president models.Login
	require.NoError(t, db.First(&president, "username = ?", "president").Error)
	require.Equal(t, "vice_president", president.Role)
	require.Equal(t, models.LoginStatusActive, president.Status)
	var sessions int64
	require.NoError(t, db.Model(&models.Session{}).Where("username IN ?", []string{"president", "vp", "secretary", "treasurer", "outgoing", "director"}).Count(&sessions).Error)
	require.EqualValues(t, 6, sessions)
}

func TestTransitionCompletionLosesToGuardedCancelWithoutPartialWrites(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)
	db := container.GetDB()
	transition, tokens := seedActivationTransition(t, db)

	lock := db.Begin()
	require.NoError(t, lock.Error)
	var locked models.OfficerTransition
	require.NoError(t, lock.Raw("SELECT * FROM officer_transitions WHERE id = ? FOR UPDATE", transition.ID).Scan(&locked).Error)
	result := make(chan error, 1)
	go func() {
		_, _, e := services.NewAccountActivationService(postgres.NewStore(db)).Complete(tokens["president"], "newpassword")
		result <- e
	}()
	// Completion reads the token then blocks on this row lock. #400 cancellation
	// claims the transition before invalidating its transition-bound tokens.
	cancel := lock.Exec("UPDATE officer_transitions SET status = 'cancelled' WHERE id = ? AND status = 'pending'", transition.ID)
	require.NoError(t, cancel.Error)
	require.EqualValues(t, 1, cancel.RowsAffected)
	require.NoError(t, lock.Where("transition_id = ?", transition.ID).Delete(&models.AccountActivation{}).Error)
	require.NoError(t, lock.Commit().Error)
	require.ErrorIs(t, <-result, services.ErrActivationNotFound)

	var cancelled models.OfficerTransition
	require.NoError(t, db.First(&cancelled, "id = ?", transition.ID).Error)
	require.Equal(t, "cancelled", cancelled.Status)
	var activation models.AccountActivation
	require.ErrorIs(t, db.First(&activation, "token_hash = ?", testTokenHash(tokens["president"])).Error, gorm.ErrRecordNotFound)
	var president models.Login
	require.NoError(t, db.First(&president, "username = ?", "president").Error)
	require.Equal(t, "vice_president", president.Role)
	require.Equal(t, models.LoginStatusActive, president.Status)
}

func seedActivationTransition(t *testing.T, db *gorm.DB) (models.OfficerTransition, map[string]string) {
	t.Helper()
	require.NoError(t, db.Exec("TRUNCATE account_activations, officer_transitions, sessions, logins CASCADE").Error)
	logins := []models.Login{
		{Username: "president", Password: "old", Role: "vice_president", Status: models.LoginStatusActive},
		{Username: "vp", Password: "old", Role: "executive", Status: models.LoginStatusPendingActivation},
		{Username: "secretary", Password: "old", Role: "executive", Status: models.LoginStatusActive},
		{Username: "treasurer", Password: "old", Role: "executive", Status: models.LoginStatusPendingActivation},
		{Username: "outgoing", Password: "old", Role: "president", Status: models.LoginStatusActive},
		{Username: "director", Password: "old", Role: "tournament_director", Status: models.LoginStatusActive},
		{Username: "webmaster", Password: "old", Role: "webmaster", Status: models.LoginStatusActive},
		{Username: "bot", Password: "old", Role: "bot", Status: models.LoginStatusActive},
	}
	for i := range logins {
		require.NoError(t, db.Create(&logins[i]).Error)
	}
	transition := models.OfficerTransition{ID: uuid.New(), InitiatedBy: "outgoing", PresidentUsername: "president", VicePresidentUsername: "vp", SecretaryUsername: "secretary", TreasurerUsername: "treasurer", Status: models.OfficerTransitionPending}
	require.NoError(t, db.Create(&transition).Error)
	for _, username := range []string{"president", "vp", "secretary", "treasurer", "outgoing", "director", "webmaster", "bot"} {
		require.NoError(t, db.Create(&models.Session{Username: username, Role: "executive", StartedAt: time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2100, 1, 1, 8, 0, 0, 0, time.UTC)}).Error)
	}
	return transition, map[string]string{
		"president": createBoundToken(t, db, "president", transition.ID),
		"vp":        createBoundToken(t, db, "vp", transition.ID),
	}
}

func createBoundToken(t *testing.T, db *gorm.DB, username string, transitionID uuid.UUID) string {
	t.Helper()
	token := uuid.NewString()
	require.NoError(t, db.Create(&models.AccountActivation{TokenHash: testTokenHash(token), Username: username, TransitionID: &transitionID, ExpiresAt: time.Date(2100, 1, 2, 0, 0, 0, 0, time.UTC)}).Error)
	return token
}

func assertUnusedTransitionToken(t *testing.T, db *gorm.DB, token string) {
	t.Helper()
	var activation models.AccountActivation
	require.NoError(t, db.First(&activation, "token_hash = ?", testTokenHash(token)).Error)
	require.Nil(t, activation.UsedAt)
}

func testTokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
