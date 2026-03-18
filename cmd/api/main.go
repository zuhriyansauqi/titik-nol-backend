package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/delivery/http"
	"github.com/mzhryns/titik-nol-backend/internal/delivery/http/middleware"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/database"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/logger"
	"github.com/mzhryns/titik-nol-backend/internal/repository"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 1.1 Initialize Logger
	logger.Initialize(cfg)

	// 2. Initialize Database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// 3. Database Migrations
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get sql.DB", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(sqlDB); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// 4. Initialize Repository
	userRepo := repository.NewUserRepository(db)

	// 5. Initialize Usecase
	userUsecase := usecase.NewUserUsecase(userRepo)

	// 6. Initialize Gin Engine
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 7. Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "UP",
			"message": "Health is good",
		})
	})

	// 8. Setup Handlers
	http.NewUserHandler(r, userUsecase)

	// 9. Run Server
	slog.Info("Starting server", "port", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		slog.Error("Failed to run server", "error", err)
		os.Exit(1)
	}
}
