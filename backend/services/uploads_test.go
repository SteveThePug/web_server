package services

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Covered here: the type table (the upload trust boundary's data), Save's
// validation, naming and routing, which store NewUploads picks, and the
// fallback — including against a real S3 client talking to a server that
// refuses the write, since "Garage said no" is the case it exists for.
//
// Not covered: a successful PutObject against Garage. That needs the garage
// container; it is checked by hand in the stack (DEPLOYMENT.md, "Garage").

// fileHeader builds the *multipart.FileHeader a handler would get from
// ctx.FormFile, by round-tripping a real multipart body.
func fileHeader(t *testing.T, name string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, _ := w.CreateFormFile("file", name)
	part.Write(content)
	w.Close()
	form, err := multipart.NewReader(&body, w.Boundary()).ReadForm(32 << 20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { form.RemoveAll() })
	return form.File["file"][0]
}

func pngBytes() []byte {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	return b.Bytes()
}

// recordingStore remembers what it was asked to store.
type recordingStore struct {
	key, contentType string
	body             []byte
}

func (r *recordingStore) Put(_ context.Context, key, contentType string, rd io.ReadSeeker, _ int64) error {
	r.key, r.contentType = key, contentType
	r.body, _ = io.ReadAll(rd)
	return nil
}

func TestUploadTypes(t *testing.T) {
	t.Run("nothing executable or markup is allowed", func(t *testing.T) {
		// Served from our own origin with no login, so anything a browser
		// would run or render as a page must never be storable.
		for _, ext := range []string{".html", ".htm", ".svg", ".js", ".mjs", ".php", ".exe", ".sh", ".xml", ".xhtml", ""} {
			if _, ok := UploadTypes[ext]; ok {
				t.Errorf("%q must not be an allowed upload type", ext)
			}
		}
	})

	for ext, kind := range UploadTypes {
		if ext == "" || ext[0] != '.' || ext != strings.ToLower(ext) {
			t.Errorf("key %q: Save looks up the lower-cased extension with its dot, so this is unreachable", ext)
		}
		if kind.ContentType == "" {
			// Stored and served as-is; under nosniff an empty type renders nothing.
			t.Errorf("%q has no ContentType", ext)
		}
	}

	t.Run("audio and gif are not sniffed", func(t *testing.T) {
		// DetectContentType misreads these; checking them would reject
		// legitimate files.
		for _, ext := range []string{".mp3", ".ogg", ".gif"} {
			if UploadTypes[ext].SniffPrefix != "" {
				t.Errorf("%q should not be content-checked", ext)
			}
		}
	})
}

func TestUploadsSave(t *testing.T) {
	ctx := context.Background()
	public, private := &recordingStore{}, &recordingStore{}
	u := &Uploads{public: public, private: private}

	t.Run("public file", func(t *testing.T) {
		img := pngBytes()
		url, err := u.Save(ctx, fileHeader(t, "Holiday Photo.PNG", img), false)
		if err != nil {
			t.Fatal(err)
		}
		// The name is replaced by 32 hex chars; the lower-cased extension stays.
		if !strings.HasPrefix(url, "/uploads/") || len(strings.TrimPrefix(url, "/uploads/")) != 32+len(".png") || !strings.HasSuffix(url, ".png") {
			t.Errorf("url = %q, want /uploads/<32 hex>.png", url)
		}
		if public.key != strings.TrimPrefix(url, "/uploads/") || public.contentType != "image/png" {
			t.Errorf("stored as %q (%q)", public.key, public.contentType)
		}
		// The whole file, not just what was left after sniffing.
		if !bytes.Equal(public.body, img) {
			t.Errorf("stored %d bytes, want %d", len(public.body), len(img))
		}
	})

	t.Run("private file goes to the private store", func(t *testing.T) {
		url, err := u.Save(ctx, fileHeader(t, "note.txt", []byte("hello")), true)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(url, "/uploads/private/") || private.key == "" {
			t.Errorf("url = %q, private store got %q", url, private.key)
		}
	})

	t.Run("rejections", func(t *testing.T) {
		cases := []struct {
			name    string
			file    string
			content []byte
			want    error
		}{
			{"disallowed type", "page.html", []byte("<html></html>"), ErrUploadType},
			{"markup dressed as an image", "evil.png", []byte("<html><script>alert(1)</script></html>"), ErrUploadContent},
			{"empty sniffed file", "empty.png", nil, ErrUploadUnreadable},
			{"too large", "big.mp3", make([]byte, MaxUploadSize+1), ErrUploadTooLarge},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				_, err := u.Save(ctx, fileHeader(t, c.file, c.content), false)
				if !errors.Is(err, c.want) {
					t.Errorf("err = %v, want %v", err, c.want)
				}
			})
		}
	})
}

func TestNewUploads(t *testing.T) {
	dir := t.TempDir()

	if _, ok := NewUploads(S3Config{}, dir, dir).public.(DiskFileStore); !ok {
		t.Error("no S3 endpoint should mean the disk store")
	}
	if _, ok := NewUploads(S3Config{Endpoint: "http://garage:3900", Bucket: "uploads"}, dir, dir).public.(DiskFileStore); !ok {
		t.Error("an endpoint without credentials should fall back to the disk store")
	}
	full := S3Config{Endpoint: "http://garage:3900", Region: "garage", Bucket: "uploads", AccessKeyID: "GK0", SecretAccessKey: "s"}
	u := NewUploads(full, dir, dir)
	if _, ok := u.public.(FallbackFileStore); !ok {
		t.Error("a full S3 config should mean the bucket with a disk fallback")
	}
	if _, ok := u.private.(DiskFileStore); !ok {
		t.Error("private files must always stay on disk")
	}
}

func TestDiskFileStorePut(t *testing.T) {
	// A directory that doesn't exist yet: the private one is created on demand.
	dir := filepath.Join(t.TempDir(), "private")
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
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	dir := t.TempDir()
	u := NewUploads(S3Config{Endpoint: srv.URL, Region: "garage", Bucket: "uploads", AccessKeyID: "GK0", SecretAccessKey: "s"}, dir, filepath.Join(dir, "private"))

	img := pngBytes()
	url, err := u.Save(context.Background(), fileHeader(t, "pic.png", img), false)
	if err != nil {
		t.Fatalf("Save should succeed through the fallback, got %v", err)
	}
	if hits == 0 {
		t.Fatal("the bucket was never tried")
	}
	// The whole file, not just what the S3 attempt left unread: this is what
	// the rewind in FallbackFileStore is for.
	got, err := os.ReadFile(filepath.Join(dir, strings.TrimPrefix(url, "/uploads/")))
	if err != nil || !bytes.Equal(got, img) {
		t.Fatalf("fallback file has %d bytes (%v), want %d", len(got), err, len(img))
	}
}
