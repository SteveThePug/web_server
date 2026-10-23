package graph

// Hand-written helpers for box.resolvers.go. Kept out of that file because
// gqlgen moves unknown code in a resolvers file to the end, under a warning
// comment, on every regenerate.

import (
	"errors"

	"adam-french.co.uk/backend/models"
)

// minBoxSize stops a box being resized to nothing, after which it could not
// be grabbed again on the canvas.
const minBoxSize = 40

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
	box.X = max(box.X, 0)
	box.Y = max(box.Y, 0)
	box.W = max(box.W, minBoxSize)
	box.H = max(box.H, minBoxSize)
	return nil
}
