package database

import (
	"context"
	"time"

	"lifeline_backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	ConnTimeoutSecs int
	AutoMigrate     bool
}

func Connect(cfg DBConfig) (*gorm.DB, error) {
	// Connect to Postgres with GORM
	// Use Silent mode — all application logging goes through zap
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Get underlying *sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxOpenConns / 2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnTimeoutSecs)*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, err
	}

	if cfg.AutoMigrate {
		// Ensure PostGIS is enabled
		_ = db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")

		// Auto-Migrate (Create Tables)
		if err := db.AutoMigrate(&models.User{}, &models.PharmacyProfile{}, &models.EmergencyRequest{}, &models.PharmacyResponse{}); err != nil {
			sqlDB.Close()
			return nil, err
		}

		// Seed Database with Vadodara coordinates pharmacies
		if err := SeedDatabase(db); err != nil {
			sqlDB.Close()
			return nil, err
		}
	}

	return db, nil
}

func SeedDatabase(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Where("email = ?", "pharma1@lifeline.com").Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	// Create pharmacy users
	pharmaEmails := []string{
		"pharma1@lifeline.com",
		"pharma2@lifeline.com",
		"pharma3@lifeline.com",
	}

	pharmaNames := []string{
		"City Care Vadodara",
		"HealWay Pharmacy Vadodara",
		"Metro Health Vadodara",
	}

	pharmaLocations := []struct {
		Lat     float64
		Lon     float64
		Address string
	}{
		{22.346083, 73.227611, "Sector 3, Main Vasna Road, Vadodara"},
		{22.336083, 73.217611, "Shop 14, Sun Plaza, OP Road, Vadodara"},
		{22.349083, 73.219611, "Near Nilamber Circle, Gotri, Vadodara"},
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashStr := string(hash)

	for i, email := range pharmaEmails {
		user := models.User{
			Name:         pharmaNames[i],
			Email:        email,
			PasswordHash: hashStr,
			Phone:        "9876543210",
			Role:         "PHARMACY",
			CreatedAt:    time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			return err
		}

		profile := models.PharmacyProfile{
			UserID:       user.ID,
			PharmacyName: pharmaNames[i],
			Address:      pharmaLocations[i].Address,
			Latitude:     pharmaLocations[i].Lat,
			Longitude:    pharmaLocations[i].Lon,
		}

		if err := db.Create(&profile).Error; err != nil {
			return err
		}
	}

	return nil
}
