package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	JWTRefreshSecret  string
	Port              string
	Env               string
}

func Load() (*Config, error) {
	// .env varsa yükle (production'da gereksiz ama zarar vermez)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		RedisURL:         os.Getenv("REDIS_URL"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
		Port:             os.Getenv("PORT"),
		Env:              os.Getenv("ENV"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET required")
	}
	if cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("JWT_REFRESH_SECRET required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.Env == "" {
		cfg.Env = "production"
	}

	return cfg, nil
}
