package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rwcarlsen/goexif/exif"

	"adam-french.co.uk/backend/models"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gin-gonic/gin"
)

// Rowing log. The interesting endpoint is CreateRowing: rather than a form,
// the admin uploads a photo of the rowing machine display and Claude reads the
// numbers off it. The session date comes from the photo's EXIF metadata, which
// also doubles as the deduplication key.

// ExtractedRowingData is the JSON contract for what Claude reads off the
// display; the keys must match the prompt in CreateRowing.
type ExtractedRowingData struct {
	TimeMinutes uint64 `json:"timeMinutes"`
	TimeSeconds uint64 `json:"timeSeconds"`
	Distance    uint64 `json:"distance"`
}

// GetRowing backs the public GET /rowing with every logged session, newest
// first.
func (store *Store) GetRowing(ctx *gin.Context) {
	var rowing []models.Rowing
	// "Created_At" works because Postgres folds unquoted identifiers to lower
	// case; the column is created_at.
	if err := store.DB.Order("Created_At DESC").Find(&rowing).Error; err != nil {
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, rowing)
}

// CreateRowing backs the admin-only POST /rowing: it takes a photo of the
// rowing machine display, dates it from EXIF, has Claude read the time and
// distance, sanity-checks the result and stores a session.
//
// Note it is admin-gated and calls a paid API once per request, so the
// validation below exists to catch misreadings rather than hostile input.
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

	// EXIF is required, not optional: it is both the session date and the
	// duplicate key, so a stripped-metadata image (anything that has been
	// through most messaging apps) is rejected outright.
	exifData, err := exif.Decode(f)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no EXIF data found"})
		return
	}

	dateTaken, err := exifData.DateTime()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no date found in EXIF data"})
		return
	}

	// exif.Decode consumed part of the stream, so rewind before reading the
	// bytes to send to Claude.
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
	// This is the browser's claimed type, not a sniffed one; it is passed
	// straight through to the Anthropic API, so the allow-list is there to
	// keep the API call well-formed rather than to police the file.
	mediaType := file.Header.Get("Content-Type")
	if !allowedMediaTypes[mediaType] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type"})
		return
	}
	encoded := base64.StdEncoding.EncodeToString(data)

	// Duplicate guard: two photos of the same session share an EXIF capture
	// time to the second. Checked before the Claude call so a re-upload costs
	// nothing. Note this reads err == nil as "found" — an actual database
	// error is indistinguishable from "no match" here and lets the insert
	// proceed.
	var existing models.Rowing
	if err := store.DB.Where("date = ?", dateTaken).First(&existing).Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "duplicate entry for this date"})
		return
	}

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
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process image"})
		return
	}

	if len(message.Content) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "empty response from image processor"})
		return
	}

	// Strip a markdown code fence if the model wrapped its JSON in one
	// despite the prompt. Same defensive dance as in services/email_sync.go.
	extractedData := ExtractedRowingData{}
	raw := message.Content[0].Text

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	err = json.Unmarshal([]byte(raw), &extractedData)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse image data"})
		return
	}

	if extractedData.Distance == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid distance in image"})
		return
	}

	totalSeconds := extractedData.TimeMinutes*60 + extractedData.TimeSeconds

	// Sanity bounds on what Claude claims to have read. A misread digit
	// typically lands far outside human rowing performance, and the pace
	// check catches the case where time and distance are individually
	// plausible but inconsistent with each other.
	const (
		minDistance    = 100    // metres
		maxDistance    = 100000 // metres
		minTotalSecs   = 30     // 30 seconds
		maxTotalSecs   = 7200   // 2 hours
		minPacePer500m = 80     // ~1:20 /500m (faster than any human)
		maxPacePer500m = 150    // ~2:30 /500m (slow, not important)
	)
	if extractedData.Distance < minDistance || extractedData.Distance > maxDistance {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "anomalous distance value"})
		return
	}
	if totalSeconds < minTotalSecs || totalSeconds > maxTotalSecs {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "anomalous time value"})
		return
	}

	per500m := float64(totalSeconds) / float64(extractedData.Distance) * 500.0
	if per500m < minPacePer500m || per500m > maxPacePer500m {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "anomalous pace value"})
		return
	}

	// Rough calorie estimate from distance alone: roughly 500 kcal per
	// 7500 m. Not from the machine and not physiologically derived — it is a
	// display figure only.
	calories := float64(extractedData.Distance) / 7500.0 * 500.0

	rowing := models.Rowing{
		Date:        dateTaken,
		Time:        totalSeconds,
		TimePer500m: per500m,
		Distance:    extractedData.Distance,
		Calories:    calories,
	}

	if err := store.DB.Create(&rowing).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save rowing"})
		return
	}

	ctx.JSON(http.StatusCreated, rowing)
}
