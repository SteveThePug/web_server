package graph

import (
	"testing"

	"adam-french.co.uk/backend/models"
)

func TestValidateBox(t *testing.T) {
	str := func(s string) *string { return &s }

	for _, c := range []struct {
		name    string
		box     models.Box
		wantErr bool
	}{
		{"widget with a name", models.Box{Kind: "widget", Widget: str("Chat")}, false},
		{"widget without a name", models.Box{Kind: "widget"}, true},
		{"widget with an empty name", models.Box{Kind: "widget", Widget: str("")}, true},
		{"html with html", models.Box{Kind: "html", HTML: str("<p>hi</p>")}, false},
		{"html with empty html", models.Box{Kind: "html", HTML: str("")}, false},
		{"html without html", models.Box{Kind: "html"}, true},
		{"unknown kind", models.Box{Kind: "iframe", HTML: str("x")}, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := validateBox(&c.box)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
		})
	}

	// Geometry is clamped, not rejected: a drag past the canvas edge lands
	// at the edge, and a box can never shrink below a grabbable size.
	box := models.Box{Kind: "widget", Widget: str("Chat"), X: -30, Y: -1, W: 5, H: 0}
	if err := validateBox(&box); err != nil {
		t.Fatal(err)
	}
	if box.X != 0 || box.Y != 0 || box.W != minBoxSize || box.H != minBoxSize {
		t.Fatalf("clamped to %d,%d %dx%d; want 0,0 %dx%d", box.X, box.Y, box.W, box.H, minBoxSize, minBoxSize)
	}
}
