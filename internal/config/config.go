package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Application
	AppName    string
	AppVersion string

	// HTTP Server
	HTTPPort string
	HTTPHost string

	// Database
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBConnMaxIdleTime int
	DBConnMaxLifetime int
	DBMaxOpenConns    int
	DBMaxIdleConns    int
}

func Load() *Config {
	return &Config{
		AppName:    getEnv("APP_NAME", "chat-service"),
		AppVersion: getEnv("APP_VERSION", "0.1.0"),

		// HTTP
		HTTPPort: getEnv("PORT", "8080"),
		HTTPHost: getEnv("HOST", "0.0.0.0"),

		// Database
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "chats"),
		DBSSLMode:         getEnv("DB_SSL_MODE", "disable"),
		DBConnMaxIdleTime: getIntEnv("DB_CONN_MAX_IDLE_TIME", 300),
		DBConnMaxLifetime: getIntEnv("DB_CONN_MAX_LIFETIME", 3600),
		DBMaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil && v > 0 {
			return v
		}
	}
	return defaultValue
}
