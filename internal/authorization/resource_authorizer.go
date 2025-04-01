package authorization

// ResourceAuthorizer is an interface that defines the methods for authorizing resources.
type ResourceAuthorizer interface {
	// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a resource.
	IsAuthorized(role string, action string) bool
}

// ResourceAuthorizerMap is a map of resource names to their respective authorizers.
type ResourceAuthorizerMap map[string]ResourceAuthorizer
