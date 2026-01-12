package services

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NotesConfig struct {
	Dir string
}

type Note struct {
	Filename   string    `json:"title"`
	Contents   string    `json:"contents"`
	LastEdited time.Time `json:"last_edited"`
}

type Notes struct {
	Config NotesConfig
}

func InitNotes(config *NotesConfig) *Notes {
	return &Notes{
		Config: *config,
	}
}

var ErrPathTraversal = errors.New("invalid path")

func (notes *Notes) GetNote(path string) (*Note, error) {
	baseDir, err := filepath.Abs(notes.Config.Dir)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(baseDir, path)
	fullPath, err = filepath.Abs(fullPath)
	if err != nil {
		return nil, err
	}

	// Enforce directory boundary
	if !strings.HasPrefix(fullPath, baseDir+string(os.PathSeparator)) {
		return nil, ErrPathTraversal
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	return &Note{
		Filename:   info.Name(),
		Contents:   string(data),
		LastEdited: info.ModTime(),
	}, nil
}
