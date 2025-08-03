package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        int
	LogLevel    string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/civilregistry?sslmode=disable"),
		Port:        getEnvAsInt("PORT", 8080),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

// LoadTest loads test-specific configuration
func LoadTest() *Config {
	// Load .env.test file if it exists
	if err := godotenv.Load(".env.test"); err != nil {
		log.Println("No .env.test file found, trying .env file")
		// Fallback to regular .env file
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables or defaults")
		}
	}

	return &Config{
		DatabaseURL: getEnv("TEST_DATABASE_URL", getEnv("DATABASE_URL", "postgres://postgres:root@localhost:5432/civilregistrytest?sslmode=disable")),
		Port:        getEnvAsInt("PORT", 8080),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
