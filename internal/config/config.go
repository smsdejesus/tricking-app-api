package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration
type Config struct {
	// The connection string format: postgres://user:password@host:port/database?sslmode=disable
	DatabaseURL string

	Port string

	Environment string

	InternalAPIKey string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Database URL is required
	// Uncomment the following lines to require DATABASE_URL env var for Production

	env := getEnv("ENVIRONMENT", "dev")
	dbURL := ""
	var err error
	if env == "dev" {
		dbURL, err = getDevDBUrl()
	} else {
		dbURL, err = getEnvRequired("DATABASE_URL")
	}
	if err != nil {
		return nil, err
	}

	internalKey, err := getEnvRequired("INTERNAL_API_KEY")
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL:    dbURL,
		Port:           getEnv("PORT", "8080"), // Default to 8080 if not set
		Environment:    env,
		InternalAPIKey: internalKey,
	}, nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnv is a helper that returns a default if the env var is not set
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired returns an error if the env var is not set
func getEnvRequired(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		// fmt.Errorf creates a formatted error message
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}

func getDevDBUrl() (string, error) {
	dbURL := getEnv("POSTGRES_DSN", "")
	if dbURL == "" {
		// Fallback: build DSN from parts
		host := getEnv("POSTGRES_HOST", "localhost")
		dbPort := getEnv("POSTGRES_PORT", "5432")
		user := getEnv("POSTGRES_USER", "postgres")
		pass := getEnv("POSTGRES_PASSWORD", "")
		name := getEnv("POSTGRES_DB", "tricking")
		sslmode := getEnv("POSTGRES_SSLMODE", "disable")

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, pass, host, dbPort, name, sslmode,
		)
	}
	return dbURL, nil
}
