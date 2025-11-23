package services

import (
	"fmt"

	"adam-french.co.uk/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SQLConfig struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     string
}

func connectToPostgreSQL(config SQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		config.User, config.Password, config.DBName, config.Host, config.Port,
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

func InitDatabase(config SQLConfig) (*gorm.DB, error) {
	db, err := connectToPostgreSQL(config)
	if err != nil {
		return nil, err
	}

	err = migrateDatabase(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
