package authorization

// RankingsAuthorizer is an interface that defines the methods for authorizing rankings.
type rankingsAuthorizer struct {
	actions []string
}

// NewRankingsAuthorizer creates a new rankings authorizer.
func NewRankingsAuthorizer() ResourceAuthorizer {
	return &rankingsAuthorizer{
		actions: []string{"get", "list", "export"},
	}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on rankings.
func (svc *rankingsAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "get":
		return HasAtleastRole(ROLE_BOT, role)
	case "list":
		return HasAtleastRole(ROLE_BOT, role)
	case "export":
		return HasAtleastRole(ROLE_SECRETARY, role)
	}

	return false
}

func (svc *rankingsAuthorizer) GetPermissions(role string) map[string]any {
	permissions := make(map[string]any)

	for _, action := range svc.actions {
		permissions[action] = svc.IsAuthorized(role, action)
	}

	return permissions
}
