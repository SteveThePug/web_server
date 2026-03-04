package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Username  string         `gorm:"uniqueIndex" json:"username"`
	Password  []byte         `json:"-"`
	Admin     bool           `json:"admin"`
}

type Post struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Title     string         `gorm:"not null" json:"title"`
	AuthorID  uint           `json:"-"`
	Author    *User          `gorm:"foreignKey:AuthorID" json:"author"`
	Content   string         `json:"content"`
}

type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Content   string         `json:"text"`
	AuthorID  uint           `json:"-"`
	Author    *User          `gorm:"foreignKey:AuthorID" json:"author"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

type Activity struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Link      *string        `json:"link"`
}

type Favorite struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Link      *string        `json:"link"`
}

type Rowing struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Date        time.Time      `json:"date"`
	Time        time.Duration  `json:"time"`
	TimePer500m time.Duration  `json:"timePer500m"`
	Distance    float64        `json:"distance"`
	Calories    float64        `json:"calories"`
}
