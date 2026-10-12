package services

import (
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

// The job-status vocabulary, as GraphQL sees it.
//
// The schema declares `enum JobStatus` (graph/schema/job_application.graphql)
// and gqlgen.yml binds it to the two functions below instead of letting
// gqlgen generate a Go enum type. That keeps models.JobApplication.Status a
// plain string column — no migration, and the email pipeline and the capture
// endpoint go on writing strings — while the GraphQL boundary still refuses a
// status nobody ranked.
//
// statusOrder (email_sync.go) is the vocabulary; the enum in the schema has to
// list the same values, and TestJobStatusEnumMatchesStatusOrder fails when
// they differ. The functions live in this package, rather than in graph,
// because this is where statusOrder is and graph already imports services.

// KnownStatus reports whether s is exactly one of the tracked statuses. No
// trimming or case folding: see NormalizeStatus for the forgiving version the
// email pipeline and the capture endpoint use on values that did not come
// through GraphQL.
func KnownStatus(s string) bool {
	_, ok := statusOrder[s]
	return ok
}

// MarshalJobStatus writes a stored status to a GraphQL response.
//
// It is deliberately lenient, where the input side is strict. The column is
// older than the enum, so a row can hold a status outside the vocabulary: the
// tracker's early Title Case values ("Rejected"), or something written by
// hand. A generated enum type would fail to serialise such a row, and because
// JobApplication.status is non-null that one error would null out the whole
// jobApplications list — one stale row hiding every application. So:
//
//   - a value that is in the vocabulary once trimmed and lowered is emitted
//     in that canonical form, which also means an edit form sends back
//     something the input side accepts;
//   - anything else is emitted exactly as stored. That is not a legal
//     JobStatus, strictly, but the response is JSON and the tracker page
//     shows an unlisted status as-is so that it can be corrected, which is
//     more use than an error or a made-up substitute.
//
// The value is written with gqlgen's own String marshaller, so it is the same
// quoted JSON string the field produced when it was declared String.
func MarshalJobStatus(s string) graphql.Marshaler {
	if canonical := strings.ToLower(strings.TrimSpace(s)); KnownStatus(canonical) {
		s = canonical
	}
	return graphql.MarshalString(s)
}

// UnmarshalJobStatus reads a status from a mutation's input, and is where
// drift is refused: only the exact lowercase values are accepted, so nothing
// outside the vocabulary can be written through GraphQL.
//
// The value arrives as a string whether the client sent a JSON string
// variable ("applied", which is what the tracker page does) or wrote the enum
// literal in the query document (applied).
func UnmarshalJobStatus(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("status must be a string")
	}
	if !KnownStatus(s) {
		return "", fmt.Errorf("%q is not a valid JobStatus", s)
	}
	return s, nil
}
