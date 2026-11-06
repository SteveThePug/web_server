package graph

import (
	"strings"
	"testing"
)

func TestCleanCoverLetterName(t *testing.T) {
	for _, c := range []struct {
		name, in, want string
		wantErr        bool
	}{
		{"plain", "Acme", "Acme", false},
		{"trimmed", "  Acme \n", "Acme", false},
		{"empty", "", "", true},
		{"only spaces", "   ", "", true},
		{"at the limit", strings.Repeat("a", maxCoverLetterNameLen), strings.Repeat("a", maxCoverLetterNameLen), false},
		{"too long", strings.Repeat("a", maxCoverLetterNameLen+1), "", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := cleanCoverLetterName(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestCheckCoverLetterBody(t *testing.T) {
	for _, c := range []struct {
		name, in string
		wantErr  bool
	}{
		{"object", `{"v":1,"sections":[]}`, false},
		{"empty object", `{}`, false},
		{"array", `[]`, true},
		{"string", `"hi"`, true},
		{"null", `null`, true},
		{"not json", `{`, true},
		{"empty", ``, true},
		{"too big", `{"x":"` + strings.Repeat("a", maxCoverLetterBodyLen) + `"}`, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if err := checkCoverLetterBody(c.in); (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}

func TestCleanSnippetText(t *testing.T) {
	for _, c := range []struct {
		name, in, want string
		wantErr        bool
	}{
		{"plain", "I like Go.", "I like Go.", false},
		{"trimmed", "\n I like Go. ", "I like Go.", false},
		{"empty", "  ", "", true},
		// Characters, not bytes: 5000 two-byte runes still fit.
		{"multibyte at the limit", strings.Repeat("é", maxSnippetTextLen), strings.Repeat("é", maxSnippetTextLen), false},
		{"too long", strings.Repeat("a", maxSnippetTextLen+1), "", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := cleanSnippetText(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
