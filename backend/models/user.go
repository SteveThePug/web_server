package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        uint         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `gorm:"index" json:"deletedAt"`
	Username  string       `gorm:"uniqueIndex" json:"username"`
	Password  []byte       `json:"-"`
	Admin     bool         `json:"admin"`
}
