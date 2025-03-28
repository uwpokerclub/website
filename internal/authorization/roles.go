package authorization

type role int

const (
	ROLE_EXECUTIVE role = iota
)

func (r role) ToString() string {
	switch r {
	case ROLE_EXECUTIVE:
		return "executive"
	}
	return ""
}

func stringToRole(r string) role {
	switch r {
	case "executive":
		return ROLE_EXECUTIVE
	}
	return -1
}

func HasRole(r role, userRole string) bool {
	return r == stringToRole(userRole)
}
