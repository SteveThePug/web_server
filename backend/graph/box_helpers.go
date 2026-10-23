package graph

// Hand-written helpers for box.resolvers.go. Kept out of that file because
// gqlgen moves unknown code in a resolvers file to the end, under a warning
// comment, on every regenerate.

import (
	"errors"
	"fmt"
	"strings"

	"adam-french.co.uk/backend/models"
)

// minBoxSize stops a box being resized to nothing, after which it could not
// be grabbed again on the canvas.
const minBoxSize = 40

// maxShadowLen bounds a box's shadow, which is a single CSS value: even a
// stack of a few comma-separated shadows is well under it.
const maxShadowLen = 200

func setInt(dst *int, src *int) {
	if src != nil {
		*dst = *src
	}
}

// validateBox checks the fields a box needs to render and clamps geometry
// rather than rejecting it, because a drag that overshoots the canvas edge
// should land at the edge, not fail.
func validateBox(box *models.Box) error {
	if !validPageKey(box.Page) {
		return errors.New("a box needs a valid page")
	}
	switch box.Kind {
	case "widget":
		if box.Widget == nil || *box.Widget == "" {
			return errors.New("a widget box needs a widget name")
		}
	case "html":
		// A linked box's html lives on its component (fillLinked).
		if box.HTML == nil && box.ComponentID == nil {
			return errors.New("an html box needs html")
		}
	default:
		return errors.New(`box kind must be "widget" or "html"`)
	}
	if box.ComponentID != nil && box.Kind != "html" {
		return errors.New("only an html box can be linked to a component")
	}
	if err := validateShadow(&box.Shadow); err != nil {
		return err
	}
	box.X = max(box.X, 0)
	box.Y = max(box.Y, 0)
	box.W = max(box.W, minBoxSize)
	box.H = max(box.H, minBoxSize)
	return nil
}

// validateShadow trims a box-shadow value and bounds its length. The value is
// not parsed: the front end sets it through a style binding, so it can only
// ever be the one property, and an invalid value just draws no shadow.
func validateShadow(shadow *string) error {
	*shadow = strings.TrimSpace(*shadow)
	if len(*shadow) > maxShadowLen {
		return fmt.Errorf("a shadow is at most %d characters", maxShadowLen)
	}
	return nil
}
