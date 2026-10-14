package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

func newEnvManager(t *testing.T) *EnvManager {
	t.Helper()
	m, err := NewEnvManager(t.TempDir())
	if err != nil {
		t.Fatalf("NewEnvManager: %v", err)
	}
	return m
}

func readRequestFile(t *testing.T, m *EnvManager) envApplyRequest {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(m.dir, envRequestFile))
	if err != nil {
		t.Fatalf("reading request: %v", err)
	}
	var request envApplyRequest
	if err := json.Unmarshal(raw, &request); err != nil {
		t.Fatalf("parsing request: %v", err)
	}
	return request
}

// The registry and .env.example describe the same variables; a name added to
// one and not the other is either invisible in the panel or undocumented.
func TestEnvRegistryMatchesEnvExample(t *testing.T) {
	raw, err := os.ReadFile("../../.env.example")
	if err != nil {
		t.Fatalf("reading .env.example: %v", err)
	}
	var example []string
	for _, match := range regexp.MustCompile(`(?m)^([A-Z][A-Z0-9_]*)=`).FindAllStringSubmatch(string(raw), -1) {
		example = append(example, match[1])
	}

	var registry []string
	for _, v := range newEnvManager(t).Vars() {
		registry = append(registry, v.Name)
	}

	for _, name := range example {
		if !slices.Contains(registry, name) {
			t.Errorf("%s is in .env.example but not in env_registry.json", name)
		}
	}
	for _, name := range registry {
		if !slices.Contains(example, name) {
			t.Errorf("%s is in env_registry.json but not in .env.example", name)
		}
	}
}

// composeServiceEnv returns the variable names in the `environment:` list of
// one service in docker-compose.yml.
//
// It reads the file as text rather than through a YAML library — there is
// none among this module's direct dependencies, and the shape needed is
// narrow: a two-space-indented `<service>:` key under `services:`, a
// four-space-indented `environment:` under it, and `- NAME=value` items. If
// the compose file is ever restructured so that this finds nothing, the test
// below fails loudly rather than passing on an empty list.
func composeServiceEnv(t *testing.T, service string) []string {
	t.Helper()
	raw, err := os.ReadFile("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("reading docker-compose.yml: %v", err)
	}

	item := regexp.MustCompile(`^      - ["']?([A-Z][A-Z0-9_]*)(=|["']?\s*$)`)
	var names []string
	inService, inEnvironment := false, false
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		switch {
		case indent <= 2:
			// A new service (or a top-level key) ends the block.
			inService, inEnvironment = line == "  "+service+":", false
		case inService && indent == 4:
			inEnvironment = trimmed == "environment:"
		case inService && inEnvironment:
			match := item.FindStringSubmatch(line)
			if match == nil {
				t.Fatalf("%s.environment has an entry this test cannot read: %q", service, line)
			}
			names = append(names, match[1])
		}
	}
	if len(names) == 0 {
		t.Fatalf("found no %s.environment list in docker-compose.yml; has the file's layout changed?", service)
	}
	return names
}

// explicitEnvServices are the containers given an explicit `environment:`
// list instead of the whole .env, so that they never hold the database
// password or the JWT secret: the Python service, and the job tracker (an
// image from its own repo, which gets its own database role besides). The
// cost is a third place a new variable for one of them has to be added — the
// registry, .env.example and that list — and forgetting the third is silent:
// the variable is simply empty inside the container.
var explicitEnvServices = []string{"python", "jobs"}

// So the registry's claim "this service reads this" and the compose list must
// agree, in both directions:
//
//   - a variable the registry says the service reads must be passed to it;
//   - an editable variable that is passed to it must say so in the registry,
//     or changing it in the admin panel would not recreate the container.
//
// Locked variables are exempt from the second rule. They are never applied
// through the panel, so the registry gives none of them a services list
// (PYTHON_PORT is passed to python and is one of these). Names in the compose
// list that are not registry variables at all, such as the literal ROOT_PATH
// or the jobs service's composed DATABASE_URL, are not this test's business.
func TestEnvRegistryMatchesComposeEnvironments(t *testing.T) {
	registry := map[string]EnvVar{}
	for _, v := range newEnvManager(t).Vars() {
		registry[v.Name] = v
	}

	for _, service := range explicitEnvServices {
		compose := composeServiceEnv(t, service)

		for _, v := range registry {
			if slices.Contains(v.Services, service) && !slices.Contains(compose, v.Name) {
				t.Errorf("%s lists the %s service in env_registry.json but is not in "+
					"%s.environment in docker-compose.yml, so the container never receives it", v.Name, service, service)
			}
		}

		for _, name := range compose {
			v, known := registry[name]
			if !known || !v.Editable {
				continue
			}
			if !slices.Contains(v.Services, service) {
				t.Errorf("%s is passed to the %s service in docker-compose.yml but env_registry.json "+
					"does not list %s among its services, so editing it would not recreate the container", name, service, service)
			}
		}
	}
}

func TestEnvRegistryEntriesAreWellFormed(t *testing.T) {
	kinds := []string{"secret", "generated", "text", "url", "bool", "duration", "number", "enum"}
	for _, v := range newEnvManager(t).Vars() {
		if !slices.Contains(kinds, v.Kind) {
			t.Errorf("%s has unknown kind %q", v.Name, v.Kind)
		}
		if v.Group == "" || v.Purpose == "" {
			t.Errorf("%s is missing its group or purpose", v.Name)
		}
		if v.Editable && len(v.Services) == 0 {
			t.Errorf("%s is editable but names no service to recreate", v.Name)
		}
		if !v.Editable && v.LockedReason == "" {
			t.Errorf("%s is locked without a reason", v.Name)
		}
		if v.Kind == "enum" && len(v.Options) == 0 {
			t.Errorf("%s is an enum with no options", v.Name)
		}
	}
}

func TestEnvValidate(t *testing.T) {
	m := newEnvManager(t)
	cases := []struct {
		name, value string
		ok          bool
	}{
		{"CLAUDE_API_KEY", "sk-ant-abc123", true},
		{"CLAUDE_API_KEY", "", true},
		// The injection case: a newline would start a second variable.
		{"CLAUDE_API_KEY", "x\nPOSTGRES_PASSWORD=owned", false},
		{"CLAUDE_API_KEY", "x\rY", false},
		{"CLAUDE_API_KEY", "it's", false},
		{"CLAUDE_API_KEY", strings.Repeat("a", envValueMaxLen+1), false},
		{"POSTGRES_PASSWORD", "anything", false},
		{"NOT_A_VARIABLE", "anything", false},
		{"EMAIL_BACKEND", "graph", true},
		{"EMAIL_BACKEND", "pop3", false},
		{"EMAIL_SYNC_INTERVAL", "45m", true},
		{"EMAIL_SYNC_INTERVAL", "soon", false},
		{"EMAIL_SYNC_INTERVAL", "-5m", false},
		{"EMAIL_SYNC_ENABLED", "true", true},
		{"EMAIL_SYNC_ENABLED", "yes", false},
		{"IMAP_PORT", "993", true},
		{"IMAP_PORT", "imaps", false},
		{"SPOTIFY_REDIRECT_URI", "https://example.com/api/spotify/callback", true},
		{"SPOTIFY_REDIRECT_URI", "example.com/callback", false},
		{"SPOTIFY_REDIRECT_URI", "javascript:alert(1)", false},
	}
	for _, c := range cases {
		err := m.Validate(c.name, c.value)
		if (err == nil) != c.ok {
			t.Errorf("Validate(%s, %q) error = %v, want ok = %v", c.name, c.value, err, c.ok)
		}
	}
}

func TestEnvSubmitWritesRequest(t *testing.T) {
	m := newEnvManager(t)
	generated, services, err := m.Submit(
		map[string]string{"TFL_APP_KEY": "  tfl-key  ", "STEAM_ID": ""},
		[]string{"ICECAST_RELAY_PASSWORD"},
	)
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if len(generated["ICECAST_RELAY_PASSWORD"]) != 64 {
		t.Fatalf("generated value = %q, want 64 hex characters", generated["ICECAST_RELAY_PASSWORD"])
	}
	if want := []string{"backend", "icecast2", "python"}; !slices.Equal(services, want) {
		t.Fatalf("services = %v, want %v", services, want)
	}

	request := readRequestFile(t, m)
	want := map[string]string{
		"TFL_APP_KEY":            "tfl-key", // surrounding whitespace from a paste is dropped
		"STEAM_ID":               "",
		"ICECAST_RELAY_PASSWORD": generated["ICECAST_RELAY_PASSWORD"],
	}
	if len(request.Set) != len(want) {
		t.Fatalf("request.Set = %v, want %v", request.Set, want)
	}
	for name, value := range want {
		if request.Set[name] != value {
			t.Errorf("request.Set[%s] = %q, want %q", name, request.Set[name], value)
		}
	}

	info, err := os.Stat(filepath.Join(m.dir, envRequestFile))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("request mode = %v, want 0600: it holds secrets in the clear", info.Mode().Perm())
	}
}

func TestEnvSubmitRejectsWholeBatch(t *testing.T) {
	m := newEnvManager(t)
	cases := []struct {
		name     string
		set      map[string]string
		generate []string
	}{
		{"one bad value", map[string]string{"STEAM_ID": "1", "EMAIL_BACKEND": "pop3"}, nil},
		{"a locked variable", map[string]string{"STEAM_ID": "1", "DOMAIN": "evil.example"}, nil},
		{"generating a pasted secret", nil, []string{"CLAUDE_API_KEY"}},
		{"generating a locked secret", nil, []string{"POSTGRES_PASSWORD"}},
		{"nothing at all", nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, err := m.Submit(c.set, c.generate); err == nil {
				t.Fatal("Submit succeeded, want an error")
			}
			if _, err := os.Stat(filepath.Join(m.dir, envRequestFile)); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("a rejected batch left a request file behind")
			}
		})
	}
}

func TestEnvSubmitMergesUnclaimedRequest(t *testing.T) {
	m := newEnvManager(t)
	if _, _, err := m.Submit(map[string]string{"STEAM_ID": "1", "TFL_APP_KEY": "old"}, nil); err != nil {
		t.Fatalf("first Submit: %v", err)
	}
	if _, _, err := m.Submit(map[string]string{"TFL_APP_KEY": "new"}, nil); err != nil {
		t.Fatalf("second Submit: %v", err)
	}
	request := readRequestFile(t, m)
	if request.Set["STEAM_ID"] != "1" || request.Set["TFL_APP_KEY"] != "new" {
		t.Fatalf("merged request = %v, want the earlier STEAM_ID kept and TFL_APP_KEY replaced", request.Set)
	}
}

func TestEnvSubmitWithoutMount(t *testing.T) {
	m, err := NewEnvManager(filepath.Join(t.TempDir(), "absent"))
	if err != nil {
		t.Fatalf("NewEnvManager: %v", err)
	}
	if _, _, err := m.Submit(map[string]string{"STEAM_ID": "1"}, nil); err == nil {
		t.Fatal("Submit succeeded with no shared directory")
	}
	if got := m.Status().State; got != "idle" {
		t.Fatalf("Status().State = %q, want idle", got)
	}
}

func writeStatusFile(t *testing.T, m *EnvManager, state string, at time.Time) {
	t.Helper()
	data, _ := json.Marshal(EnvApplyStatus{State: state, ID: "abc", At: &at})
	if err := os.WriteFile(filepath.Join(m.dir, envStatusFile), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestEnvStatus(t *testing.T) {
	m := newEnvManager(t)
	if got := m.Status(); got.State != "idle" || !got.Watcher {
		t.Fatalf("fresh status = %+v, want idle with watcher assumed", got)
	}

	writeStatusFile(t, m, "applied", time.Now())
	if got := m.Status(); got.State != "applied" || got.ID != "abc" {
		t.Fatalf("status = %+v, want the host's applied report", got)
	}

	// An unclaimed request outranks the report of an older change.
	if _, _, err := m.Submit(map[string]string{"CLAUDE_API_KEY": "sk-secret"}, nil); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	got := m.Status()
	if got.State != "pending" || !slices.Equal(got.Pending, []string{"CLAUDE_API_KEY"}) || !got.Watcher {
		t.Fatalf("status = %+v, want pending CLAUDE_API_KEY", got)
	}
	if encoded, _ := json.Marshal(got); strings.Contains(string(encoded), "sk-secret") {
		t.Fatal("the status leaks a pending value")
	}

	// Backdate the request past the grace period: nothing is watching.
	request := readRequestFile(t, m)
	request.RequestedAt = time.Now().Add(-2 * envWatcherGrace)
	if err := m.writeRequest(&request); err != nil {
		t.Fatal(err)
	}
	if m.Status().Watcher {
		t.Fatal("a long-unclaimed request still reports a watcher")
	}

	if err := m.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if got := m.Status().State; got != "applied" {
		t.Fatalf("after Cancel state = %q, want the earlier applied report", got)
	}
	if err := m.Cancel(); err != nil {
		t.Fatalf("Cancel with nothing pending: %v", err)
	}
}

func TestEnvSubmitRefusedWhileApplying(t *testing.T) {
	m := newEnvManager(t)
	writeStatusFile(t, m, "applying", time.Now())
	if _, _, err := m.Submit(map[string]string{"STEAM_ID": "1"}, nil); !errors.Is(err, ErrEnvApplying) {
		t.Fatalf("Submit error = %v, want ErrEnvApplying", err)
	}

	// A host script that died mid-run must not block the panel for ever.
	writeStatusFile(t, m, "applying", time.Now().Add(-2*envApplyingStale))
	if _, _, err := m.Submit(map[string]string{"STEAM_ID": "1"}, nil); err != nil {
		t.Fatalf("Submit after a stale applying status: %v", err)
	}
}
