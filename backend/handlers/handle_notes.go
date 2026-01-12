package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (store *Store) GetNote(ctx *gin.Context) {
	path := ctx.Param("path")

	note, err := store.Notes.GetNote(path)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, note)
}
