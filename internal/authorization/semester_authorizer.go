package authorization

import "strings"

// semesterAuthorizer is an interface that defines the methods for authorizing semesters.
type semesterAuthorizer struct {
	resourceAuthorizers ResourceAuthorizerMap
}

// NewSemesterAuthorizer creates a new semester authorizer.
func NewSemesterAuthorizer(resourceAuthorizers ResourceAuthorizerMap) ResourceAuthorizer {
	return &semesterAuthorizer{
		resourceAuthorizers: resourceAuthorizers,
	}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a semester.
func (svc *semesterAuthorizer) IsAuthorized(role, action string) bool {
	// Validate input
	if action == "" {
		return false
	}

	// Split out the sub-resource from the action string if possible
	var resource string
	parts := strings.SplitAfterN(action, ".", 2)
	if len(parts) == 1 {
		action = parts[0]
	} else {
		// User wants to authorized for a sub-resource
		resource = strings.TrimSuffix(parts[0], ".")
		action = parts[1]

		authorizer, exists := svc.resourceAuthorizers[resource]
		if !exists {
			return false
		}
		return authorizer.IsAuthorized(role, action)
	}

	// Check if user only wants to perform an action on this resource
	switch action {
	case "create":
		return HasAtleastRole(ROLE_VICE_PRESIDENT, role)
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasAtleastRole(ROLE_BOT, role)
	}

	return false
}
