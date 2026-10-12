package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Job leads: the pipeline that finds postings, as opposed to tracking
// applications. On a timer it polls every enabled models.JobSource
// (job_sources.go), files what it has not seen before as a models.JobLead,
// marks which leads pass a keyword pre-filter, and asks Claude — through the
// Python service, like every Claude call — to score those against the
// models.JobSearchProfile.
//
// Leads are a separate table from the tracker on purpose (see models.JobLead).
// The one way a lead reaches the tracker is promotion, a deliberate click on
// the leads page (graph/job_lead.resolvers.go), which goes through
// UpsertJobApplication so it cannot duplicate a posting already captured.

// Lead states, stored in JobLead.State.
const (
	LeadNew       = "new"
	LeadDismissed = "dismissed"
	LeadPromoted  = "promoted"
)

const (
	// jobScorePath is internal in the same way as the other /internal/
	// routes: nginx refuses /py/internal/ and the route refuses anything
	// nginx forwarded.
	jobScorePath = "/internal/jobs/score"

	// maxScoresPerRun bounds the Claude calls one run can make, and so both
	// its cost and its length. A first poll of a few boards can match
	// hundreds of leads; the rest are picked up by later runs, newest first.
	maxScoresPerRun = 40

	// maxScoreAttempts caps how many times a lead Claude failed to score is
	// retried, for the same reason as maxEmailAttempts.
	maxScoreAttempts = 3

	// maxRunDuration bounds a whole run. Generous: forty scores at a worst
	// case of two minutes each would exceed it, but a healthy call is a few
	// seconds, and a run that is cut short loses nothing — unscored leads are
	// simply still unscored next time.
	maxRunDuration = 20 * time.Minute

	// leadBatch sizes the IN lists and inserts, well inside Postgres's limit
	// of 65535 bound parameters per statement.
	leadBatch = 500
)

// ErrLeadSyncRunning is returned by Run when another run holds the lock.
var ErrLeadSyncRunning = errors.New("lead sync already in progress")

// JobLeadConfig configures the pipeline. The API keys are optional: a search
// source whose key is missing reports that as its own error and the rest of
// the run carries on.
type JobLeadConfig struct {
	Enabled      bool
	Interval     time.Duration
	AdzunaAppID  string
	AdzunaAppKey string
	ReedAPIKey   string
}

// JobLeadService owns the lead pipeline.
type JobLeadService struct {
	Config *JobLeadConfig
	DB     *gorm.DB
	// PythonURL is the Python service's base URL on the Docker network, the
	// same value as handlers.Store.PythonURL.
	PythonURL string
	// HTTPClient fetches the sources. Its timeout is per request, and has to
	// allow for a multi-megabyte board arriving over the Pi's connection.
	HTTPClient *http.Client

	// mu is used with TryLock, as in EmailSyncService: a run requested while
	// another is going is refused, not queued.
	mu sync.Mutex

	// statusMu guards the three fields below, which the leads page polls
	// while a run it started is in flight.
	statusMu  sync.Mutex
	running   bool
	lastRunAt *time.Time
	lastError string
}

// LeadSyncStatus is what the leads page shows beside its "Check now" button.
type LeadSyncStatus struct {
	Running bool
	// LastRunAt is when the most recent run finished; nil until one has, in
	// this process — it is not persisted, so a restart clears it.
	LastRunAt *time.Time
	// LastError is "" after a run in which nothing failed outright. One
	// source failing is not here; it is on that JobSource.
	LastError string
}

// InitJobLeads builds the service. It makes no network or database calls.
func InitJobLeads(config *JobLeadConfig, db *gorm.DB, pythonURL string) *JobLeadService {
	return &JobLeadService{
		Config:     config,
		DB:         db,
		PythonURL:  pythonURL,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Status reports whether a run is in flight and how the last one ended.
func (s *JobLeadService) Status() LeadSyncStatus {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	return LeadSyncStatus{Running: s.running, LastRunAt: s.lastRunAt, LastError: s.lastError}
}

// StartScheduler runs the pipeline on every tick until ctx is cancelled. It
// is meant to be run in its own goroutine and returns straight away when
// JOB_LEADS_ENABLED is not "true".
//
// Unlike the email scheduler it does not run at start-up: every deploy
// restarts the backend, and a deploy should not spend forty Claude calls. The
// first run is one interval in, or whenever "Check now" is pressed — which
// works whether or not the scheduler is enabled.
func (s *JobLeadService) StartScheduler(ctx context.Context) {
	if !s.Config.Enabled {
		log.Println("[JobLeads] Scheduler disabled; runs only on request")
		return
	}
	log.Printf("[JobLeads] Starting scheduler, interval: %s", s.Config.Interval)

	ticker := time.NewTicker(s.Config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[JobLeads] Scheduler stopped")
			return
		case <-ticker.C:
			if err := s.Run(ctx); err != nil {
				log.Printf("[JobLeads] Run error: %v", err)
			}
		}
	}
}

// RunInBackground starts a run and returns at once, reporting false when one
// was already going. It backs the page's "Check now": a run can take minutes,
// far longer than a request should be held open, so the page polls Status
// instead of waiting.
func (s *JobLeadService) RunInBackground() bool {
	if s.Status().Running {
		return false
	}
	go func() {
		// Not the request's context: the request is over long before this is.
		if err := s.Run(context.Background()); err != nil && !errors.Is(err, ErrLeadSyncRunning) {
			log.Printf("[JobLeads] Manual run error: %v", err)
		}
	}()
	return true
}

// Run is one pass of the pipeline: fetch, store, pre-filter, score.
func (s *JobLeadService) Run(ctx context.Context) (err error) {
	if !s.mu.TryLock() {
		return ErrLeadSyncRunning
	}
	defer s.mu.Unlock()

	s.statusMu.Lock()
	s.running = true
	s.statusMu.Unlock()

	// Runs before the unlock above. It also absorbs a panic, turning it into
	// this run's error, so one malformed posting cannot take the backend down
	// or leave `running` stuck true.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[JobLeads] PANIC recovered: %v", r)
			err = fmt.Errorf("internal error during the run")
		}
		now := time.Now()
		s.statusMu.Lock()
		s.running = false
		s.lastRunAt = &now
		s.lastError = ""
		if err != nil {
			s.lastError = err.Error()
		}
		s.statusMu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(ctx, maxRunDuration)
	defer cancel()

	profile, err := LoadJobSearchProfile(s.DB)
	if err != nil {
		return fmt.Errorf("loading the search profile: %w", err)
	}
	filter := newLeadFilter(profile)

	var sources []models.JobSource
	if err := s.DB.Where("enabled = ?", true).Order("id ASC").Find(&sources).Error; err != nil {
		return fmt.Errorf("loading sources: %w", err)
	}
	tracked, err := s.trackedURLs()
	if err != nil {
		return fmt.Errorf("loading tracked applications: %w", err)
	}

	log.Printf("[JobLeads] Starting run over %d sources", len(sources))
	added := 0
	for _, src := range sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		jobs, fetchErr := s.fetchSource(ctx, src)
		var n int
		if fetchErr == nil {
			n, fetchErr = s.storeLeads(ctx, src, jobs, filter, tracked)
			added += n
		}
		s.recordPoll(src, len(jobs), fetchErr)
	}

	if err := s.refilter(filter); err != nil {
		return fmt.Errorf("re-applying the filter: %w", err)
	}
	scored, err := s.scoreLeads(ctx, profile)
	log.Printf("[JobLeads] Run finished: %d new leads, %d scored", added, scored)
	return err
}

// recordPoll writes a poll's outcome onto its source for the page to show.
// Updates rather than Save, so a source edited on the page mid-run is not
// overwritten with the copy this run loaded.
func (s *JobLeadService) recordPoll(src models.JobSource, count int, fetchErr error) {
	now := time.Now()
	updates := map[string]any{"last_fetched_at": &now, "last_error": ""}
	if fetchErr != nil {
		log.Printf("[JobLeads] %s %q failed: %v", src.Kind, src.Label, fetchErr)
		updates["last_error"] = truncate(fetchErr.Error(), 300)
	} else {
		updates["last_count"] = count
	}
	if err := s.DB.Model(&models.JobSource{}).Where("id = ?", src.ID).Updates(updates).Error; err != nil {
		log.Printf("[JobLeads] Recording the poll of source %d failed: %v", src.ID, err)
	}
}

// trackedURLs maps the canonical URL of every live application that has one
// to its id, so a fetched posting that is already in the tracker can be told
// apart from a fresh lead.
func (s *JobLeadService) trackedURLs() (map[string]uint, error) {
	var apps []models.JobApplication
	if err := s.DB.Select("id", "url").Where("url IS NOT NULL AND url <> ''").Find(&apps).Error; err != nil {
		return nil, err
	}
	tracked := make(map[string]uint, len(apps))
	for _, app := range apps {
		if canonical, ok := CanonicalJobURL(*app.URL); ok {
			tracked[canonical] = app.ID
		}
	}
	return tracked, nil
}

// storeLeads files the postings a source returned and reports how many were
// new. A posting seen before only has its LastSeenAt moved: its state, score
// and text are left exactly as they are, which is what keeps a dismissed
// lead dismissed.
func (s *JobLeadService) storeLeads(ctx context.Context, src models.JobSource, jobs []fetchedJob, filter leadFilter, tracked map[string]uint) (int, error) {
	now := time.Now()
	added := 0
	for start := 0; start < len(jobs); start += leadBatch {
		batch := jobs[start:min(start+leadBatch, len(jobs))]
		ids := make([]string, len(batch))
		for i, job := range batch {
			ids[i] = job.ExternalID
		}

		var known []string
		if err := s.DB.Model(&models.JobLead{}).
			Where("source_kind = ? AND external_id IN ?", src.Kind, ids).
			Pluck("external_id", &known).Error; err != nil {
			return added, fmt.Errorf("looking up existing leads: %w", err)
		}
		if len(known) > 0 {
			if err := s.DB.Model(&models.JobLead{}).
				Where("source_kind = ? AND external_id IN ?", src.Kind, known).
				UpdateColumn("last_seen_at", now).Error; err != nil {
				return added, fmt.Errorf("updating existing leads: %w", err)
			}
		}
		seen := make(map[string]bool, len(known))
		for _, id := range known {
			seen[id] = true
		}

		var fresh []models.JobLead
		for _, job := range batch {
			// `seen` also catches a posting listed twice in one response.
			if seen[job.ExternalID] {
				continue
			}
			seen[job.ExternalID] = true
			lead, ok := newLead(src, job, filter, tracked, now)
			if !ok {
				continue
			}
			// The full text costs a request per job, so it is fetched only
			// for a lead that will be shown and scored. A failure leaves the
			// snippet in place rather than losing the lead.
			if job.detailID != "" && lead.Matched && lead.State == LeadNew {
				if full, err := s.fetchReedDescription(ctx, job.detailID); err == nil && full != "" {
					lead.Description = truncateRunes(full, maxLeadRunes)
				}
			}
			fresh = append(fresh, lead)
		}
		if len(fresh) == 0 {
			continue
		}
		// DoNothing covers the one way a row can appear between the lookup
		// above and this insert: a second search source of the same kind
		// returning the same advert. Runs themselves never overlap.
		result := s.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&fresh)
		if result.Error != nil {
			return added, fmt.Errorf("saving new leads: %w", result.Error)
		}
		added += int(result.RowsAffected)
	}
	return added, nil
}

// newLead turns a fetched posting into the row to insert. ok is false for a
// posting with no id, title or company, which could be neither deduplicated
// nor promoted.
//
// A posting whose URL is already in the tracker is filed as promoted and
// linked to that application: it still has to be a row, or the next poll
// would consider it again, but it must not be offered in the inbox.
func newLead(src models.JobSource, job fetchedJob, filter leadFilter, tracked map[string]uint, now time.Time) (models.JobLead, bool) {
	title := strings.TrimSpace(job.Title)
	company := strings.TrimSpace(job.Company)
	if job.ExternalID == "" || title == "" || company == "" {
		return models.JobLead{}, false
	}
	lead := models.JobLead{
		SourceID:    src.ID,
		SourceKind:  src.Kind,
		ExternalID:  job.ExternalID,
		JobTitle:    title,
		Company:     company,
		Location:    strings.TrimSpace(job.Location),
		URL:         strings.TrimSpace(job.URL),
		Description: truncateRunes(job.Description, maxLeadRunes),
		PostedAt:    job.PostedAt,
		LastSeenAt:  now,
		State:       LeadNew,
	}
	lead.Matched = filter.matches(lead.JobTitle, lead.Location)
	if canonical, ok := CanonicalJobURL(lead.URL); ok {
		lead.URL = canonical
		if appID, found := tracked[canonical]; found {
			lead.State = LeadPromoted
			lead.JobAppID = &appID
		}
	}
	return lead, true
}

// refilter brings Matched up to date for every lead still in the inbox, so
// that editing the keywords on the page takes effect on the next run instead
// of only for postings found afterwards.
func (s *JobLeadService) refilter(filter leadFilter) error {
	var leads []models.JobLead
	if err := s.DB.Select("id", "job_title", "location", "matched").
		Where("state = ?", LeadNew).Find(&leads).Error; err != nil {
		return err
	}
	var toMatch, toUnmatch []uint
	for _, lead := range leads {
		switch matched := filter.matches(lead.JobTitle, lead.Location); {
		case matched && !lead.Matched:
			toMatch = append(toMatch, lead.ID)
		case !matched && lead.Matched:
			toUnmatch = append(toUnmatch, lead.ID)
		}
	}
	for value, ids := range map[bool][]uint{true: toMatch, false: toUnmatch} {
		for start := 0; start < len(ids); start += leadBatch {
			if err := s.DB.Model(&models.JobLead{}).
				Where("id IN ?", ids[start:min(start+leadBatch, len(ids))]).
				UpdateColumn("matched", value).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// PromoteJobLead adds a lead to the tracker as a saved application and marks
// it promoted. It goes through UpsertJobApplication, so a posting that was
// captured from the browser in the meantime is linked to, not duplicated —
// the lead's text and score fill any fields that row was missing.
//
// Promoting twice is harmless: the second call finds the lead already
// promoted and returns it.
func PromoteJobLead(db *gorm.DB, id uint, now time.Time) (*models.JobLead, error) {
	var lead models.JobLead
	if err := db.First(&lead, id).Error; err != nil {
		return nil, err
	}
	if lead.State == LeadPromoted && lead.JobAppID != nil {
		return &lead, nil
	}

	app, _, err := UpsertJobApplication(db, JobCaptureInput{
		JobTitle:    lead.JobTitle,
		Company:     lead.Company,
		Location:    lead.Location,
		URL:         lead.URL,
		Status:      "saved",
		Notes:       leadNotes(&lead),
		Description: lead.Description,
	}, now)
	if err != nil {
		return nil, err
	}
	lead.State = LeadPromoted
	lead.JobAppID = &app.ID
	if err := db.Model(&lead).Updates(map[string]any{"state": lead.State, "job_app_id": lead.JobAppID}).Error; err != nil {
		return nil, fmt.Errorf("marking the lead promoted: %w", err)
	}
	return &lead, nil
}

// leadNotes is the line a promoted lead leaves in the application's notes, so
// why it looked worth applying to is still there once the lead is forgotten.
func leadNotes(lead *models.JobLead) string {
	if lead.Score == nil {
		return ""
	}
	note := fmt.Sprintf("Lead score %d/100", *lead.Score)
	if lead.ScoreReasons != nil && strings.TrimSpace(*lead.ScoreReasons) != "" {
		note += ": " + strings.TrimSpace(*lead.ScoreReasons)
	}
	return note
}

// ValidateJobSource tidies a source's fields and reports the first thing
// wrong with it, in words for the page. Which fields are required depends on
// the kind; the ones a kind does not use are cleared so a source switched
// from one kind to another does not carry the old kind's settings around.
func ValidateJobSource(src *models.JobSource) error {
	src.Kind = strings.ToLower(strings.TrimSpace(src.Kind))
	src.Label = strings.TrimSpace(src.Label)
	src.Token = strings.TrimSpace(src.Token)
	src.Query = strings.TrimSpace(src.Query)
	src.Location = strings.TrimSpace(src.Location)

	switch {
	case IsBoardSource(src.Kind):
		// The token goes into a URL path. Board slugs are letters, digits
		// and a little punctuation; anything else is a pasted URL or a typo.
		if src.Token == "" || strings.ContainsAny(src.Token, "/?#& ") {
			return fmt.Errorf("token must be the board's slug, e.g. the \"acme\" in boards.greenhouse.io/acme")
		}
		if src.Label == "" {
			return fmt.Errorf("label is required: it is the company name its leads are filed under")
		}
		src.Query, src.Location = "", ""
	case IsSearchSource(src.Kind):
		if src.Query == "" {
			return fmt.Errorf("query is required for a search source")
		}
		if src.Label == "" {
			src.Label = src.Query
		}
		src.Token = ""
	default:
		return fmt.Errorf("kind must be one of greenhouse, lever, ashby, adzuna, reed")
	}
	return nil
}

// Pre-filter

// leadFilter is the keyword test a lead must pass before it is shown or
// costs a Claude call. It is intentionally crude and cheap: its job is to
// throw out the obviously irrelevant ninety percent of a company board
// (sales, legal, other countries), not to judge fit — that is the score.
type leadFilter struct {
	include   []string
	exclude   []string
	locations []string
}

func newLeadFilter(profile *models.JobSearchProfile) leadFilter {
	return leadFilter{
		include:   splitKeywords(profile.IncludeKeywords),
		exclude:   splitKeywords(profile.ExcludeKeywords),
		locations: splitKeywords(profile.Locations),
	}
}

// splitKeywords reads a keyword list as the page stores it: one entry per
// line, commas also accepted, case and surrounding space ignored.
func splitKeywords(list string) []string {
	var out []string
	for _, entry := range strings.FieldsFunc(list, func(r rune) bool { return r == '\n' || r == ',' }) {
		if entry = strings.ToLower(strings.TrimSpace(entry)); entry != "" {
			out = append(out, entry)
		}
	}
	return out
}

// matches reports whether a posting passes. An empty list is no constraint,
// so with nothing configured every lead matches.
//
//   - the title must contain one of the include keywords, and none of the
//     exclude keywords, as whole words (see containsWord);
//   - the location must contain one of the wanted locations as a plain
//     substring ("london" finds "Cardiff, London or Remote (UK)"). A posting
//     that states no location passes: unknown is not the same as wrong.
func (f leadFilter) matches(title, location string) bool {
	title = strings.ToLower(title)
	if len(f.include) > 0 && !containsAnyWord(title, f.include) {
		return false
	}
	if containsAnyWord(title, f.exclude) {
		return false
	}
	location = strings.ToLower(strings.TrimSpace(location))
	if len(f.locations) == 0 || location == "" {
		return true
	}
	for _, wanted := range f.locations {
		if strings.Contains(location, wanted) {
			return true
		}
	}
	return false
}

func containsAnyWord(text string, words []string) bool {
	for _, word := range words {
		if containsWord(text, word) {
			return true
		}
	}
	return false
}

// containsWord reports whether word occurs in text on word boundaries, with
// an optional plural "s". Boundaries are what let "intern" be excluded
// without also excluding "Internal Tools Engineer", and the plural is what
// lets "engineer" find "Engineers". Both arguments are already lowercase.
func containsWord(text, word string) bool {
	isWordRune := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
	for from := 0; ; {
		i := strings.Index(text[from:], word)
		if i < 0 {
			return false
		}
		start := from + i
		end := start + len(word)
		before := []rune(text[:start])
		after := []rune(strings.TrimPrefix(text[end:], "s"))
		if (len(before) == 0 || !isWordRune(before[len(before)-1])) &&
			(len(after) == 0 || !isWordRune(after[0])) {
			return true
		}
		from = start + 1
	}
}

// Scoring

// LeadScore is the JSON contract for Claude's verdict on one lead: the body
// of the Python service's reply (LeadScore in python/app/job_score.py).
type LeadScore struct {
	Score   int    `json:"score"`
	Reasons string `json:"reasons"`
}

// scoreRequest is the body of POST /internal/jobs/score; the field names are
// JobScoreRequest's in python/app/main.py.
type scoreRequest struct {
	Profile     string `json:"profile"`
	JobTitle    string `json:"jobTitle"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

// scoreLeads asks Claude to score the matched leads that have no score yet,
// newest first, up to maxScoresPerRun, and reports how many it scored.
//
// With no profile text there is nothing to score against, so it does nothing
// — leads still arrive and can be read, they just carry no score.
func (s *JobLeadService) scoreLeads(ctx context.Context, profile *models.JobSearchProfile) (int, error) {
	if strings.TrimSpace(profile.Profile) == "" {
		return 0, nil
	}
	var leads []models.JobLead
	if err := s.DB.
		Where("state = ? AND matched = ? AND score IS NULL AND score_attempts < ?", LeadNew, true, maxScoreAttempts).
		Order("created_at DESC, id DESC").Limit(maxScoresPerRun).
		Find(&leads).Error; err != nil {
		return 0, fmt.Errorf("loading leads to score: %w", err)
	}

	scored := 0
	for i := range leads {
		lead := &leads[i]
		result, err := s.scoreLead(ctx, profile.Profile, lead)

		var readErr *PythonError
		switch {
		case err == nil:
			reasons := strings.TrimSpace(result.Reasons)
			score := min(max(result.Score, 0), 100)
			if dbErr := s.DB.Model(lead).Updates(map[string]any{"score": score, "score_reasons": reasons}).Error; dbErr != nil {
				return scored, fmt.Errorf("saving a score: %w", dbErr)
			}
			scored++
		case errors.As(err, &readErr):
			// Claude answered badly for this posting. Count the attempt and
			// move on; maxScoreAttempts stops it being retried forever.
			log.Printf("[JobLeads] Scoring lead %d failed: %v", lead.ID, err)
			if dbErr := s.DB.Model(lead).UpdateColumn("score_attempts", gorm.Expr("score_attempts + 1")).Error; dbErr != nil {
				return scored, fmt.Errorf("recording a failed score: %w", dbErr)
			}
		default:
			// Anything else is not about this lead: no API key, the Python
			// service down, the run cancelled. The next lead would fail the
			// same way, so stop, and spend no attempt on it.
			// The detail names internal addresses and goes to the log; the
			// returned error is shown on the leads page.
			log.Printf("[JobLeads] Scoring stopped at lead %d: %v", lead.ID, err)
			return scored, errors.New("scoring is unavailable: the Python service did not answer, or has no Claude key (see the backend log)")
		}
	}
	return scored, nil
}

func (s *JobLeadService) scoreLead(ctx context.Context, profile string, lead *models.JobLead) (LeadScore, error) {
	return CallPython[LeadScore](ctx, s.PythonURL, jobScorePath, scoreRequest{
		Profile:     truncateRunes(profile, maxProfileRunes),
		JobTitle:    lead.JobTitle,
		Company:     lead.Company,
		Location:    lead.Location,
		Description: lead.Description,
	})
}

// Search profile

// maxProfileRunes caps the profile text, both as stored and as sent. It
// mirrors MAX_PROFILE_CHARS in python/app/job_score.py.
const maxProfileRunes = 12_000

// jobSearchProfileID is the one row's id; see models.JobSearchProfile.
const jobSearchProfileID = 1

// LoadJobSearchProfile returns the single profile row, creating it empty the
// first time it is asked for.
func LoadJobSearchProfile(db *gorm.DB) (*models.JobSearchProfile, error) {
	profile := models.JobSearchProfile{ID: jobSearchProfileID}
	if err := db.FirstOrCreate(&profile, models.JobSearchProfile{ID: jobSearchProfileID}).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// SaveJobSearchProfile stores an edited profile, bounding the free text.
func SaveJobSearchProfile(db *gorm.DB, profile *models.JobSearchProfile) error {
	profile.ID = jobSearchProfileID
	profile.Profile = truncateRunes(strings.TrimSpace(profile.Profile), maxProfileRunes)
	profile.MinScore = min(max(profile.MinScore, 0), 100)
	return db.Save(profile).Error
}
