package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model        // includes ID, CreatedAt, UpdatedAt, DeletedAt
	Title      string `gorm:"not null"`
	Content    string `gorm:"type:text; not null"`
	AuthorID   uint   // foreign key to User
	Author     User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // optional relation
}
