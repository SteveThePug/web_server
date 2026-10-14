package graph

import "testing"

func TestValidateSlug(t *testing.T) {
	for _, c := range []struct {
		slug string
		ok   bool
	}{
		{"notes-to-self", true},
		{"shrines/evangelion", true},
		{"shrines/neon-genesis/rei2", true},
		{"", false},
		{"Shrines/GTO", false},      // upper case
		{"shrines/", false},         // empty segment
		{"/shrines/gto", false},     // leading slash
		{"shrines//gto", false},     // empty segment
		{"shrines/gto-", false},     // trailing hyphen
		{"shrines/../admin", false}, // dots
		{"shrines", false},          // the SPA's own index
		{"shrines/demoman", true},   // no longer an SPA route
		{"admin/pages", false},      // under an SPA route
		{"jobs/x", false},           // under an nginx route
		{"stp", false},              // a built-in canvas's key
		{"index", false},            // likewise
		{"stp-fan-page", true},      // only whole segments are reserved
	} {
		err := validateSlug(c.slug)
		if (err == nil) != c.ok {
			t.Errorf("validateSlug(%q) = %v, want ok=%v", c.slug, err, c.ok)
		}
	}
}
