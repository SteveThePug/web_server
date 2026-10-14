package graph

// Hand-written helpers for background.resolvers.go, kept out of the resolver
// file for the same reason as box_helpers.go.

import (
	"errors"
	"regexp"
	"strings"

	"adam-french.co.uk/backend/graph/model"
	"adam-french.co.uk/backend/models"
)

// The SPA puts these values straight into a style binding, as
// `background-color` and `url("<image>")`. Vue sets them through the CSSOM,
// so nothing can escape into other properties, but a quote or bracket in the
// image could still end the url() early — hence the narrow shapes.
var (
	// A hex colour, a colour name, or one of the colour functions — named,
	// so that url() and friends don't pass as a "function".
	bgColorPattern = regexp.MustCompile(`^(#[0-9a-fA-F]{3,8}|[a-zA-Z]+|(rgba?|hsla?|hwb|lab|lch|oklab|oklch)\([0-9a-z.,%/ -]*\))$`)
	// A site path (an upload, /img/...) or an http(s) URL, with no quotes,
	// brackets, backslashes or spaces.
	bgImagePattern = regexp.MustCompile(`^(/|https?://)[^\s"'()\\<>]*$`)
)

const maxBgImageLen = 1000

var bgFits = map[string]bool{"cover": true, "contain": true, "tile": true}

// backgroundFrom validates input into a Background for page.
func backgroundFrom(page string, input model.BackgroundInput) (models.Background, error) {
	bg := models.Background{
		Page:  page,
		Color: strings.TrimSpace(input.Color),
		Image: strings.TrimSpace(input.Image),
		Fit:   input.Fit,
	}
	if !validPageKey(page) {
		return bg, errors.New("a background needs a valid page")
	}
	if bg.Color != "" && (len(bg.Color) > 64 || !bgColorPattern.MatchString(bg.Color)) {
		return bg, errors.New(`colour must be a CSS colour such as #1b110e, black or rgb(0 0 0 / 50%)`)
	}
	if bg.Image != "" && (len(bg.Image) > maxBgImageLen || !bgImagePattern.MatchString(bg.Image)) {
		return bg, errors.New("image must be a path on this site (e.g. /uploads/...) or an http(s) URL, without spaces or quotes")
	}
	if !bgFits[bg.Fit] {
		return bg, errors.New(`fit must be "cover", "contain" or "tile"`)
	}
	return bg, nil
}
