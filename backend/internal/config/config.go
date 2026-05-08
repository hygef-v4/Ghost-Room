package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	AllowedOrigins   string
	Environment      string
	WebSocketTimeout int
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		AllowedOrigins:   getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		WebSocketTimeout: getEnvInt("WS_TIMEOUT", 60),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
