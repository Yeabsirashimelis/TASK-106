package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort             string
	DatabaseURL            string
	JWTSecret              string
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	DeviceFingerprintSalt  string
	BcryptCost             int
	MaxLoginAttempts       int
	LockoutDuration        time.Duration
	RateLimitRPS           float64
	RateLimitBurst         int
	StoragePath            string
}

func Load() *Config {
	return &Config{
		ServerPort:            getEnv("SERVER_PORT", ":8080"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://authuser:authpass@localhost:5432/authdb?sslmode=disable"),
		JWTSecret:             getEnv("JWT_SECRET", "default-secret-change-me-at-least-32-bytes!!"),
		AccessTokenTTL:        getDuration("ACCESS_TOKEN_TTL", 30*time.Minute),
		RefreshTokenTTL:       getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		DeviceFingerprintSalt: getEnv("DEVICE_FINGERPRINT_SALT", "default-device-salt"),
		BcryptCost:            getInt("BCRYPT_COST", 12),
		MaxLoginAttempts:      getInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:       getDuration("LOCKOUT_DURATION", 15*time.Minute),
		RateLimitRPS:          getFloat("RATE_LIMIT_RPS", 10.0),
		RateLimitBurst:        getInt("RATE_LIMIT_BURST", 20),
		StoragePath:           getEnv("STORAGE_PATH", "./storage"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
