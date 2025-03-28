package authorization

type loginAuthorizer struct{}

func NewLoginAuthorizer() ResourceAuthorizer {
	return &loginAuthorizer{}
}

func (svc *loginAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasRole(ROLE_EXECUTIVE, role)
	}

	return false
}
