package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// This file backs the admin diagnostics panel's log tail. It is an admin-only
// REST endpoint rather than GraphQL because it reads the filesystem — outside
// the data model the GraphQL schema describes.

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
