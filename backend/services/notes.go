package services

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// NotesConfig points at the directory of markdown notes mounted into the
// container.
type NotesConfig struct {
	Dir string
}

// Notes serves files out of a single directory, guarding against path
// traversal. It is admin-only at the route level (see main.go).
type Notes struct {
	Config NotesConfig
}

// InitNotes builds the Notes service.
func InitNotes(config *NotesConfig) *Notes {
	return &Notes{
		Config: *config,
	}
}

// ParsePath resolves a user-supplied request path to an absolute path inside
// the notes directory, or returns an error if it would escape.
//
// The traversal guard is the prefix check at the end. filepath.Join already
// cleans "..", but the separator is appended to baseDir before comparing so
// that a sibling directory whose name merely starts with the same characters
// (e.g. /backend/notes-private next to /backend/notes) cannot pass the check.
func (notes *Notes) ParsePath(path string) (string, error) {
	// The route is /notes/*path, so a bare /notes/ arrives as "" or "/";
	// treat both as a request for the wiki index.
	if path == "" || path == "/" {
		path = "Index.md"
	}

	baseDir, err := filepath.Abs(notes.Config.Dir)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(baseDir, path)
	fullPath, err = filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	// Enforce directory boundary
	if !strings.HasPrefix(fullPath, baseDir+string(os.PathSeparator)) {
		return "", errors.New("Invalid path")
	}

	return fullPath, nil
}
