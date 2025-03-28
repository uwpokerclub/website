package authorization

import "strings"

type semesterAuthorizer struct {
	resourceAuthorizers ResourceAuthorizerMap
}

func NewSemesterAuthorizer(resourceAuthorizers ResourceAuthorizerMap) ResourceAuthorizer {
	return &semesterAuthorizer{
		resourceAuthorizers: resourceAuthorizers,
	}
}

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
		return HasRole(ROLE_EXECUTIVE, role)
	case "get":
		return HasRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasRole(ROLE_EXECUTIVE, role)
	}

	return false
}
