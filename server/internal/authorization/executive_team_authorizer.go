package authorization

type executiveTeamAuthorizer struct{}

func NewExecutiveTeamAuthorizer() ResourceAuthorizer {
	return &executiveTeamAuthorizer{}
}

func (a *executiveTeamAuthorizer) IsAuthorized(role, action string) bool {
	return (action == "get" || action == "create" || action == "update") && HasAtleastRole(ROLE_SECRETARY, role)
}

func (a *executiveTeamAuthorizer) GetPermissions(role string) map[string]any {
	return map[string]any{
		"get":    a.IsAuthorized(role, "get"),
		"create": a.IsAuthorized(role, "create"),
		"update": a.IsAuthorized(role, "update"),
	}
}
