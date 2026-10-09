package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rwcarlsen/goexif/exif"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Rowing log. The interesting endpoint is CreateRowing: rather than a form,
// the admin uploads a photo of the rowing machine display and Claude reads the
// numbers off it. The session date comes from the photo's EXIF metadata, which
// also doubles as the deduplication key.
//
// The Claude call itself lives in the Python service (python/app/rowing.py),
// reached directly over the Docker network. Everything that needs the session
// or the database — the admin gate, EXIF, the duplicate check, the sanity
// bounds, the insert — stays here.

// ExtractedRowingData is the JSON contract for what Claude reads off the
// display: the body of the Python service's reply, whose keys are the ones
// its prompt asks the model for (RowingReading in python/app/rowing.py).
type ExtractedRowingData struct {
	TimeMinutes uint64 `json:"timeMinutes"`
	TimeSeconds uint64 `json:"timeSeconds"`
	Distance    uint64 `json:"distance"`
}

// rowingReadPath is the Python route. It is internal: nginx refuses
// /py/internal/ and the route itself refuses anything nginx forwarded, so the
// only way in is this direct container-to-container call.
const rowingReadPath = "/internal/rowing/read"

// rowingReadClient bounds the wait on the Python service. Two minutes sits
// above that side's own worst case (a 30s Claude timeout, retried twice with
// backoff), so a slow Claude surfaces as Python's error rather than as a
// timeout here.
var rowingReadClient = &http.Client{Timeout: 2 * time.Minute}

// displayReadError is a failure the Python service reported in words that are
// safe and useful to show the admin (its 502 detail), as opposed to a
// transport or decoding failure, which is only logged.
type displayReadError struct{ detail string }

func (e *displayReadError) Error() string { return e.detail }

// readRowingDisplay sends the image to the Python service and returns what
// Claude read off it. No plausibility checking happens on either side of this
// call; that is CreateRowing's job.
func (store *Store) readRowingDisplay(ctx context.Context, mediaType string, image []byte) (ExtractedRowingData, error) {
	var extracted ExtractedRowingData

	// encoding/json writes a []byte as standard base64, which is the form the
	// Anthropic API wants and so what the Python side passes straight through.
	body, err := json.Marshal(struct {
		MediaType string `json:"media_type"`
		Data      []byte `json:"data"`
	}{mediaType, image})
	if err != nil {
		return extracted, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, store.PythonURL+rowingReadPath, bytes.NewReader(body))
	if err != nil {
		return extracted, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rowingReadClient.Do(req)
	if err != nil {
		return extracted, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// FastAPI errors are {"detail": ...}. Only a 502 carries one of the
		// reader's own fixed messages; a 422's detail is a list, which fails
		// to decode into the string and falls through to the generic error.
		var failure struct {
			Detail string `json:"detail"`
		}
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		if resp.StatusCode == http.StatusBadGateway && json.Unmarshal(raw, &failure) == nil && failure.Detail != "" {
			return extracted, &displayReadError{detail: failure.Detail}
		}
		return extracted, fmt.Errorf("rowing reader returned %d: %s", resp.StatusCode, raw)
	}

	if err := json.NewDecoder(resp.Body).Decode(&extracted); err != nil {
		return extracted, fmt.Errorf("decoding rowing reader response: %w", err)
	}
	return extracted, nil
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
// distance (via the Python service), sanity-checks the result and stores a
// session.
//
// Note it is admin-gated and costs a paid API call once per request, so the
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
	// bytes to send on for reading.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to seek image"})
		return
	}

	// Cap the upload before buffering it (and then base64-encoding it, 1.33x
	// bigger, on its way to the Python service): the route is admin-only but
	// an unbounded read would still let a giant file exhaust the Pi's memory
	// and produce a huge paid API call. 10MB is well above any photo this
	// endpoint is meant to take; the Python side's own limit is sized from it.
	if file.Size > 10<<20 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image too large (max 10MB)"})
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
	// keep the API call well-formed rather than to police the file. The
	// Python service enforces the same list; checking here first gives the
	// admin a clear 400 instead of that side's validation error.
	mediaType := file.Header.Get("Content-Type")
	if !allowedMediaTypes[mediaType] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type"})
		return
	}

	// Duplicate guard: two photos of the same session share an EXIF capture
	// time to the second. Checked before the Claude call so a re-upload costs
	// nothing.
	//
	// Fails closed: only gorm.ErrRecordNotFound means "no duplicate". Any
	// other error is a real database problem, and treating it as "not found"
	// would both insert a duplicate and spend a paid Claude call to do it.
	var existing models.Rowing
	err = store.DB.Where("date = ?", dateTaken).First(&existing).Error
	switch {
	case err == nil:
		ctx.JSON(http.StatusConflict, gin.H{"error": "duplicate entry for this date"})
		return
	case !errors.Is(err, gorm.ErrRecordNotFound):
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	extractedData, err := store.readRowingDisplay(ctx.Request.Context(), mediaType, data)
	if err != nil {
		log.Println(err)
		// The reader's own messages ("failed to parse image data", ...) are
		// passed on; anything else is a plumbing failure the admin can do
		// nothing with.
		msg := "failed to process image"
		var readErr *displayReadError
		if errors.As(err, &readErr) {
			msg = readErr.detail
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
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
