package graph

// Hand-written helpers for cover_letter.resolvers.go, kept out of the
// resolver file for the same reason as box_helpers.go.

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"
)

const (
	maxCoverLetterNameLen = 100
	// A letter is a page of text; 100 KB is far beyond any real one and
	// still small enough that a runaway client can't fill the table.
	maxCoverLetterBodyLen = 100 * 1024
	maxSnippetTextLen     = 5000
	maxSnippetSectionLen  = 100
)

// cleanCoverLetterName trims name and reports why it can't be a letter's
// name.
func cleanCoverLetterName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return name, errors.New("a cover letter needs a name")
	}
	if len(name) > maxCoverLetterNameLen {
		return name, fmt.Errorf("a cover letter name is at most %d characters", maxCoverLetterNameLen)
	}
	return name, nil
}

// coverLetterNameTaken reports whether a live letter other than exceptID has
// this name. Checked up front, like layoutNameTaken, so the admin gets a
// sentence rather than a constraint violation.
func coverLetterNameTaken(db *gorm.DB, name string, exceptID uint) (bool, error) {
	var count int64
	err := db.Model(&models.CoverLetter{}).Where("name = ? AND id <> ?", name, exceptID).Count(&count).Error
	return count > 0, err
}

// checkCoverLetterBody reports why body can't be stored. The server never
// reads inside it, so it only has to be a JSON object of a sane size — an
// array or a bare string would be a frontend bug worth refusing early.
func checkCoverLetterBody(body string) error {
	if len(body) > maxCoverLetterBodyLen {
		return fmt.Errorf("a cover letter is at most %d KB", maxCoverLetterBodyLen/1024)
	}
	var obj map[string]json.RawMessage
	// null unmarshals into a nil map without error, so it is caught here too.
	if err := json.Unmarshal([]byte(body), &obj); err != nil || obj == nil {
		return errors.New("a cover letter body must be a JSON object")
	}
	return nil
}

// cleanSnippetText trims text and reports why it can't be a bank paragraph.
// Counted in characters, not bytes, since the limit is about how much prose
// fits rather than storage.
func cleanSnippetText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return text, errors.New("a paragraph can't be empty")
	}
	if utf8.RuneCountInString(text) > maxSnippetTextLen {
		return text, fmt.Errorf("a paragraph is at most %d characters", maxSnippetTextLen)
	}
	return text, nil
}

// cleanSnippetSection trims a snippet's section title.
func cleanSnippetSection(section string) (string, error) {
	section = strings.TrimSpace(section)
	if utf8.RuneCountInString(section) > maxSnippetSectionLen {
		return section, fmt.Errorf("a section title is at most %d characters", maxSnippetSectionLen)
	}
	return section, nil
}
