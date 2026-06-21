package config

import "os"

type Config struct {
	Debug       bool
	Port        string
	AutoMigrate bool
	CORSOrigins []string
	PostgresURL string
}

func Load() *Config {
	return &Config{
		Debug:       getEnv("PB_DEBUG", "false") == "true",
		Port:        getEnv("PORT", "8090"),
		AutoMigrate: getEnv("PB_AUTO_MIGRATE", "true") == "true",
		CORSOrigins: []string{getEnv("PB_ALLOWED_ORIGINS", "http://localhost:4321,http://localhost:5173")},
		PostgresURL: getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/delcarpio?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
