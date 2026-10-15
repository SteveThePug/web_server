package services

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Covered here: which store NewUploadStore picks, the disk store, and the
// fallback path — including against a real S3 client talking to a server
// that refuses the write, since "Garage said no" is the case the fallback
// exists for.
//
// Not covered: a successful PutObject against Garage. That needs the garage
// container; it is checked by hand in the stack (see DEPLOYMENT.md, "Garage").

func TestNewUploadStore(t *testing.T) {
	dir := t.TempDir()

	if _, ok := NewUploadStore(S3Config{}, dir).(DiskFileStore); !ok {
		t.Error("no S3 endpoint should mean the disk store")
	}
	if _, ok := NewUploadStore(S3Config{Endpoint: "http://garage:3900", Bucket: "uploads"}, dir).(DiskFileStore); !ok {
		t.Error("an endpoint without credentials should fall back to the disk store")
	}
	full := S3Config{Endpoint: "http://garage:3900", Region: "garage", Bucket: "uploads", AccessKeyID: "GK0", SecretAccessKey: "s"}
	if _, ok := NewUploadStore(full, dir).(FallbackFileStore); !ok {
		t.Error("a full S3 config should mean the bucket with a disk fallback")
	}
}

func TestDiskFileStorePut(t *testing.T) {
	dir := t.TempDir()
	body := []byte("hello")
	if err := (DiskFileStore{Dir: dir}).Put(context.Background(), "a.txt", "text/plain", bytes.NewReader(body), int64(len(body))); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "a.txt"))
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("read back %q, %v; want %q", got, err, body)
	}
}

func TestFallbackWhenBucketRefuses(t *testing.T) {
	// 403 rather than 500: the SDK retries 5xx with backoff, which would only
	// make the test slow. Either way PutObject returns an error.
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		// Drain the body so a test failure can't be a half-read request.
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	dir := t.TempDir()
	store := NewUploadStore(S3Config{Endpoint: srv.URL, Region: "garage", Bucket: "uploads", AccessKeyID: "GK0", SecretAccessKey: "s"}, dir)

	body := bytes.Repeat([]byte("x"), 4096)
	if err := store.Put(context.Background(), "f.bin", "application/octet-stream", bytes.NewReader(body), int64(len(body))); err != nil {
		t.Fatalf("Put should succeed through the fallback, got %v", err)
	}
	if hits == 0 {
		t.Fatal("the bucket was never tried")
	}
	// The whole file, not just what was left after the S3 attempt read it:
	// this is what the rewind in FallbackFileStore is for.
	got, err := os.ReadFile(filepath.Join(dir, "f.bin"))
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("fallback file has %d bytes (%v), want %d", len(got), err, len(body))
	}
}
