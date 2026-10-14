// Package graph is the GraphQL layer: the single POST /graphql endpoint that
// serves most of the site's data.
//
// How the pieces fit:
//
//   - graph/schema/*.graphql is the source of truth. Running
//     `go run github.com/99designs/gqlgen generate` regenerates generated.go
//     and model/models_gen.go, and adds stubs for any new field to the
//     matching *.resolvers.go. Never edit those two generated files.
//   - gqlgen.yml maps most GraphQL object types straight onto the GORM
//     structs in backend/models, so model/ holds only input and payload
//     types. A model's `ID uint` binds to `id: ID!` directly, because
//     gqlgen.yml lists graphql.UintID for the ID scalar; no resolver is
//     needed for it. A field resolver is only written where the Go type has
//     no binding (Message.authorId and Rowing's time/distance are unsigned in
//     Go and Int in the schema).
//   - Resolvers reach the database and the third-party caches through
//     Resolver.Store, the same handlers.Store the REST endpoints use.
//   - context.go and middleware.go carry the caller's identity from Gin into
//     the plain context.Context a resolver receives.
//
// Authorisation, which is the thing most likely to catch out a newcomer:
// there is NO middleware guarding this endpoint. AuthContextMiddleware only
// annotates the context and lets everything through. What guards a field is a
// directive on it in the schema:
//
//	createUser(input: CreateUserInput!): User! @admin
//	places: [Place!]! @signedIn
//
// @admin requires an admin's access token, @signedIn any valid one. Both are
// implemented once, in directives.go, and run before the resolver, so a
// guarded resolver contains no check of its own. A field with NEITHER
// directive is fully public.
//
// The rule of thumb in this schema is that queries are public and mutations
// are @admin, with these exceptions: login, logout and refreshToken are
// public; the places mutations and query are @signedIn; and the users
// queries are @admin (they are private data).
// directives_test.go holds that list and fails when the schema departs from
// it — in particular when a mutation is added with no directive.
//
// A directive is only a gate. Anything finer stays in the resolver and uses
// IsAdminFromCtx / UserIDFromCtx directly: a post can only be edited by its
// author, `me` answers for the caller (with its own "unauthorized" error),
// the messages query hides private messages from non-admins rather than
// refusing, and spotifyNeedsReauth is simply false for non-admins.
//
// The files ending in _helpers.go are hand-written and are not touched by
// gqlgen; the *.resolvers.go files are partly generated, so anything added to
// them by hand may be moved to the end of the file on the next generate.
package graph
