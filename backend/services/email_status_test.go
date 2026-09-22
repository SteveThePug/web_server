package services

import "testing"

// These tests cover the status progression rule only — statusAdvances, the
// predicate extracted out of updateJobApplication. The rest of
// updateJobApplication (note appending, the DB.Save, the returned counters)
// needs a live *gorm.DB and the only driver in go.mod is postgres, so it is
// deliberately not exercised here. Nothing in this file touches Claude or the
// Graph API either; email classification is a network call by construction.

func TestStatusAdvances(t *testing.T) {
	cases := []struct {
		name    string
		current string
		next    string
		want    bool
	}{
		{"one step forward is the normal case", "applied", "screening", true},
		{"skipping intermediate stages still advances", "applied", "offer", true},
		{"adjacent step backwards is refused", "interviewing", "assessment", false},
		{"a long way backwards is refused", "offer", "applied", false},
		{"same status is not progression, so a duplicate email is a no-op", "screening", "screening", false},

		// rejected and withdrawn sit at the top of statusOrder: once an
		// application lands there, no later email may move it.
		{"rejected cannot fall back to interviewing", "rejected", "interviewing", false},
		{"rejected cannot fall back to offer", "rejected", "offer", false},
		{"rejected cannot repeat", "rejected", "rejected", false},
		{"withdrawn cannot fall back to offer", "withdrawn", "offer", false},
		{"withdrawn cannot fall back to rejected", "withdrawn", "rejected", false},
		{"withdrawn cannot repeat", "withdrawn", "withdrawn", false},
		// The one ordering asymmetry between the two terminals: withdrawn
		// ranks above rejected, so this single pair is allowed.
		{"rejected may still become withdrawn because withdrawn ranks higher", "rejected", "withdrawn", true},

		{"an unknown new status is never applied", "applied", "hired", false},
		{"an empty new status is never applied", "applied", "", false},
		{"an unknown new status cannot escape a terminal state either", "rejected", "ghosted", false},
		{"a recognised status takes over from an unknown current status", "ghosted", "applied", true},
		{"a recognised status takes over from an empty current status", "", "offer", true},
		{"unknown to unknown stays put", "ghosted", "hired", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := statusAdvances(c.current, c.next); got != c.want {
				t.Errorf("statusAdvances(%q, %q) = %v, want %v", c.current, c.next, got, c.want)
			}
		})
	}
}

// Guards the claim the other tests rest on: rejected and withdrawn really are
// the top of the ranking, so adding a stage below them cannot silently make
// them non-terminal.
func TestTerminalStatusesRankHighest(t *testing.T) {
	for status, order := range statusOrder {
		if status == "rejected" || status == "withdrawn" {
			continue
		}
		if order >= statusOrder["rejected"] {
			t.Errorf("%q ranks %d, at or above rejected (%d)", status, order, statusOrder["rejected"])
		}
	}
}

// NormalizeStatus is the guard that keeps Claude's occasional casing drift
// ("Interview", "on hold") out of the status column: an unrecognised value
// stored verbatim makes statusAdvances treat the row's current status as
// unknown, which switches the forward-only rule off for that row entirely.
func TestNormalizeStatus(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"a tracked status passes through", "interviewing", "interviewing"},
		{"casing drift is lowered, not rejected", "Interviewing", "interviewing"},
		{"surrounding whitespace is trimmed", "  offer  ", "offer"},
		{"an untracked status falls back to the pipeline default", "on hold", "applied"},
		{"the UI's old Title Case value is recognised once lowered", "Rejected", "rejected"},
		{"a value the UI once allowed but the pipeline never had is rejected", "Interview", "applied"},
		{"empty falls back rather than storing a blank status", "", "applied"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeStatus(c.in); got != c.want {
				t.Errorf("NormalizeStatus(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
