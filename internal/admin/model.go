package admin

import (
	"gorm.io/gorm"
)

type AdminUser struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	Password     string `gorm:"not null" json:"-"`
	Name         string `json:"name"`
	IsSuperAdmin bool   `gorm:"column:is_super_admin;default:false;" json:"is_super_admin"`
	TenantID     *uint  `json:"tenant_id"` // null for superadmin
}

type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}
