package config

import "os"

type Config struct {
	SupabaseURL          string
	SupabaseAnonKey      string
	SupabaseServiceRole  string
	SupabaseJWTSecret    string
	Port                 string
}

func Load() *Config {
	return &Config{
		SupabaseURL:         getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:     getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceRole: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseJWTSecret:   getEnv("SUPABASE_JWT_SECRET", ""),
		Port:                getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
