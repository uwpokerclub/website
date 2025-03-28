package authorization

type userAuthorizer struct{}

func NewUserAuthorizer() ResourceAuthorizer {
	return &userAuthorizer{}
}

func (svc *userAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasRole(ROLE_EXECUTIVE, role)
	case "get":
		return HasRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasRole(ROLE_EXECUTIVE, role)
	case "edit":
		return HasRole(ROLE_EXECUTIVE, role)
	case "delete":
		return HasRole(ROLE_EXECUTIVE, role)
	}

	return false
}
