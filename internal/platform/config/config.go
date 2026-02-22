package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port	 string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:   os.Getenv("PORT"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),
		DBName: os.Getenv("DB_NAME"),
	}
}

func getEnv(key, defaultback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultback
}