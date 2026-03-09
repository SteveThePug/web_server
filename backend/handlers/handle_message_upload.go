package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".mp4": true, ".webm": true, ".mp3": true, ".ogg": true,
	".pdf": true, ".txt": true,
}

var extensionToMIMEPrefix = map[string]string{
	".jpg": "image/", ".jpeg": "image/", ".png": "image/", ".gif": "image/", ".webp": "image/",
	".mp4": "video/", ".webm": "video/",
	".pdf": "application/pdf", ".txt": "text/",
}

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

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate filename"})
		return
	}
	filename := hex.EncodeToString(b) + ext

	uploadDir := "/backend/uploads/"
	dest := filepath.Join(uploadDir, filename)
	if err := ctx.SaveUploadedFile(file, dest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": "/uploads/" + filename})
}
