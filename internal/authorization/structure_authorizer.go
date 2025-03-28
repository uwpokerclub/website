package authorization

type structureAuthorizer struct{}

func NewStructureAuthorizer() ResourceAuthorizer {
	return &structureAuthorizer{}
}

func (svc *structureAuthorizer) IsAuthorized(role, action string) bool {
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
