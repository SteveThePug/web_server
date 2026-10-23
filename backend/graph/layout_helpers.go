package graph

// Hand-written helpers for layout.resolvers.go, kept out of the resolver file
// for the same reason as box_helpers.go.

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"adam-french.co.uk/backend/graph/model"
	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"
)

const maxLayoutNameLen = 100

// layoutBox is one box as a layout stores it: what makes it look the way it
// does, without the id and page that tie a row to one canvas. Changing these
// field names would make older layouts load with the renamed fields empty.
type layoutBox struct {
	Kind         string  `json:"kind"`
	Widget       *string `json:"widget,omitempty"`
	HTML         *string `json:"html,omitempty"`
	Title        *string `json:"title,omitempty"`
	X            int     `json:"x"`
	Y            int     `json:"y"`
	W            int     `json:"w"`
	H            int     `json:"h"`
	Z            int     `json:"z"`
	Framed       bool    `json:"framed"`
	HideOnMobile bool    `json:"hideOnMobile"`
	AdminOnly    bool    `json:"adminOnly"`
}

type layoutBackground struct {
	Color string `json:"color"`
	Image string `json:"image"`
	Fit   string `json:"fit"`
}

// cleanLayoutName trims name and reports why it can't be a layout's name.
func cleanLayoutName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return name, errors.New("a layout needs a name")
	}
	if len(name) > maxLayoutNameLen {
		return name, fmt.Errorf("a layout name is at most %d characters", maxLayoutNameLen)
	}
	return name, nil
}

// layoutNameTaken reports whether a live layout other than exceptID has this
// name. Checked up front, like slugTaken, so the admin gets a sentence rather
// than a constraint violation.
func layoutNameTaken(db *gorm.DB, name string, exceptID uint) (bool, error) {
	var count int64
	err := db.Model(&models.Layout{}).Where("name = ? AND id <> ?", name, exceptID).Count(&count).Error
	return count > 0, err
}

// canvasExists reports whether page is a canvas boxes can be put on: the
// landing page, or a live Page. Loading a layout onto any other key would
// leave its boxes where nothing shows them.
func canvasExists(db *gorm.DB, page string) (bool, error) {
	if page == "index" {
		return true, nil
	}
	if !validPageKey(page) {
		return false, nil
	}
	var count int64
	err := db.Model(&models.Page{}).Where("slug = ?", page).Count(&count).Error
	return count > 0, err
}

// pageBoxes is every box on page, adminOnly ones included, in the order the
// boxes query returns them. Only for admin-only callers.
func pageBoxes(db *gorm.DB, page string) ([]models.Box, error) {
	var boxes []models.Box
	err := db.Where("page = ?", page).Order("z ASC, id ASC").Find(&boxes).Error
	return boxes, err
}

// snapshotPage is an unsaved Layout holding page's boxes and background as
// they are now. empty is true when there was nothing to keep.
func snapshotPage(db *gorm.DB, page, name string) (layout models.Layout, empty bool, err error) {
	boxes, err := pageBoxes(db, page)
	if err != nil {
		return layout, false, err
	}
	stored := make([]layoutBox, len(boxes))
	for i, b := range boxes {
		stored[i] = layoutBox{
			Kind: b.Kind, Widget: b.Widget, HTML: b.HTML, Title: b.Title,
			X: b.X, Y: b.Y, W: b.W, H: b.H, Z: b.Z,
			Framed: b.Framed, HideOnMobile: b.HideOnMobile, AdminOnly: b.AdminOnly,
		}
	}
	boxesJSON, err := json.Marshal(stored)
	if err != nil {
		return layout, false, err
	}
	layout = models.Layout{Name: name, SourcePage: page, Boxes: string(boxesJSON)}

	var bg models.Background
	err = db.Where("page = ?", page).First(&bg).Error
	switch {
	case err == nil:
		bgJSON, err := json.Marshal(layoutBackground{Color: bg.Color, Image: bg.Image, Fit: bg.Fit})
		if err != nil {
			return layout, false, err
		}
		s := string(bgJSON)
		layout.Background = &s
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return layout, false, err
	}
	return layout, len(boxes) == 0 && layout.Background == nil, nil
}

func decodeLayoutBoxes(layout *models.Layout) ([]layoutBox, error) {
	var boxes []layoutBox
	if err := json.Unmarshal([]byte(layout.Boxes), &boxes); err != nil {
		return nil, fmt.Errorf("layout %q is unreadable: %w", layout.Name, err)
	}
	return boxes, nil
}

// pageLabel is how a page key reads in a generated layout name.
func pageLabel(page string) string {
	if page == "index" {
		return "/"
	}
	return "/" + page
}

// uniqueLayoutName is base, or base with " (2)", " (3)"… if that is taken.
func uniqueLayoutName(db *gorm.DB, base string) (string, error) {
	name := base
	for n := 2; ; n++ {
		taken, err := layoutNameTaken(db, name, 0)
		if err != nil || !taken {
			return name, err
		}
		name = fmt.Sprintf("%s (%d)", base, n)
	}
}

// applyLayoutTo copies layout onto page inside tx; see applyLayout in
// layout.graphql for what replace does. The copies are new rows, so the page
// owns them outright.
func applyLayoutTo(tx *gorm.DB, layout *models.Layout, page string, replace bool) error {
	stored, err := decodeLayoutBoxes(layout)
	if err != nil {
		return err
	}

	// Added on top, the copies keep their own stacking order but all sit
	// above what the page already has.
	zShift := 0
	if replace {
		// What is being replaced is kept first, so a replace can be undone by
		// loading that layout back. A page with nothing on it has nothing to
		// keep.
		name, err := uniqueLayoutName(tx, fmt.Sprintf("%s before %s, %s",
			pageLabel(page), layout.Name, time.Now().Format("2 Jan 2006 15:04")))
		if err != nil {
			return err
		}
		backup, empty, err := snapshotPage(tx, page, name)
		if err != nil {
			return err
		}
		if !empty {
			if err := tx.Create(&backup).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("page = ?", page).Delete(&models.Box{}).Error; err != nil {
			return err
		}
		if err := tx.Where("page = ?", page).Delete(&models.Background{}).Error; err != nil {
			return err
		}
	} else if len(stored) > 0 {
		var top *int
		if err := tx.Model(&models.Box{}).Where("page = ?", page).Select("MAX(z)").Scan(&top).Error; err != nil {
			return err
		}
		if top != nil {
			lowest := stored[0].Z
			for _, b := range stored {
				lowest = min(lowest, b.Z)
			}
			zShift = *top + 1 - lowest
		}
	}

	boxes := make([]models.Box, len(stored))
	for i, b := range stored {
		boxes[i] = models.Box{
			Page: page, Kind: b.Kind, Widget: b.Widget, HTML: b.HTML, Title: b.Title,
			X: b.X, Y: b.Y, W: b.W, H: b.H, Z: b.Z + zShift,
			Framed: b.Framed, HideOnMobile: b.HideOnMobile, AdminOnly: b.AdminOnly,
		}
		if err := validateBox(&boxes[i]); err != nil {
			return err
		}
	}
	if len(boxes) > 0 {
		if err := tx.Create(&boxes).Error; err != nil {
			return err
		}
	}

	// Adding on top leaves the page's own background alone.
	if replace && layout.Background != nil {
		var saved layoutBackground
		if err := json.Unmarshal([]byte(*layout.Background), &saved); err != nil {
			return fmt.Errorf("layout %q is unreadable: %w", layout.Name, err)
		}
		bg, err := backgroundFrom(page, model.BackgroundInput{Color: saved.Color, Image: saved.Image, Fit: saved.Fit})
		if err != nil {
			return err
		}
		if err := tx.Create(&bg).Error; err != nil {
			return err
		}
	}
	return nil
}
