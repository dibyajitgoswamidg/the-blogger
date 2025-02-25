package admin

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db        *gorm.DB
	jwtSecret []byte
}

func NewService(db *gorm.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) CreateSuperAdmin(req CreateAdminRequest) (*AdminUser, error) {
	// Check if any superadmin exists
	var count int64
	s.db.Model(&AdminUser{}).Where("is_super_admin = ?", true).Count(&count)
	if count > 0 {
		return nil, errors.New("superadmin already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := AdminUser{
		Email:        req.Email,
		Password:     string(hashedPassword),
		Name:         req.Name,
		IsSuperAdmin: true,
	}

	if err := s.db.Create(&admin).Error; err != nil {
		return nil, err
	}

	return &admin, nil
}

func (s *Service) CreateTenantAdmin(tenantID uint, req CreateAdminRequest) (*AdminUser, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := AdminUser{
		Email:        req.Email,
		Password:     string(hashedPassword),
		Name:         req.Name,
		IsSuperAdmin: false,
		TenantID:     &tenantID,
	}

	if err := s.db.Create(&admin).Error; err != nil {
		return nil, err
	}

	return &admin, nil
}

func (s *Service) LoginAdmin(email, password string) (string, error) {
	var admin AdminUser
	if err := s.db.Where("email = ?", email).First(&admin).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"sub":            admin.ID,
		"email":          admin.Email,
		"is_super_admin": admin.IsSuperAdmin,
		"tenant_id":      admin.TenantID,
		"exp":            time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
