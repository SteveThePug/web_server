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
	AuthorID  uint           `json:"authorId"`
	FileURL   string         `json:"fileUrl,omitempty"`
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
	Time        uint64         `json:"time"`
	Distance    uint64         `json:"distance"`
	TimePer500m float64        `json:"timePer500m"`
	Calories    float64        `json:"calories"`
}

type Bookmark struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Category  string         `gorm:"not null" json:"category"`
	Name      string         `gorm:"not null" json:"name"`
	Link      string         `gorm:"not null" json:"link"`
}

type JobAppReference struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Category  string         `gorm:"not null" json:"category"`
	Label     string         `gorm:"not null" json:"label"`
	Value     string         `gorm:"not null" json:"value"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
}

type JobApplication struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	JobTitle  string         `gorm:"not null" json:"jobTitle"`
	Company   string         `gorm:"not null" json:"company"`
	Location  *string        `json:"location"`
	URL       *string        `json:"url"`
	Status    string         `gorm:"not null" json:"status"`
	Notes     *string        `json:"notes"`
	AppliedAt *time.Time     `json:"appliedAt"`
}
