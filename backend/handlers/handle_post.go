package handlers

import (
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
)

type CreatePostInput struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	AuthorID uint   `json:"authorID" binding:"required"`
}

func (h *Handler) GetPosts(c *gin.Context) {
	var posts []models.Post
	if err := h.DB.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.Post{}})
}

func (h *Handler) CreatePost(c *gin.Context) {
	var input CreatePostInput
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create post
	post := models.Post{Title: input.Title, Content: input.Content, AuthorID: input.AuthorID}
	h.DB.Create(&post)

	c.JSON(http.StatusOK, gin.H{"data": post})
}

func (h *Handler) UpdatePost(c *gin.Context) {

}
