package services

import (
	"testing"

	"adam-french.co.uk/backend/models"
)

func TestValidPermission(t *testing.T) {
	for _, p := range PermissionTypes {
		if !ValidPermission(p.Key) {
			t.Errorf("listed key %q reported invalid", p.Key)
		}
	}
	for _, key := range []string{"", "admin", "Receipt", "receipt "} {
		if ValidPermission(key) {
			t.Errorf("unlisted key %q reported valid", key)
		}
	}
}

// Keys must be URL-path safe: nginx splices them into /auth/validate/<key>.
func TestPermissionKeysArePathSafe(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range PermissionTypes {
		if seen[p.Key] {
			t.Errorf("duplicate key %q", p.Key)
		}
		seen[p.Key] = true
		for _, r := range p.Key {
			if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
				t.Errorf("key %q has %q; use lowercase letters, digits and '-'", p.Key, r)
			}
		}
	}
}

func TestHasPermission(t *testing.T) {
	cases := []struct {
		name string
		user models.User
		want bool
	}{
		{"admin passes without a grant", models.User{Admin: true}, true},
		{"granted user passes", models.User{Permissions: []string{"receipt"}}, true},
		{"other grant does not count", models.User{Permissions: []string{"other"}}, false},
		{"no grants (NULL column) fails", models.User{}, false},
	}
	for _, c := range cases {
		if got := HasPermission(&c.user, "receipt"); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
