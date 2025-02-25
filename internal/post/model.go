package post

import (
	"time"

	"github.com/dibyajitgoswamidg/the-blogger/internal/auth"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title       string    `json:"title" gorm:"not null"`
	Content     string    `json:"content" gorm:"not null"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	User        auth.User `json:"user" gorm:"foreignKey:UserID"`
	Status      string    `json:"status" gorm:"default:'draft'"` // draft, published
	PublishedAt time.Time `json:"published_at,omitempty"`
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=3"`
	Content string `json:"content" binding:"required,min=10"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"omitempty,min=3"`
	Content string `json:"content" binding:"omitempty,min=10"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published"`
}
