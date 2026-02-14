package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	ConnTimeoutSecs int
}

func ConnectPostgres(cfg DBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxOpenConns / 2)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnTimeoutSecs)*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
