package authorization

import "testing"

func TestExecutiveTeamAuthorizer(t *testing.T) {
	a := NewExecutiveTeamAuthorizer()
	for role, allowed := range map[string]bool{
		"bot": false, "executive": false, "tournament_director": false,
		"secretary": true, "treasurer": true, "vice_president": true,
		"president": true, "webmaster": true, "unknown": false,
	} {
		for _, action := range []string{"get", "create", "update", "delete"} {
			if got := a.IsAuthorized(role, action); got != allowed {
				t.Errorf("IsAuthorized(%q, %s) = %v, want %v", role, action, got, allowed)
			}
		}
	}
	if a.IsAuthorized("president", "list") {
		t.Fatal("unsupported action was authorized")
	}
	for role, allowed := range map[string]bool{"secretary": true, "president": true, "executive": false, "tournament_director": false, "bot": false} {
		permissions := a.GetPermissions(role)
		if permissions["create"] != allowed {
			t.Errorf("GetPermissions(%q).create = %v, want %v", role, permissions["create"], allowed)
		}
		if permissions["update"] != allowed {
			t.Errorf("GetPermissions(%q).update = %v, want %v", role, permissions["update"], allowed)
		}
		if permissions["delete"] != allowed {
			t.Errorf("GetPermissions(%q).delete = %v, want %v", role, permissions["delete"], allowed)
		}
	}
}
