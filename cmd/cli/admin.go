package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/config"
	"github.com/mzhryns/titik-nol-backend/internal/infrastructure/database"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/crypto"
	"gorm.io/gorm"
)

type adminFlags struct {
	action   string
	email    string
	password string
	name     string
}

func parseFlags() adminFlags {
	action := flag.String("action", "", "Action to perform: register | remove")
	email := flag.String("email", "", "Admin email")
	password := flag.String("password", "", "Admin password")
	name := flag.String("name", "System Administrator", "Admin name")

	flag.Parse()

	if *action == "" || *email == "" {
		fmt.Println("Usage: go run cmd/cli/admin.go -action [register|remove] -email <email> [-password <password>] [-name <name>]")
		os.Exit(1)
	}

	return adminFlags{
		action:   *action,
		email:    *email,
		password: *password,
		name:     *name,
	}
}

func initDB() *gorm.DB {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	return db
}

func registerAdmin(db *gorm.DB, f adminFlags) {
	if f.password == "" {
		log.Fatal("Password is required for registration")
	}

	hashedPassword, err := crypto.HashPassword(f.password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := domain.User{
		ID:         uuid.New(),
		Email:      f.email,
		Name:       f.name,
		Provider:   domain.ProviderLocal,
		ProviderID: "local-admin-" + f.email,
		Role:       domain.RoleAdmin,
		Password:   &hashedPassword,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Failed to register admin: %v", err)
	}
	fmt.Println("Administrator registered successfully!")
}

func removeAdmin(db *gorm.DB, email string) {
	if err := db.Where("email = ? AND role = ?", email, domain.RoleAdmin).Delete(&domain.User{}).Error; err != nil {
		log.Fatalf("Failed to remove admin: %v", err)
	}
	fmt.Println("Administrator removed successfully!")
}

func main() {
	f := parseFlags()
	db := initDB()

	switch f.action {
	case "register":
		registerAdmin(db, f)
	case "remove":
		removeAdmin(db, f.email)
	default:
		log.Fatalf("Unknown action: %s", f.action)
	}
}
