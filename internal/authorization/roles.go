package authorization

type role int

// Role represents the role of a user in the system.
// The roles are ordered from lowest to highest privilege.
const (
	ROLE_BOT                 role = iota
	ROLE_EXECUTIVE           role = iota
	ROLE_TOURNAMENT_DIRECTOR role = iota
	ROLE_SECRETARY           role = iota
	ROLE_TREASURER           role = iota
	ROLE_VICE_PRESIDENT      role = iota
	ROLE_PRESIDENT           role = iota
	ROLE_WEBMASTER           role = iota
)

// ToString converts a role to a string.
// This is used for storing the role in the database.
func (r role) ToString() string {
	switch r {
	case ROLE_BOT:
		return "bot"
	case ROLE_EXECUTIVE:
		return "executive"
	case ROLE_TOURNAMENT_DIRECTOR:
		return "tournament_director"
	case ROLE_SECRETARY:
		return "secretary"
	case ROLE_TREASURER:
		return "treasurer"
	case ROLE_VICE_PRESIDENT:
		return "vice_president"
	case ROLE_PRESIDENT:
		return "president"
	case ROLE_WEBMASTER:
		return "webmaster"
	}
	return ""
}

// ToRole converts a string to a role.
// This is used for converting the role from the database to a role.
func stringToRole(r string) role {
	switch r {
	case ROLE_BOT.ToString():
		return ROLE_BOT
	case ROLE_EXECUTIVE.ToString():
		return ROLE_EXECUTIVE
	case ROLE_TOURNAMENT_DIRECTOR.ToString():
		return ROLE_TOURNAMENT_DIRECTOR
	case ROLE_SECRETARY.ToString():
		return ROLE_SECRETARY
	case ROLE_TREASURER.ToString():
		return ROLE_TREASURER
	case ROLE_VICE_PRESIDENT.ToString():
		return ROLE_VICE_PRESIDENT
	case ROLE_PRESIDENT.ToString():
		return ROLE_PRESIDENT
	case ROLE_WEBMASTER.ToString():
		return ROLE_WEBMASTER
	}
	return -1
}

// HasRole checks if the user has the given role.
func HasRole(r role, userRole string) bool {
	return r == stringToRole(userRole)
}

// HasAtleastRole checks if the user has at least the given role.
func HasAtleastRole(r role, userRole string) bool {
	return r <= stringToRole(userRole)
}
