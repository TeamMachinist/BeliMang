package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"belimang/internal/app/image"
	"belimang/internal/app/items"
	"belimang/internal/app/merchant"
	"belimang/internal/app/purchase"
	"belimang/internal/app/user"
	"belimang/internal/config"
	"belimang/internal/infrastructure/cache"
	"belimang/internal/infrastructure/database"
	"belimang/internal/observability/metrics"
	"belimang/internal/pkg/jwt"
	logger "belimang/internal/pkg/logging"
	"belimang/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sys/unix"
)

var startTime time.Time

func main() {
	startTime = time.Now()
	ctx := context.Background()

	// OPTIMIZED: Go runtime tuning for 60k RPS
	// GOMAXPROCS already set via env (GOMAXPROCS=2 per pod)
	// But ensure it's not overridden
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	debug.SetGCPercent(200)

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Init()

	// Initialize database with automatic metrics
	db, err := database.NewDatabase(ctx, cfg.Database.DbUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisCache := cache.NewRedisCache(cfg.Cache)
	defer redisCache.Close()

	// Initialize services
	jwtService := jwt.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.Issuer)
	passwordService := utils.NewPasswordService()
	validator := validator.New()

	// Set app info metrics (once)
	metrics.AppInfo.WithLabelValues("1.0.0", cfg.Server.Env, runtime.Version()).Set(1)

	// OPTIMIZED: Update uptime less frequently (30s instead of 10s)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics.AppUptime.Set(time.Since(startTime).Seconds())
		}
	}()

	// CRITICAL: Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// IMPORTANT: Add recovery middleware first (for panic handling)
	router.Use(gin.Recovery())

	router.Use(metrics.PrometheusMiddleware())

	metricsHandler := promhttp.Handler()
	router.GET("/metrics", func(c *gin.Context) {
		metricsHandler.ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/healthz/deep", func(c *gin.Context) {
		if err := db.HealthCheck(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "connected"})
	})

	// Register routes
	userService := user.NewUserService(db.Queries, redisCache, jwtService, passwordService)
	userHandler := user.NewUserHandler(userService, validator)
	user.RegisterRoutes(router, userHandler)

	itemService := items.NewItemService(db.Queries, redisCache)
	itemHandler := items.NewItemHandler(itemService)
	items.ItemRoutes(router, itemHandler, jwtService)

	purchaseService := purchase.NewPurchaseService(db.Queries, db)
	purchaseHandler := purchase.NewPurchaseHandler(purchaseService, validator)
	purchase.PurchaseRoutes(router, purchaseHandler, jwtService)

	merchantService := merchant.NewMerchantService(redisCache, db.Queries)
	merchantHandler := merchant.NewMerchantHandler(merchantService, validator)
	merchant.MerchantRoutes(router, merchantHandler, jwtService)

	imageHandler := image.NewImageHandler()
	image.RegisterRoutes(router, imageHandler)

	// OPTIMIZED: Create listener with SO_REUSEPORT
	ln, err := createListener(ctx, cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	// OPTIMIZED: HTTP server configuration for high throughput
	srv := &http.Server{
		Handler: router,

		// Connection timeouts
		ReadTimeout:       3 * time.Second,  // Time to read request
		ReadHeaderTimeout: 2 * time.Second,  // Time to read headers only
		WriteTimeout:      5 * time.Second,  // Time to write response
		IdleTimeout:       90 * time.Second, // Keep-alive timeout

		// Limits
		MaxHeaderBytes: 1 << 20, // 1 MB max header size

		// OPTIONAL: Disable HTTP/2 if you don't need it (slight performance gain)
		// TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("🚀 Server starting on port %d", cfg.Server.Port)
	log.Printf("⚙️  GOMAXPROCS: %d (CPU cores: %d)", runtime.GOMAXPROCS(0), runtime.NumCPU())
	log.Printf("🧠 GC: %d%% (Memory: %v)", debug.SetGCPercent(-1), debug.SetMemoryLimit(-1))
	log.Printf("📊 Metrics: http://localhost:%d/metrics", cfg.Server.Port)
	log.Printf("❤️  Health: http://localhost:%d/healthz", cfg.Server.Port)

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// createListener creates a TCP listener with SO_REUSEPORT for better performance
// This allows multiple processes to bind to the same port (Linux/BSD only)
func createListener(ctx context.Context, port int) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var opErr error
			err := c.Control(func(fd uintptr) {
				// Enable SO_REUSEPORT for load balancing across processes
				opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if opErr != nil {
					return
				}

				// OPTIONAL: Additional socket optimizations
				// Enable TCP_NODELAY (disable Nagle's algorithm)
				opErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_NODELAY, 1)
				if opErr != nil {
					return
				}

				// Set TCP_QUICKACK for faster ACKs
				opErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_QUICKACK, 1)
			})
			if err != nil {
				return err
			}
			return opErr
		},
	}

	return lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", port))
}

// OPTIONAL: Custom access log middleware (only if needed for debugging)
/*
func accessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Only log slow requests or errors
		if latency > 100*time.Millisecond || statusCode >= 400 {
			log.Printf("[%s] %d %s %s %v",
				c.Request.Method,
				statusCode,
				path,
				query,
				latency,
			)
		}
	}
}
*/
