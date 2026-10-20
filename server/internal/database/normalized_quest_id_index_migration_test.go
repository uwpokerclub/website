package database_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"api/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestNormalizedQuestIDIndexOnMigratedSchema(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	db := container.GetSQLDB()
	var indexDefinition string
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT indexdef FROM pg_indexes
		WHERE schemaname = current_schema() AND indexname = 'users_quest_id_unique'
	`).Scan(&indexDefinition))
	require.Contains(t, indexDefinition, "lower(btrim(quest_id))")
	require.Contains(t, indexDefinition, "WHERE (btrim(quest_id) <> ''::text)")

	insert := func(id int64, questID any) error {
		_, err := db.ExecContext(ctx, `INSERT INTO users (id, quest_id) VALUES ($1, $2)`, id, questID)
		return err
	}
	require.NoError(t, insert(1, "CaseTest"))
	require.Error(t, insert(2, "casetest"), "case variants must collide")
	require.NoError(t, insert(3, " space test "))
	require.Error(t, insert(4, "space test"), "btrim variants must collide")
	require.NoError(t, insert(5, "update-target"))
	require.NoError(t, insert(6, "update-source"))
	_, err = db.ExecContext(ctx, `UPDATE users SET quest_id = ' UPDATE-TARGET ' WHERE id = 6`)
	require.Error(t, err, "an update that creates a normalized collision must fail")
	var sourceQuestID string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT quest_id FROM users WHERE id = 6`).Scan(&sourceQuestID))
	require.Equal(t, "update-source", sourceQuestID, "a failed update must preserve the old row value")

	for id, questID := range []any{nil, nil, "", "", "   ", "   "} {
		require.NoError(t, insert(int64(10+id), questID), "blank Quest IDs remain unrestricted")
	}
}

func TestNormalizedQuestIDIndexMigrationWithLegacyRows(t *testing.T) {
	ctx := context.Background()
	container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
	require.NoError(t, err)
	defer container.Close(ctx)

	conn, err := container.GetSQLDB().Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationPaths, err := filepath.Glob(filepath.Join(filepath.Dir(currentFile), "../../atlas/migrations/20261019120000_normalized_user_quest_id_index.sql"))
	require.NoError(t, err)
	require.Len(t, migrationPaths, 1)
	migrationSQL, err := os.ReadFile(migrationPaths[0])
	require.NoError(t, err)

	createLegacyTable := func() {
		_, err := conn.ExecContext(ctx, `CREATE TEMP TABLE users (id bigint PRIMARY KEY, quest_id text NULL)`)
		require.NoError(t, err)
	}
	_, err = conn.ExecContext(ctx, `CREATE TEMP TABLE users (id bigint PRIMARY KEY, quest_id text NULL)`)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `INSERT INTO users (id, quest_id) VALUES
		(1, 'CleanID'), (2, ' clean-other '), (3, NULL), (4, NULL), (5, ''), (6, ''), (7, '   '), (8, '   ')`)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err, "legacy clean data should accept the migration")
	_, err = conn.ExecContext(ctx, `INSERT INTO users (id, quest_id) VALUES (9, ' cleanid ')`)
	require.Error(t, err, "the newly applied index must enforce normalized uniqueness")
	_, err = conn.ExecContext(ctx, `DROP TABLE users`)
	require.NoError(t, err)

	createLegacyTable()
	_, err = conn.ExecContext(ctx, `INSERT INTO users (id, quest_id) VALUES (11, 'DuplicateID'), (12, ' duplicateid ')`)
	require.NoError(t, err)
	const preflight = `SELECT lower(btrim(quest_id)) AS normalized, count(*), array_agg(id)::text
		FROM users WHERE btrim(quest_id) <> '' GROUP BY 1 HAVING count(*) > 1`
	var normalized, ids string
	var count int
	require.NoError(t, conn.QueryRowContext(ctx, preflight).Scan(&normalized, &count, &ids))
	require.Equal(t, "duplicateid", normalized)
	require.Equal(t, 2, count)
	require.Equal(t, "{11,12}", ids)

	_, err = conn.ExecContext(ctx, string(migrationSQL))
	require.Error(t, err, "the index must fail while normalized duplicates remain")
	var unchanged int
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE id IN (11, 12) AND quest_id IN ('DuplicateID', ' duplicateid ')`).Scan(&unchanged))
	require.Equal(t, 2, unchanged, "failed index creation must preserve every legacy row")
	var indexCount int
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT count(*) FROM pg_indexes WHERE schemaname LIKE 'pg_temp_%' AND indexname = 'users_quest_id_unique'`).Scan(&indexCount))
	require.Zero(t, indexCount, "failed index creation must not leave a partial index")

}
