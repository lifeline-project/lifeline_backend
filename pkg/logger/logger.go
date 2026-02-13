package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// Config holds logger configuration
// Service and Env are included in every log entry
// Level is the minimum log level
// Dev enables development mode (pretty logs)
type Config struct {
	Service string
	Env     string
	Level   string
	Dev     bool
}

// Init initializes the global logger
func Init(cfg Config) error {
	var zapCfg zap.Config
	if cfg.Dev {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
		zapCfg.Encoding = "json"
	}

	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	// Set log level
	level := zapcore.InfoLevel
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// Add base fields
	zapCfg.InitialFields = map[string]interface{}{
		"service": cfg.Service,
		"env":     cfg.Env,
	}

	l, err := zapCfg.Build()
	if err != nil {
		return err
	}
	Logger = l
	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
