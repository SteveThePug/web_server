package models

import "gorm.io/gorm"

type User struct {
	gorm.Model        // includes ID, CreatedAt, UpdatedAt, DeletedAt
	Username   string `gorm:"uniqueIndex"`
	Password   []byte
	Admin      bool
}
