package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dibyajitgoswamidg/the-blogger/internal/admin"
	"github.com/dibyajitgoswamidg/the-blogger/internal/auth"
	"github.com/dibyajitgoswamidg/the-blogger/internal/platform/middleware"
	"github.com/dibyajitgoswamidg/the-blogger/internal/post"
	"github.com/dibyajitgoswamidg/the-blogger/internal/tenant"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Basic health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	if err := db.Exec("SET search_path TO public").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&tenant.Tenant{}, &admin.AdminUser{}); err != nil {
		log.Fatal(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	// Initialize admin service and handler
	adminService := admin.NewService(db, jwtSecret)
	adminHandler := admin.NewHandler(adminService)

	// Initialize tenant middleware
	tenantMiddleware := middleware.NewTenantMiddleware(db)

	// Initialize tenant service
	tenantService := tenant.NewService(db)

	// API routes
	api := r.Group("/api/v1")
	{
		// Admin routes for tenant management (no tenant middleware)
		admin := api.Group("/admin")
		{
			// Public admin routes
			admin.POST("/login", adminHandler.AdminLogin)

			// First-time setup route (should be disabled after first use)
			if os.Getenv("ALLOW_SETUP") == "true" {
				admin.POST("/setup", adminHandler.CreateSuperAdmin)
			}

			// Protected admin routes
			adminProtected := admin.Group("")
			adminProtected.Use(middleware.AuthMiddleware(jwtSecret))
			adminProtected.Use(func(c *gin.Context) {
				if !c.GetBool("is_super_admin") {
					c.JSON(http.StatusForbidden, gin.H{"error": "superadmin access required"})
					c.Abort()
					return
				}
				c.Next()
			})

			{
				// Add tenant management endpoints here
				adminProtected.POST("/tenants", func(c *gin.Context) {
					var req struct {
						Name      string `json:"name" binding:"required"`
						Subdomain string `json:"subdomain" binding:"required"`
					}
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(400, gin.H{"error": err.Error()})
						return
					}

					tenant, err := tenantService.CreateTenant(req.Name, req.Subdomain)
					if err != nil {
						c.JSON(500, gin.H{"error": err.Error()})
						return
					}

					c.JSON(201, tenant)
				})
			}
		}

		// Initialize auth service and handler
		authService := auth.NewService(db, jwtSecret)
		authHandler := auth.NewHandler(authService)

		// Initialize post service and handler
		postService := post.NewService(db)
		postHandler := post.NewHandler(postService)

		// Multi-tenant routes
		tenantRoutes := api.Group("")
		tenantRoutes.Use(tenantMiddleware.IdentifyTenant())
		{
			// Public routes
			tenantRoutes.POST("/register", authHandler.Register)
			tenantRoutes.POST("/login", authHandler.Login)
			tenantRoutes.GET("/posts", postHandler.List)

			// Protected routes
			protected := tenantRoutes.Group("")
			protected.Use(middleware.AuthMiddleware(jwtSecret))
			{
				protected.GET("/me", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"user_id": c.GetString("id"),
						"email":   c.GetString("email"),
						"role":    c.GetString("role"),
					})
				})

				protected.POST("/posts", postHandler.Create)
				protected.GET("/posts/:id", postHandler.Get)
				protected.PUT("/posts/:id", postHandler.Update)
				protected.DELETE("/posts/:id", postHandler.Delete)
			}
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
