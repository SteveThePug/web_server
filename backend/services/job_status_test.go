package services

import (
	"bytes"
	"os"
	"slices"
	"sort"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// jobStatusEnumValues reads `enum JobStatus` out of the schema file itself.
// The test lives here, not in graph, because statusOrder is unexported and
// graph imports this package; parsing the file directly avoids both an
// export that exists only for a test and an import cycle.
func jobStatusEnumValues(t *testing.T) []string {
	t.Helper()
	const path = "../graph/schema/job_application.graphql"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the schema: %v", err)
	}
	// The file extends Query and Mutation and uses Time and the two
	// authorisation directives, all declared in schema.graphql; give it just
	// enough of those to parse on its own.
	const prelude = "scalar Time\ndirective @admin on FIELD_DEFINITION\ndirective @signedIn on FIELD_DEFINITION\ntype Query\ntype Mutation\n"
	schema, gqlErr := gqlparser.LoadSchema(
		&ast.Source{Name: "prelude", Input: prelude},
		&ast.Source{Name: path, Input: string(raw)},
	)
	if gqlErr != nil {
		t.Fatalf("parsing the schema: %v", gqlErr)
	}
	def := schema.Types["JobStatus"]
	if def == nil || def.Kind != ast.Enum {
		t.Fatal("the schema has no `enum JobStatus`")
	}
	var values []string
	for _, v := range def.EnumValues {
		values = append(values, v.Name)
	}
	return values
}

// One vocabulary, two places it has to be written down: the enum the GraphQL
// API validates against and the map the email pipeline ranks with. A status
// in the enum but not the map could be saved from the tracker and then never
// advanced; one in the map but not the enum could be stored by the pipeline
// and then not edited.
func TestJobStatusEnumMatchesStatusOrder(t *testing.T) {
	enum := jobStatusEnumValues(t)
	var ranked []string
	for status := range statusOrder {
		ranked = append(ranked, status)
	}
	sort.Strings(ranked)

	sorted := slices.Clone(enum)
	sort.Strings(sorted)
	if !slices.Equal(sorted, ranked) {
		t.Fatalf("enum JobStatus = %v\nstatusOrder    = %v\nthe two must list the same statuses", sorted, ranked)
	}

	// The schema lists them in lifecycle order, which is the order the
	// tracker shows; keep it agreeing with the ranks.
	for i, status := range enum {
		if statusOrder[status] != i {
			t.Errorf("enum lists %q at position %d but statusOrder ranks it %d", status, i, statusOrder[status])
		}
	}
}

// The exact strings, spelled out once: these are what is stored in the
// database and sent by the front end and the userscript, so a rename is a
// data migration, not an edit.
func TestJobStatusVocabulary(t *testing.T) {
	want := []string{"saved", "applied", "screening", "assessment", "interviewing", "offer", "rejected", "withdrawn"}
	if got := jobStatusEnumValues(t); !slices.Equal(got, want) {
		t.Fatalf("enum JobStatus = %v, want %v", got, want)
	}
}

func TestUnmarshalJobStatus(t *testing.T) {
	for status := range statusOrder {
		if got, err := UnmarshalJobStatus(status); err != nil || got != status {
			t.Errorf("UnmarshalJobStatus(%q) = %q, %v", status, got, err)
		}
	}
	// Strict on the way in: none of the forgiveness NormalizeStatus shows the
	// email pipeline. In particular nothing is coerced to "applied".
	for _, bad := range []any{"hired", "Applied", " applied", "", "on hold", 3, nil, true} {
		if got, err := UnmarshalJobStatus(bad); err == nil {
			t.Errorf("UnmarshalJobStatus(%#v) = %q, want an error", bad, got)
		}
	}
}

func TestMarshalJobStatus(t *testing.T) {
	cases := []struct {
		name   string
		stored string
		want   string
	}{
		{"a tracked status is a quoted string, as when the field was a String", "interviewing", `"interviewing"`},
		{"the tracker's old Title Case rows read back canonical", "Rejected", `"rejected"`},
		{"stray whitespace is dropped for a tracked status", " offer ", `"offer"`},
		// The reason this is a custom marshaller: a generated enum type would
		// error here, and status is non-null, so one such row would null the
		// whole jobApplications list.
		{"an untracked status is passed through, not refused", "Interview", `"Interview"`},
		{"an untracked status is not rewritten to the default", "on hold", `"on hold"`},
		{"an empty status stays empty", "", `""`},
		{"quotes in a stray value are escaped", `a"b`, `"a\"b"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var b bytes.Buffer
			MarshalJobStatus(c.stored).MarshalGQL(&b)
			if b.String() != c.want {
				t.Fatalf("MarshalJobStatus(%q) wrote %s, want %s", c.stored, &b, c.want)
			}
		})
	}
}
