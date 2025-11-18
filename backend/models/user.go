package models

import "gorm.io/gorm"

type User struct {
	gorm.Model        // includes ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string `gorm:"uniqueIndex"`
	Email      string `gorm:"uniqueIndex"`
	Password   string
}
