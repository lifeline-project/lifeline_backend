package database

import (
	"context"
	"time"

	"lifeline_backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	ConnTimeoutSecs int
	AutoMigrate     bool
}

func Connect(cfg DBConfig) (*gorm.DB, error) {
	// Connect to Postgres with GORM
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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
		// Auto-Migrate (Create Tables)
		if err := db.AutoMigrate(&models.User{}, &models.PharmacyProfile{}, &models.EmergencyRequest{}); err != nil {
			sqlDB.Close()
			return nil, err
		}
	}

	return db, nil
}
