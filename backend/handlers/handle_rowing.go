package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"

	"adam-french.co.uk/backend/models"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gin-gonic/gin"
)

type ExtractedRowingData struct {
	TimeMinutes float64 `json:"timeMinutes"`
	TimeSeconds float64 `json:"timeSeconds"`
	Distance    float64 `json:"distance"`
}

func (store *Store) GetRowing(ctx *gin.Context) {
	var rowing []models.Rowing
	if err := store.DB.Order("Created_At DESC").Find(&rowing).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, rowing)
}

func (store *Store) CreateRowing(ctx *gin.Context) {
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open image"})
		return
	}
	defer f.Close()

	// Get the date taken from the EXIF data
	x, err := exif.Decode(f)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no EXIF data found"})
		return
	}

	dateTaken, err := x.DateTime()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no date found in EXIF data"})
		return
	}

	// Seek back to start since exif.Decode advanced the cursor
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to seek image"})
		return
	}

	data, err := io.ReadAll(f)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image"})
		return
	}

	allowedMediaTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	mediaType := file.Header.Get("Content-Type")
	if !allowedMediaTypes[mediaType] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type"})
		return
	}
	encoded := base64.StdEncoding.EncodeToString(data)

	// Build the message with an image + text prompt
	message, err := store.ClaudeClient.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.ContentBlockParamUnion{
					// Image block
					anthropic.NewImageBlock(anthropic.Base64ImageSourceParam{
						Type:      "base64",
						MediaType: anthropic.Base64ImageSourceMediaType(mediaType),
						Data:      encoded,
					}),
					// Text prompt requesting exactly 2 variables
					anthropic.NewTextBlock(
						`Look at this rowing machine display. Extract the total elapsed time and total distance.

Return ONLY a JSON object with these exact keys and numeric values:
- "timeMinutes": total minutes (e.g. 2:30 = 2)
- "timeSeconds": total seconds (e.g. 2:30 = 30)
- "distance": distance in meters as a number (e.g. 5000)

No text, no markdown, no explanation. Just the JSON object.`),
				},
			},
		},
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze image"})
		return
	}

	if len(message.Content) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "empty response from Claude"})
		return
	}

	extractedData := ExtractedRowingData{}
	raw := message.Content[0].Text

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	err = json.Unmarshal([]byte(raw), &extractedData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to parse JSON response",
			"detail": err.Error(),
			"raw":    raw,
		})
		return
	}

	if extractedData.Distance == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid distance in image"})
		return
	}

	totalSeconds := extractedData.TimeMinutes*60 + extractedData.TimeSeconds
	totalDuration := time.Duration(totalSeconds * float64(time.Second))
	per500m := time.Duration(totalSeconds / extractedData.Distance * 500 * float64(time.Second))

	rowing := models.Rowing{
		Date:        dateTaken,
		Time:        totalDuration,
		TimePer500m: per500m,
		Distance:    extractedData.Distance,
		Calories:    extractedData.Distance / 7500.0 * 500.0,
	}

	if err := store.DB.Create(&rowing).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save rowing"})
		return
	}

	ctx.JSON(http.StatusCreated, rowing)
}
