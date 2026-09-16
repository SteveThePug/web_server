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
//     types. That is also why almost every type has a hand-written ID
//     resolver: model IDs are uint, GraphQL's Int is a Go int.
//   - Resolvers reach the database and the third-party caches through
//     Resolver.Store, the same handlers.Store the REST endpoints use.
//   - context.go and middleware.go carry the caller's identity from Gin into
//     the plain context.Context a resolver receives.
//
// Authorisation, which is the thing most likely to catch out a newcomer:
// there is NO middleware guarding this endpoint. AuthContextMiddleware only
// annotates the context and lets everything through, so each resolver checks
// for itself with IsAdminFromCtx or UserIDFromCtx. A resolver that omits that
// check is fully public. The rule of thumb in this schema is that queries are
// public and mutations are admin-only, with these exceptions: the job
// application and job reference queries are admin-only (they are private
// data), the places mutations need only a signed-in user, and the messages
// query hides private messages from non-admins rather than refusing.
//
// The files ending in _helpers.go are hand-written and are not touched by
// gqlgen; the *.resolvers.go files are partly generated, so anything added to
// them by hand may be moved to the end of the file on the next generate.
package graph
