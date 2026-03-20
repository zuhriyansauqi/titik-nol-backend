package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	delivery "github.com/mzhryns/titik-nol-backend/internal/delivery/http"
	"github.com/mzhryns/titik-nol-backend/internal/delivery/http/middleware"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/database"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/logger"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/google"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/jwt"
	"github.com/mzhryns/titik-nol-backend/internal/repository"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
)

// @title           Titik Nol API
// @version         1.0
// @description     API Documentation for the Titik Nol Backend
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		// Note: logger not yet initialized, uses Go's default handler
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Initialize Logger (must be early, before any other slog calls)
	logger.Initialize(cfg)

	// 3. Initialize Database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// 4. Database Migrations
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get sql.DB", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(sqlDB, "file://migrations"); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// 5. Initialize Repository
	userRepo := repository.NewUserRepository(db)
	accRepo := repository.NewAccountRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	catRepo := repository.NewCategoryRepository(db)

	// 6. Initialize Services & Usecase
	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpirySeconds)
	googleSSO := google.NewGoogleSSOService(cfg.GoogleClientID)

	userUsecase := usecase.NewUserUsecase(userRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, googleSSO, jwtService)
	accountUsecase := usecase.NewAccountUsecase(accRepo, txRepo, db)
	transactionUsecase := usecase.NewTransactionUsecase(txRepo, accRepo, catRepo, db)
	onboardingUsecase := usecase.NewOnboardingUsecase(accRepo, txRepo, db)
	dashboardUsecase := usecase.NewDashboardUsecase(accRepo, txRepo, catRepo)
	categoryUsecase := usecase.NewCategoryUsecase(catRepo, db)

	// 7. Initialize Middleware
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// 8. Initialize Gin Engine
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	if cfg.CORSAllowOrigins != "" {
		corsConfig.AllowOrigins = []string{cfg.CORSAllowOrigins}
	} else {
		corsConfig.AllowAllOrigins = true
	}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Configure Rate Limiter
	r.Use(middleware.RateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst))

	// 9. API Documentation (Swagger + Scalar)
	r.StaticFile("/docs/swagger.json", "./docs/swagger.json")
	r.GET("/docs/api", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<!doctype html>
			<html>
			  <head>
			    <title>API Reference</title>
			    <meta charset="utf-8" />
			    <meta name="viewport" content="width=device-width, initial-scale=1" />
			  </head>
			  <body>
			    <script id="api-reference" data-url="/docs/swagger.json"></script>
			    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
			  </body>
			</html>
		`))
	})

	// 10. Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "UP",
			"message": "Health is good",
		})
	})

	// 10. Setup Handlers

	delivery.NewAuthHandler(r, authUsecase, authMiddleware)

	// API v1 routes (authenticated)
	v1 := r.Group("/api/v1")
	v1.Use(authMiddleware)
	delivery.NewUserHandler(v1, userUsecase)
	delivery.NewAccountHandler(v1, accountUsecase)
	delivery.NewTransactionHandler(v1, transactionUsecase)
	delivery.NewOnboardingHandler(v1, onboardingUsecase)
	delivery.NewDashboardHandler(v1, dashboardUsecase)
	delivery.NewCategoryHandler(v1, categoryUsecase)

	// 11. Graceful Shutdown
	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("Starting server", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to run server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}

