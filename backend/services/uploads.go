package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Uploaded files: everything the site stores from a user and serves back at
// /uploads/. One entry point for any handler that takes a file:
//
//	url, err := store.Uploads.Save(ctx, fileHeader, private)
//
// Save validates the file against UploadTypes and MaxUploadSize, gives it a
// random name, stores it, and returns the URL to put in the page. An
// *UploadError is the uploader's fault and its message is safe to show them;
// any other error is ours (log it, answer 500).
//
// WHERE FILES GO
//   Public files go to the Garage bucket `uploads` (garage/garage.toml), with
//   the `uploads` volume as a fallback for when Garage refuses a write. nginx
//   mirrors that: /uploads/<name> is fetched from the bucket first and the
//   volume second (nginx.conf.template), so the URL is the same wherever the
//   file landed. With S3_ENDPOINT unset there is no bucket and the volume is
//   the only store.
//   Private files (admin-only chat attachments) always go to the volume, under
//   /uploads/private/, which nginx gates behind an admin check. nginx can't
//   sign S3 requests, so it has no way to gate a bucket.
//
// Everything public is reachable by anyone who has the URL, with no login.
// That is why the type checks below are strict, and why names are random.

// MaxUploadSize caps one upload. It bounds how fast one account can fill the
// disk or the bucket; nginx rate-limits the route as well.
const MaxUploadSize = 10 << 20 // 10MB

// UploadType describes one allowed file extension.
type UploadType struct {
	// ContentType is stored with the file and served back with it. Garage
	// returns exactly this, and the browser renders by it: nginx adds
	// `nosniff`, so a wrong type means the file doesn't display.
	ContentType string
	// SniffPrefix is what http.DetectContentType must report for the file's
	// first 512 bytes, so that e.g. HTML can't be uploaded as a .png. Empty
	// means "not checked": DetectContentType is too unreliable for audio and
	// gif files, and checking would reject legitimate ones.
	SniffPrefix string
}

// UploadTypes is the allow-list, keyed by lower-cased extension with its dot.
// Nothing a browser would execute or render as markup (.html, .svg, .js, ...)
// may ever be added: these files are served from our own origin. The
// frontend's file pickers (vue/src/uploads.js, ACCEPT) list the same types.
var UploadTypes = map[string]UploadType{
	".jpg":  {ContentType: "image/jpeg", SniffPrefix: "image/"},
	".jpeg": {ContentType: "image/jpeg", SniffPrefix: "image/"},
	".png":  {ContentType: "image/png", SniffPrefix: "image/"},
	".gif":  {ContentType: "image/gif"},
	".webp": {ContentType: "image/webp", SniffPrefix: "image/"},
	".mp4":  {ContentType: "video/mp4", SniffPrefix: "video/"},
	".webm": {ContentType: "video/webm", SniffPrefix: "video/"},
	".mp3":  {ContentType: "audio/mpeg"},
	".ogg":  {ContentType: "audio/ogg"},
	".pdf":  {ContentType: "application/pdf", SniffPrefix: "application/pdf"},
	".txt":  {ContentType: "text/plain; charset=utf-8", SniffPrefix: "text/"},
}

// UploadError is a rejected upload: the uploader's mistake, with a message
// meant for them. The handler answers 400 with it.
type UploadError struct{ Message string }

func (e *UploadError) Error() string { return e.Message }

var (
	ErrUploadTooLarge   = &UploadError{"file too large (the limit is 10MB)"}
	ErrUploadType       = &UploadError{"file type not allowed"}
	ErrUploadContent    = &UploadError{"file content does not match extension"}
	ErrUploadUnreadable = &UploadError{"failed to read file content"}
)

// Uploads stores files. Build it with NewUploads.
type Uploads struct {
	public  FileStore
	private FileStore
}

// URL prefixes the files are served under; nginx owns both paths.
const (
	publicURLPrefix  = "/uploads/"
	privateURLPrefix = "/uploads/private/"
)

// NewUploads picks the stores. publicDir and privateDir are directories on the
// shared `uploads` volume (/backend/uploads/ and its private/ subdirectory in
// the container).
//
// With no S3 endpoint the public store is the volume alone, the behaviour
// before Garage. An endpoint with missing credentials is a configuration
// mistake: it is logged and the volume is used rather than failing every
// upload.
func NewUploads(cfg S3Config, publicDir, privateDir string) *Uploads {
	disk := DiskFileStore{Dir: publicDir}
	u := &Uploads{public: disk, private: DiskFileStore{Dir: privateDir}}

	switch {
	case cfg.Endpoint == "":
		log.Printf("[upload] S3_ENDPOINT unset; storing uploads on disk in %s", publicDir)
	case cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Bucket == "":
		log.Printf("[upload] S3_ENDPOINT is set but S3_BUCKET or the S3 key is missing; storing uploads on disk in %s", publicDir)
	default:
		log.Printf("[upload] storing uploads in bucket %s at %s (disk fallback: %s)", cfg.Bucket, cfg.Endpoint, publicDir)
		u.public = FallbackFileStore{Primary: NewS3FileStore(cfg), Fallback: disk}
	}
	return u
}

// Save validates and stores one uploaded file and returns its URL
// (/uploads/<name>, or /uploads/private/<name>). Authorisation is the
// caller's job: only an admin may save a private file.
func (u *Uploads) Save(ctx context.Context, fh *multipart.FileHeader, private bool) (string, error) {
	if fh.Size > MaxUploadSize {
		return "", ErrUploadTooLarge
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	kind, ok := UploadTypes[ext]
	if !ok {
		return "", ErrUploadType
	}

	f, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("opening upload: %w", err)
	}
	defer f.Close()

	if kind.SniffPrefix != "" {
		// 512 bytes is all http.DetectContentType looks at. A short read is
		// fine; sniffing works on a prefix, so only an empty file is an error.
		buf := make([]byte, 512)
		n, err := io.ReadFull(f, buf)
		if n == 0 {
			return "", ErrUploadUnreadable
		}
		if err != nil && err != io.ErrUnexpectedEOF {
			return "", fmt.Errorf("reading upload: %w", err)
		}
		if !strings.HasPrefix(http.DetectContentType(buf[:n]), kind.SniffPrefix) {
			return "", ErrUploadContent
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("rewinding upload: %w", err)
		}
	}

	// The uploaded name is thrown away for 16 random bytes. That rules out
	// path traversal and collisions in one go, and makes the URL unguessable.
	// Only the extension survives, and it is on the allow-list.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating filename: %w", err)
	}
	name := hex.EncodeToString(b) + ext

	store, prefix := u.public, publicURLPrefix
	if private {
		store, prefix = u.private, privateURLPrefix
	}
	if err := store.Put(ctx, name, kind.ContentType, f, fh.Size); err != nil {
		return "", err
	}
	return prefix + name, nil
}

// StoredFile is one public upload, as listed for the admin's upload picker.
type StoredFile struct {
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// List returns every public upload, newest first. Private files are never
// listed: they live in their own store, which this doesn't touch.
//
// The caller must be an admin. An upload's URL is public but unguessable, and
// that is the only thing keeping a file nobody linked to from being found; a
// listing anyone could read would hand every one of them out.
//
// Only names with an extension on the allow-list are returned, so anything
// else that ends up in the bucket or the directory (by hand, through
// garage-webui) never reaches a page through the picker.
func (u *Uploads) List(ctx context.Context) ([]StoredFile, error) {
	files, err := u.public.List(ctx)
	if err != nil {
		return nil, err
	}
	out := files[:0]
	for _, f := range files {
		if _, ok := UploadTypes[strings.ToLower(filepath.Ext(f.Name))]; ok {
			f.URL = publicURLPrefix + f.Name
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Modified.After(out[j].Modified) })
	return out, nil
}

// ---- Stores ----------------------------------------------------------------

// FileStore stores one file under key, and lists what it holds. r is a
// ReadSeeker because the S3 client must hash the body before sending it over
// plain HTTP, and because FallbackFileStore rewinds it to retry; a
// multipart.File satisfies it.
//
// List returns top-level files only, with URL unset (Uploads.List fills it).
type FileStore interface {
	Put(ctx context.Context, key, contentType string, r io.ReadSeeker, size int64) error
	List(ctx context.Context) ([]StoredFile, error)
}

// DiskFileStore writes files into Dir, creating it if needed.
type DiskFileStore struct{ Dir string }

func (d DiskFileStore) Put(_ context.Context, key, _ string, r io.ReadSeeker, _ int64) error {
	// 0755 and 0644 so nginx, a different user, can reach and read the files.
	if err := os.MkdirAll(d.Dir, 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(filepath.Join(d.Dir, key), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// List skips subdirectories, which is what keeps the private/ directory
// (inside the public one on the volume) out of the listing. A directory that
// doesn't exist yet just holds nothing.
func (d DiskFileStore) List(_ context.Context) ([]StoredFile, error) {
	entries, err := os.ReadDir(d.Dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var files []StoredFile
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue // deleted between ReadDir and here
		}
		files = append(files, StoredFile{Name: e.Name(), Size: info.Size(), Modified: info.ModTime()})
	}
	return files, nil
}

// S3FileStore writes files into one bucket of an S3-compatible server.
type S3FileStore struct {
	client *s3.Client
	bucket string
}

// S3Config is the S3_* block of .env.
type S3Config struct {
	Endpoint        string // e.g. http://garage:3900
	Region          string // must match garage.toml's s3_region
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

// s3Timeout bounds one upload. Uploads are capped at 10MB and Garage is on
// the same Docker network, so anything near this is a hang, and the fallback
// to disk is better than holding the request open.
const s3Timeout = 30 * time.Second

func NewS3FileStore(cfg S3Config) *S3FileStore {
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.Endpoint),
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		// Garage addresses buckets by path (/bucket/key); virtual-host style
		// would need a DNS name per bucket.
		UsePathStyle: true,
		// Recent SDKs add CRC32 checksums to every PUT by default. Only send
		// them where S3 requires them, which is what third-party servers
		// handle reliably.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &S3FileStore{client: client, bucket: cfg.Bucket}
}

func (s *S3FileStore) Put(ctx context.Context, key, contentType string, r io.ReadSeeker, size int64) error {
	ctx, cancel := context.WithTimeout(ctx, s3Timeout)
	defer cancel()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("s3 put %s/%s: %w", s.bucket, key, err)
	}
	return nil
}

// List pages through the whole bucket. Keys containing "/" are skipped: they
// are not uploads (Save never makes one) but something put under a prefix by
// hand, and nginx's /uploads/<name> only ever serves a top-level name anyway.
func (s *S3FileStore) List(ctx context.Context) ([]StoredFile, error) {
	ctx, cancel := context.WithTimeout(ctx, s3Timeout)
	defer cancel()

	var files []StoredFile
	pages := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket)})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("s3 list %s: %w", s.bucket, err)
		}
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if key == "" || strings.Contains(key, "/") {
				continue
			}
			files = append(files, StoredFile{Name: key, Size: aws.ToInt64(obj.Size), Modified: aws.ToTime(obj.LastModified)})
		}
	}
	return files, nil
}

// FallbackFileStore tries Primary and, if that fails, logs it and writes to
// Fallback instead. The upload still succeeds and its URL still works (nginx
// looks in both places), so a Garage outage costs a log line rather than
// broken uploads.
type FallbackFileStore struct {
	Primary  FileStore
	Fallback FileStore
}

func (f FallbackFileStore) Put(ctx context.Context, key, contentType string, r io.ReadSeeker, size int64) error {
	err := f.Primary.Put(ctx, key, contentType, r, size)
	if err == nil {
		return nil
	}
	log.Printf("[upload] primary store failed, writing %s to fallback: %v", key, err)
	if _, seekErr := r.Seek(0, io.SeekStart); seekErr != nil {
		return fmt.Errorf("rewinding for fallback after %v: %w", err, seekErr)
	}
	return f.Fallback.Put(ctx, key, contentType, r, size)
}

// List is the union of both stores, since nginx serves from both: the
// fallback holds whatever was written while the primary was down (and, on
// the live volume, the pre-Garage copies). A name in both is listed once,
// with the primary's details. A primary that can't list is logged and the
// fallback's files are returned alone, as for Put.
func (f FallbackFileStore) List(ctx context.Context) ([]StoredFile, error) {
	primary, err := f.Primary.List(ctx)
	if err != nil {
		log.Printf("[upload] primary store failed to list, listing fallback only: %v", err)
		primary = nil
	}
	fallback, fbErr := f.Fallback.List(ctx)
	if fbErr != nil {
		if err != nil {
			return nil, fmt.Errorf("listing both stores: %v; %w", err, fbErr)
		}
		log.Printf("[upload] fallback store failed to list: %v", fbErr)
	}
	seen := make(map[string]bool, len(primary))
	for _, file := range primary {
		seen[file.Name] = true
	}
	for _, file := range fallback {
		if !seen[file.Name] {
			primary = append(primary, file)
		}
	}
	return primary, nil
}
