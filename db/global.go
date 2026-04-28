package db

import (
	"sync"

	"gorm.io/gorm"
)

var (
	globalDB   *gorm.DB
	globalDBMu sync.RWMutex
)

// SetDB stores the global database handle for packages that need DB access.
func SetDB(db *gorm.DB) {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()
	globalDB = db
}

// GetDB returns the global database handle if it has been set.
func GetDB() *gorm.DB {
	globalDBMu.RLock()
	defer globalDBMu.RUnlock()
	return globalDB
}
