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
		{"shrines", true},           // the index is a canvas page too
		{"shrines/demoman", true},   // no longer an SPA route
		{"admin/pages", false},      // under an SPA route
		{"jobs/x", false},           // under an nginx route
		{"stp", true},               // the home page is a canvas page too
		{"stp2", false},             // an SPA route
		{"index", false},            // the landing canvas's key
		{"stp-fan-page", true},      // only whole segments are reserved
	} {
		err := validateSlug(c.slug)
		if (err == nil) != c.ok {
			t.Errorf("validateSlug(%q) = %v, want ok=%v", c.slug, err, c.ok)
		}
	}
}
