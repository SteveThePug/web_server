package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (store *Store) GetNoteFile(ctx *gin.Context) {
	path := ctx.Param("path")

	path, err := store.Notes.ParsePath(path)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get the filename from path
	filename := filepath.Base(path)
	ctx.FileAttachment(path, filename)
}
