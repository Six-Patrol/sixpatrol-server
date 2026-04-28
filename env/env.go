package env

import (
	"os"

	"github.com/joho/godotenv"
)

// Load attempts to load a .env file if present. It returns any error from godotenv.Load,
// but the environment variables can still come from the real environment.
func Load() error {
	// Ignore error if file not present; return only real errors.
	_ = godotenv.Load()
	return nil
}

// Get returns the environment variable or the provided default.
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
