package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"adam-french.co.uk/backend/handlers"
	"adam-french.co.uk/backend/models"
)

func connectToPostgreSQL() (*gorm.DB, error) {
	dsn := "user=postgres password=password dbname=simplebackend host=localhost port=5432 sslmode=disable"
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
	h := handlers.Handler{DB: db}

	r.GET("/posts", h.GetPosts)
	r.POST("/posts", h.CreatePost)
	r.PUT("/posts/:id", h.UpdatePost)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	r.Run()
}
