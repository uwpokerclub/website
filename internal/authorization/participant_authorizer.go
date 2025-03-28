package authorization

type participantAuthorizer struct{}

func NewParticipantAuthorizer() ResourceAuthorizer {
	return &participantAuthorizer{}
}

func (svc *participantAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasRole(ROLE_EXECUTIVE, role)
	case "get":
		return HasRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasRole(ROLE_EXECUTIVE, role)
	case "signin":
		return HasRole(ROLE_EXECUTIVE, role)
	case "signout":
		return HasRole(ROLE_EXECUTIVE, role)
	case "delete":
		return HasRole(ROLE_EXECUTIVE, role)
	}

	return false
}
