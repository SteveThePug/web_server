package services

import (
	"fmt"
	"os"

	"adam-french.co.uk/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func InitDatabase() (*gorm.DB, error) {
	db, err := connectToPostgreSQL()
	if err != nil {
		return nil, err
	}

	err = migrateDatabase(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
