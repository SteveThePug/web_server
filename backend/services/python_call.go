package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Every Claude call the site makes lives in the Python service, behind an
// /internal/ route. This file is the one way the Go backend reaches those
// routes. The only caller now is the rowing photo reader
// (handlers/handle_rowing.go); the job posting reader, email classifier and
// lead scorer went with the job tracker to its own repo.
//
// The routes are internal in a specific sense: nginx answers 404 for
// /py/internal/, and each route itself refuses a request carrying a proxy
// header, so the only way in is this direct container-to-container call. That
// is why the request below sets nothing but Content-Type.

// pythonClient bounds the wait on the Python service. Two minutes sits above
// that side's own worst case (a 30s Claude timeout, retried twice with
// backoff), so a slow Claude surfaces as Python's error rather than as a
// timeout here. One client for every route, so they share a connection pool.
var pythonClient = &http.Client{Timeout: 2 * time.Minute}

// PythonError is a failure the Python service reported in words that are safe
// and useful to show a person: the string `detail` of a 502, which is always
// one of that route's own fixed messages ("failed to parse image data", ...).
//
// Everything else — a refused connection, a 422, a 503 for a missing API key,
// a reply that will not decode — comes back as an ordinary error instead. The
// distinction is the point: a caller relays a PythonError and only logs the
// rest, because the rest describe plumbing the user can do nothing about and
// may quote internals.
type PythonError struct{ Detail string }

func (e *PythonError) Error() string { return e.Detail }

// CallPython POSTs body as JSON to baseURL+path and decodes a 200 reply into
// T. baseURL is the service's address on the Docker network with no trailing
// slash (handlers.Store.PythonURL); path is the route, starting with a slash.
//
// Any non-200 is an error. See PythonError for which ones are relayable. A
// cancelled ctx is wrapped, not replaced, so errors.Is(err, context.Canceled)
// still holds.
func CallPython[T any](ctx context.Context, baseURL, path string, body any) (T, error) {
	var out T

	// Note for callers sending bytes: encoding/json writes a []byte as
	// standard base64, which is the form the Anthropic API wants and so what
	// the Python side passes straight through.
	payload, err := json.Marshal(body)
	if err != nil {
		return out, fmt.Errorf("encoding request for %s: %w", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := pythonClient.Do(req)
	if err != nil {
		return out, fmt.Errorf("calling %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// FastAPI errors are {"detail": ...}. Only a 502 carries one of the
		// route's own fixed messages; a 422's detail is a list, which fails
		// to decode into the string and falls through to the generic error,
		// and a 503's detail names an environment variable, which is for the
		// log rather than the page. The body is capped because it is only
		// ever quoted in a log line.
		var failure struct {
			Detail string `json:"detail"`
		}
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		if resp.StatusCode == http.StatusBadGateway && json.Unmarshal(raw, &failure) == nil && failure.Detail != "" {
			return out, &PythonError{Detail: failure.Detail}
		}
		return out, fmt.Errorf("%s returned %d: %s", path, resp.StatusCode, raw)
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		// A fresh zero value, so a half-decoded T never escapes with an error.
		var zero T
		return zero, fmt.Errorf("decoding %s response: %w", path, err)
	}
	return out, nil
}
