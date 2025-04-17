package env

import (
	"log"

	"github.com/joho/godotenv"
)

var (
	ApiEnv            string
	ApiPort           string
	DbUser            string
	DbPassword        string
	DbHost            string
	DbName            string
	DbPort            string
	DB_MAX_OPEN_CONNS int
	DB_MAX_IDLE_CONNS int
	DB_MAX_IDLE_TIME  string
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ApiEnv = getEnvWithDefault("API_ENV", "development")
	ApiPort = getEnvWithDefault("API_PORT", ":8080")

	DbUser = getEnvWithDefault("DB_USER", "root")
	DbPassword = getEnvWithDefault("DB_PASSWORD", "password")
	DbHost = getEnvWithDefault("DB_HOST", "localhost")
	DbName = getEnvWithDefault("DB_NAME", "gopher_social")
	DbPort = getEnvWithDefault("DB_PORT", "5432")

	DB_MAX_OPEN_CONNS = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	DB_MAX_IDLE_CONNS = getEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	DB_MAX_IDLE_TIME = getEnvWithDefault("DB_MAX_IDLE_TIME", "15m")
}
