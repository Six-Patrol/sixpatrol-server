package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Tenant represents an organization using the service.
type Tenant struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name           string    `gorm:"size:255;not null;unique"`
	DashboardEmail string    `gorm:"size:255;not null"`
	PasswordHash   string    `gorm:"size:512;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// ApiKey stores API keys for a tenant.
type ApiKey struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TenantID        uuid.UUID `gorm:"type:uuid;index;not null"`
	ApiKeyString    string    `gorm:"size:512;not null;unique"`
	SecretKeyString string    `gorm:"size:512;not null"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

func (k *ApiKey) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}

// TenantConfig stores arbitrary tenant-specific configuration using JSONB.
// This supports dynamic placeholders and templating for alert thresholds, webhooks, etc.
type TenantConfig struct {
	ID        uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TenantID  uuid.UUID         `gorm:"type:uuid;index;not null"`
	Config    datatypes.JSONMap `gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *TenantConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TelemetryUsage tracks usage events for billing/analytics.
type TelemetryUsage struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TenantID        uuid.UUID `gorm:"type:uuid;index;not null"`
	FeatureUsed     string    `gorm:"size:255;not null"`
	FramesProcessed int64     `gorm:"not null;default:0"`
	Timestamp       time.Time `gorm:"not null;index"`
}

func (u *TelemetryUsage) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// PiracyDetection stores matches reported by async workers.
type PiracyDetection struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TenantID        uuid.UUID `gorm:"type:uuid;index;not null"`
	PirateURL       string    `gorm:"size:2048;not null"`
	ConfidenceScore float64   `gorm:"not null"`
	CreatedAt       time.Time
}

func (p *PiracyDetection) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// AutoMigrate runs GORM auto-migrations for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Tenant{}, &ApiKey{}, &TenantConfig{}, &TelemetryUsage{}, &PiracyDetection{})
}
