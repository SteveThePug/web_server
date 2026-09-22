package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tailLogFile's rules that are easy to get wrong: the line cap, and dropping
// the partial first line when the read started mid-file.

func writeLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "go.log")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	return path
}

func TestTailLogFileCapsLines(t *testing.T) {
	var b strings.Builder
	for range logTailLines + 50 {
		b.WriteString("line\n")
	}
	lines, _, err := tailLogFile(writeLog(t, b.String()))
	if err != nil {
		t.Fatalf("tailLogFile: %v", err)
	}
	if len(lines) != logTailLines {
		t.Fatalf("got %d lines, want the cap of %d", len(lines), logTailLines)
	}
}

func TestTailLogFileDropsPartialFirstLine(t *testing.T) {
	// A file larger than the byte window: the read starts mid-file, so the
	// first line it sees is a fragment and must be discarded.
	filler := strings.Repeat("x", 100) + "\n"
	var b strings.Builder
	for b.Len() < logTailBytes+len(filler) {
		b.WriteString(filler)
	}
	b.WriteString("FINAL\n")

	lines, _, err := tailLogFile(writeLog(t, b.String()))
	if err != nil {
		t.Fatalf("tailLogFile: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected some lines")
	}
	if lines[len(lines)-1] != "FINAL" {
		t.Fatalf("last line = %q, want FINAL", lines[len(lines)-1])
	}
	// Every surviving line must be whole, i.e. a full filler line.
	if got := lines[0]; got != strings.TrimSuffix(filler, "\n") {
		t.Fatalf("first line = %q, want a complete filler line", got)
	}
}

func TestTailLogFileEmpty(t *testing.T) {
	lines, size, err := tailLogFile(writeLog(t, ""))
	if err != nil {
		t.Fatalf("tailLogFile: %v", err)
	}
	if len(lines) != 0 || size != 0 {
		t.Fatalf("got %d lines / %d bytes, want 0 / 0", len(lines), size)
	}
}

func TestTailLogFileMissing(t *testing.T) {
	if _, _, err := tailLogFile(filepath.Join(t.TempDir(), "absent.log")); err == nil {
		t.Fatal("expected an error for a missing log file")
	}
}
