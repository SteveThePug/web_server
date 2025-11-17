package main

import (
	"fmt"
	"net/http"

	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

func main() {
	db, err := connectToPostgreSQL()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Perform database migration
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&models.Post{})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Go!")
	})

	fmt.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
