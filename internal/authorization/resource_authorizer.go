package authorization

type ResourceAuthorizer interface {
	IsAuthorized(role string, action string) bool
}

type ResourceAuthorizerMap map[string]ResourceAuthorizer
