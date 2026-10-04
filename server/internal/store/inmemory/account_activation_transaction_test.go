package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAccountActivationTransaction_RejectsStaleCommit(t *testing.T) {
	st := NewStore()
	hash := []byte("token-hash")
	require.NoError(t, st.AccountActivations().Create(&models.AccountActivation{
		TokenHash: hash, Username: "questid", ExpiresAt: time.Now().UTC().Add(time.Hour),
	}))

	first, err := st.BeginTx()
	require.NoError(t, err)
	second, err := st.BeginTx()
	require.NoError(t, err)
	require.NoError(t, consume(first, hash))
	require.NoError(t, consume(second, hash))
	require.NoError(t, first.Commit())
	require.ErrorIs(t, second.Commit(), store.ErrTransactionConflict)
}

func consume(tx store.Store, hash []byte) error {
	_, err := tx.AccountActivations().Consume(hash)
	return err
}
