package database


import (
	"fmt"
	"log"
	"os"

	"github.com/meetsuhagiya/lifeline-backend/internal/models"// <--- MAKE SURE THIS MATCHES YOUR go.mod NAME

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	// 1. Load .env file
	// We look for .env in the root directory
	err := godotenv.Load() 
	if err != nil {
		log.Println("Warning: Error loading .env file, checking system environment variables")
	}

	// 2. Build Connection String
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("SSL_MODE"),
	)

	// 3. Connect to Postgres
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Show SQL in terminal
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully")
	
	// 4. Auto-Migrate (Create Tables)
	// This magically creates the tables in DB based on your Structs
	log.Println("Running Migrations...")
	db.AutoMigrate(&models.User{}, &models.PharmacyProfile{}, &models.EmergencyRequest{})

	DB = db
}