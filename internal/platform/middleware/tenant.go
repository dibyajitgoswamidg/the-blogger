package middleware

import (
	"strings"

	"github.com/dibyajitgoswamidg/the-blogger/internal/platform/database"
	"github.com/dibyajitgoswamidg/the-blogger/internal/tenant"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantMiddleware struct {
	db       *gorm.DB
	tenantDB *database.TenantDB
}

func NewTenantMiddleware(db *gorm.DB) *TenantMiddleware {
	return &TenantMiddleware{
		db:       db,
		tenantDB: database.NewTenantDB(db),
	}
}

func (tm *TenantMiddleware) IdentifyTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		subdomain := strings.Split(host, ".")[0]

		// Skip tenant check for the main domain
		if subdomain == "www" || subdomain == "api" {
			c.Next()
			return
		}

		var tenant tenant.Tenant
		tm.tenantDB.GetConnectionForSchema("public")
		if err := tm.db.Where("subdomain = ? AND is_active = ?", subdomain, true).First(&tenant).Error; err != nil {
			c.JSON(404, gin.H{"error": "tenant not found"})
			c.Abort()
			return
		}

		// Set tenant information in context
		c.Set("tenant_id", tenant.ID)
		c.Set("tenant_schema", tenant.Schema)

		// Set database search path for the tenant
		_, err := tm.tenantDB.GetConnectionForSchema(tenant.Schema)
		if err != nil {
			c.JSON(500, gin.H{"error": "database error"})
			c.Abort()
			return
		}

		c.Next()
	}
}
