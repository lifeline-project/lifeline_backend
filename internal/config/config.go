package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	DBMaxOpenConns        int
	DBConnTimeoutSecs     int
	JWTSecret             string
	CloudflareR2Endpoint  string
	CloudflareR2AccessKey string
	CloudflareR2SecretKey string
	CloudflareR2Bucket    string
	LoggerService         string
	LoggerEnv             string
	LoggerLevel           string
	LoggerDev             bool
}

func LoadConfig() *Config {
	// Load .env file from configs directory
	_ = godotenv.Load("configs/.env")

	return &Config{
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		DBMaxOpenConns:        getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBConnTimeoutSecs:     getEnvInt("DB_CONN_TIMEOUT_SECS", 10),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		CloudflareR2Endpoint:  getEnv("CLOUDFLARE_R2_ENDPOINT", ""),
		CloudflareR2AccessKey: getEnv("CLOUDFLARE_R2_ACCESS_KEY_ID", ""),
		CloudflareR2SecretKey: getEnv("CLOUDFLARE_R2_SECRET_ACCESS_KEY", ""),
		CloudflareR2Bucket:    getEnv("CLOUDFLARE_R2_BUCKET", ""),
		LoggerService:         getEnv("LOGGER_SERVICE", "lifeline-backend"),
		LoggerEnv:             getEnv("LOGGER_ENV", "dev"),
		LoggerLevel:           getEnv("LOGGER_LEVEL", "info"),
		LoggerDev:             getEnvBool("LOGGER_DEV", true),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if value == "true" || value == "1" {
			return true
		}
		if value == "false" || value == "0" {
			return false
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if num, err := strconv.Atoi(value); err == nil {
			return num
		}
	}
	return fallback
}
