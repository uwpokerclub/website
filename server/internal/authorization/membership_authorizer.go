package authorization

// ResourceAuthorizer is an interface that defines the methods for authorizing resources.
type membershipAuthorizer struct {
	actions []string
}

// NewMembershipAuthorizer creates a new membership authorizer.
func NewMembershipAuthorizer() ResourceAuthorizer {
	return &membershipAuthorizer{
		actions: []string{"create", "get", "list", "edit"},
	}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a membership.
func (svc *membershipAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role) || HasRole(ROLE_BOT, role)
	case "get":
		return HasAtleastRole(ROLE_BOT, role)
	case "list":
		return HasAtleastRole(ROLE_BOT, role)
	case "edit":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	}

	return false
}

func (svc *membershipAuthorizer) GetPermissions(role string) map[string]any {
	permissions := make(map[string]any)

	for _, action := range svc.actions {
		permissions[action] = svc.IsAuthorized(role, action)
	}

	return permissions
}
