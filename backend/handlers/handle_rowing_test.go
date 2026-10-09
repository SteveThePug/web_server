package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// readRowingDisplay is the Go half of the contract with the Python service's
// /internal/rowing/read. CreateRowing itself needs a database and a real
// multipart photo with EXIF, so what is covered is the call: the request the
// Python side receives, and how each kind of reply comes back.

func rowingReader(t *testing.T, handler http.HandlerFunc) *Store {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Store{PythonURL: srv.URL}
}

func TestReadRowingDisplayRequestAndReply(t *testing.T) {
	var got struct {
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	}
	var path string
	var forwarded bool
	store := rowingReader(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		// The Python route treats either header as "came through nginx" and
		// answers 404, so this call must never carry them.
		forwarded = r.Header.Get("X-Real-IP") != "" || r.Header.Get("X-Forwarded-For") != ""
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		w.Write([]byte(`{"timeMinutes": 20, "timeSeconds": 5, "distance": 5000}`))
	})

	data, err := store.readRowingDisplay(context.Background(), "image/jpeg", []byte("hello"))
	if err != nil {
		t.Fatalf("readRowingDisplay: %v", err)
	}
	if data != (ExtractedRowingData{TimeMinutes: 20, TimeSeconds: 5, Distance: 5000}) {
		t.Fatalf("got %+v", data)
	}
	if path != rowingReadPath {
		t.Fatalf("path = %q, want %q", path, rowingReadPath)
	}
	if forwarded {
		t.Fatal("request carried a proxy header the Python route rejects")
	}
	// "hello" in standard base64, the encoding the Anthropic API expects.
	if got.MediaType != "image/jpeg" || got.Data != "aGVsbG8=" {
		t.Fatalf("request body = %+v", got)
	}
}

func TestReadRowingDisplayErrors(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		body       string
		wantDetail string // non-empty: a displayReadError with this message
	}{
		{"reader failure is relayed", http.StatusBadGateway, `{"detail": "failed to parse image data"}`, "failed to parse image data"},
		{"validation error is not relayed", http.StatusUnprocessableEntity, `{"detail": [{"msg": "bad"}]}`, ""},
		{"missing key is not relayed", http.StatusServiceUnavailable, `{"detail": "CLAUDE_API_KEY is not set"}`, ""},
		{"non-JSON 502 is not relayed", http.StatusBadGateway, `<html>bad gateway</html>`, ""},
		{"malformed success body", http.StatusOK, `not json`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := rowingReader(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				w.Write([]byte(c.body))
			})

			_, err := store.readRowingDisplay(context.Background(), "image/jpeg", []byte("x"))
			if err == nil {
				t.Fatal("expected an error")
			}
			var readErr *displayReadError
			if relayed := errors.As(err, &readErr); relayed != (c.wantDetail != "") {
				t.Fatalf("relayed = %v, err = %v", relayed, err)
			}
			if readErr != nil && readErr.detail != c.wantDetail {
				t.Fatalf("detail = %q, want %q", readErr.detail, c.wantDetail)
			}
		})
	}
}
