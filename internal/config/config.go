package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	JwtSecret   string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // silently ignore if .env not found

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		JwtSecret:   os.Getenv("JWT_SECRET"),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Default port
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
