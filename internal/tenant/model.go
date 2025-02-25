package tenant

import (
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model
	Subdomain string `gorm:"uniqueIndex;not null" json:"subdomain"`
	Name      string `gorm:"not null" json:"name"`
	Schema    string `gorm:"uniqueIndex;not null" json:"schema"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`
}
