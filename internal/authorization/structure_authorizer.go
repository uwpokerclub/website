package authorization

// StructureAuthorizer is an interface that defines the methods for authorizing structure resources.
type structureAuthorizer struct{}

// NewStructureAuthorizer creates a new structure authorizer.
func NewStructureAuthorizer() ResourceAuthorizer {
	return &structureAuthorizer{}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a structure.
func (svc *structureAuthorizer) IsAuthorized(role, action string) bool {
	switch action {
	case "create":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "edit":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	}

	return false
}
