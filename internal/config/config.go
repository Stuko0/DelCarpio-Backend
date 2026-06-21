package config

import "os"

type Config struct {
	DatabaseURL       string
	DatabaseDirectURL string
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseJWTSecret string
	Port              string
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		DatabaseDirectURL: getEnv("DATABASE_DIRECT_URL", ""),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:   getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
		Port:              getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
