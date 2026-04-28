package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sixpatrol/sixpatrol-server/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	gormDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	if err := createTestSchema(gormDB); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	db.SetDB(gormDB)
	t.Cleanup(func() {
		db.SetDB(nil)
		_ = os.Remove(dbPath)
	})

	return gormDB
}

func createTestSchema(gormDB *gorm.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			api_key_string TEXT NOT NULL,
			secret_key_string TEXT NOT NULL,
			created_at TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS telemetry_usages (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			feature_used TEXT NOT NULL,
			frames_processed INTEGER NOT NULL,
			timestamp TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS piracy_detections (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			pirate_url TEXT NOT NULL,
			confidence_score REAL NOT NULL,
			created_at TIMESTAMP
		);`,
	}

	for _, statement := range statements {
		if err := gormDB.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
