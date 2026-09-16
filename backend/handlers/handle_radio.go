package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Admin management of the radio's fallback music. These handlers only move
// files around on disk; the Python service is what actually streams them, and
// it picks up the directory contents on its own.
//
// Disabling a song moves it into a "disabled" subdirectory rather than
// deleting it — so presence in that directory IS the disabled flag, and there
// is no database state for the radio at all.

// fallbackMusicDir is a shared volume, also mounted by the radio service.
const fallbackMusicDir = "/backend/fallback_music"

// allowedAudioExtensions is the upload allow-list, keyed by lower-cased
// extension. Unlike chat uploads, the content is not sniffed: the route is
// admin-only and audio formats sniff unreliably.
var allowedAudioExtensions = map[string]bool{
	".mp3": true, ".ogg": true, ".flac": true, ".wav": true, ".m4a": true, ".opus": true,
}

// UploadRadioSong backs the admin-only POST /radio/upload. Unlike chat
// uploads the original filename is kept, because it is what the admin sees in
// the song list — hence the sanitisation below.
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

	// filepath.Base strips any directory component, so "../../etc/passwd"
	// becomes "passwd". The absolute-path check below is a second, redundant
	// guard in case Base ever fails to neutralise some input.
	filename := filepath.Base(file.Filename)
	dest := filepath.Join(fallbackMusicDir, filename)

	// The separator is appended before comparing so a sibling directory whose
	// name merely starts with the same characters cannot pass.
	absDest, err := filepath.Abs(dest)
	if err != nil || !strings.HasPrefix(absDest, fallbackMusicDir+string(os.PathSeparator)) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	// Refuse to overwrite. There is a check-then-write race here, but the
	// route is admin-only and single-user in practice.
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

// ListRadioSongs backs the admin-only GET /radio/songs, returning both the
// enabled songs and those parked in the disabled subdirectory, flagged.
func (store *Store) ListRadioSongs(ctx *gin.Context) {
	// Declared inside the function because it is purely the JSON shape of
	// this one response.
	type songInfo struct {
		Name     string `json:"name"`
		Size     int64  `json:"size"`
		Modified int64  `json:"modified"`
		Disabled bool   `json:"disabled"`
	}

	// Initialised empty rather than nil so the response is [] and not null.
	songs := []songInfo{}

	// Read enabled songs
	entries, err := os.ReadDir(fallbackMusicDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read music directory"})
		return
	}
	// Skipping directories is also what keeps the "disabled" subdirectory out
	// of the enabled list; dotfiles skip editor and OS cruft.
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

	// The disabled directory is created lazily by DisableRadioSong, so a read
	// failure here is normally just "nothing has ever been disabled" and is
	// ignored rather than reported.
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

// DisableRadioSong moves a song into the disabled subdirectory so the radio
// stops playing it, keeping the file.
func (store *Store) DisableRadioSong(ctx *gin.Context) {
	// filepath.Base neutralises traversal in the :filename route parameter.
	// It returns "." for an empty input and "/" for "/", neither of which is
	// a real file, so both are rejected explicitly.
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

// EnableRadioSong moves a song back out of the disabled subdirectory.
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

// DeleteRadioSong permanently removes an enabled song. It cannot reach the
// disabled subdirectory, so a disabled song must be enabled before it can be
// deleted.
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
