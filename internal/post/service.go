package post

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) CreatePost(userID uint, req CreatePostRequest) (*Post, error) {
	post := Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
		Status:  req.Status,
	}

	if req.Status == "published" {
		post.PublishedAt = time.Now()
	}

	if err := s.db.Create(&post).Error; err != nil {
		return nil, err
	}

	// Load the associated user
	if err := s.db.Preload("User").First(&post, post.ID).Error; err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *Service) GetPost(id uint) (*Post, error) {
	var post Post
	if err := s.db.Preload("User").First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *Service) ListPosts(page, pageSize int) ([]Post, int64, error) {
	var posts []Post
	var total int64

	query := s.db.Model(&Post{}).Where("status = ?", "published")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Preload("User").
		Order("published_at DESC").
		Find(&posts).Error

	return posts, total, err
}

func (s *Service) UpdatePost(id, userID uint, req UpdatePostRequest) (*Post, error) {
	post := Post{}
	if err := s.db.First(&post, id).Error; err != nil {
		return nil, err
	}

	if post.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Status != "" {
		updates["status"] = req.Status
		if req.Status == "published" && post.Status != "published" {
			updates["published_at"] = time.Now()
		}
	}

	if err := s.db.Model(&post).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload the post with user information
	if err := s.db.Preload("User").First(&post, id).Error; err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *Service) DeletePost(id, userID uint) error {
	post := Post{}
	if err := s.db.First(&post, id).Error; err != nil {
		return err
	}

	if post.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.db.Delete(&post).Error
}
