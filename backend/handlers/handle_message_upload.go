package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Attachment upload for the chat. Files land in a directory that nginx serves
// directly at /uploads/, so anything written here is publicly reachable by
// URL — which is why the extension and content checks below matter.

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

// UploadMessageFile backs POST /messages/upload (login required) and returns
// the public URL of the stored file, which the client then sends over the
// WebSocket as a message's fileUrl.
func (store *Store) UploadMessageFile(ctx *gin.Context) {
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

	uploadDir := "/backend/uploads/"
	dest := filepath.Join(uploadDir, filename)

	// Reopened rather than rewound: the earlier handle was closed after
	// sniffing, and multipart files may be backed by either memory or a temp
	// file, so a fresh reader is the simplest correct option.
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	defer src.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": "/uploads/" + filename})
}
