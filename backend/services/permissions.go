package services

import (
	"slices"

	"adam-french.co.uk/backend/models"
)

// Per-service permissions: how a non-admin account is let into one gated
// service (the receipt printer, say) without being made an admin.
//
// A permission is a key from PermissionTypes stored in models.User.Permissions.
// nginx gates a service on one key by asking GET /auth/validate/<key>
// (handlers.Store.ValidatePermission); admins pass every such check, so
// moving a service from the admin gate to a permission gate never locks an
// admin out. Services that are not listed here stay admin-only.
//
// Adding a grantable service is one entry below plus swapping that location's
// nginx gate from /internal/auth/admin-validate to the permission gate (see
// "Permission gate" in nginx/nginx_routes.conf.template).

// PermissionType describes one grantable service. Key is what is stored and
// what nginx names; Label and Path are for the admin panel's checkboxes and
// the signed-in user's /account list of links.
type PermissionType struct {
	Key   string
	Label string
	Path  string
}

// PermissionTypes is the full set of grantable permissions, in display order.
// Removing an entry leaves the key in users' rows, where it is inert: the
// validate endpoint 404s an unknown key, so nothing can be gated on it.
var PermissionTypes = []PermissionType{
	{Key: "receipt", Label: "Receipt printer", Path: "/receipt/"},
}

// ValidPermission reports whether key names a permission in PermissionTypes.
func ValidPermission(key string) bool {
	return slices.ContainsFunc(PermissionTypes, func(p PermissionType) bool {
		return p.Key == key
	})
}

// HasPermission reports whether user may use the service gated on key. It
// reads the row, not a token, so Admin here is the current database value.
func HasPermission(user *models.User, key string) bool {
	return user.Admin || slices.Contains(user.Permissions, key)
}
