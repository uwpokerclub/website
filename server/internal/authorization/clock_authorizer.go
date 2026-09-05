package authorization

// clockAuthorizer is an interface that defines the methods for authorizing an
// event's tournament clock.
type clockAuthorizer struct {
	actions []string
}

// NewClockAuthorizer creates a new clock authorizer.
func NewClockAuthorizer() ResourceAuthorizer {
	return &clockAuthorizer{
		actions: []string{"get", "control"},
	}
}

// IsAuthorized checks if a user with the given role is authorized to perform
// the specified action on an event's clock.
func (svc *clockAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "control":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	}

	return false
}

func (svc *clockAuthorizer) GetPermissions(role string) map[string]any {
	permissions := make(map[string]any)

	for _, action := range svc.actions {
		permissions[action] = svc.IsAuthorized(role, action)
	}

	return permissions
}
