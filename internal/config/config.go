package config

import (
	"os"
)

type Config struct {
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	CloudflareR2Endpoint   string
	CloudflareR2AccessKey  string
	CloudflareR2SecretKey  string
	CloudflareR2Bucket     string
}

func LoadConfig() *Config {
	return &Config{
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		CloudflareR2Endpoint:  getEnv("CLOUDFLARE_R2_ENDPOINT", ""),
		CloudflareR2AccessKey: getEnv("CLOUDFLARE_R2_ACCESS_KEY_ID", ""),
		CloudflareR2SecretKey: getEnv("CLOUDFLARE_R2_SECRET_ACCESS_KEY", ""),
		CloudflareR2Bucket:    getEnv("CLOUDFLARE_R2_BUCKET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
