package authorization

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type authorizationService struct {
	role                string
	resourceAuthorizers ResourceAuthorizerMap
}

func NewAuthorizationService(db *gorm.DB, username string, resourceAuthorizers ResourceAuthorizerMap) (*authorizationService, error) {
	user := models.Login{Username: username}

	// Attempt to find the users role
	res := db.Find(&user)
	err := res.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, e.Unauthorized("You do not have permission to perform this action.")
		}
		return nil, e.InternalServerError("An internal problem has occurred. Please try again later.")
	}

	svc := authorizationService{
		role:                user.Role,
		resourceAuthorizers: resourceAuthorizers,
	}

	return &svc, nil
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
