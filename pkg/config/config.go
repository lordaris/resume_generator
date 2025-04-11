package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all application configuration
type Config struct {
	Port      string
	DBUrl     string
	JWTSecret string
	CSRFKey   string
}

// Load loads configuration from environment variables with validation
func Load() (*Config, error) {
	// Try to load .env file, continue if not found
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	}

	config := &Config{
		Port:      os.Getenv("PORT"),
		DBUrl:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		CSRFKey:   os.Getenv("CSRF_KEY"),
	}

	// Validate configuration
	var missingVars []string

	if config.Port == "" {
		missingVars = append(missingVars, "PORT")
	}

	if config.DBUrl == "" {
		missingVars = append(missingVars, "DB_URL")
	}

	if config.JWTSecret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	if config.CSRFKey == "" {
		missingVars = append(missingVars, "CSRF_KEY")
	}

	if len(missingVars) > 0 {
		return nil, errors.New("missing required environment variables: " + strings.Join(missingVars, ", "))
	}

	return config, nil
}
