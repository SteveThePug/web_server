package graph

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
)

// The schema's two authorisation directives. A field marked @admin or
// @signedIn in graph/schema/*.graphql has the matching function here run
// before its resolver; if it returns an error the resolver is never called
// and the error is what the client receives for that field.
//
// This replaces a check that used to be copied into the top of every guarded
// resolver. The rule itself is unchanged — it is the same IsAdminFromCtx /
// UserIDFromCtx test against the claims AuthContextMiddleware verified — but
// it is now declared next to the field, where a missing one can be seen and
// tested for (directives_test.go), instead of being a line someone has to
// remember to paste.
//
// The two error strings are part of the contract with the front end, which
// shows them as they are; the tracker page, for one, explains "admin access
// required" as a lapsed session. Do not reword them.
//
// This file is hand-written and never touched by gqlgen.

const (
	errAdminRequired    = "admin access required"
	errSignedInRequired = "authentication required"
)

// Directives is the set gqlgen needs in graph.Config. It is a function so
// that main and the tests wire exactly the same implementations.
//
// Leaving a directive nil in the Config does not open the fields that use it:
// the generated code answers "directive admin is not implemented" for each.
func Directives() DirectiveRoot {
	return DirectiveRoot{
		Admin:    requireAdmin,
		SignedIn: requireSignedIn,
	}
}

// requireAdmin implements @admin: the request must carry an access token
// whose admin claim is true. Admin comes from the token, not the database, so
// demoting a user only bites once their access token is replaced.
func requireAdmin(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	if !IsAdminFromCtx(ctx) {
		return nil, fmt.Errorf(errAdminRequired)
	}
	return next(ctx)
}

// requireSignedIn implements @signedIn: any verified access token will do,
// admin or not. An admin is signed in, so this never needs pairing with
// @admin.
func requireSignedIn(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	if _, ok := UserIDFromCtx(ctx); !ok {
		return nil, fmt.Errorf(errSignedInRequired)
	}
	return next(ctx)
}
