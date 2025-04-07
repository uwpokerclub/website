package authorization

// ResourceAuthorizer is an interface that defines the methods for authorizing resources.
type ResourceAuthorizer interface {
	// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a resource.
	IsAuthorized(role string, action string) bool

	// GetPermissions returns a map of permissions for the resource.
	// The key is an action or sub-resource name, and the value is either
	// a boolean or a map of permissions for the sub-resource.
	GetPermissions(role string) map[string]any
}

// ResourceAuthorizerMap is a map of resource names to their respective authorizers.
type ResourceAuthorizerMap map[string]ResourceAuthorizer

var DefaultAuthorizerMap = ResourceAuthorizerMap{
	"login": NewLoginAuthorizer(),
	"user":  NewUserAuthorizer(),
	"semester": NewSemesterAuthorizer(ResourceAuthorizerMap{
		"rankings":    NewRankingsAuthorizer(),
		"transaction": NewTransactionAuthorizer(),
	}),
	"membership": NewMembershipAuthorizer(),
	"structure":  NewStructureAuthorizer(),
	"event": NewEventAuthorizer(ResourceAuthorizerMap{
		"participant": NewParticipantAuthorizer(),
	}),
}
