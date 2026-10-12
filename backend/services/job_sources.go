package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"adam-french.co.uk/backend/models"
)

// The job sources the lead fetcher (job_leads.go) polls: one function per
// models.JobSource.Kind, each turning that source's JSON into the common
// fetchedJob. Nothing here touches the database.
//
// The three board kinds are public, keyless endpoints that return a whole
// company board, description included, in one response. Their shapes were
// checked against live boards. The two search kinds need API keys and their
// shapes are taken from the providers' documentation — if one of them starts
// reporting zero results or a decode error, its struct here is the first
// thing to compare against a real response.

// Source kinds. The strings are stored in JobSource.Kind and JobLead.SourceKind
// and repeated in the leads page's kind picker (vue/src/views/CV/JobSources.vue).
const (
	SourceGreenhouse = "greenhouse"
	SourceLever      = "lever"
	SourceAshby      = "ashby"
	SourceAdzuna     = "adzuna"
	SourceReed       = "reed"
)

// IsBoardSource reports whether kind is one company's board, identified by a
// token, rather than a keyword search.
func IsBoardSource(kind string) bool {
	return kind == SourceGreenhouse || kind == SourceLever || kind == SourceAshby
}

// IsSearchSource reports whether kind is a keyword search.
func IsSearchSource(kind string) bool {
	return kind == SourceAdzuna || kind == SourceReed
}

// Base URLs, variables so the tests can point each at a stub server.
var (
	greenhouseBase = "https://boards-api.greenhouse.io"
	leverBase      = "https://api.lever.co"
	ashbyBase      = "https://api.ashbyhq.com"
	adzunaBase     = "https://api.adzuna.com"
	reedBase       = "https://www.reed.co.uk"
)

const (
	// maxSourceBody bounds one response. A large board with every description
	// inline runs to several megabytes (Palantir's Lever board is about six);
	// this is far above that and only there so a misbehaving endpoint cannot
	// fill the Pi's memory.
	maxSourceBody = 64 << 20

	// searchPages × searchPageSize is how deep a keyword search is read on
	// each poll. Results are newest first and polls are hours apart, so the
	// first hundred is where anything new will be.
	searchPages    = 2
	searchPageSize = 50

	// adzunaMaxDaysOld keeps a search to recent adverts; older ones were seen
	// by an earlier poll or are not worth surfacing now.
	adzunaMaxDaysOld = 14

	// maxLeadRunes caps a stored description. It mirrors MAX_DESCRIPTION_CHARS
	// in python/app/job_score.py, which rejects anything longer.
	maxLeadRunes = 40_000
)

// fetchedJob is one posting as a source reported it, before it becomes a
// models.JobLead.
type fetchedJob struct {
	// ExternalID is unique within the source kind; see models.JobLead.
	ExternalID  string
	Title       string
	Company     string
	Location    string
	URL         string
	Description string
	PostedAt    *time.Time
	// detailID is set when Description is only a snippet and the full text is
	// one more request away (Reed). The fetcher makes that request only for
	// a new lead that passed the pre-filter.
	detailID string
}

// fetchSource returns everything src currently lists.
func (s *JobLeadService) fetchSource(ctx context.Context, src models.JobSource) ([]fetchedJob, error) {
	switch src.Kind {
	case SourceGreenhouse:
		return s.fetchGreenhouse(ctx, src)
	case SourceLever:
		return s.fetchLever(ctx, src)
	case SourceAshby:
		return s.fetchAshby(ctx, src)
	case SourceAdzuna:
		return s.fetchAdzuna(ctx, src)
	case SourceReed:
		return s.fetchReed(ctx, src)
	}
	return nil, fmt.Errorf("unknown source kind %q", src.Kind)
}

// getJSON GETs target and decodes a 200 reply into out. basicUser, when set,
// is sent as the username of HTTP basic auth with an empty password, which is
// how Reed takes its API key.
//
// The error never includes the URL: Adzuna's carries the API key in its query
// string, and this error is stored in JobSource.LastError and shown on the
// page.
func (s *JobLeadService) getJSON(ctx context.Context, target, basicUser string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("building request")
	}
	req.Header.Set("Accept", "application/json")
	if basicUser != "" {
		req.SetBasicAuth(basicUser, "")
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		// A *url.Error's message quotes the full URL; keep only its cause.
		if urlErr, ok := err.(*url.Error); ok {
			err = urlErr.Err
		}
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("not found (404): check the board token")
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("refused (%d): check the API key", resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("source returned %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxSourceBody)).Decode(out); err != nil {
		return fmt.Errorf("unexpected response: %v", err)
	}
	return nil
}

// Greenhouse

type greenhouseBoard struct {
	Jobs []struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		AbsoluteURL string `json:"absolute_url"`
		// Content is HTML that has itself been entity-escaped ("&lt;p&gt;").
		Content        string `json:"content"`
		FirstPublished string `json:"first_published"`
		UpdatedAt      string `json:"updated_at"`
		Location       struct {
			Name string `json:"name"`
		} `json:"location"`
	} `json:"jobs"`
}

func (s *JobLeadService) fetchGreenhouse(ctx context.Context, src models.JobSource) ([]fetchedJob, error) {
	var board greenhouseBoard
	target := fmt.Sprintf("%s/v1/boards/%s/jobs?content=true", greenhouseBase, url.PathEscape(src.Token))
	if err := s.getJSON(ctx, target, "", &board); err != nil {
		return nil, err
	}
	jobs := make([]fetchedJob, 0, len(board.Jobs))
	for _, j := range board.Jobs {
		posted := j.FirstPublished
		if posted == "" {
			posted = j.UpdatedAt
		}
		jobs = append(jobs, fetchedJob{
			ExternalID: boardExternalID(src.Token, strconv.FormatInt(j.ID, 10)),
			Title:      j.Title,
			Company:    src.Label,
			Location:   j.Location.Name,
			URL:        j.AbsoluteURL,
			// Unescape once to get the HTML back; htmlToText handles the rest.
			Description: htmlToText(html.UnescapeString(j.Content)),
			PostedAt:    parseTime(time.RFC3339, posted),
		})
	}
	return jobs, nil
}

// Lever

type leverPosting struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	HostedURL  string `json:"hostedUrl"`
	CreatedAt  int64  `json:"createdAt"` // Unix milliseconds
	Categories struct {
		Location string `json:"location"`
	} `json:"categories"`
	WorkplaceType    string `json:"workplaceType"`
	DescriptionPlain string `json:"descriptionPlain"`
	AdditionalPlain  string `json:"additionalPlain"`
	// Lists are the posting's headed bullet sections ("What you'll do"),
	// which descriptionPlain does not include.
	Lists []struct {
		Text    string `json:"text"`
		Content string `json:"content"` // HTML <li> items
	} `json:"lists"`
}

func (s *JobLeadService) fetchLever(ctx context.Context, src models.JobSource) ([]fetchedJob, error) {
	var postings []leverPosting
	target := fmt.Sprintf("%s/v0/postings/%s?mode=json", leverBase, url.PathEscape(src.Token))
	if err := s.getJSON(ctx, target, "", &postings); err != nil {
		return nil, err
	}
	jobs := make([]fetchedJob, 0, len(postings))
	for _, p := range postings {
		parts := []string{strings.TrimSpace(p.DescriptionPlain)}
		for _, list := range p.Lists {
			parts = append(parts, strings.TrimSpace(list.Text)+"\n"+htmlToText(list.Content))
		}
		parts = append(parts, strings.TrimSpace(p.AdditionalPlain))

		var posted *time.Time
		if p.CreatedAt > 0 {
			t := time.UnixMilli(p.CreatedAt).UTC()
			posted = &t
		}
		jobs = append(jobs, fetchedJob{
			ExternalID:  boardExternalID(src.Token, p.ID),
			Title:       p.Text,
			Company:     src.Label,
			Location:    withWorkplace(p.Categories.Location, p.WorkplaceType),
			URL:         p.HostedURL,
			Description: joinNonEmpty(parts, "\n\n"),
			PostedAt:    posted,
		})
	}
	return jobs, nil
}

// Ashby

type ashbyBoard struct {
	Jobs []struct {
		ID               string `json:"id"`
		Title            string `json:"title"`
		Location         string `json:"location"`
		WorkplaceType    string `json:"workplaceType"`
		PublishedAt      string `json:"publishedAt"`
		IsListed         bool   `json:"isListed"`
		JobURL           string `json:"jobUrl"`
		DescriptionPlain string `json:"descriptionPlain"`
	} `json:"jobs"`
}

func (s *JobLeadService) fetchAshby(ctx context.Context, src models.JobSource) ([]fetchedJob, error) {
	var board ashbyBoard
	target := fmt.Sprintf("%s/posting-api/job-board/%s", ashbyBase, url.PathEscape(src.Token))
	if err := s.getJSON(ctx, target, "", &board); err != nil {
		return nil, err
	}
	jobs := make([]fetchedJob, 0, len(board.Jobs))
	for _, j := range board.Jobs {
		// Unlisted postings are reachable by direct link only; the company
		// has chosen not to advertise them.
		if !j.IsListed {
			continue
		}
		jobs = append(jobs, fetchedJob{
			ExternalID:  boardExternalID(src.Token, j.ID),
			Title:       strings.TrimSpace(j.Title),
			Company:     src.Label,
			Location:    withWorkplace(j.Location, j.WorkplaceType),
			URL:         j.JobURL,
			Description: strings.TrimSpace(j.DescriptionPlain),
			PostedAt:    parseTime(time.RFC3339, j.PublishedAt),
		})
	}
	return jobs, nil
}

// Adzuna

type adzunaResults struct {
	Results []struct {
		// A string in the API, though it is all digits.
		ID          json.Number `json:"id"`
		Title       string      `json:"title"`
		Description string      `json:"description"` // a snippet, about 500 characters
		RedirectURL string      `json:"redirect_url"`
		Created     string      `json:"created"`
		Company     struct {
			DisplayName string `json:"display_name"`
		} `json:"company"`
		Location struct {
			DisplayName string `json:"display_name"`
		} `json:"location"`
	} `json:"results"`
}

// fetchAdzuna runs one UK keyword search. Adzuna only ever returns a snippet
// of the description and has no endpoint for the rest, so these leads are
// scored on less text than the others.
func (s *JobLeadService) fetchAdzuna(ctx context.Context, src models.JobSource) ([]fetchedJob, error) {
	if s.Config.AdzunaAppID == "" || s.Config.AdzunaAppKey == "" {
		return nil, fmt.Errorf("ADZUNA_APP_ID and ADZUNA_APP_KEY are not set")
	}
	var jobs []fetchedJob
	for page := 1; page <= searchPages; page++ {
		query := url.Values{
			"app_id":           {s.Config.AdzunaAppID},
			"app_key":          {s.Config.AdzunaAppKey},
			"what":             {src.Query},
			"results_per_page": {strconv.Itoa(searchPageSize)},
			"max_days_old":     {strconv.Itoa(adzunaMaxDaysOld)},
			"sort_by":          {"date"},
		}
		if src.Location != "" {
			query.Set("where", src.Location)
		}
		var results adzunaResults
		target := fmt.Sprintf("%s/v1/api/jobs/gb/search/%d?%s", adzunaBase, page, query.Encode())
		if err := s.getJSON(ctx, target, "", &results); err != nil {
			return nil, err
		}
		for _, r := range results.Results {
			jobs = append(jobs, fetchedJob{
				ExternalID: r.ID.String(),
				// Adzuna wraps the matched search terms in <strong>.
				Title:       htmlToText(r.Title),
				Company:     r.Company.DisplayName,
				Location:    r.Location.DisplayName,
				URL:         r.RedirectURL,
				Description: htmlToText(r.Description),
				PostedAt:    parseTime(time.RFC3339, r.Created),
			})
		}
		if len(results.Results) < searchPageSize {
			break
		}
	}
	return jobs, nil
}

// Reed

type reedResults struct {
	Results []struct {
		JobID          int64  `json:"jobId"`
		EmployerName   string `json:"employerName"`
		JobTitle       string `json:"jobTitle"`
		LocationName   string `json:"locationName"`
		Date           string `json:"date"` // dd/mm/yyyy
		JobDescription string `json:"jobDescription"`
		JobURL         string `json:"jobUrl"`
	} `json:"results"`
}

type reedDetail struct {
	JobDescription string `json:"jobDescription"` // HTML
}

// fetchReed runs one keyword search. The search returns a snippet of each
// description; the full text is fetchReedDescription, one request per job.
func (s *JobLeadService) fetchReed(ctx context.Context, src models.JobSource) ([]fetchedJob, error) {
	if s.Config.ReedAPIKey == "" {
		return nil, fmt.Errorf("REED_API_KEY is not set")
	}
	var jobs []fetchedJob
	for page := 0; page < searchPages; page++ {
		query := url.Values{
			"keywords":      {src.Query},
			"resultsToTake": {strconv.Itoa(searchPageSize)},
			"resultsToSkip": {strconv.Itoa(page * searchPageSize)},
		}
		if src.Location != "" {
			query.Set("locationName", src.Location)
		}
		var results reedResults
		target := fmt.Sprintf("%s/api/1.0/search?%s", reedBase, query.Encode())
		if err := s.getJSON(ctx, target, s.Config.ReedAPIKey, &results); err != nil {
			return nil, err
		}
		for _, r := range results.Results {
			id := strconv.FormatInt(r.JobID, 10)
			jobs = append(jobs, fetchedJob{
				ExternalID:  id,
				Title:       r.JobTitle,
				Company:     r.EmployerName,
				Location:    r.LocationName,
				URL:         r.JobURL,
				Description: htmlToText(r.JobDescription),
				PostedAt:    parseTime("02/01/2006", r.Date),
				detailID:    id,
			})
		}
		if len(results.Results) < searchPageSize {
			break
		}
	}
	return jobs, nil
}

// fetchReedDescription returns the full text of one Reed advert.
func (s *JobLeadService) fetchReedDescription(ctx context.Context, jobID string) (string, error) {
	var detail reedDetail
	target := fmt.Sprintf("%s/api/1.0/jobs/%s", reedBase, url.PathEscape(jobID))
	if err := s.getJSON(ctx, target, s.Config.ReedAPIKey, &detail); err != nil {
		return "", err
	}
	return htmlToText(detail.JobDescription), nil
}

// Helpers

// boardExternalID namespaces a board's own job id with the board, because
// ids are only unique within one company's board.
func boardExternalID(token, id string) string {
	return strings.ToLower(strings.TrimSpace(token)) + "/" + id
}

// withWorkplace appends a posting's remote/hybrid/on-site label to its
// location, where the source reports one separately. The pre-filter matches
// on the result, so "remote" in the wanted locations finds these.
func withWorkplace(location, workplace string) string {
	location = strings.TrimSpace(location)
	workplace = strings.ToLower(strings.TrimSpace(workplace))
	if workplace == "" || workplace == "unspecified" || strings.Contains(strings.ToLower(location), workplace) {
		return location
	}
	if location == "" {
		return workplace
	}
	return location + " (" + workplace + ")"
}

func joinNonEmpty(parts []string, sep string) string {
	kept := parts[:0:0]
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, sep)
}

// parseTime returns nil for a value that does not parse: a missing posting
// date is not worth failing a fetch over.
func parseTime(layout, value string) *time.Time {
	t, err := time.Parse(layout, strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &t
}

var (
	// Tags that end a line in rendered HTML. Replaced with a newline before
	// the rest are stripped, so paragraphs and bullets survive as lines.
	htmlBlockTag  = regexp.MustCompile(`(?i)</?(p|div|br|li|ul|ol|h[1-6]|tr|section)\b[^>]*>`)
	htmlAnyTag    = regexp.MustCompile(`<[^>]*>`)
	htmlSpaces    = regexp.MustCompile(`[ \t\x{00a0}]+`)
	htmlBlankRuns = regexp.MustCompile(`\n[ \t]*(\n[ \t]*)+`)
)

// htmlToText reduces a posting's HTML to readable plain text. Like
// cleanEmailBody it is a regex strip, not a parser, which is enough for text
// that is read by a person and by a language model and never rendered as
// HTML. Unlike cleanEmailBody it keeps the whole text and its line structure.
func htmlToText(content string) string {
	text := htmlBlockTag.ReplaceAllString(content, "\n")
	text = htmlAnyTag.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = htmlSpaces.ReplaceAllString(text, " ")
	text = htmlBlankRuns.ReplaceAllString(text, "\n\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// truncateRunes cuts s to at most max runes, on a rune boundary so the result
// is still valid UTF-8 for Postgres and for JSON.
func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
