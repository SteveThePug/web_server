package graph

import (
	"testing"

	"adam-french.co.uk/backend/graph/model"
)

func TestBackgroundFrom(t *testing.T) {
	for _, c := range []struct {
		page  string
		input model.BackgroundInput
		ok    bool
	}{
		{"index", model.BackgroundInput{Color: "#102030", Fit: "cover"}, true},
		{"shrines/gto", model.BackgroundInput{Image: "/uploads/abc.gif", Fit: "tile"}, true},
		{"stp", model.BackgroundInput{Color: "rgb(0 0 0 / 50%)", Image: "https://example.com/a.png", Fit: "contain"}, true},
		{"index", model.BackgroundInput{Color: "rebeccapurple", Fit: "cover"}, true},
		{"index", model.BackgroundInput{Fit: "cover"}, true},    // plain theme colour
		{"Index", model.BackgroundInput{Fit: "cover"}, false},   // bad page key
		{"index", model.BackgroundInput{Fit: "stretch"}, false}, // unknown fit
		{"index", model.BackgroundInput{Color: "red; color: blue", Fit: "cover"}, false},
		{"index", model.BackgroundInput{Color: "url(/x.png)", Fit: "cover"}, false}, // an image in disguise
		{"index", model.BackgroundInput{Image: `/x.png") , url("/y`, Fit: "cover"}, false},
		{"index", model.BackgroundInput{Image: "javascript:alert(1)", Fit: "cover"}, false},
		{"index", model.BackgroundInput{Image: "uploads/relative.png", Fit: "cover"}, false},
	} {
		_, err := backgroundFrom(c.page, c.input)
		if (err == nil) != c.ok {
			t.Errorf("backgroundFrom(%q, %+v) = %v, want ok=%v", c.page, c.input, err, c.ok)
		}
	}
}
