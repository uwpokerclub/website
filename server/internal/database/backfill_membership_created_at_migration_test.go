package database_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"api/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestBackfillMembershipCreatedAtMigrationIsSemesterScopedAndTimezoneIndependent(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	conn, err := container.GetSQLDB().Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	setupStatements := []string{
		`CREATE TEMP TABLE memberships (
			id integer PRIMARY KEY,
			semester_id integer NOT NULL,
			created_at timestamp NULL
		)`,
		`CREATE TEMP TABLE events (
			id integer PRIMARY KEY,
			semester_id integer NOT NULL,
			start_date timestamptz NOT NULL
		)`,
		`CREATE TEMP TABLE participants (
			membership_id integer NOT NULL,
			event_id integer NOT NULL
		)`,
		`INSERT INTO memberships (id, semester_id, created_at) VALUES
			(1, 10, NULL),
			(2, 10, NULL),
			(3, 10, NULL),
			(4, 10, TIMESTAMP '2026-09-06 18:00:00')`,
		`INSERT INTO events (id, semester_id, start_date) VALUES
			(101, 10, TIMESTAMPTZ '2025-11-02 04:30:00+00'),
			(102, 10, TIMESTAMPTZ '2025-11-02 05:30:00+00'),
			(201, 11, TIMESTAMPTZ '2025-10-01 19:00:00+00')`,
		`INSERT INTO participants (membership_id, event_id) VALUES
			(1, 101),
			(1, 102),
			(2, 201),
			(4, 101)`,
	}
	for _, statement := range setupStatements {
		_, err = conn.ExecContext(ctx, statement)
		require.NoError(t, err)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationPaths, err := filepath.Glob(filepath.Join(filepath.Dir(currentFile), "../../atlas/migrations/*_backfill_membership_created_at.sql"))
	require.NoError(t, err)
	require.Len(t, migrationPaths, 1)
	migrationSQL, err := os.ReadFile(migrationPaths[0])
	require.NoError(t, err)

	for _, timezone := range []string{"UTC", "America/Toronto"} {
		_, err = conn.ExecContext(ctx, "SET TIME ZONE '"+timezone+"'")
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, string(migrationSQL))
		require.NoError(t, err, "migration should execute with PostgreSQL TimeZone=%s", timezone)

		var dated, crossSemester, noEvent, realInsert sql.NullString
		require.NoError(t, conn.QueryRowContext(ctx, `
			SELECT
				(SELECT created_at::text FROM memberships WHERE id = 1),
				(SELECT created_at::text FROM memberships WHERE id = 2),
				(SELECT created_at::text FROM memberships WHERE id = 3),
				(SELECT created_at::text FROM memberships WHERE id = 4)
		`).Scan(&dated, &crossSemester, &noEvent, &realInsert))
		require.Equal(t, "2025-11-02 04:30:00", dated.String, "store the earliest event as a UTC wall-clock timestamp")
		require.False(t, crossSemester.Valid, "a malformed participant link to another semester must not backfill this membership")
		require.False(t, noEvent.Valid, "a membership without an event remains undated")
		require.Equal(t, "2026-09-06 18:00:00", realInsert.String, "preserve an existing real insertion time")

		_, err = conn.ExecContext(ctx, `UPDATE memberships SET created_at = NULL WHERE id IN (1, 2, 3)`)
		require.NoError(t, err)
	}
}
