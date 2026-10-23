package authorization

import "testing"

func TestExecutiveTeamAuthorizer(t *testing.T) {
	a := NewExecutiveTeamAuthorizer()
	for role, allowed := range map[string]bool{
		"bot": false, "executive": false, "tournament_director": false,
		"secretary": true, "treasurer": true, "vice_president": true,
		"president": true, "webmaster": true, "unknown": false,
	} {
		if got := a.IsAuthorized(role, "get"); got != allowed {
			t.Errorf("IsAuthorized(%q, get) = %v, want %v", role, got, allowed)
		}
	}
	if a.IsAuthorized("president", "list") {
		t.Fatal("unsupported action was authorized")
	}
}
