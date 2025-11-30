package handlers

import (
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CreatePostInput struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (store *Store) GetPosts(ctx *gin.Context) {
	var posts []models.Post
	if err := store.DB.Find(&posts).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": posts})
}

func (store *Store) GetPost(ctx *gin.Context) {
	postID := ctx.Param("id")
	var post models.Post
	if err := store.DB.First(&post, postID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": post})
}

func (store *Store) CreatePost(ctx *gin.Context) {
	var input CreatePostInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claimsVal, ok := ctx.Get("userClaims")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user claims could not be found"})
		return
	}

	claims, ok := claimsVal.(*jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims"})
		return
	}

	userID, ok := (*claims)["id"].(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}

	// Create post
	post := models.Post{Title: input.Title, Content: input.Content, AuthorID: userID}
	store.DB.Create(&post)

	ctx.JSON(http.StatusCreated, gin.H{"data": post})
}

func (store *Store) UpdatePost(ctx *gin.Context) {
	postID := ctx.Param("id")
	var post models.Post
	if err := store.DB.First(&post, postID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	claimsVal, ok := ctx.Get("userClaims")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user claims could not be found"})
		return
	}

	claims, ok := claimsVal.(*jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims"})
		return
	}

	userID, ok := (*claims)["id"].(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}

	if !(userID == post.AuthorID) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user and post author id missmatch"})
	}

	var input CreatePostInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.Title = input.Title
	post.Content = input.Content
	store.DB.Save(&post)

	ctx.JSON(http.StatusOK, gin.H{"data": post})
}

func (store *Store) DeletePost(ctx *gin.Context) {
	postID := ctx.Param("id")
	var post models.Post
	if err := store.DB.First(&post, postID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	claimsVal, ok := ctx.Get("userClaims")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user claims could not be found"})
		return
	}

	claims, ok := claimsVal.(*jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims"})
		return
	}

	userID, ok := (*claims)["id"].(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}

	if !(userID == post.AuthorID) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user and post author id missmatch"})
	}

	store.DB.Delete(&post)
	ctx.JSON(http.StatusOK, gin.H{"data": post})
}
