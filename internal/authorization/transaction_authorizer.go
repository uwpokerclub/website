package authorization

// TransactionAuthorizer is an interface that defines the methods for authorizing transactions.
type transactionAuthorizer struct {
	actions []string
}

// NewTransactionAuthorizer creates a new transaction authorizer.
func NewTransactionAuthorizer() ResourceAuthorizer {
	return &transactionAuthorizer{
		actions: []string{"create", "get", "list", "edit", "delete"},
	}
}

// IsAuthorized checks if a user with the given role is authorized to perform the specified action on a transaction.
func (svc *transactionAuthorizer) IsAuthorized(role string, action string) bool {
	switch action {
	case "create":
		return HasRole(ROLE_TREASURER, role) || HasRole(ROLE_WEBMASTER, role)
	case "get":
		return HasAtleastRole(ROLE_SECRETARY, role)
	case "list":
		return HasAtleastRole(ROLE_SECRETARY, role)
	case "edit":
		return HasRole(ROLE_TREASURER, role) || HasRole(ROLE_WEBMASTER, role)
	case "delete":
		return HasRole(ROLE_TREASURER, role) || HasRole(ROLE_WEBMASTER, role)
	}

	return false
}

func (svc *transactionAuthorizer) GetPermissions(role string) map[string]any {
	permissions := make(map[string]any)

	for _, action := range svc.actions {
		permissions[action] = svc.IsAuthorized(role, action)
	}

	return permissions
}
