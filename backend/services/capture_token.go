package services

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CAPTURE_TOKEN_JSON_PATH sits beside the Spotify and Graph tokens, in the
// directory docker-compose bind-mounts so that it survives a rebuild.
const CAPTURE_TOKEN_JSON_PATH = "/backend/token/job_capture_token.json"

// CaptureToken holds the bearer token the job capture userscript presents
// (handlers/handle_job_capture.go). The token is minted from the admin panel
// rather than set in .env, so it can be replaced without a redeploy when the
// browser profile holding it is lost.
//
// Only a SHA-256 of the token is kept, in memory and on disk: the token is
// shown once, when generated, and cannot be read back afterwards. A plain
// hash is enough because the token is 32 random bytes, not a password — there
// is nothing to brute-force.
//
// There is at most one token. Generating replaces it, which is also how a
// leaked one is rotated.
type CaptureToken struct {
	path string

	mu        sync.RWMutex
	hash      []byte
	createdAt time.Time
}

type captureTokenFile struct {
	SHA256    string    `json:"sha256"`
	CreatedAt time.Time `json:"created_at"`
}

// InitCaptureToken loads the token file if there is one. A missing file is
// the normal first-run state and means "no token": the capture routes stay
// off until one is generated. A file that cannot be read or parsed is
// reported, and likewise leaves the routes off rather than open.
func InitCaptureToken(path string) (*CaptureToken, error) {
	t := &CaptureToken{path: path}

	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return t, nil
	}
	if err != nil {
		return t, fmt.Errorf("reading capture token: %w", err)
	}
	var stored captureTokenFile
	if err := json.Unmarshal(raw, &stored); err != nil {
		return t, fmt.Errorf("parsing capture token: %w", err)
	}
	hash, err := hex.DecodeString(stored.SHA256)
	if err != nil || len(hash) != sha256.Size {
		return t, fmt.Errorf("capture token file holds no valid hash")
	}
	t.hash, t.createdAt = hash, stored.CreatedAt
	return t, nil
}

// Generate mints a new token, replacing any existing one, and returns it.
// This is the only moment the token itself exists on the server.
//
// The file is written before the in-memory hash is swapped, so a failed write
// leaves the old token working instead of handing out one that would stop
// working at the next restart.
func (t *CaptureToken) Generate() (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	now := time.Now().UTC()

	data, err := json.MarshalIndent(captureTokenFile{SHA256: hex.EncodeToString(sum[:]), CreatedAt: now}, "", "  ")
	if err != nil {
		return "", time.Time{}, err
	}
	if err := os.MkdirAll(filepath.Dir(t.path), 0755); err != nil {
		return "", time.Time{}, fmt.Errorf("creating token directory: %w", err)
	}
	if err := os.WriteFile(t.path, data, 0600); err != nil {
		return "", time.Time{}, fmt.Errorf("writing capture token: %w", err)
	}

	t.mu.Lock()
	t.hash, t.createdAt = sum[:], now
	t.mu.Unlock()
	return token, now, nil
}

// Revoke deletes the token, switching the capture routes off.
func (t *CaptureToken) Revoke() error {
	if err := os.Remove(t.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing capture token: %w", err)
	}
	t.mu.Lock()
	t.hash, t.createdAt = nil, time.Time{}
	t.mu.Unlock()
	return nil
}

// Status reports whether a token exists and when it was generated.
func (t *CaptureToken) Status() (set bool, createdAt time.Time) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.hash != nil, t.createdAt
}

// Verify reports whether presented is the current token. With no token it is
// false for every input, the empty string included.
func (t *CaptureToken) Verify(presented string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.hash == nil {
		return false
	}
	sum := sha256.Sum256([]byte(presented))
	return subtle.ConstantTimeCompare(sum[:], t.hash) == 1
}
