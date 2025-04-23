package env

import (
	"log"
	"os"
	"strconv"
	"time"
)

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.ParseBool(value); err == nil {
			return v
		}
	}
	return fallback
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	value := getEnvWithDefault(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Error parsing duration for %s, using default: %v", key, err)
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}
