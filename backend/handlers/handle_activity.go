package handlers

import (
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
)

type CreateActivityInput struct {
	Type string  `json:"type" binding:"required"`
	Name string  `json:"name" binding:"required"`
	Link *string `json:"link"`
}

func (store *Store) GetActivity(ctx *gin.Context) {
	var activitys []models.Activity
	if err := store.DB.Order("Created_At DESC").Find(&activitys).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, activitys)
}

func (store *Store) CreateActivity(ctx *gin.Context) {
	var input CreateActivityInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	activity := models.Activity{Type: input.Type, Name: input.Name, Link: input.Link}
	tx := store.DB.Create(&activity)
	if tx.Error != nil {
		ctx.JSON(http.StatusInternalServerError, tx.Error.Error())
		return
	}

	ctx.JSON(http.StatusCreated, activity)
}
