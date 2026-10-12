package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// CallPython is the one path from this backend to the Python service's
// internal routes, so its contract is tested here once rather than per
// caller: what the request looks like, and which replies come back as a
// relayable *PythonError, which as a plain error.

type pythonReply struct {
	Name  string  `json:"name"`
	Count int     `json:"count"`
	Note  *string `json:"note"`
}

func pythonStub(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestCallPythonRequestAndReply(t *testing.T) {
	var (
		method, path, contentType string
		forwarded                 bool
		got                       map[string]any
	)
	url := pythonStub(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, contentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		// The Python routes treat either header as "came through nginx" and
		// answer 404, so the helper must never send them.
		forwarded = r.Header.Get("X-Real-IP") != "" || r.Header.Get("X-Forwarded-For") != ""
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		w.Write([]byte(`{"name": "x", "count": 3, "note": null}`))
	})

	reply, err := CallPython[pythonReply](context.Background(), url, "/internal/thing", struct {
		Text string `json:"text"`
		Data []byte `json:"data"`
	}{"hello", []byte("hello")})
	if err != nil {
		t.Fatalf("CallPython: %v", err)
	}
	if reply.Name != "x" || reply.Count != 3 || reply.Note != nil {
		t.Fatalf("reply = %+v", reply)
	}
	if method != http.MethodPost || path != "/internal/thing" || contentType != "application/json" {
		t.Fatalf("request was %s %s (%s)", method, path, contentType)
	}
	if forwarded {
		t.Fatal("request carried a proxy header the Python routes reject")
	}
	// A []byte travels as standard base64, which is what the rowing reader
	// relies on for the image.
	if got["text"] != "hello" || got["data"] != "aGVsbG8=" {
		t.Fatalf("request body = %v", got)
	}
}

func TestCallPythonErrors(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		body       string
		wantDetail string // non-empty: a *PythonError with this message
	}{
		{"a 502's string detail is relayed", http.StatusBadGateway, `{"detail": "failed to parse image data"}`, "failed to parse image data"},
		{"a 502 with an empty detail is not", http.StatusBadGateway, `{"detail": ""}`, ""},
		{"a 502 that is not JSON is not", http.StatusBadGateway, `<html>bad gateway</html>`, ""},
		// A 422's detail is a list of validation errors, not a message.
		{"a validation error is not relayed", http.StatusUnprocessableEntity, `{"detail": [{"msg": "bad"}]}`, ""},
		// Its detail is a string, but it names an environment variable: for
		// the log, not for the page.
		{"a missing API key is not relayed", http.StatusServiceUnavailable, `{"detail": "CLAUDE_API_KEY is not set"}`, ""},
		{"a 404 is not relayed", http.StatusNotFound, `{"detail": "Not Found"}`, ""},
		{"a malformed success body is an error", http.StatusOK, `not json`, ""},
		{"a success body of the wrong shape is an error", http.StatusOK, `{"count": "three"}`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			url := pythonStub(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				w.Write([]byte(c.body))
			})

			reply, err := CallPython[pythonReply](context.Background(), url, "/internal/thing", struct{}{})
			if err == nil {
				t.Fatal("expected an error")
			}
			if reply != (pythonReply{}) {
				t.Fatalf("reply = %+v alongside an error, want the zero value", reply)
			}
			var pyErr *PythonError
			if relayed := errors.As(err, &pyErr); relayed != (c.wantDetail != "") {
				t.Fatalf("relayed = %v, err = %v", relayed, err)
			}
			if pyErr != nil && (pyErr.Detail != c.wantDetail || pyErr.Error() != c.wantDetail) {
				t.Fatalf("detail = %q, want %q", pyErr.Detail, c.wantDetail)
			}
			// A plumbing error names the route, so the log line says which
			// of the three calls it was.
			if pyErr == nil && !strings.Contains(err.Error(), "/internal/thing") {
				t.Fatalf("err = %v, want it to name the route", err)
			}
		})
	}
}

func TestCallPythonUnreachable(t *testing.T) {
	url := pythonStub(t, func(http.ResponseWriter, *http.Request) {})
	// Port 1 on localhost: nothing listens there.
	_, err := CallPython[pythonReply](context.Background(), "http://127.0.0.1:1", "/internal/thing", struct{}{})
	if err == nil {
		t.Fatal("expected an error")
	}
	var pyErr *PythonError
	if errors.As(err, &pyErr) {
		t.Fatalf("a transport failure was reported as relayable: %v", err)
	}

	// A cancelled call must still read as context.Canceled: the email
	// pipeline uses that to avoid counting an abandoned sync as a failure.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := CallPython[pythonReply](ctx, url, "/internal/thing", struct{}{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want it to wrap context.Canceled", err)
	}
}

func TestCallPythonUnencodableBody(t *testing.T) {
	// Channels cannot be marshalled; the point is that this fails before any
	// request is made, so the URL is never dialled.
	if _, err := CallPython[pythonReply](context.Background(), "http://127.0.0.1:1", "/internal/thing", make(chan int)); err == nil {
		t.Fatal("expected an error")
	}
}
