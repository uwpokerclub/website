package authorization

import "strings"

type eventAuthorizer struct {
	resourceAuthorizers ResourceAuthorizerMap
}

func NewEventAuthorizer(resourceAuthorizers ResourceAuthorizerMap) ResourceAuthorizer {
	return &eventAuthorizer{
		resourceAuthorizers: resourceAuthorizers,
	}
}

func (svc *eventAuthorizer) IsAuthorized(role string, action string) bool {
	// Validate input
	if action == "" || role == "" {
		return false
	}

	// Split out the resource from the action string if possible
	var resource string
	parts := strings.SplitAfterN(action, ".", 2)
	if len(parts) == 1 {
		action = parts[0]
	} else {
		// User wants to authorize against a sub-resource
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
	case "edit":
		return HasRole(ROLE_EXECUTIVE, role)
	case "end":
		return HasRole(ROLE_EXECUTIVE, role)
	case "restart":
		return HasRole(ROLE_EXECUTIVE, role)
	case "rebuy":
		return HasRole(ROLE_EXECUTIVE, role)
	}

	return false
}
