package authorization

// ResourceAuthorizer is an interface that defines the methods for authorizing resources.
type membershipAuthorizer struct{}

// NewMembershipAuthorizer creates a new membership authorizer.
func NewMembershipAuthorizer() ResourceAuthorizer {
	return &membershipAuthorizer{}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a membership.
func (svc *membershipAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasAtleastRole(ROLE_BOT, role)
	case "edit":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	}

	return false
}
