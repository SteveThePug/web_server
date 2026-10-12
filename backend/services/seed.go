package services

import (
	"log"
	"time"

	"adam-french.co.uk/backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDatabase populates an empty database with demo content for local
// development. It is called from main only when SEED_DB=true, which
// docker-compose.dev.yml sets; it must never run in production because it
// creates an admin account with the password "password".
func SeedDatabase(db *gorm.DB) {
	// "Any user exists" is the whole idempotency check — if the users table
	// has a row, nothing is seeded, even if the other tables were emptied.
	var user models.User
	if db.First(&user).Error == nil {
		log.Println("Database already has data, skipping seed")
		return
	}

	log.Println("Seeding database with test data...")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash seed password:", err)
	}

	// Dev-only credentials: testuser / password.
	testUser := models.User{
		Username: "testuser",
		Password: hashedPassword,
		Admin:    true,
	}
	// GORM writes the generated primary key back into testUser, which the
	// posts below then reference as their author.
	db.Create(&testUser)

	posts := []models.Post{
		{Title: "Welcome to my blog", Content: "This is the first test post with some example content.", AuthorID: testUser.ID},
		{Title: "Learning Go", Content: "Go is a great language for building web servers and APIs.", AuthorID: testUser.ID},
		{Title: "Vue 3 Tips", Content: "The composition API makes Vue components much more flexible.", AuthorID: testUser.ID},
	}
	db.Create(&posts)

	link1 := "https://example.com/project"
	link2 := "https://example.com/book"
	activities := []models.Activity{
		{Type: "project", Name: "coding"},
		{Type: "hobby", Name: "reading", Link: &link1},
		{Type: "fitness", Name: "exercise"},
	}
	db.Create(&activities)

	favorites := []models.Favorite{
		{Type: "language", Name: "Go"},
		{Type: "book", Name: "Designing Data-Intensive Applications", Link: &link2},
		{Type: "framework", Name: "Vue"},
	}
	db.Create(&favorites)

	now := time.Now()
	rowingEntries := []models.Rowing{
		{Date: now.AddDate(0, 0, -14), Time: 1800, Distance: 5000, TimePer500m: 120.0, Calories: 300},
		{Date: now.AddDate(0, 0, -7), Time: 1750, Distance: 5200, TimePer500m: 118.5, Calories: 315},
		{Date: now, Time: 1700, Distance: 5400, TimePer500m: 116.2, Calories: 330},
	}
	db.Create(&rowingEntries)

	// Job leads, so /cv/leads has something to show without a poll: the dev
	// stack usually has no Claude key, and the scheduler is off by default.
	// The sources are real public boards, so "Check now" works against them.
	db.Create(&models.JobSearchProfile{
		ID:              jobSearchProfileID,
		Profile:         "Five years of backend work in Go and Python. Looking for a mid-level or senior backend or platform role, in London or remote in the UK.",
		IncludeKeywords: "engineer\ndeveloper",
		ExcludeKeywords: "intern\nmanager",
		Locations:       "london\nremote",
		MinScore:        60,
	})
	sources := []models.JobSource{
		{Kind: SourceGreenhouse, Label: "Monzo", Token: "monzo", Enabled: true},
		{Kind: SourceAshby, Label: "Ramp", Token: "ramp", Enabled: false},
	}
	db.Create(&sources)
	high, mid, low := 86, 64, 35
	highWhy := "Your Go and platform experience matches the core requirements; no payments background."
	midWhy := "Backend fit is good, but the role wants Kotlin, which your profile does not mention."
	lowWhy := "The role is a staff-level position leading several teams, above the level you are after."
	posted := now.AddDate(0, 0, -3)
	leads := []models.JobLead{
		{SourceID: sources[0].ID, SourceKind: SourceGreenhouse, ExternalID: "monzo/seed-1", JobTitle: "Backend Engineer, Platform", Company: "Monzo", Location: "London or Remote (UK)", URL: "https://example.com/jobs/1", Description: "We are looking for a backend engineer to join our platform team.\n\nYou will build and run Go services.", PostedAt: &posted, LastSeenAt: now, Matched: true, State: LeadNew, Score: &high, ScoreReasons: &highWhy},
		{SourceID: sources[0].ID, SourceKind: SourceGreenhouse, ExternalID: "monzo/seed-2", JobTitle: "Senior Backend Engineer", Company: "Monzo", Location: "London", URL: "https://example.com/jobs/2", Description: "Kotlin services at scale.", LastSeenAt: now, Matched: true, State: LeadNew, Score: &mid, ScoreReasons: &midWhy},
		{SourceID: sources[0].ID, SourceKind: SourceGreenhouse, ExternalID: "monzo/seed-3", JobTitle: "Staff Engineer", Company: "Monzo", Location: "London", URL: "https://example.com/jobs/3", LastSeenAt: now, Matched: true, State: LeadNew, Score: &low, ScoreReasons: &lowWhy},
		{SourceID: sources[0].ID, SourceKind: SourceGreenhouse, ExternalID: "monzo/seed-4", JobTitle: "Software Engineer, Data", Company: "Monzo", Location: "Remote (UK)", URL: "https://example.com/jobs/4", LastSeenAt: now, Matched: true, State: LeadNew},
		{SourceID: sources[0].ID, SourceKind: SourceGreenhouse, ExternalID: "monzo/seed-5", JobTitle: "Account Executive", Company: "Monzo", Location: "London", URL: "https://example.com/jobs/5", LastSeenAt: now, Matched: false, State: LeadNew},
		{SourceID: sources[0].ID, SourceKind: SourceGreenhouse, ExternalID: "monzo/seed-6", JobTitle: "Android Engineer", Company: "Monzo", Location: "London", URL: "https://example.com/jobs/6", LastSeenAt: now, Matched: true, State: LeadDismissed, Score: &low},
	}
	db.Create(&leads)

	log.Println("Database seeded successfully")
}
