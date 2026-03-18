package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/delivery/http"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/database"
	"github.com/mzhryns/titik-nol-backend/internal/repository"
	"github.com/mzhryns/titik-nol-backend/internal/usecase"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 3. Database Migrations
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	if err := database.RunMigrations(sqlDB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 4. Initialize Repository
	userRepo := repository.NewUserRepository(db)

	// 5. Initialize Usecase
	userUsecase := usecase.NewUserUsecase(userRepo)

	// 6. Initialize Gin Engine
	r := gin.Default()

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
	log.Printf("Starting server on port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
