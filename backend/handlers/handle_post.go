package handlers

import (
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
)

type CreatePostInput struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Author  string `json:"author" binding:"required"`
}

func (store *Store) GetPosts(ctx *gin.Context) {
	var posts []models.Post
	if err := store.DB.Find(&posts).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": models.Post{}})
}

func (store *Store) CreatePost(ctx *gin.Context) {
	var input CreatePostInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create post
	post := models.Post{Title: input.Title, Content: input.Content, Author: input.Author}
	store.DB.Create(&post)

	ctx.JSON(http.StatusOK, gin.H{"data": post})
}

func (store *Store) UpdatePost(ctx *gin.Context) {
	id := ctx.Param("id")
	var post models.Post
	if err := store.DB.First(&post, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var input CreatePostInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.Title = input.Title
	post.Content = input.Content
	post.Author = input.Author
	store.DB.Save(&post)

	ctx.JSON(http.StatusOK, gin.H{"data": post})
}
