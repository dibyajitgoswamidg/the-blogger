package tenant

import (
	"fmt"

	"github.com/dibyajitgoswamidg/the-blogger/internal/auth"
	"github.com/dibyajitgoswamidg/the-blogger/internal/platform/database"
	"github.com/dibyajitgoswamidg/the-blogger/internal/post"
	"gorm.io/gorm"
)

type Service struct {
	db       *gorm.DB
	tenantDB *database.TenantDB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:       db,
		tenantDB: database.NewTenantDB(db),
	}
}

func (s *Service) CreateTenant(name, subdomain string) (*Tenant, error) {
	// Generate schema name
	schema := fmt.Sprintf("tenant_%s", subdomain)

	tenant := Tenant{
		Name:      name,
		Subdomain: subdomain,
		Schema:    schema,
		IsActive:  true,
	}

	// First create the tenant record in the main schema
	if err := s.db.Create(&tenant).Error; err != nil {
		return nil, err
	}

	// Then create the schema
	if err := s.tenantDB.CreateSchema(schema); err != nil {
		return nil, err
	}

	// Set search path setting for a specific connection
	// Get a new DB connection with the specific schema set
	tenantDB, err := s.tenantDB.GetConnectionForSchema(schema)
	if err != nil {
		return nil, err
	}

	// Migrate tables for the tenant using the tenant-specific connection
	if err := s.migrateTenantTables(tenantDB); err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (s *Service) migrateTenantTables(db *gorm.DB) error {
	// Migrate tenant-specific tables
	return db.AutoMigrate(
		&auth.User{},
		&post.Post{},
		// Add other tenant-specific models here
	)
}
