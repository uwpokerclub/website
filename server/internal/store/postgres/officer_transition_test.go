package postgres

import (
	"api/internal/store"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestMapOfficerTransitionCreateError(t *testing.T) {
	err := mapOfficerTransitionCreateError(&pgconn.PgError{Code: "23505", ConstraintName: "one_pending_transition"})
	require.ErrorIs(t, err, store.ErrConflict)
	require.Error(t, mapOfficerTransitionCreateError(&pgconn.PgError{Code: "23505", ConstraintName: "other_unique_index"}))
}
