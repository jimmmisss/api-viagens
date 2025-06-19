package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APIPort            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecretKey       string
	JWTExpirationHours int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Log but don't fail, as env vars could be set by the system
		fmt.Println("Warning: .env file not found, reading from environment")
	}

	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "72"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION_HOURS: %w", err)
	}

	return &Config{
		APIPort:            getEnv("API_PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "user"),
		DBPassword:         getEnv("DB_PASSWORD", "password"),
		DBName:             getEnv("DB_NAME", "tripdb"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTSecretKey:       getEnv("JWT_SECRET_KEY", "default-secret"),
		JWTExpirationHours: jwtExp,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
