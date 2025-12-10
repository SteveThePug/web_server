package models

import (
	"database/sql"
	"time"
)

type Post struct {
	ID        uint         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `gorm:"index" json:"deletedAt"`
	Title     string       `gorm:"not null" json:"title"`
	AuthorID  uint         `json:"-"`
	Author    *User        `gorm:"foreignKey:AuthorID" json:"author"`
	Content   string       `json:"content"`
}
