package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"belimang/internal/app/image"
	"belimang/internal/app/items"
	"belimang/internal/app/merchant"
	"belimang/internal/app/purchase"
	"belimang/internal/app/user"
	"belimang/internal/config"
	"belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	"belimang/internal/pkg/jwt"
	logger "belimang/internal/pkg/logging"
	"belimang/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init()

	// Initialize database
	db, err := database.NewDatabase(ctx, cfg.Database.DbUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(cfg.Cache)
	defer redisCache.Close()

	// Initialize shared services using configuration
	jwtService := jwt.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.Issuer)
	passwordService := utils.NewPasswordService()
	validator := validator.New()

	// Initialize Gin router
	router := gin.Default()

	// Setup routes with shared dependencies
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Initialize user components with shared dependencies
	userService := user.NewUserService(db.Queries, redisCache, jwtService, passwordService)
	userHandler := user.NewUserHandler(userService, validator)
	user.RegisterRoutes(router, userHandler)

	// Item
	itemService := items.NewItemService(db.Queries, redisCache)
	itemHandler := items.NewItemHandler(itemService)
	items.ItemRoutes(router, itemHandler, jwtService)

	// Purchase
	purhcaseService := purchase.NewPurchaseService(db.Queries, db)
	purchaseHandler := purchase.NewPurchaseHandler(purhcaseService, validator)
	purchase.PurchaseRoutes(router, purchaseHandler, jwtService)

	// Initialize merchant components with shared dependencies
	merchantService := merchant.NewMerchantService(redisCache, db.Queries)
	merchantHandler := merchant.NewMerchantHandler(merchantService, validator)
	merchant.MerchantRoutes(router, merchantHandler, jwtService)

	// Image
	imageHandler := image.NewImageHandler()
	image.RegisterRoutes(router, imageHandler)

	// Start HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	log.Printf("Server starting on port %d", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
