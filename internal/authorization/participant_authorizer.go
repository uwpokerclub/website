package authorization

// participantAuthorizer is an interface that defines the methods for authorizing participants.
type participantAuthorizer struct{}

// NewParticipantAuthorizer creates a new participant authorizer.
func NewParticipantAuthorizer() ResourceAuthorizer {
	return &participantAuthorizer{}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a participant.
func (svc *participantAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "signin":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "signout":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "delete":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	}

	return false
}
