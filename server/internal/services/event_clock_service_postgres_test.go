package services_test

import (
	"context"
	"testing"
	"time"

	"api/internal/models"
	"api/internal/services"
	"api/internal/store/postgres"
	"api/internal/testutils"

	"github.com/stretchr/testify/require"
)

// TestEventClockService_Pause_RetriesAfterLosingLazyCreateRace deterministically
// forces the exact race in-memory transactions cannot reproduce (they snapshot
// at BeginTx and never observe another transaction's later commit): a
// concurrent transaction commits the clock's very first row while our
// transaction is mid-flight, and our own INSERT ... ON CONFLICT DO NOTHING
// only resolves - as a conflict - after that commit. The service must retry
// against the winner's row instead of propagating store.ErrAlreadyExists.
func TestEventClockService_Pause_RetriesAfterLosingLazyCreateRace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	t.Cleanup(func() { container.Close(ctx) })

	db := container.GetDB()

	semester, err := testutils.CreateTestSemester(db, "Race Retry Semester")
	require.NoError(t, err)
	structure, err := testutils.CreateTestStructure(db, "Race Retry Structure")
	require.NoError(t, err)
	event, err := testutils.CreateTestEvent(db, semester.ID, structure.ID, "Race Retry Event")
	require.NoError(t, err)

	// The "winner": a separate, still-open transaction that gets to the
	// clock row first. It creates a *running* clock, so the losing
	// transaction's Pause - once it retries against this row - performs a
	// real, observable state change rather than a no-op.
	winnerTx := db.Begin()
	t.Cleanup(func() { winnerTx.Rollback() })
	winnerRepo := postgres.NewEventClockRepository(winnerTx)
	now := time.Now().UTC()
	require.NoError(t, winnerRepo.Create(&models.EventClock{
		EventID:     event.ID,
		LevelIndex:  0,
		LevelEndsAt: now.Add(10 * time.Minute),
		PausedAt:    nil,
		Version:     1,
		UpdatedAt:   now,
	}))

	// The "loser": the service, given a fresh (uncommitted-winner-blind)
	// store, running in its own goroutine. Its internal transaction's
	// lazy-create INSERT blocks on the winner's uncommitted row.
	svc := services.NewEventClockService(postgres.NewStore(db))
	result := make(chan error, 1)
	go func() {
		_, err := svc.Pause(event.ID)
		result <- err
	}()

	select {
	case err := <-result:
		t.Fatalf("Pause must block on the winner's uncommitted row, not return early (returned: %v)", err)
	case <-time.After(200 * time.Millisecond):
	}

	require.NoError(t, winnerTx.Commit().Error)

	select {
	case err := <-result:
		require.NoError(t, err, "losing the lazy-create race must retry against the winner's row, not fail")
	case <-time.After(5 * time.Second):
		t.Fatal("Pause never returned after the winning transaction committed")
	}

	stored, err := postgres.NewEventClockRepository(db).FindByEventID(event.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.PausedAt, "the retried Pause must have applied to the winner's running clock")
	require.Equal(t, int64(2), stored.Version, "the retried Pause is a real state change on top of the winner's row, not a no-op")
}
