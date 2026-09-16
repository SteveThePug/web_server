package graph

import "adam-french.co.uk/backend/handlers"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

// Resolver is the root of the GraphQL schema and the only dependency every
// resolver has. Rather than listing dependencies individually it holds the
// same handlers.Store the REST handlers use, so GraphQL and REST share one
// database handle and one set of third-party caches.
//
// The generated code embeds *Resolver in per-type resolver structs
// (queryResolver, mutationResolver, postResolver, ...), which is why those
// one-line `struct{ *Resolver }` declarations appear at the bottom of each
// resolvers file.
type Resolver struct {
	Store *handlers.Store
}
