package authorization

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransitionAuthorizerCancelAndReissue(t *testing.T) {
	a := NewTransitionAuthorizer()
	require.True(t, a.IsAuthorized(ROLE_PRESIDENT.ToString(), "cancel"))
	require.True(t, a.IsAuthorized(ROLE_WEBMASTER.ToString(), "cancel"))
	require.False(t, a.IsAuthorized(ROLE_VICE_PRESIDENT.ToString(), "cancel"))

	require.True(t, a.IsAuthorized(ROLE_VICE_PRESIDENT.ToString(), "reissue"))
	require.True(t, a.IsAuthorized(ROLE_PRESIDENT.ToString(), "reissue"))
	require.True(t, a.IsAuthorized(ROLE_WEBMASTER.ToString(), "reissue"))
}

func TestTransitionAuthorizerGetIsLimitedToOfficersAndWebmaster(t *testing.T) {
	a := NewTransitionAuthorizer()

	for _, role := range []string{
		ROLE_SECRETARY.ToString(),
		ROLE_TREASURER.ToString(),
		ROLE_VICE_PRESIDENT.ToString(),
		ROLE_PRESIDENT.ToString(),
		ROLE_WEBMASTER.ToString(),
	} {
		require.True(t, a.IsAuthorized(role, "get"), role)
	}

	for _, role := range []string{
		ROLE_BOT.ToString(),
		ROLE_EXECUTIVE.ToString(),
		ROLE_TOURNAMENT_DIRECTOR.ToString(),
	} {
		require.False(t, a.IsAuthorized(role, "get"), role)
	}
}
