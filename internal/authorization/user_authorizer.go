package authorization

// UserAuthorizer is an interface that defines the methods for authorizing user resources.
type userAuthorizer struct {
	actions []string
}

// NewUserAuthorizer creates a new user authorizer.
func NewUserAuthorizer() ResourceAuthorizer {
	return &userAuthorizer{
		actions: []string{"create", "get", "list", "edit", "delete"},
	}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a user.
func (svc *userAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "edit":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "delete":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	}

	return false
}

func (svc *userAuthorizer) GetPermissions(role string) map[string]any {
	permissions := make(map[string]any)

	for _, action := range svc.actions {
		permissions[action] = svc.IsAuthorized(role, action)
	}

	return permissions
}
