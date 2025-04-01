package authorization

// loginAuthorizer is a struct that implements the ResourceAuthorizer interface.
type loginAuthorizer struct{}

// NewLoginAuthorizer creates a new login authorizer.
func NewLoginAuthorizer() ResourceAuthorizer {
	return &loginAuthorizer{}
}

// IsAuthorized checks if the user is authorized to perform the action.
func (svc *loginAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasRole(ROLE_WEBMASTER, role)
	}

	return false
}
