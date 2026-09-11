package authorization

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransitionAuthorizerCancelAndReissue(t *testing.T) {
	a := NewTransitionAuthorizer()
	for _, action := range []string{"cancel", "reissue"} {
		require.True(t, a.IsAuthorized(ROLE_PRESIDENT.ToString(), action))
		require.True(t, a.IsAuthorized(ROLE_WEBMASTER.ToString(), action))
		require.False(t, a.IsAuthorized(ROLE_VICE_PRESIDENT.ToString(), action))
	}
}
