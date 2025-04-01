package authorization

import "strings"

// eventAuthorizer is an interface that defines the methods for authorizing events.
type eventAuthorizer struct {
	resourceAuthorizers ResourceAuthorizerMap
}

// NewEventAuthorizer creates a new event authorizer.
// It takes a ResourceAuthorizerMap as an argument, which is a map of resource names to their respective authorizers.
func NewEventAuthorizer(resourceAuthorizers ResourceAuthorizerMap) ResourceAuthorizer {
	return &eventAuthorizer{
		resourceAuthorizers: resourceAuthorizers,
	}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on an event.
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
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "get":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "list":
		return HasAtleastRole(ROLE_EXECUTIVE, role)
	case "edit":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	case "end":
		return HasAtleastRole(ROLE_SECRETARY, role)
	case "restart":
		return HasAtleastRole(ROLE_SECRETARY, role)
	case "rebuy":
		return HasAtleastRole(ROLE_TOURNAMENT_DIRECTOR, role)
	}

	return false
}
