package services

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ENV_APPLY_DIR is the directory shared with the host (./env-apply/run in the
// checkout, bind-mounted by docker-compose). It is the whole interface
// between this process and the host script that edits .env.
const ENV_APPLY_DIR = "/backend/env-apply"

// The registry is JSON rather than a Go literal because the host script
// (env-apply/apply.py) reads the same file from the checkout: one list
// decides both what the panel offers and what the host will accept.
//
//go:embed env_registry.json
var envRegistryJSON []byte

// EnvVar is one registry entry.
//
// Kind decides how a value is validated and how the panel renders it:
// "secret" and "generated" are never sent to the browser; "generated" can
// also be minted here. Services lists the compose services that read the
// variable, which is what the host recreates after a change.
type EnvVar struct {
	Name         string   `json:"name"`
	Group        string   `json:"group"`
	Purpose      string   `json:"purpose"`
	Kind         string   `json:"kind"`
	Editable     bool     `json:"editable"`
	LockedReason string   `json:"lockedReason,omitempty"`
	Services     []string `json:"services,omitempty"`
	DocsURL      string   `json:"docsUrl,omitempty"`
	Options      []string `json:"options,omitempty"`
	Warning      string   `json:"warning,omitempty"`
}

// Secret reports whether the variable's value must stay on the server.
func (v EnvVar) Secret() bool {
	return v.Kind == "secret" || v.Kind == "generated"
}

// EnvManager turns an edit made in the admin panel into a request for the
// host to apply.
//
// The backend deliberately cannot write .env or restart containers itself:
// it is the internet-facing process, and either ability would make a
// compromised admin session a compromised host. Instead it validates the
// change and leaves request.json in the shared directory; the host's watcher
// (env-apply/apply.py, run by a systemd path unit) merges it into .env,
// recreates the affected containers and reports back through status.json.
type EnvManager struct {
	dir    string
	vars   []EnvVar
	byName map[string]EnvVar

	// Serialises Submit and Cancel so two requests cannot interleave their
	// read-merge-write of request.json. The host is a separate process and is
	// handled by renames instead (see Submit).
	mu sync.Mutex
}

// envApplyRequest is request.json. Clearing a variable is setting it to "".
type envApplyRequest struct {
	ID          string            `json:"id"`
	RequestedAt time.Time         `json:"requestedAt"`
	Set         map[string]string `json:"set"`
}

// EnvApplyStatus is what the panel polls. State is one of idle, pending
// (request written, host has not picked it up), applying, applied, failed.
type EnvApplyStatus struct {
	State string `json:"state"`
	ID    string `json:"id,omitempty"`
	// When the state was entered.
	At *time.Time `json:"at,omitempty"`
	// Names in a pending request. Values are never reported.
	Pending  []string `json:"pending,omitempty"`
	Services []string `json:"services,omitempty"`
	Changed  []string `json:"changed,omitempty"`
	Output   string   `json:"output,omitempty"`
	// False once a request has waited long enough that nothing on the host
	// can be watching for it (a dev machine, or the units not installed).
	Watcher bool `json:"watcher"`
}

const (
	envRequestFile = "request.json"
	envStatusFile  = "status.json"

	// The path unit fires within a second, so a request still unclaimed
	// after this long means there is no watcher.
	envWatcherGrace = 15 * time.Second
	// An "applying" status older than this is a host script that died
	// mid-run; stop letting it block new requests.
	envApplyingStale = 5 * time.Minute

	envValueMaxLen = 4096
)

// ErrEnvApplying is returned by Submit while the host is mid-apply.
var ErrEnvApplying = errors.New("a change is being applied; try again in a moment")

// NewEnvManager parses the embedded registry. dir need not exist yet: without
// the bind mount, reads report "idle" and Submit fails with a clear error.
func NewEnvManager(dir string) (*EnvManager, error) {
	var vars []EnvVar
	if err := json.Unmarshal(envRegistryJSON, &vars); err != nil {
		return nil, fmt.Errorf("parsing env registry: %w", err)
	}
	byName := make(map[string]EnvVar, len(vars))
	for _, v := range vars {
		if _, dup := byName[v.Name]; dup {
			return nil, fmt.Errorf("env registry lists %s twice", v.Name)
		}
		byName[v.Name] = v
	}
	return &EnvManager{dir: dir, vars: vars, byName: byName}, nil
}

// Vars returns the registry in file order, which is the order the panel
// shows.
func (m *EnvManager) Vars() []EnvVar {
	return m.vars
}

// Validate checks one proposed value. The empty string clears the variable
// and is always acceptable for an editable one.
//
// The control-character rule is the security-relevant one: .env is line
// oriented, so a newline in a value would let an edit to an allowed variable
// write a second, arbitrary one.
func (m *EnvManager) Validate(name, value string) error {
	v, ok := m.byName[name]
	if !ok {
		return fmt.Errorf("%s is not a known variable", name)
	}
	if !v.Editable {
		return fmt.Errorf("%s cannot be changed from the panel", name)
	}
	if value == "" {
		return nil
	}
	if len(value) > envValueMaxLen {
		return fmt.Errorf("%s is too long", name)
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("%s contains a control character", name)
		}
	}
	// The host writes values single-quoted, and a dotenv single-quoted
	// string has no escape for the quote itself.
	if strings.ContainsRune(value, '\'') {
		return fmt.Errorf("%s cannot contain a single quote", name)
	}

	switch v.Kind {
	case "url":
		u, err := url.Parse(value)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("%s must be an http(s) URL", name)
		}
	case "bool":
		if value != "true" && value != "false" {
			return fmt.Errorf("%s must be true or false", name)
		}
	case "duration":
		if d, err := time.ParseDuration(value); err != nil || d <= 0 {
			return fmt.Errorf("%s must be a duration such as 30m", name)
		}
	case "number":
		if n, err := strconv.Atoi(value); err != nil || n <= 0 {
			return fmt.Errorf("%s must be a positive whole number", name)
		}
	case "enum":
		if !slices.Contains(v.Options, value) {
			return fmt.Errorf("%s must be one of: %s", name, strings.Join(v.Options, ", "))
		}
	}
	return nil
}

// ServicesFor is the set of compose services a change to names recreates.
// The backend is always among them: the panel reports state from this
// process's own environment, which only a recreate refreshes.
func (m *EnvManager) ServicesFor(names []string) []string {
	services := []string{"backend"}
	for _, name := range names {
		for _, s := range m.byName[name].Services {
			if !slices.Contains(services, s) {
				services = append(services, s)
			}
		}
	}
	slices.Sort(services)
	return services
}

// Submit validates a batch of edits and hands it to the host. set holds
// explicit values ("" clears); generate names variables to mint a random
// value for. The generated values are returned and exist nowhere else on
// this side — the caller shows them once.
//
// Nothing is written unless the whole batch is valid.
//
// A request the host has not yet claimed is merged into rather than
// refused, so edits are not stuck behind an absent watcher. The host claims
// a request by renaming it, so the worst a race can do is hand it values it
// has already applied.
func (m *EnvManager) Submit(set map[string]string, generate []string) (map[string]string, []string, error) {
	changes := make(map[string]string, len(set)+len(generate))
	for name, value := range set {
		value = strings.TrimSpace(value)
		if err := m.Validate(name, value); err != nil {
			return nil, nil, err
		}
		changes[name] = value
	}
	generated := make(map[string]string, len(generate))
	for _, name := range generate {
		if v, ok := m.byName[name]; !ok || !v.Editable || v.Kind != "generated" {
			return nil, nil, fmt.Errorf("%s cannot be generated", name)
		}
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return nil, nil, err
		}
		generated[name] = hex.EncodeToString(raw)
		changes[name] = generated[name]
	}
	if len(changes) == 0 {
		return nil, nil, errors.New("no changes given")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if status := m.readStatus(); status != nil && status.State == "applying" &&
		status.At != nil && time.Since(*status.At) < envApplyingStale {
		return nil, nil, ErrEnvApplying
	}

	request := envApplyRequest{Set: changes}
	if pending := m.readRequest(); pending != nil {
		for name, value := range pending.Set {
			if _, replaced := changes[name]; !replaced {
				request.Set[name] = value
			}
		}
	}
	id := make([]byte, 8)
	if _, err := rand.Read(id); err != nil {
		return nil, nil, err
	}
	request.ID = hex.EncodeToString(id)
	request.RequestedAt = time.Now().UTC()

	if err := m.writeRequest(&request); err != nil {
		return nil, nil, err
	}

	names := make([]string, 0, len(request.Set))
	for name := range request.Set {
		names = append(names, name)
	}
	return generated, m.ServicesFor(names), nil
}

// writeRequest writes request.json so the host only ever sees a complete
// file it can read: temp name, owner and mode set, then rename.
//
// This process runs as root in the container while the host script runs as
// the checkout's owner, so the file is handed to whoever owns the shared
// directory. 0600 because it holds the new secrets in the clear until the
// host consumes it.
func (m *EnvManager) writeRequest(request *envApplyRequest) error {
	info, err := os.Stat(m.dir)
	if err != nil {
		return fmt.Errorf("the env-apply directory is not mounted: %w", err)
	}
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	tmp := filepath.Join(m.dir, envRequestFile+".tmp")
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("writing env request: %w", err)
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if err := os.Chown(tmp, int(stat.Uid), int(stat.Gid)); err != nil {
			os.Remove(tmp)
			return fmt.Errorf("handing env request to the host user: %w", err)
		}
	}
	if err := os.Rename(tmp, filepath.Join(m.dir, envRequestFile)); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("writing env request: %w", err)
	}
	return nil
}

// Cancel withdraws a request the host has not claimed. Once claimed it is
// too late, and this is a no-op.
func (m *EnvManager) Cancel() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := os.Remove(filepath.Join(m.dir, envRequestFile)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing env request: %w", err)
	}
	return nil
}

func (m *EnvManager) readRequest() *envApplyRequest {
	raw, err := os.ReadFile(filepath.Join(m.dir, envRequestFile))
	if err != nil {
		return nil
	}
	var request envApplyRequest
	if json.Unmarshal(raw, &request) != nil {
		return nil
	}
	return &request
}

func (m *EnvManager) readStatus() *EnvApplyStatus {
	raw, err := os.ReadFile(filepath.Join(m.dir, envStatusFile))
	if err != nil {
		return nil
	}
	var status EnvApplyStatus
	if json.Unmarshal(raw, &status) != nil {
		return nil
	}
	return &status
}

// Status reports where the latest change has got to. An unclaimed request
// takes precedence over status.json, which then describes an older change.
func (m *EnvManager) Status() EnvApplyStatus {
	if request := m.readRequest(); request != nil {
		names := make([]string, 0, len(request.Set))
		for name := range request.Set {
			names = append(names, name)
		}
		slices.Sort(names)
		at := request.RequestedAt
		return EnvApplyStatus{
			State:    "pending",
			ID:       request.ID,
			At:       &at,
			Pending:  names,
			Services: m.ServicesFor(names),
			Watcher:  time.Since(at) < envWatcherGrace,
		}
	}
	if status := m.readStatus(); status != nil {
		status.Watcher = true
		return *status
	}
	return EnvApplyStatus{State: "idle", Watcher: true}
}
