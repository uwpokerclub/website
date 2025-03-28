package authorization

type rankingsAuthorizer struct{}

func NewRankingsAuthorizer() ResourceAuthorizer {
	return &rankingsAuthorizer{}
}

func (svc *rankingsAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "get":
		return HasRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasRole(ROLE_EXECUTIVE, role)
	case "export":
		return HasRole(ROLE_EXECUTIVE, role)
	}

	return false
}
