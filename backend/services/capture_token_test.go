package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newCaptureToken(t *testing.T) (*CaptureToken, string) {
	t.Helper()
	// A nested path: the token directory may not exist on a fresh checkout.
	path := filepath.Join(t.TempDir(), "token", "job_capture_token.json")
	tok, err := InitCaptureToken(path)
	if err != nil {
		t.Fatalf("InitCaptureToken: %v", err)
	}
	return tok, path
}

func TestCaptureTokenStartsUnsetAndRejectsEverything(t *testing.T) {
	tok, _ := newCaptureToken(t)
	if set, _ := tok.Status(); set {
		t.Fatal("a fresh store reports a token")
	}
	// The empty string is the credential of a request with no token at all.
	for _, presented := range []string{"", "anything"} {
		if tok.Verify(presented) {
			t.Fatalf("Verify(%q) passed with no token generated", presented)
		}
	}
}

func TestCaptureTokenGenerateVerifyAndRotate(t *testing.T) {
	tok, path := newCaptureToken(t)

	first, createdAt, err := tok.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(first) != 64 {
		t.Fatalf("token is %d chars, want 64 hex chars (32 bytes)", len(first))
	}
	if set, at := tok.Status(); !set || !at.Equal(createdAt) {
		t.Fatalf("Status = %v, %v", set, at)
	}
	if !tok.Verify(first) || tok.Verify(first+"x") || tok.Verify("") {
		t.Fatal("Verify does not accept exactly the generated token")
	}

	// The file must be useless to someone who reads it, and private anyway.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading token file: %v", err)
	}
	if strings.Contains(string(raw), first) {
		t.Fatal("the token itself was written to disk")
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("token file mode = %v, want 0600", info.Mode().Perm())
	}

	second, _, err := tok.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if tok.Verify(first) || !tok.Verify(second) {
		t.Fatal("generating again must replace the old token")
	}
}

func TestCaptureTokenSurvivesRestart(t *testing.T) {
	tok, path := newCaptureToken(t)
	token, createdAt, _ := tok.Generate()

	reloaded, err := InitCaptureToken(path)
	if err != nil {
		t.Fatalf("InitCaptureToken: %v", err)
	}
	if !reloaded.Verify(token) {
		t.Fatal("token no longer verifies after a reload")
	}
	if _, at := reloaded.Status(); !at.Equal(createdAt) {
		t.Fatalf("createdAt = %v, want %v", at, createdAt)
	}
}

func TestCaptureTokenRevoke(t *testing.T) {
	tok, path := newCaptureToken(t)
	token, _, _ := tok.Generate()

	if err := tok.Revoke(); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if tok.Verify(token) {
		t.Fatal("token still verifies after Revoke")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("token file still exists, so the token would return on restart")
	}
	if err := tok.Revoke(); err != nil {
		t.Fatalf("revoking with nothing to revoke: %v", err)
	}
}

// A damaged file must leave the routes off, never open and never fatal.
func TestCaptureTokenCorruptFile(t *testing.T) {
	for name, content := range map[string]string{
		"not json":     "{",
		"not a hash":   `{"sha256": "zz"}`,
		"short hash":   `{"sha256": "abcd"}`,
		"empty object": `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "t.json")
			os.WriteFile(path, []byte(content), 0o600)
			tok, err := InitCaptureToken(path)
			if err == nil {
				t.Fatal("expected an error")
			}
			if tok == nil || tok.Verify("") {
				t.Fatal("a corrupt file must yield a usable store that rejects everything")
			}
		})
	}
}
