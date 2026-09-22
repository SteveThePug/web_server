package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// This file backs the admin diagnostics panel: a configuration report and a
// log tail. Both are admin-only REST endpoints rather than GraphQL because
// they read the process environment and the filesystem — things outside the
// data model the GraphQL schema describes.

// ConfigVar describes one piece of configuration the backend can report on.
// Required marks vars whose absence visibly breaks a feature; Optional vars
// have a working default.
type ConfigVar struct {
	Name     string `json:"name"`
	Purpose  string `json:"purpose"`
	Required bool   `json:"required"`
}

// configVars is the checklist the panel renders. Sensitive values are never
// returned — only whether each var is present — so the endpoint cannot leak
// credentials even to an admin client.
var configVars = []ConfigVar{
	{"POSTGRES_HOST", "Database connectivity", true},
	{"POSTGRES_USER", "Database connectivity", true},
	{"POSTGRES_PASSWORD", "Database connectivity", true},
	{"POSTGRES_DB", "Database connectivity", true},
	{"DOMAIN", "Site origin, JWT cookie domain", true},
	{"AUTH_SECRET", "JWT signing", true},
	{"SPOTIFY_CLIENT_ID", "Spotify listening widget", true},
	{"SPOTIFY_CLIENT_SECRET", "Spotify listening widget", true},
	{"SPOTIFY_REDIRECT_URI", "Spotify listening widget", true},
	{"STEAM_API_KEY", "Steam status widget", false},
	{"STEAM_ID", "Steam status widget", false},
	{"CLAUDE_API_KEY", "Email summarisation", false},
	{"EMAIL_BACKEND", "Email sync (graph/imap)", false},
	{"MSGRAPH_CLIENT_ID", "Email sync (Microsoft Graph)", false},
	{"MSGRAPH_CLIENT_SECRET", "Email sync (Microsoft Graph)", false},
	{"MSGRAPH_REDIRECT_URI", "Email sync (Microsoft Graph)", false},
	{"MSGRAPH_TENANT_ID", "Email sync (Microsoft Graph)", false},
	{"IMAP_HOST", "Email sync (IMAP fallback)", false},
	{"IMAP_EMAIL", "Email sync (IMAP fallback)", false},
	{"IMAP_PASSWORD", "Email sync (IMAP fallback)", false},
	{"GITEA_HOST", "Gitea activity feed", false},
	{"GITEA_PORT", "Gitea activity feed", false},
	{"JWT_COOKIE_SECURE", "Cookie security in dev", false},
	{"ICECAST_URL", "Radio queue widget", false},
}

// ConfigCheck backs GET /admin/config. It reports one entry per known var:
// set or empty, plus the value only for vars flagged non-sensitive (none of
// the credentials are). A var not in the list at all is not reported — the
// point is a checklist, not a full env dump.
func (store *Store) ConfigCheck(ctx *gin.Context) {
	type entry struct {
		ConfigVar
		Set    bool   `json:"set"`
		Value  string `json:"value,omitempty"`
		Secret bool   `json:"secret"`
	}

	// Vars whose *value* may come back (bad values like "consumers" are fine;
	// they contain no credential).
	visible := map[string]bool{
		"EMAIL_BACKEND":     true,
		"MSGRAPH_TENANT_ID": true,
		"ICECAST_URL":       true,
		"STEAM_ID":          true,
	}

	// Explicitly-credentialled vars get the secret badge so the panel can
	// label them "present" instead of showing a masked string.
	secrets := map[string]bool{
		"POSTGRES_PASSWORD":     true,
		"AUTH_SECRET":           true,
		"SPOTIFY_CLIENT_SECRET": true,
		"MSGRAPH_CLIENT_SECRET": true,
		"CLAUDE_API_KEY":        true,
		"STEAM_API_KEY":         true,
		"IMAP_PASSWORD":         true,
	}

	result := make([]entry, 0, len(configVars))
	for _, v := range configVars {
		val := os.Getenv(v.Name)
		e := entry{ConfigVar: v, Set: strings.TrimSpace(val) != "", Secret: secrets[v.Name]}
		if e.Set && visible[v.Name] {
			e.Value = val
		}
		result = append(result, e)
	}

	ctx.JSON(http.StatusOK, gin.H{"vars": result})
}

// logTailBytes and logTailLines bound how much of the log file one request
// moves: 256 KiB is beyond any useful screen at log-ish line lengths, and is
// cheap even on the Pi.
const (
	logTailBytes = 256 << 10
	logTailLines = 500
)

// Logs backs GET /admin/logs. It returns the tail of the backend's own log
// file (go.log — Gin traffic plus application log lines), which lives on the
// bind-mounted ./logs volume.
//
// Why tail and not offer a range: the panel wants "what happened recently",
// and the file can grow unboundedly between rotations, so anything else
// invites reading the whole thing into memory.
func (store *Store) Logs(ctx *gin.Context) {
	lines, size, err := tailLogFile(logPath)
	if err != nil {
		// The mount being absent is a deployment problem worth surfacing
		// distinctly from "log file empty".
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "log unavailable: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"lines": lines, "total_bytes": size})
}

// logPath matches the file main.go opens for gin.DefaultWriter.
const logPath = "/backend/logs/go.log"

// tailLogFile returns the last logTailLines lines of a file, reading at most
// logTailBytes from the end. Split out from the handler so the offset and
// partial-line rules are testable without a filesystem mount.
func tailLogFile(path string) ([]string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	// Read from near the end rather than head-to-toe so the read stays
	// constant-size regardless of how old the file is.
	start := int64(0)
	if info.Size() > logTailBytes {
		start = info.Size() - logTailBytes
	}
	buf := make([]byte, info.Size()-start)
	if len(buf) > 0 {
		if _, err := f.ReadAt(buf, start); err != nil {
			return nil, 0, err
		}
	}

	// Starting mid-file almost certainly lands mid-line; drop that fragment
	// so the panel never renders a truncated first entry.
	text := string(buf)
	if start > 0 {
		if i := strings.IndexByte(text, '\n'); i >= 0 {
			text = text[i+1:]
		}
	}

	text = strings.TrimRight(text, "\n")
	if text == "" {
		return []string{}, info.Size(), nil
	}

	lines := strings.Split(text, "\n")
	if len(lines) > logTailLines {
		lines = lines[len(lines)-logTailLines:]
	}
	return lines, info.Size(), nil
}
