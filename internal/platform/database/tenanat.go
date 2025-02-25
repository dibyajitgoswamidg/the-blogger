package database

import (
	"fmt"

	"gorm.io/gorm"
)

type TenantDB struct {
	DB *gorm.DB
}

func NewTenantDB(db *gorm.DB) *TenantDB {
	return &TenantDB{DB: db}
}

func (t *TenantDB) CreateSchema(schema string) error {
	return t.DB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error
}

func (t *TenantDB) SetSearchPath(schema string) error {
	return t.DB.Exec(fmt.Sprintf("SET search_path TO %s", schema)).Error
}

func (t *TenantDB) GetConnectionForSchema(schema string) (*gorm.DB, error) {
	// Get a new connection with the schema set
	db := t.DB.Session(&gorm.Session{})

	// Set the search path for this connection
	if err := db.Exec(fmt.Sprintf("SET search_path TO %s", schema)).Error; err != nil {
		return nil, err
	}

	return db, nil
}
