package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model        // includes ID, CreatedAt, UpdatedAt, DeletedAt
	Title      string `gorm:"not null"`
	AuthorID   uint
	Author     User `gorm:"foreignKey:AuthorID"`
	Content    string
}
