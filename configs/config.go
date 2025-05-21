package configs

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BASE_URL	string
	DBDSN         string
	JWTSecret     string
	SMTP_HOST     string
	SMTP_PORT     int
	SMTP_ACCOUNT  string
	SMTP_PASSWORD string
}

func LoadConfig() (*Config, error) {
	// Load .env file in development
	if os.Getenv("APP_ENV") == "" || os.Getenv("APP_ENV") == "development" {
		err := godotenv.Load()
		if err != nil {
			return nil, fmt.Errorf("error loading .env file: %v", err)
		}
	}

	// Parse SMTP_PORT
	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT value: %v", err)
	}

	return &Config{
		BASE_URL:      getEnv("BASE_URL", fmt.Sprintf("http://localhost:8080")),
		DBDSN:         getEnv("DB_DSN", "file:urls.db"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		SMTP_HOST:     getEnv("SMTP_HOST", ""),
		SMTP_PORT:     smtpPort,
		SMTP_ACCOUNT:  getEnv("SMTP_ACCOUNT", ""),
		SMTP_PASSWORD: getEnv("SMTP_PASSWORD", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == "" {
			panic(fmt.Sprintf("Environment variable %s is required", key))
		}
		return defaultValue
	}
	return value
}
