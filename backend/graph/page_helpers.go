package graph

// Hand-written helpers for page.resolvers.go and box.resolvers.go, kept out
// of the resolver files for the same reason as box_helpers.go.

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"
)

// pageKeyPattern is the shape of a box's page key and of a page slug: lower
// case words joined by hyphens, in one or more /-separated segments. It is
// URL-safe as it stands, so the front end can use a slug as a path without
// escaping it.
var pageKeyPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*(/[a-z0-9]+(-[a-z0-9]+)*)*$`)

const maxSlugLen = 100

// reservedFirstSegments are paths a page could never be reached at, because
// nginx or one of the SPA's own routes answers them first (see
// nginx/nginx_routes.conf.template and vue/src/router/index.js), plus the
// built-in canvases' page keys, whose boxes a page of the same slug would
// share. "shrines" is deliberately absent: shrine pages live under it.
var reservedFirstSegments = map[string]bool{
	"stp": true, "stp2": true, "index": true, "admin": true, "cv": true,
	"automata": true, "hotels": true, "api": true, "py": true, "radio": true,
	"gitea": true, "notes": true, "sb": true, "jobs": true, "tracker": true,
	"usaco": true, "uploads": true, "img": true, "fonts": true, "pdf": true,
	"sound": true, "assets": true, "internal": true,
}

// reservedSlugs are individual paths the SPA routes itself.
var reservedSlugs = map[string]bool{
	"shrines":         true,
	"shrines/demoman": true,
}

func validPageKey(key string) bool {
	return len(key) <= maxSlugLen && pageKeyPattern.MatchString(key)
}

// validateSlug reports why slug can't be a page's address, or nil.
func validateSlug(slug string) error {
	if !validPageKey(slug) {
		return errors.New("a page address is lower-case letters, digits and hyphens, in /-separated parts, e.g. shrines/evangelion")
	}
	first, _, _ := strings.Cut(slug, "/")
	if reservedFirstSegments[first] || reservedSlugs[slug] {
		return errors.New("that address is already used by the site")
	}
	return nil
}

// pageHiddenFrom reports whether the page with this key is an adminOnly page
// and the caller is not an admin. A key with no Page row (a built-in canvas,
// or a page that doesn't exist) is not hidden.
func pageHiddenFrom(ctx context.Context, db *gorm.DB, key string) (bool, error) {
	if IsAdminFromCtx(ctx) {
		return false, nil
	}
	var count int64
	err := db.Model(&models.Page{}).Where("slug = ? AND admin_only = ?", key, true).Count(&count).Error
	return count > 0, err
}

// slugTaken reports whether a live page other than exceptID has this slug.
// Checked up front so the admin gets a sentence rather than a constraint
// violation; the partial unique index still backs it up.
func slugTaken(db *gorm.DB, slug string, exceptID uint) (bool, error) {
	var count int64
	err := db.Model(&models.Page{}).Where("slug = ? AND id <> ?", slug, exceptID).Count(&count).Error
	return count > 0, err
}
