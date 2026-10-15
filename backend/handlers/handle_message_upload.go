package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Attachment upload for the chat (and, through it, the canvas editors).
//
// Files land in one of two places. Public files go to Store.Files: the Garage
// bucket when S3_ENDPOINT is configured, /backend/uploads/ otherwise (see
// services/file_store.go). Either way nginx serves them at /uploads/ with no
// authentication, so anything stored is publicly reachable by URL — which is
// why the extension and content checks below matter.
// /backend/uploads/private/ holds attachments for private (admin-only)
// messages and is served at /uploads/private/ behind an nginx auth_request
// gate; only an admin may upload there. Private files always stay on disk:
// nginx cannot sign S3 requests, so it has no way to gate a bucket.

// allowedExtensions is the upload allow-list, keyed by lower-cased extension.
var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".mp4": true, ".webm": true, ".mp3": true, ".ogg": true,
	".pdf": true, ".txt": true,
}

// extensionToMIMEPrefix cross-checks the sniffed content type against the
// claimed extension, to stop e.g. HTML being served from a .png URL.
//
// The audio extensions (.mp3, .ogg) and .gif are deliberately absent:
// http.DetectContentType's answers for them are unreliable enough to reject
// legitimate files, so a missing entry means "extension allowed, content not
// verified".
var extensionToMIMEPrefix = map[string]string{
	".jpg": "image/", ".jpeg": "image/", ".png": "image/", ".gif": "image/", ".webp": "image/",
	".mp4": "video/", ".webm": "video/",
	".pdf": "application/pdf", ".txt": "text/",
}

// extensionContentType is the Content-Type a file is stored with. It comes
// from the (already allow-listed) extension rather than from sniffing,
// because sniffing is unreliable for exactly the audio and gif files the
// table above leaves unverified. It matters for the bucket: Garage serves
// back the stored type, where nginx serving from disk inferred its own.
var extensionContentType = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".gif": "image/gif", ".webp": "image/webp",
	".mp4": "video/mp4", ".webm": "video/webm", ".mp3": "audio/mpeg", ".ogg": "audio/ogg",
	".pdf": "application/pdf", ".txt": "text/plain; charset=utf-8",
}

// PublicUploadDir and privateUploadDir are the two directories on the shared
// `uploads` volume; the private one is created on demand because nothing else
// guarantees it exists. PublicUploadDir is exported for main, which hands it
// to services.NewUploadStore as the disk store (or the bucket's fallback).
const (
	PublicUploadDir  = "/backend/uploads/"
	privateUploadDir = "/backend/uploads/private/"
)

// UploadMessageFile backs POST /messages/upload (login required) and returns
// the URL of the stored file, which the client then sends over the WebSocket
// as a message's fileUrl.
//
// The multipart form takes an optional "private" field: "true" stores the file
// under /uploads/private/ for use as a private message's attachment. That is
// admin-only, because private messages themselves are — an attachment URL
// leaking out is the same disclosure as the message leaking out.
func (store *Store) UploadMessageFile(ctx *gin.Context) {
	private := ctx.PostForm("private") == "true"
	if private && !requestIsAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	const maxSize = 10 << 20 // 10MB
	if file.Size > maxSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file type not allowed"})
		return
	}

	// Validate actual content type matches extension
	f, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	// 512 bytes is exactly what http.DetectContentType inspects; reading more
	// would be wasted. A single Read may return fewer bytes than asked for,
	// which is fine — sniffing works on a short prefix — so only a read that
	// produced nothing at all is an error.
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	f.Close()
	if err != nil && n == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file content"})
		return
	}
	detectedType := http.DetectContentType(buf[:n])

	expectedPrefix, ok := extensionToMIMEPrefix[ext]
	if ok && !strings.HasPrefix(detectedType, expectedPrefix) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file content does not match extension"})
		return
	}

	// The uploaded name is discarded entirely in favour of 16 random bytes
	// from crypto/rand. That kills path traversal and collisions in one go,
	// and makes the resulting URL unguessable. Only the extension is carried
	// over, and it has already been checked against the allow-list.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate filename"})
		return
	}
	filename := hex.EncodeToString(b) + ext

	// Reopened rather than rewound: the earlier handle was closed after
	// sniffing, and multipart files may be backed by either memory or a temp
	// file, so a fresh reader is the simplest correct option.
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	defer src.Close()

	var dest services.FileStore
	urlPrefix := "/uploads/"
	if private {
		urlPrefix = "/uploads/private/"
		// 0755 so nginx (a different user) can traverse the directory to
		// serve the files once its auth_request gate has approved.
		if err := os.MkdirAll(privateUploadDir, 0755); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}
		dest = services.DiskFileStore{Dir: privateUploadDir}
	} else {
		dest = store.Files
		// A Store built without one (the tests, mostly) keeps the old
		// behaviour rather than panicking.
		if dest == nil {
			dest = services.DiskFileStore{Dir: PublicUploadDir}
		}
	}

	if err := dest.Put(ctx.Request.Context(), filename, extensionContentType[ext], src, file.Size); err != nil {
		log.Printf("[upload] storing %s failed: %v", filename, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": urlPrefix + filename})
}

// requestIsAdmin reports whether the caller's access token carries admin=true.
//
// This route sits in the `protected` group, which runs AuthMiddlewear but not
// AdminMiddleware — the endpoint has to stay usable by any signed-in user for
// public attachments — so the admin check is made here from the same
// "userClaims" value AdminMiddleware would have read.
func requestIsAdmin(ctx *gin.Context) bool {
	claims, exists := ctx.Get("userClaims")
	if !exists {
		return false
	}
	mapClaims, ok := claims.(*jwt.MapClaims)
	if !ok {
		return false
	}
	admin, ok := (*mapClaims)["admin"].(bool)

	return ok && admin
}
