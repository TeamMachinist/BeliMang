package main

import (
	"fmt"
	"log"
	"net/http"

	"sinibeli/internal/app/user"
	"sinibeli/internal/config"
	"sinibeli/internal/infrastructure/cache"
	"sinibeli/internal/infrastructure/database"
	logger "sinibeli/internal/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init()

	// Initialize database
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(cfg.Cache)
	defer redisCache.Close()

	// Initialize shared services using configuration
	// jwtService := jwt.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.Issuer)
	// passwordService := utils.NewPasswordService()
	validator := validator.New()

	// Initialize Gin router
	router := gin.Default()

	// Setup routes with shared dependencies
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create API v1 group
	v1 := router.Group("/api/v1")

	// Initialize user components with shared dependencies
	userRepo := user.NewUserRepository(db.DB)
	userService := user.NewUserService(userRepo, redisCache)
	userHandler := user.NewUserHandler(userService, validator)
	user.UserRoutes(v1, userHandler)

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
