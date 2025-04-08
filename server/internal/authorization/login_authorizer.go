package authorization

// loginAuthorizer is a struct that implements the ResourceAuthorizer interface.
type loginAuthorizer struct {
	actions []string
}

// NewLoginAuthorizer creates a new login authorizer.
func NewLoginAuthorizer() ResourceAuthorizer {
	return &loginAuthorizer{
		actions: []string{"create"},
	}
}

// IsAuthorized checks if the user is authorized to perform the action.
func (svc *loginAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasRole(ROLE_WEBMASTER, role)
	}

	return false
}

func (svc *loginAuthorizer) GetPermissions(role string) map[string]any {
	permissions := make(map[string]any)

	for _, action := range svc.actions {
		permissions[action] = svc.IsAuthorized(role, action)
	}

	return permissions
}
