package services

import (
	"log"
	"time"

	"adam-french.co.uk/backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) {
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

	testUser := models.User{
		Username: "testuser",
		Password: hashedPassword,
		Admin:    true,
	}
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

	log.Println("Database seeded successfully")
}
