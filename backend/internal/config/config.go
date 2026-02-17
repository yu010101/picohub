package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DBPath         string
	JWTSecret      string
	JWTExpiry      time.Duration
	UploadDir      string
	MaxUploadSize  int64
	AllowedOrigins []string
	BcryptCost     int
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "picohub.db"),
		JWTSecret:      getEnv("JWT_SECRET", "picohub-dev-secret-change-in-production"),
		JWTExpiry:      getDurationEnv("JWT_EXPIRY", 24*time.Hour),
		UploadDir:      getEnv("UPLOAD_DIR", "uploads"),
		MaxUploadSize:  getInt64Env("MAX_UPLOAD_SIZE", 10<<20), // 10MB
		AllowedOrigins: []string{getEnv("CORS_ORIGIN", "http://localhost:3000")},
		BcryptCost:     getIntEnv("BCRYPT_COST", 12),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getInt64Env(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
