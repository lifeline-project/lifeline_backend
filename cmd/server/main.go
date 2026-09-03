package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lifeline_backend/internal/app"
	"lifeline_backend/internal/config"
	"lifeline_backend/internal/database"
	"lifeline_backend/internal/middleware"
	"lifeline_backend/internal/routes"
	wshub "lifeline_backend/internal/websocket"
	"lifeline_backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		panic("Invalid configuration: " + err.Error())
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

	db, err := database.Connect(database.DBConfig{
		DSN:             cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		ConnTimeoutSecs: cfg.DBConnTimeoutSecs,
		AutoMigrate:     cfg.DBAutoMigrate,
	})
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Logger.Info("Database connected successfully", zap.Int("max_conns", cfg.DBMaxOpenConns))
	// Get underlying *sql.DB for cleanup
	sqlDB, err := db.DB()
	if err != nil {
		logger.Logger.Fatal("Failed to get database instance", zap.Error(err))
	}

	if cfg.LoggerDev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger(logger.Logger))

	// Initialize and run real-time broadcast websocket hub
	hub := wshub.NewHub()
	go hub.Run()

	// Build dependency container and register ALL routes (system + versioned API)
	application := &app.App{
		DB:     db,
		Logger: logger.Logger,
		Router: router,
		Config: cfg,
		Hub:    hub,
	}
	routes.Register(application)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Logger.Info("Server starting", zap.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Logger.Info("Shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Logger.Error("Server shutdown failed", zap.Error(err))
	}

	if err := sqlDB.Close(); err != nil {
		logger.Logger.Error("Database close failed", zap.Error(err))
	}
}
