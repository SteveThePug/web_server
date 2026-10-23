package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vektah/gqlparser/v2/ast"

	"adam-french.co.uk/backend/handlers"
)

// Authorisation is declared in the schema (@admin / @signedIn) and enforced
// by directives.go. These tests cover both halves: that the two functions
// decide correctly, and that the schema marks the fields it should — the
// second being the one that matters day to day, since the way this goes
// wrong is a new field added without a directive.
//
// Nothing here needs a database: a refused field never reaches its resolver,
// and the tests only ever drive fields that are refused or that return before
// touching the Store.

// withClaims builds the context AuthContextMiddleware would for a verified
// token. Claims are decoded from JSON in production, hence float64 for the id.
func withClaims(id float64, admin bool) context.Context {
	claims := jwt.MapClaims{"id": id, "admin": admin}
	return context.WithValue(context.Background(), userClaimsKey, &claims)
}

func TestDirectives(t *testing.T) {
	anonymous := context.Background()
	user := withClaims(2, false)
	admin := withClaims(1, true)
	// A token that verified but carries no id: not something GenerateJWT
	// mints, but it must not count as signed in.
	noID := func() context.Context {
		claims := jwt.MapClaims{"admin": false}
		return context.WithValue(context.Background(), userClaimsKey, &claims)
	}()

	cases := []struct {
		name      string
		directive func(context.Context, any, graphql.Resolver) (any, error)
		ctx       context.Context
		wantErr   string // "" means the resolver must run
	}{
		{"@admin refuses an anonymous request", requireAdmin, anonymous, "admin access required"},
		{"@admin refuses a signed-in non-admin", requireAdmin, user, "admin access required"},
		{"@admin admits an admin", requireAdmin, admin, ""},
		{"@signedIn refuses an anonymous request", requireSignedIn, anonymous, "authentication required"},
		{"@signedIn refuses a token with no id", requireSignedIn, noID, "authentication required"},
		{"@signedIn admits a non-admin", requireSignedIn, user, ""},
		{"@signedIn admits an admin", requireSignedIn, admin, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			called := false
			next := func(ctx context.Context) (any, error) {
				called = true
				return "resolved", nil
			}
			res, err := c.directive(c.ctx, nil, next)

			if c.wantErr == "" {
				if err != nil || !called || res != "resolved" {
					t.Fatalf("res = %v, err = %v, resolver called = %v; want the resolver's result", res, err, called)
				}
				return
			}
			// The exact strings are what the front end displays.
			if err == nil || err.Error() != c.wantErr {
				t.Fatalf("err = %v, want %q", err, c.wantErr)
			}
			if called {
				t.Fatal("the resolver ran for a refused request")
			}
			if res != nil {
				t.Fatalf("res = %v, want nil alongside the error", res)
			}
		})
	}
}

// Directives() must supply every directive the schema declares; a nil one
// makes gqlgen fail each field that uses it.
func TestDirectivesAreAllImplemented(t *testing.T) {
	d := Directives()
	if d.Admin == nil || d.SignedIn == nil {
		t.Fatal("Directives() left a directive unimplemented")
	}
}

// guardOf reports which authorisation directive a field carries: "admin",
// "signedIn", or "" for none. Carrying both is a mistake (@signedIn adds
// nothing to @admin) and fails the test rather than being quietly accepted.
func guardOf(t *testing.T, field *ast.FieldDefinition) string {
	t.Helper()
	admin := field.Directives.ForName("admin") != nil
	signedIn := field.Directives.ForName("signedIn") != nil
	switch {
	case admin && signedIn:
		t.Errorf("%s carries both @admin and @signedIn", field.Name)
		return "admin"
	case admin:
		return "admin"
	case signedIn:
		return "signedIn"
	}
	return ""
}

func schemaUnderTest() *ast.Schema {
	return NewExecutableSchema(Config{}).Schema()
}

// publicMutations is the complete list of mutations that anyone may call.
// A mutation is public only by being named here, so adding one to the schema
// without a directive fails TestEveryMutationIsGuarded until it is either
// guarded or deliberately listed.
//
// All three are how a session is obtained or given up, so they cannot require
// one: login checks a password, refreshToken checks the refresh cookie, and
// logout is a no-op for a visitor who is not signed in.
var publicMutations = []string{"login", "logout", "refreshToken"}

func TestEveryMutationIsGuarded(t *testing.T) {
	schema := schemaUnderTest()
	var seen []string
	for _, field := range schema.Mutation.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		seen = append(seen, field.Name)
		guard := guardOf(t, field)
		public := slices.Contains(publicMutations, field.Name)
		switch {
		case guard == "" && !public:
			t.Errorf("mutation %s has neither @admin nor @signedIn, so anyone can call it; "+
				"add a directive in the schema, or add it to publicMutations if that is intended", field.Name)
		case guard != "" && public:
			t.Errorf("mutation %s is listed as public but carries @%s", field.Name, guard)
		}
	}
	// A renamed or removed mutation must not leave a stale exemption behind.
	for _, name := range publicMutations {
		if !slices.Contains(seen, name) {
			t.Errorf("publicMutations lists %s, which is not in the schema", name)
		}
	}
}

// The mutations that need only a session rather than an admin. Pinned so that
// loosening a mutation from @admin to @signedIn — which the test above would
// accept — has to be done here too.
func TestSignedInMutationsArePinned(t *testing.T) {
	want := []string{"createPlace", "deletePlace", "updatePlace"}
	var got []string
	for _, field := range schemaUnderTest().Mutation.Fields {
		if guardOf(t, field) == "signedIn" {
			got = append(got, field.Name)
		}
	}
	sort.Strings(got)
	if !slices.Equal(got, want) {
		t.Fatalf("@signedIn mutations = %v, want %v", got, want)
	}
}

// Queries are public by default — the home page is built from them — so the
// rule cannot be "every field is guarded". Instead the guarded set is pinned
// exactly: removing a directive from one of these, or adding a query that
// exposes private data without one, has to be a visible change to this map.
//
// Three queries are sensitive to who is asking without being in it, because
// they answer rather than refuse: `me` returns "unauthorized" itself,
// `messages` filters out private rows for non-admins, and
// `spotifyNeedsReauth` is false for non-admins.
func TestGuardedQueriesArePinned(t *testing.T) {
	want := map[string]string{
		"users":           "admin",
		"user":            "admin",
		"places":          "signedIn",
		"layouts":         "admin",
		"permissionTypes": "signedIn",
	}
	got := map[string]string{}
	for _, field := range schemaUnderTest().Query.Fields {
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		if guard := guardOf(t, field); guard != "" {
			got[field.Name] = guard
		}
	}
	for name, guard := range want {
		if got[name] != guard {
			t.Errorf("query %s: guard = %q, want @%s", name, got[name], guard)
		}
	}
	for name, guard := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("query %s carries @%s but is not in the pinned list; add it there if the guard is intended", name, guard)
		}
	}
}

// The directives are FIELD_DEFINITION only and are meant for the two root
// types. One on an object's field would be enforced, but nothing above would
// notice it being removed, so keep them where the tests look.
func TestDirectivesOnlyOnRootFields(t *testing.T) {
	schema := schemaUnderTest()
	for name, def := range schema.Types {
		if def == schema.Query || def == schema.Mutation || strings.HasPrefix(name, "__") {
			continue
		}
		for _, field := range def.Fields {
			if guardOf(t, field) != "" {
				t.Errorf("%s.%s carries an authorisation directive outside Query/Mutation", name, field.Name)
			}
		}
	}
}

// execute runs one operation through the real executable schema, wired the
// way main wires it, and returns the response's error messages. The Store is
// empty: every operation below is refused, or returns, before its resolver
// would touch the database.
func execute(t *testing.T, ctx context.Context, query string, variables map[string]any) []string {
	t.Helper()
	srv := handler.New(NewExecutableSchema(Config{
		Resolvers:  &Resolver{Store: &handlers.Store{}},
		Directives: Directives(),
	}))
	srv.AddTransport(transport.POST{})

	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response %s: %v", rec.Body, err)
	}
	var messages []string
	for _, e := range resp.Errors {
		messages = append(messages, e.Message)
	}
	return messages
}

// End to end: the message a client receives for a refused field is the bare
// string the resolvers used to return, with nothing wrapped around it.
func TestGuardedFieldsRefuseOverTheWire(t *testing.T) {
	cases := []struct {
		name  string
		ctx   context.Context
		query string
		want  string
	}{
		{"anonymous admin query", context.Background(), `{ users { id } }`, "admin access required"},
		{"non-admin admin query", withClaims(2, false), `{ users { id } }`, "admin access required"},
		{"anonymous admin mutation", context.Background(), `mutation { deleteBookmark(id: 1) { id } }`, "admin access required"},
		{"non-admin admin mutation", withClaims(2, false), `mutation { deleteUser(id: 1) { id } }`, "admin access required"},
		{"anonymous signed-in query", context.Background(), `{ places { id } }`, "authentication required"},
		{"anonymous signed-in mutation", context.Background(), `mutation { deletePlace(id: 1) { id } }`, "authentication required"},
		// Not a directive, and a different message on purpose.
		{"me keeps its own message", context.Background(), `{ me { id } }`, "unauthorized"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := execute(t, c.ctx, c.query, nil)
			if !slices.Equal(got, []string{c.want}) {
				t.Fatalf("errors = %q, want exactly %q", got, c.want)
			}
		})
	}
}

// IDs: the wire form must be what it was when every type had a hand-written
// resolver returning int. Both marshallers write a quoted decimal string and
// both read a string or a number.
func TestUintIDWireFormatMatchesIntID(t *testing.T) {
	for _, id := range []uint{0, 1, 12, 4294967295} {
		var asUint, asInt bytes.Buffer
		graphql.MarshalUintID(id).MarshalGQL(&asUint)
		graphql.MarshalIntID(int(id)).MarshalGQL(&asInt)
		if asUint.String() != asInt.String() {
			t.Errorf("id %d: UintID writes %s, IntID wrote %s", id, &asUint, &asInt)
		}
	}
	if got := func() string {
		var b bytes.Buffer
		graphql.MarshalUintID(12).MarshalGQL(&b)
		return b.String()
	}(); got != `"12"` {
		t.Errorf("MarshalUintID(12) = %s, want \"12\"", got)
	}
}
