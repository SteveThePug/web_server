package services

import (
	"fmt"

	"adam-french.co.uk/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SQLConfig holds the PostgreSQL connection parameters, all read from the
// environment in main.
type SQLConfig struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     string
}

// connectToPostgreSQL opens the GORM connection. sslmode=disable is safe
// only because Postgres is reachable exclusively on the private Docker
// network, never over the internet.
func connectToPostgreSQL(config *SQLConfig) (*gorm.DB, error) {
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

// migrateDatabase brings the schema up to date. This project has no
// migration files: GORM's AutoMigrate is the only schema source, so it will
// add tables, columns and indexes but never drop or rename anything. Removing
// a field from a model therefore leaves the column behind, and a renamed
// field appears as a brand-new column with no data. Any model added to
// models.go must also be listed here or its table will never be created.
func migrateDatabase(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Activity{},
		&models.Favorite{},
		&models.Rowing{},
		&models.Message{},
		&models.Bookmark{},
		&models.Place{},
		&models.Box{},
		&models.Page{},
		&models.Background{},
		&models.Layout{},
		&models.Component{},
	)
	if err != nil {
		return err
	}

	return ensureStpPage(db)
}

// ensureStpPage turns the home page (/stp) into an ordinary canvas Page once.
// It used to be a built-in canvas with its own route; its boxes and
// background were already keyed "stp", so a Page row with that slug is all
// it takes. Unscoped, so that deleting the page on purpose sticks rather than
// bringing it back on the next start.
func ensureStpPage(db *gorm.DB) error {
	var count int64
	if err := db.Unscoped().Model(&models.Page{}).Where("slug = ?", "stp").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&models.Page{Slug: "stp", Title: "STP"}).Error
}

// InitDatabase connects to Postgres and runs the auto-migration.
func InitDatabase(config *SQLConfig) (*gorm.DB, error) {
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
