package authorization

import (
	"strings"
)

type authorizationService struct {
	role                string
	resourceAuthorizers ResourceAuthorizerMap
}

func NewAuthorizationService(role string, resourceAuthorizers ResourceAuthorizerMap) *authorizationService {
	return &authorizationService{
		role:                role,
		resourceAuthorizers: resourceAuthorizers,
	}
}

func (svc *authorizationService) IsAuthorized(action string) bool {
	// Validate input
	if action == "" {
		return false
	}

	// Split out the resource from the action string
	parts := strings.SplitAfterN(action, ".", 2)
	if len(parts) != 2 {
		return false
	}
	resource := strings.TrimSuffix(parts[0], ".")
	action = parts[1]

	// Check if the user is authorized to perform an action on this resource
	authorizer, exists := svc.resourceAuthorizers[resource]
	if !exists {
		return false
	}
	return authorizer.IsAuthorized(svc.role, action)
}

func (svc *authorizationService) Role() string {
	return svc.role
}

func (svc *authorizationService) GetPermissions() map[string]map[string]any {
	permissions := make(map[string]map[string]any)

	for resource, authorizer := range svc.resourceAuthorizers {
		permissions[resource] = authorizer.GetPermissions(svc.role)
	}

	return permissions
}
