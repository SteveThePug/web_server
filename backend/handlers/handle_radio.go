package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const fallbackMusicDir = "/backend/fallback_music"

var allowedAudioExtensions = map[string]bool{
	".mp3": true, ".ogg": true, ".flac": true, ".wav": true, ".m4a": true, ".opus": true,
}

func (store *Store) UploadRadioSong(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	const maxSize = 50 << 20 // 50MB
	if file.Size > maxSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 50MB)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedAudioExtensions[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file type not allowed (accepted: .mp3, .ogg, .flac, .wav, .m4a, .opus)"})
		return
	}

	filename := filepath.Base(file.Filename)
	dest := filepath.Join(fallbackMusicDir, filename)

	// Verify the resolved path stays within the music directory
	absDest, err := filepath.Abs(dest)
	if err != nil || !strings.HasPrefix(absDest, fallbackMusicDir+string(os.PathSeparator)) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	if _, err := os.Stat(dest); err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "file already exists"})
		return
	}

	if err := ctx.SaveUploadedFile(file, dest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"name": filename})
}

func (store *Store) ListRadioSongs(ctx *gin.Context) {
	type songInfo struct {
		Name     string `json:"name"`
		Size     int64  `json:"size"`
		Modified int64  `json:"modified"`
		Disabled bool   `json:"disabled"`
	}

	songs := []songInfo{}

	// Read enabled songs
	entries, err := os.ReadDir(fallbackMusicDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read music directory"})
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		songs = append(songs, songInfo{
			Name:     entry.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
			Disabled: false,
		})
	}

	// Read disabled songs
	disabledDir := filepath.Join(fallbackMusicDir, "disabled")
	disabledEntries, err := os.ReadDir(disabledDir)
	if err == nil {
		for _, entry := range disabledEntries {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			songs = append(songs, songInfo{
				Name:     entry.Name(),
				Size:     info.Size(),
				Modified: info.ModTime().Unix(),
				Disabled: true,
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"songs": songs})
}

func (store *Store) DisableRadioSong(ctx *gin.Context) {
	filename := filepath.Base(ctx.Param("filename"))
	if filename == "." || filename == "/" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	src := filepath.Join(fallbackMusicDir, filename)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	disabledDir := filepath.Join(fallbackMusicDir, "disabled")
	if err := os.MkdirAll(disabledDir, 0o755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create disabled directory"})
		return
	}

	dst := filepath.Join(disabledDir, filename)
	if err := os.Rename(src, dst); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable song"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"disabled": filename})
}

func (store *Store) EnableRadioSong(ctx *gin.Context) {
	filename := filepath.Base(ctx.Param("filename"))
	if filename == "." || filename == "/" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	src := filepath.Join(fallbackMusicDir, "disabled", filename)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found in disabled directory"})
		return
	}

	dst := filepath.Join(fallbackMusicDir, filename)
	if err := os.Rename(src, dst); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable song"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"enabled": filename})
}

func (store *Store) DeleteRadioSong(ctx *gin.Context) {
	filename := filepath.Base(ctx.Param("filename"))
	if filename == "." || filename == "/" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	path := filepath.Join(fallbackMusicDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if err := os.Remove(path); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": filename})
}
