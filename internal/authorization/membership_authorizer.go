package authorization

type membershipAuthorizer struct{}

func NewMembershipAuthorizer() ResourceAuthorizer {
	return &membershipAuthorizer{}
}

func (svc *membershipAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasRole(ROLE_EXECUTIVE, role)
	case "get":
		return HasRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasRole(ROLE_EXECUTIVE, role)
	case "edit":
		return HasRole(ROLE_EXECUTIVE, role)
	}

	return false
}
