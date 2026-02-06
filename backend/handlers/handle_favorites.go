package handlers

import (
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
)

type CreateFavoriteInput struct {
	Type string  `json:"type" binding:"required"`
	Name string  `json:"name" binding:"required"`
	Link *string `json:"link"`
}

func (store *Store) GetFavorites(ctx *gin.Context) {
	var favorites []models.Favorite
	if err := store.DB.Order("Created_At DESC").Find(&favorites).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, favorites)
}

func (store *Store) CreateFavorite(ctx *gin.Context) {
	var input CreateFavoriteInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	favorite := models.Favorite{Type: input.Type, Name: input.Name, Link: input.Link}
	tx := store.DB.Create(&favorite)
	if tx.Error != nil {
		ctx.JSON(http.StatusInternalServerError, tx.Error.Error())
		return
	}

	ctx.JSON(http.StatusCreated, favorite)
}
