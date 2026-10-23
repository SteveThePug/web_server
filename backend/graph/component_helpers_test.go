package graph

import (
	"strings"
	"testing"
)

func TestCleanComponentName(t *testing.T) {
	for _, c := range []struct {
		name, in, want string
		wantErr        bool
	}{
		{"plain", "Clock", "Clock", false},
		{"trimmed", "  Clock \n", "Clock", false},
		{"empty", "", "", true},
		{"only spaces", "   ", "", true},
		{"at the limit", strings.Repeat("a", maxComponentNameLen), strings.Repeat("a", maxComponentNameLen), false},
		{"too long", strings.Repeat("a", maxComponentNameLen+1), "", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := cleanComponentName(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
