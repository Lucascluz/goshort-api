package configs

import (
	"fmt"
	"os"
	"strconv"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv    string
	Port      int
	DBDSN     string
	JWTSecret string
}

func LoadConfig() (*Config, error) {
	// Load .env file in development
	if os.Getenv("APP_ENV") == "" || os.Getenv("APP_ENV") == "development" {
		err := godotenv.Load()
		if err != nil {
			return nil, fmt.Errorf("error loading .env file: %v", err)
		}
	}

	// Parse port to integer
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value: %v", err)
	}

	return &Config{
		AppEnv:    getEnv("APP_ENV", "development"),
		Port:      port,
		DBDSN:     getEnv("DB_DSN", "file:urls.db"),
		JWTSecret: getEnv("JWT_SECRET", ""),
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