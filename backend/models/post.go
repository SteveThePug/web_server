package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model        // includes ID, CreatedAt, UpdatedAt, DeletedAt
	Title      string `gorm:"not null" json:"title"`
	AuthorID   uint   `json:"-"`
	Author     User   `gorm:"foreignKey:AuthorID;references:ID" json:"author"`
	Content    string `json:"content"`
}
