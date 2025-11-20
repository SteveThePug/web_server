package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"adam-french.co.uk/backend/handlers"
	"adam-french.co.uk/backend/models"
)

func connectToPostgreSQL() (*gorm.DB, error) {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		user, password, dbname, host, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDatabase(db *gorm.DB) error {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&models.Post{})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	db, err := connectToPostgreSQL()
	if err != nil {
		log.Fatal(err)
	}

	err = migrateDatabase(db)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	store := handlers.Store{DB: db}

	r.GET("/posts", store.GetPosts)
	r.POST("/posts", store.CreatePost)
	r.PUT("/posts/:id", store.UpdatePost)

	r.GET("/spotify", store.ListeningTo)
	r.POST("/spotify", store.SendSong)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	port := os.Getenv("BACKEND_PORT")
	r.Run(fmt.Sprintf(":%s", port))
}
