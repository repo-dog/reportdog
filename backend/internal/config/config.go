package config

import (
	"fmt"
	"os"
)

// Config holds application configuration.
type Config struct {
	Port                string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	CORSAllowOrigin     string
	MaxUploadSize       int64
	DisableManualUpload bool
	AutoMigrate         bool
	MigrationsDir       string
}

// Load reads config from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "reportdog"),
		DBPassword:          getEnv("DB_PASSWORD", "reportdog"),
		DBName:              getEnv("DB_NAME", "reportdog"),
		CORSAllowOrigin:     getEnv("CORS_ALLOW_ORIGIN", "http://localhost:3000"),
		MaxUploadSize:       50 << 20, // 50 MB
		DisableManualUpload: getEnv("DISABLE_MANUAL_UPLOAD", "") == "true",
		AutoMigrate:         getEnv("AUTO_MIGRATE", "true") == "true",
		MigrationsDir:       getEnv("MIGRATIONS_DIR", "migrations"),
	}
}

// DSN returns a PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
