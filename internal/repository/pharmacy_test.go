package repository

import (
	"os"
	"path/filepath"
	"testing"

	"lifeline_backend/internal/database"
	"lifeline_backend/internal/models"

	"github.com/joho/godotenv"
)

func TestFindNearbyPharmacies(t *testing.T) {
	// Load config from configs/.env (which is located in the root)
	envPath, _ := filepath.Abs("../../configs/.env")
	_ = godotenv.Load(envPath)

	// Since we need to test against the real DB URL, we'll read it from env (fallback to empty)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping database integration test because DATABASE_URL is not set")
	}

	db, err := database.Connect(database.DBConfig{
		DSN:             dbURL,
		MaxOpenConns:    2,
		ConnTimeoutSecs: 5,
		AutoMigrate:     true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Run within a transaction to rollback modifications
	tx := db.Begin()
	defer tx.Rollback()

	repo := NewPharmacyRepository(tx)

	// 1. Create a patient / test users
	userPatient := models.User{Name: "Test Patient", Email: "testpatient@lifeline.com", Role: "PATIENT"}
	if err := tx.Create(&userPatient).Error; err != nil {
		t.Fatalf("Failed to create test patient user: %v", err)
	}

	userPharm1 := models.User{Name: "Close Pharmacy Owner", Email: "owner1@lifeline.com", Role: "PHARMACY"}
	if err := tx.Create(&userPharm1).Error; err != nil {
		t.Fatalf("Failed to create user pharm 1: %v", err)
	}

	userPharm2 := models.User{Name: "Far Pharmacy Owner", Email: "owner2@lifeline.com", Role: "PHARMACY"}
	if err := tx.Create(&userPharm2).Error; err != nil {
		t.Fatalf("Failed to create user pharm 2: %v", err)
	}

	// 2. Create Pharmacy Profiles with locations
	// Patient coordinates: (13.0827, 80.2707) - e.g. Chennai, India
	patientLat := 13.0827
	patientLon := 80.2707

	// Pharmacy 1: ~1.2 km away
	pharm1 := models.PharmacyProfile{
		UserID:       userPharm1.ID,
		PharmacyName: "Close Care Pharmacy",
		Address:      "123 Close St",
		Latitude:     13.0880,
		Longitude:    80.2790,
	}
	if err := tx.Create(&pharm1).Error; err != nil {
		t.Fatalf("Failed to create profile 1: %v", err)
	}

	// Pharmacy 2: ~11 km away
	pharm2 := models.PharmacyProfile{
		UserID:       userPharm2.ID,
		PharmacyName: "Far Way Pharmacy",
		Address:      "789 Distance Rd",
		Latitude:     13.1800,
		Longitude:    80.3100,
	}
	if err := tx.Create(&pharm2).Error; err != nil {
		t.Fatalf("Failed to create profile 2: %v", err)
	}

	// Test 1: Find nearby within 3000 meters (3km)
	nearby, err := repo.FindNearby(patientLat, patientLon, 3000.0)
	if err != nil {
		t.Fatalf("Failed to find nearby: %v", err)
	}

	if len(nearby) != 1 {
		t.Errorf("Expected 1 nearby pharmacy, got %d", len(nearby))
	} else if nearby[0].UserID != userPharm1.ID {
		t.Errorf("Expected pharmacy ID %d, got %d", userPharm1.ID, nearby[0].UserID)
	}

	// Test 2: Find nearby within 15000 meters (15km)
	nearbyAll, err := repo.FindNearby(patientLat, patientLon, 15000.0)
	if err != nil {
		t.Fatalf("Failed to find nearby: %v", err)
	}

	if len(nearbyAll) != 2 {
		t.Errorf("Expected 2 nearby pharmacies within 15km, got %d", len(nearbyAll))
	}
}
