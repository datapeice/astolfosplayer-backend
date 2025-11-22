package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	SecretKey   string
	SecurityKey string
	Port        string
}

func LoadAuthConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "auth.db"),
		SecretKey:   getEnv("SECRET_KEY", "dev-secret-key"),
		SecurityKey: getEnv("SECURITY_KEY", "dev-security-key"),
		Port:        getEnv("PORT", "50051"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
