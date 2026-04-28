package db

import (
	"fmt"
	"os"
	"strconv"

	"github.com/qdrant/go-client/qdrant"
)

// InitQdrantClient initializes an official Qdrant client using environment
// variables: QDRANT_HOST (default localhost), QDRANT_PORT (default 6334),
// QDRANT_API_KEY (optional), and QDRANT_USE_TLS (optional, true/false).
func InitQdrantClient() (*qdrant.Client, error) {
	host := os.Getenv("QDRANT_HOST")
	if host == "" {
		host = "localhost"
	}
	portStr := os.Getenv("QDRANT_PORT")
	if portStr == "" {
		portStr = "6334"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid QDRANT_PORT: %w", err)
	}

	apiKey := os.Getenv("QDRANT_API_KEY")
	useTLS := false
	if v := os.Getenv("QDRANT_USE_TLS"); v != "" {
		if parsed, perr := strconv.ParseBool(v); perr == nil {
			useTLS = parsed
		}
	}

	cfg := &qdrant.Config{
		Host:   host,
		Port:   port,
		APIKey: apiKey,
		UseTLS: useTLS,
	}

	client, err := qdrant.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	return client, nil
}
