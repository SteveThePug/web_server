package graph

// Hand-written helpers for component.resolvers.go, kept out of the resolver
// file for the same reason as box_helpers.go.

import (
	"errors"
	"fmt"
	"strings"

	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"
)

const maxComponentNameLen = 100

// cleanComponentName trims name and reports why it can't be a component's
// name.
func cleanComponentName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return name, errors.New("a component needs a name")
	}
	if len(name) > maxComponentNameLen {
		return name, fmt.Errorf("a component name is at most %d characters", maxComponentNameLen)
	}
	return name, nil
}

// componentNameTaken reports whether a live component other than exceptID
// has this name. Checked up front, like layoutNameTaken, so the admin gets a
// sentence rather than a constraint violation.
func componentNameTaken(db *gorm.DB, name string, exceptID uint) (bool, error) {
	var count int64
	err := db.Model(&models.Component{}).Where("name = ? AND id <> ?", name, exceptID).Count(&count).Error
	return count > 0, err
}

// usableComponentName checks name and answers with a sentence if it can't be
// used, so the resolvers that name a component share one set of messages.
func usableComponentName(db *gorm.DB, name string, exceptID uint) (string, error) {
	name, err := cleanComponentName(name)
	if err != nil {
		return name, err
	}
	taken, err := componentNameTaken(db, name, exceptID)
	if err != nil {
		return name, err
	}
	if taken {
		return name, errors.New("there is already a component with that name")
	}
	return name, nil
}

// fillLinked sets each linked box's in-memory HTML to its component's, with
// one query for all of them. Only for boxes about to be returned or
// snapshotted: a linked row's own HTML is nil, and saving a filled box would
// freeze a copy of the component into it. A box whose component has gone
// (deleted outside deleteComponent, which detaches first) is left with
// empty HTML rather than failing the whole page.
func fillLinked(db *gorm.DB, boxes []models.Box) error {
	var ids []uint
	for _, b := range boxes {
		if b.ComponentID != nil {
			ids = append(ids, *b.ComponentID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var components []models.Component
	if err := db.Where("id IN ?", ids).Find(&components).Error; err != nil {
		return err
	}
	html := make(map[uint]string, len(components))
	for _, c := range components {
		html[c.ID] = c.HTML
	}
	for i := range boxes {
		if id := boxes[i].ComponentID; id != nil {
			h := html[*id]
			boxes[i].HTML = &h
		}
	}
	return nil
}

// fillLinkedOne is fillLinked for a single box.
func fillLinkedOne(db *gorm.DB, box *models.Box) error {
	boxes := []models.Box{*box}
	if err := fillLinked(db, boxes); err != nil {
		return err
	}
	*box = boxes[0]
	return nil
}

// componentExists reports whether id is a live component, for linking a box
// to it.
func componentExists(db *gorm.DB, id uint) (bool, error) {
	var count int64
	err := db.Model(&models.Component{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// detachFrom gives every box linked to component c its own copy of c's HTML
// and unlinks it, inside tx.
func detachFrom(tx *gorm.DB, c *models.Component) error {
	return tx.Model(&models.Box{}).Where("component_id = ?", c.ID).
		Updates(map[string]any{"html": c.HTML, "component_id": nil}).Error
}
