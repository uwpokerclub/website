package authorization

type transitionAuthorizer struct{ actions []string }

func NewTransitionAuthorizer() ResourceAuthorizer {
	return &transitionAuthorizer{actions: []string{"create", "get", "cancel", "reissue"}}
}
func (a *transitionAuthorizer) IsAuthorized(role, action string) bool {
	switch action {
	case "create", "cancel", "reissue":
		return HasAtleastRole(ROLE_PRESIDENT, role)
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	}
	return false
}
func (a *transitionAuthorizer) GetPermissions(role string) map[string]any {
	return map[string]any{"create": a.IsAuthorized(role, "create"), "get": a.IsAuthorized(role, "get"), "cancel": a.IsAuthorized(role, "cancel"), "reissue": a.IsAuthorized(role, "reissue")}
}
