package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func (store *Store) GetNoteFile(ctx *gin.Context) {
	path := ctx.Param("path")

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

	ctx.Header("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition, Last-Modified")

	ctx.FileAttachment(path, info.Name())
}
