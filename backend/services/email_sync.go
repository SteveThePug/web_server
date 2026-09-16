package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/anthropics/anthropic-sdk-go"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// This file is the job-application email pipeline. On a timer it fetches
// recent mail (via Microsoft Graph or plain IMAP), keeps the ones that look
// recruitment-related, asks Claude to turn each into structured data, and
// creates or advances a models.JobApplication row. Every email it has looked
// at is recorded as a models.ProcessedEmail so it is never handled twice.

// MSGRAPH_TOKEN_JSON_PATH lives in a Docker volume so the OAuth refresh token
// survives restarts, exactly like the Spotify token.
const MSGRAPH_TOKEN_JSON_PATH = "/backend/token/msgraph_token.json"

// EmailSyncConfig configures the pipeline. Backend picks the mail source;
// anything other than "graph" or "imap" leaves the service permanently idle.
type EmailSyncConfig struct {
	Backend      string // "graph" or "imap"
	ClientID     string
	ClientSecret string
	TenantID     string
	RedirectURI  string
	IMAP         *IMAPConfig
	SyncInterval time.Duration
	Enabled      bool
}

// EmailSyncService owns the sync pipeline.
type EmailSyncService struct {
	Config      *EmailSyncConfig
	OAuthConfig *oauth2.Config
	// HTTPClient doubles as the readiness flag: nil means "not configured or
	// not yet authenticated", and both the scheduler and the manual-trigger
	// handler refuse to run while it is nil. On the Graph backend it is a
	// real token-injecting client; on the IMAP backend it is set to
	// http.DefaultClient purely so this check passes (IMAP never uses it).
	HTTPClient   *http.Client
	DB           *gorm.DB
	ClaudeClient *anthropic.Client
	// mu is used with TryLock, not Lock: it makes a sync a no-op while
	// another one is running rather than queueing behind it, so the manual
	// admin trigger cannot pile up on top of the scheduled run.
	mu sync.Mutex
}

// Microsoft Graph API response types

// Microsoft Graph API response types. The IMAP backend parses raw RFC822
// messages into these same structs (see parseRawEmail) so that everything
// downstream of fetchEmails is backend-agnostic.

// graphMessagesResponse is one page of /me/messages; NextLink is Graph's
// cursor for the next page and is empty on the last one.
type graphMessagesResponse struct {
	Value    []graphMessage `json:"value"`
	NextLink string         `json:"@odata.nextLink"`
}

// graphMessage is a single email.
type graphMessage struct {
	ID               string    `json:"id"`
	Subject          string    `json:"subject"`
	ReceivedDateTime string    `json:"receivedDateTime"`
	From             graphFrom `json:"from"`
	Body             graphBody `json:"body"`
	BodyPreview      string    `json:"bodyPreview"`
}

// graphFrom wraps the sender, which Graph nests one level deep.
type graphFrom struct {
	EmailAddress graphEmailAddress `json:"emailAddress"`
}

// graphEmailAddress is a display name plus address.
type graphEmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// graphBody is the message body; ContentType is "html" or "text".
type graphBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

// Claude response type

// EmailAnalysis is the JSON contract Claude is asked to return; the keys here
// must stay in step with emailClassificationPrompt at the bottom of this file.
// Pointer fields are the ones Claude is allowed to answer null for.
type EmailAnalysis struct {
	IsJobEmail bool    `json:"isJobEmail"`
	Action     string  `json:"action"`
	Company    string  `json:"company"`
	JobTitle   string  `json:"jobTitle"`
	Status     string  `json:"status"`
	Location   *string `json:"location"`
	URL        *string `json:"url"`
	AppliedAt  *string `json:"appliedAt"`
	Notes      *string `json:"notes"`
}

// Email filtering config

// Cheap pre-filter, applied before any Claude call so that the bulk of a
// mailbox never costs an API request. It is intentionally over-inclusive:
// false positives are cheap (Claude rejects them), false negatives are not
// (the email is never looked at again once its window has passed).

// subjectKeywords match case-insensitively anywhere in the subject line.
var subjectKeywords = []string{
	"application", "interview", "offer", "rejected", "assessment",
	"applied", "candidate", "position", "role", "hiring",
	"thank you for applying", "we regret", "move forward",
	"next steps", "coding challenge", "take-home",
}

// senderDomains are applicant-tracking systems; mail from any of them is
// considered job-related whatever the subject says.
var senderDomains = []string{
	"greenhouse.io", "lever.co", "workday.com", "myworkday.com",
	"ashbyhq.com", "smartrecruiters.com", "icims.com",
	"jobvite.com", "taleo.net", "breezy.hr", "recruitee.com",
	"applytojob.com", "jazz.co", "dover.com",
}

// Status progression order

// statusOrder ranks the application lifecycle so updates can only move an
// application forwards (see updateJobApplication). A status missing from this
// map is treated as unknown and never applied. Note that "rejected" and
// "withdrawn" sit at the top, so once an application reaches them nothing can
// move it back.
var statusOrder = map[string]int{
	"applied":      0,
	"screening":    1,
	"assessment":   2,
	"interviewing": 3,
	"offer":        4,
	"rejected":     5,
	"withdrawn":    6,
}

// Token persistence

// SaveMSGraphToken persists an OAuth token to disk with 0600 permissions.
// Like the Spotify equivalent it copies into a local struct so the on-disk
// shape does not depend on the oauth2 package's own JSON tags.
func SaveMSGraphToken(path string, tok *oauth2.Token) error {
	data := struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		Expiry       time.Time `json:"expiry"`
	}{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonBytes, 0600)
}

// LoadMSGraphToken reads a token written by SaveMSGraphToken. A missing file
// is reported as an error and means "not yet authenticated".
func LoadMSGraphToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var saved struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		Expiry       time.Time `json:"expiry"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  saved.AccessToken,
		RefreshToken: saved.RefreshToken,
		TokenType:    saved.TokenType,
		Expiry:       saved.Expiry,
	}, nil
}

// persistingTokenSource wraps an oauth2.TokenSource to save refreshed tokens to disk.
type persistingTokenSource struct {
	base      oauth2.TokenSource
	path      string
	lastToken *oauth2.Token
}

// Token returns a valid token, refreshing through the wrapped source when
// needed and writing any new token back to disk.
//
// Without this wrapper the oauth2 client would refresh happily in memory but
// the file on disk would keep the original refresh token; Microsoft rotates
// refresh tokens, so after the original expired the service would have to be
// re-authorised by hand on the next restart.
//
// There is no mutex here even though oauth2's client may call Token from
// several goroutines. In practice syncs are serialised by EmailSyncService.mu
// so concurrent calls do not arise; the worst case if they ever did is a
// duplicate write of the same token.
func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := p.base.Token()
	if err != nil {
		return nil, err
	}
	// Save only when the token actually changed (i.e. was refreshed)
	if p.lastToken == nil || tok.AccessToken != p.lastToken.AccessToken {
		p.lastToken = tok
		if saveErr := SaveMSGraphToken(p.path, tok); saveErr != nil {
			log.Printf("[EmailSync] Warning: failed to persist refreshed token: %v", saveErr)
		}
	}
	return tok, nil
}

// InitEmailSync builds the service and, for the Graph backend, restores a
// saved OAuth session. It never fails: a misconfigured or unauthenticated
// service is returned with a nil HTTPClient, which leaves the pipeline idle
// while the rest of the backend starts normally. For Graph, the authorisation
// URL is printed to the container log for a one-off manual sign-in that
// lands on GET /email/callback.
func InitEmailSync(config *EmailSyncConfig, db *gorm.DB, claudeClient *anthropic.Client) *EmailSyncService {
	svc := &EmailSyncService{
		Config:       config,
		DB:           db,
		ClaudeClient: claudeClient,
	}

	switch config.Backend {
	case "imap":
		if config.IMAP == nil || config.IMAP.Email == "" {
			log.Println("[EmailSync] IMAP backend selected but not configured")
			return svc
		}
		// IMAP connects per-sync, no persistent client needed.
		// Mark as ready by setting a non-nil HTTPClient (used as readiness flag).
		svc.HTTPClient = http.DefaultClient
		log.Printf("[EmailSync] Configured with IMAP backend (%s)", config.IMAP.Host)

	case "graph":
		oauthConfig := &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", config.TenantID),
				TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", config.TenantID),
			},
			RedirectURL: config.RedirectURI,
			Scopes:      []string{"https://graph.microsoft.com/Mail.Read", "offline_access"},
		}
		svc.OAuthConfig = oauthConfig

		token, err := LoadMSGraphToken(MSGRAPH_TOKEN_JSON_PATH)
		if err != nil || token == nil {
			authURL := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
			log.Println("[EmailSync] No token saved. Authenticate Microsoft Graph with:")
			log.Println(authURL)
			return svc
		}

		baseSource := oauthConfig.TokenSource(context.Background(), token)
		persistSource := &persistingTokenSource{base: baseSource, path: MSGRAPH_TOKEN_JSON_PATH, lastToken: token}
		svc.HTTPClient = oauth2.NewClient(context.Background(), persistSource)
		log.Println("[EmailSync] Authenticated with Microsoft Graph")

	default:
		log.Printf("[EmailSync] Unknown backend %q, defaulting to disabled", config.Backend)
	}

	return svc
}

// AuthURL returns the OAuth2 authorization URL for initial setup.
func (s *EmailSyncService) AuthURL() string {
	return s.OAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

// CompleteAuth exchanges an authorization code for a token and initializes the HTTP client.
func (s *EmailSyncService) CompleteAuth(ctx context.Context, code string) error {
	token, err := s.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	if err := SaveMSGraphToken(MSGRAPH_TOKEN_JSON_PATH, token); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	baseSource := s.OAuthConfig.TokenSource(ctx, token)
	persistSource := &persistingTokenSource{base: baseSource, path: MSGRAPH_TOKEN_JSON_PATH, lastToken: token}
	s.HTTPClient = oauth2.NewClient(ctx, persistSource)

	log.Println("[EmailSync] Authentication completed successfully")
	return nil
}

// StartScheduler runs SyncEmails immediately and then on every tick until ctx
// is cancelled. It is meant to be run in its own goroutine and returns
// straight away when the service is disabled or unauthenticated.
//
// Because readiness is only checked once here, authenticating later via
// /email/callback does NOT start the scheduler; the backend has to be
// restarted, or syncs triggered by hand through POST /email/sync.
func (s *EmailSyncService) StartScheduler(ctx context.Context) {
	if !s.Config.Enabled {
		log.Println("[EmailSync] Disabled, skipping scheduler")
		return
	}
	if s.HTTPClient == nil {
		log.Println("[EmailSync] Not authenticated, skipping scheduler")
		return
	}

	log.Printf("[EmailSync] Starting scheduler, interval: %s", s.Config.SyncInterval)

	// Run once immediately on startup
	if err := s.SyncEmails(ctx); err != nil {
		log.Printf("[EmailSync] Initial sync error: %v", err)
	}

	ticker := time.NewTicker(s.Config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[EmailSync] Scheduler stopped")
			return
		case <-ticker.C:
			if err := s.SyncEmails(ctx); err != nil {
				log.Printf("[EmailSync] Sync error: %v", err)
			}
		}
	}
}

// SyncEmails is the main pipeline: fetch, filter, dedup, classify, create/update.
func (s *EmailSyncService) SyncEmails(ctx context.Context) error {
	// TryLock, so a manual trigger during a scheduled run is rejected
	// immediately instead of blocking a request for minutes.
	if !s.mu.TryLock() {
		return fmt.Errorf("sync already in progress")
	}
	defer s.mu.Unlock()

	// Registered after the unlock defer, therefore it runs BEFORE it: the
	// panic is absorbed here and the mutex is still released afterwards.
	// A recovered panic makes SyncEmails return nil, so the caller sees a
	// successful sync — the log line is the only evidence.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[EmailSync] PANIC recovered: %v", r)
		}
	}()

	log.Println("[EmailSync] Starting sync...")

	// Window start = when the last email was *processed*, not when it was
	// received. Those differ by up to one sync interval, which is deliberate:
	// it overlaps the windows so nothing that arrived mid-sync is missed. The
	// overlap re-fetches a few emails, and the ProcessedEmail dedup below is
	// what stops them being handled twice. On a cold database (no processed
	// emails at all) it falls back to the last 24 hours, so the pipeline
	// never tries to ingest an entire mailbox.
	since := time.Now().Add(-24 * time.Hour)
	var latest models.ProcessedEmail
	if err := s.DB.Order("created_at DESC").First(&latest).Error; err == nil {
		since = latest.CreatedAt
	}

	// Fetch emails from Microsoft Graph
	emails, err := s.fetchEmails(ctx, since)
	if err != nil {
		return fmt.Errorf("fetching emails: %w", err)
	}

	log.Printf("[EmailSync] Fetched %d emails since %s", len(emails), since.Format(time.RFC3339))

	// Filter by subject/sender patterns
	filtered := s.filterEmails(emails)
	log.Printf("[EmailSync] %d emails passed filters", len(filtered))

	var created, updated, skipped, errored int

	for _, email := range filtered {
		// Dedup check
		var existing models.ProcessedEmail
		if err := s.DB.Where("graph_message_id = ?", email.ID).First(&existing).Error; err == nil {
			skipped++
			continue
		}

		// A failed email is still recorded (with action "error") so the next
		// sync skips it rather than retrying — and, importantly, so it does
		// not hold the window open. The trade-off is that a transient Claude
		// outage permanently drops those emails.
		action, jobAppID, processErr := s.processEmail(ctx, email)
		if processErr != nil {
			log.Printf("[EmailSync] Error processing email %q: %v", email.Subject, processErr)
			s.DB.Create(&models.ProcessedEmail{
				GraphMessageID: email.ID,
				Subject:        truncate(email.Subject, 255),
				Action:         "error",
			})
			errored++
			continue
		}

		// Record as processed
		s.DB.Create(&models.ProcessedEmail{
			GraphMessageID: email.ID,
			Subject:        truncate(email.Subject, 255),
			Action:         action,
			JobAppID:       jobAppID,
		})

		switch action {
		case "created":
			created++
		case "updated":
			updated++
		default:
			skipped++
		}
	}

	log.Printf("[EmailSync] Sync complete: %d fetched, %d filtered, %d created, %d updated, %d skipped, %d errors",
		len(emails), len(filtered), created, updated, skipped, errored)

	return nil
}

// fetchEmails retrieves emails since the given time using the configured
// backend, normalising both into []graphMessage.
func (s *EmailSyncService) fetchEmails(ctx context.Context, since time.Time) ([]graphMessage, error) {
	if s.Config.Backend == "imap" {
		return s.fetchEmailsIMAP(since)
	}
	return s.fetchEmailsGraph(ctx, since)
}

// fetchEmailsGraph retrieves emails from the Microsoft Graph API since the given time.
func (s *EmailSyncService) fetchEmailsGraph(ctx context.Context, since time.Time) ([]graphMessage, error) {
	var all []graphMessage

	sinceStr := since.UTC().Format("2006-01-02T15:04:05Z")
	params := url.Values{
		"$filter":  {fmt.Sprintf("receivedDateTime ge %s", sinceStr)},
		"$select":  {"id,subject,from,receivedDateTime,body,bodyPreview"},
		"$top":     {"50"},
		"$orderby": {"receivedDateTime asc"},
	}

	// Graph paginates with @odata.nextLink, which already carries every query
	// parameter, so the loop below only builds this first URL by hand.
	nextURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/messages?%s", params.Encode())

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := s.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Status is checked before the read error so that a non-200 reports
		// Graph's own error body rather than a generic read failure.
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Graph API returned %d: %s", resp.StatusCode, string(body))
		}

		if err != nil {
			return nil, err
		}

		var result graphMessagesResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		all = append(all, result.Value...)
		nextURL = result.NextLink
	}

	return all, nil
}

// filterEmails returns only emails that match subject keywords or sender domains.
func (s *EmailSyncService) filterEmails(emails []graphMessage) []graphMessage {
	var matched []graphMessage
	for _, email := range emails {
		if s.matchesFilter(email) {
			matched = append(matched, email)
		}
	}
	return matched
}

// matchesFilter reports whether an email looks job-related on subject or
// sender alone. See subjectKeywords for why this errs towards including too
// much.
func (s *EmailSyncService) matchesFilter(email graphMessage) bool {
	subjectLower := strings.ToLower(email.Subject)
	for _, kw := range subjectKeywords {
		if strings.Contains(subjectLower, kw) {
			return true
		}
	}

	senderAddr := strings.ToLower(email.From.EmailAddress.Address)
	for _, domain := range senderDomains {
		if strings.HasSuffix(senderAddr, "@"+domain) || strings.HasSuffix(senderAddr, "."+domain) {
			return true
		}
	}

	return false
}

// processEmail sends a single email to Claude and creates/updates a JobApplication.
// Returns the action taken ("created", "updated", "skipped") and the job application ID if applicable.
func (s *EmailSyncService) processEmail(ctx context.Context, email graphMessage) (string, *uint, error) {
	cleaned := cleanEmailBody(email.Body.Content, email.Body.ContentType)

	userMsg := fmt.Sprintf("Subject: %s\nFrom: %s <%s>\nDate: %s\n\n%s",
		email.Subject,
		email.From.EmailAddress.Name,
		email.From.EmailAddress.Address,
		email.ReceivedDateTime,
		cleaned,
	)

	message, err := s.ClaudeClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 512,
		System: []anthropic.TextBlockParam{
			{Text: emailClassificationPrompt},
		},
		Messages: []anthropic.MessageParam{
			{
				Role:    "user",
				Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(userMsg)},
			},
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("Claude API error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", nil, fmt.Errorf("empty response from Claude")
	}

	// Only the first content block is read, and the prompt asks for bare JSON,
	// but models still wrap output in a markdown fence often enough that it is
	// stripped defensively here. The two TrimPrefix calls are ordered so that
	// "```json" is removed first and a plain "```" otherwise.
	raw := message.Content[0].Text
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var analysis EmailAnalysis
	if err := json.Unmarshal([]byte(raw), &analysis); err != nil {
		return "", nil, fmt.Errorf("parsing Claude response: %w (raw: %s)", err, raw)
	}

	if !analysis.IsJobEmail {
		return "skipped", nil, nil
	}

	if analysis.Company == "" || analysis.JobTitle == "" {
		return "skipped", nil, nil
	}

	// Applications are identified by (company, job title) compared
	// case-insensitively — there is no external id to match on. Claude
	// wording the same role slightly differently therefore creates a second
	// application rather than updating the first. The most recent match wins
	// when there are several.
	var existing models.JobApplication
	found := s.DB.Where("LOWER(company) = LOWER(?) AND LOWER(job_title) = LOWER(?)",
		analysis.Company, analysis.JobTitle).
		Order("created_at DESC").
		First(&existing).Error == nil

	if found && analysis.Action == "update" {
		return s.updateJobApplication(&existing, &analysis)
	}

	if !found {
		return s.createJobApplication(&analysis)
	}

	// Found but action is "create" — the record already exists, skip
	return "skipped", &existing.ID, nil
}

// createJobApplication inserts a new application from Claude's analysis.
func (s *EmailSyncService) createJobApplication(analysis *EmailAnalysis) (string, *uint, error) {
	app := models.JobApplication{
		Company:  analysis.Company,
		JobTitle: analysis.JobTitle,
		Status:   analysis.Status,
		Location: analysis.Location,
		URL:      analysis.URL,
		Notes:    analysis.Notes,
	}

	// The prompt asks for YYYY-MM-DD but full RFC3339 comes back often enough
	// to be worth trying first; an unparseable date is dropped silently.
	if analysis.AppliedAt != nil {
		if t, err := time.Parse(time.RFC3339, *analysis.AppliedAt); err == nil {
			app.AppliedAt = &t
		} else if t, err := time.Parse("2006-01-02", *analysis.AppliedAt); err == nil {
			app.AppliedAt = &t
		}
	}

	if app.Status == "" {
		app.Status = "applied"
	}

	if err := s.DB.Create(&app).Error; err != nil {
		return "", nil, fmt.Errorf("creating job application: %w", err)
	}

	log.Printf("[EmailSync] Created job application: %s at %s (status: %s)", app.JobTitle, app.Company, app.Status)
	return "created", &app.ID, nil
}

// updateJobApplication advances an existing application's status, but only
// ever forwards through statusOrder. Emails commonly arrive out of order (a
// delayed "thanks for applying" after an interview invite), and this is what
// stops a stale email dragging an application backwards.
func (s *EmailSyncService) updateJobApplication(existing *models.JobApplication, analysis *EmailAnalysis) (string, *uint, error) {
	newOrder, newExists := statusOrder[analysis.Status]
	currentOrder, currentExists := statusOrder[existing.Status]

	// Only update if the new status represents progression
	if !newExists || (currentExists && newOrder <= currentOrder) {
		return "skipped", &existing.ID, nil
	}

	existing.Status = analysis.Status

	// Append to notes if Claude provided any
	if analysis.Notes != nil && *analysis.Notes != "" {
		if existing.Notes != nil && *existing.Notes != "" {
			combined := *existing.Notes + "\n" + *analysis.Notes
			existing.Notes = &combined
		} else {
			existing.Notes = analysis.Notes
		}
	}

	if err := s.DB.Save(existing).Error; err != nil {
		return "", nil, fmt.Errorf("updating job application: %w", err)
	}

	log.Printf("[EmailSync] Updated job application: %s at %s (status: %s)", existing.JobTitle, existing.Company, existing.Status)
	return "updated", &existing.ID, nil
}

// cleanEmailBody reduces an email body to plain-ish text before it is sent to
// Claude, cutting token cost and noise.
//
// The HTML handling is a regex tag-strip, not a parser: fine for extracting
// prose, but it would mangle inline <script>/<style> content and mis-handle
// tags inside attribute values. That is acceptable because the output is only
// ever read by a language model, never rendered.
func cleanEmailBody(content string, contentType string) string {
	text := content

	if strings.EqualFold(contentType, "html") {
		// Strip HTML tags
		tagRegex := regexp.MustCompile(`<[^>]*>`)
		text = tagRegex.ReplaceAllString(text, " ")

		// Decode common HTML entities
		replacer := strings.NewReplacer(
			"&nbsp;", " ",
			"&amp;", "&",
			"&lt;", "<",
			"&gt;", ">",
			"&quot;", `"`,
			"&#39;", "'",
			"&apos;", "'",
		)
		text = replacer.Replace(text)
	}

	// Collapse whitespace
	spaceRegex := regexp.MustCompile(`[ \t]+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	nlRegex := regexp.MustCompile(`\n{3,}`)
	text = nlRegex.ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)

	// Cut everything from the first signature marker onwards. The idx > 0
	// test (not >= 0) means a body that begins with a marker is left alone,
	// which avoids blanking the whole message.
	sigPatterns := []string{
		"\n-- \n",
		"\nSent from my iPhone",
		"\nSent from my iPad",
		"\nGet Outlook for",
	}
	for _, pattern := range sigPatterns {
		if idx := strings.Index(text, pattern); idx > 0 {
			text = text[:idx]
		}
	}

	// Remove forwarded email chains
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			continue
		}
		cleaned = append(cleaned, line)
	}
	text = strings.Join(cleaned, "\n")

	// Cap the prompt size. This slices by bytes, so a cut can land mid-rune
	// and leave an invalid UTF-8 tail; harmless for a model prompt.
	if len(text) > 4000 {
		text = text[:4000]
	}

	return strings.TrimSpace(text)
}

// truncate caps a string at maxLen bytes (not runes) before it is stored in
// ProcessedEmail.Subject.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// emailClassificationPrompt must stay in sync with the EmailAnalysis struct:
// every key it names is a field there, and the status values it lists are the
// keys of statusOrder.
const emailClassificationPrompt = `You are an email classifier for job applications. You will receive the subject line, sender, and body of an email. Determine if this email is related to a job application, and if so, extract structured data.

Return ONLY a JSON object with these exact keys:
- "isJobEmail": boolean - true if this email is about a job application
- "action": "create" | "update" | "none" - whether this represents a new application confirmation or an update to an existing one
- "company": string - the company name
- "jobTitle": string - the job title/position
- "status": string - one of: "applied", "screening", "interviewing", "assessment", "offer", "rejected", "withdrawn"
- "location": string or null - job location if mentioned
- "url": string or null - any application portal URL
- "appliedAt": string or null - ISO 8601 date if an application date is mentioned (YYYY-MM-DD format)
- "notes": string or null - brief summary of what this email communicates (e.g. "Interview scheduled for March 15", "Application confirmed via Greenhouse")

If isJobEmail is false, only include that field. No text, no markdown, no explanation. Just the JSON object.`
