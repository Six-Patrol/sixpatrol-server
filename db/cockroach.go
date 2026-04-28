package db

import (
	"context"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitCockroachDB initializes a GORM DB connected to a CockroachDB
// (Postgres-compatible). It reads the DSN from COCKROACH_DSN.
func InitCockroachDB(ctx context.Context) (*gorm.DB, error) {
	dsn := os.Getenv("COCKROACH_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("environment variable COCKROACH_DSN is required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm postgres connection: %w", err)
	}

	// Optionally verify connection by getting generic database object and pinging.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql DB from gorm: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping cockroachdb: %w", err)
	}

	return db, nil
}
