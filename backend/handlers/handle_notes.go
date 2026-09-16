package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetNoteFile serves one file from the notes directory for the admin-only
// route GET /notes/*path. The wildcard means `path` arrives with a leading
// slash and may contain further slashes, so it has to be validated rather
// than trusted — Notes.ParsePath does the containment check.

func (store *Store) GetNoteFile(ctx *gin.Context) {
	path := ctx.Param("path")

	// A traversal attempt is reported as 404, not 403, so probing cannot
	// distinguish a blocked path from a missing one.
	path, err := store.Notes.ParsePath(path)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// The SPA reads Last-Modified to show a note's edit date. Browsers hide
	// non-safelisted headers from cross-origin fetch() unless they are listed
	// in Access-Control-Expose-Headers, hence the second header.
	ctx.Header("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, Last-Modified")

	ctx.FileAttachment(path, info.Name())
}
