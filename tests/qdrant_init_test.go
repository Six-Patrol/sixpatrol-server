package tests

import (
	"testing"

	"github.com/sixpatrol/sixpatrol-server/db"
)

func TestInitQdrantClient(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		t.Setenv("QDRANT_HOST", "localhost")
		t.Setenv("QDRANT_PORT", "6334")
		if _, err := db.InitQdrantClient(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		t.Setenv("QDRANT_PORT", "not-a-number")
		if _, err := db.InitQdrantClient(); err == nil {
			t.Fatalf("expected error for invalid port")
		}
	})
}
