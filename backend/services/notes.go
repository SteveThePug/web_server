package services

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type NotesConfig struct {
	Dir string
}

type Notes struct {
	Config NotesConfig
}

func InitNotes(config *NotesConfig) *Notes {
	return &Notes{
		Config: *config,
	}
}

func (notes *Notes) ParsePath(path string) (string, error) {
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
