package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Where uploaded files are written. The upload handler
// (handlers/handle_message_upload.go) validates a file and picks its name; a
// FileStore only stores the bytes under that name.
//
// Public uploads go to the Garage bucket when S3_ENDPOINT is set, and to the
// local `uploads` volume otherwise. Either way they are served at
// /uploads/<name>: nginx looks on the volume first and falls back to the
// bucket (see the /uploads/ location in nginx.conf.template). That ordering
// is what makes every failure here survivable: if Garage is down or
// misconfigured, FallbackFileStore writes to disk instead, and nginx finds
// the file there without anyone having to know which store took it.

// FileStore stores one file under key. r is a ReadSeeker because the S3
// client must hash the body before sending it over plain HTTP, and because
// FallbackFileStore rewinds it to retry; a multipart.File satisfies it.
type FileStore interface {
	Put(ctx context.Context, key, contentType string, r io.ReadSeeker, size int64) error
}

// DiskFileStore writes files into Dir, which must exist.
type DiskFileStore struct{ Dir string }

func (d DiskFileStore) Put(_ context.Context, key, _ string, r io.ReadSeeker, _ int64) error {
	// 0644 so nginx, a different user, can read what the backend wrote.
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

	// ContentType matters: Garage serves back exactly what is stored here,
	// and nginx no longer infers a type from the extension for files it
	// fetches from the bucket.
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

// FallbackFileStore tries Primary and, if that fails, logs it and writes to
// Fallback instead. The upload still succeeds and its URL still works, so a
// Garage outage costs a log line rather than broken uploads.
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

// NewUploadStore picks the store for public uploads. With no S3 endpoint
// configured it is the disk alone, exactly the behaviour before Garage, so
// deploying this code changes nothing until .env opts in. An endpoint with
// missing credentials is a configuration mistake: it is logged and the disk
// is used rather than failing every upload.
func NewUploadStore(cfg S3Config, diskDir string) FileStore {
	disk := DiskFileStore{Dir: diskDir}
	if cfg.Endpoint == "" {
		log.Printf("[upload] S3_ENDPOINT unset; storing uploads on disk in %s", diskDir)
		return disk
	}
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Bucket == "" {
		log.Printf("[upload] S3_ENDPOINT is set but S3_BUCKET or the S3 key is missing; storing uploads on disk in %s", diskDir)
		return disk
	}
	log.Printf("[upload] storing uploads in bucket %s at %s (disk fallback: %s)", cfg.Bucket, cfg.Endpoint, diskDir)
	return FallbackFileStore{Primary: NewS3FileStore(cfg), Fallback: disk}
}
