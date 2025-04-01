package authorization

// TransactionAuthorizer is an interface that defines the methods for authorizing transactions.
type transactionAuthorizer struct {
}

// NewTransactionAuthorizer creates a new transaction authorizer.
func NewTransactionAuthorizer() ResourceAuthorizer {
	return &transactionAuthorizer{}
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
