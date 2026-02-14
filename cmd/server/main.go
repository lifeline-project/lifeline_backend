package main

import (
	"fmt"
	"net/http"

	"lifeline_backend/internal/config"
	"lifeline_backend/internal/database"
	"lifeline_backend/pkg/logger"

	"go.uber.org/zap"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to LifeLine Backend!")
}

func main() {
	cfg := config.LoadConfig()

	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		panic("JWT_SECRET is required")
	}

	err := logger.Init(logger.Config{
		Service: cfg.LoggerService,
		Env:     cfg.LoggerEnv,
		Level:   cfg.LoggerLevel,
		Dev:     cfg.LoggerDev,
	})
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	db, err := database.ConnectPostgres(database.DBConfig{
		DSN:             cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		ConnTimeoutSecs: cfg.DBConnTimeoutSecs,
	})
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Logger.Info("Database connected successfully", zap.Int("max_conns", cfg.DBMaxOpenConns))
	defer db.Close()

	http.HandleFunc("/", homeHandler)

	logger.Logger.Info("Server starting", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		logger.Logger.Fatal("Server failed", zap.Error(err))
	}
}
