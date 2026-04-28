package tests

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/db"
	"gorm.io/datatypes"
)

func TestRenderConfigTemplate(t *testing.T) {
	cfg := datatypes.JSONMap{
		"alerts": map[string]any{
			"latency_ms": "{{ .LatencyMs }}",
			"webhook":    "https://hooks.example.com/{{ .TenantSlug }}",
		},
		"enabled": true,
	}

	vars := map[string]any{
		"LatencyMs":  250,
		"TenantSlug": "acme-co",
	}

	rendered, err := db.RenderConfigTemplate(cfg, vars)
	if err != nil {
		t.Fatalf("RenderConfigTemplate error: %v", err)
	}

	alerts, ok := rendered["alerts"].(map[string]any)
	if !ok {
		t.Fatalf("alerts should be an object")
	}

	if got := alerts["latency_ms"]; got != "250" {
		t.Fatalf("latency_ms mismatch: got %v", got)
	}

	if got := alerts["webhook"]; got != "https://hooks.example.com/acme-co" {
		t.Fatalf("webhook mismatch: got %v", got)
	}
}

func ExampleRenderConfigTemplate() {
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenantConfig := db.TenantConfig{
		TenantID: tenantID,
		Config: datatypes.JSONMap{
			"alerts": map[string]any{
				"threshold": "{{ .Threshold }}",
			},
			"webhook_url": "https://hooks.example.com/tenant/{{ .TenantID }}",
		},
	}

	vars := map[string]any{
		"Threshold": "0.85",
		"TenantID":  tenantID.String(),
	}

	rendered, _ := db.RenderConfigTemplate(tenantConfig.Config, vars)
	webhookURL := rendered["webhook_url"].(string)
	alerts := rendered["alerts"].(map[string]any)
	threshold := alerts["threshold"].(string)

	fmt.Println(webhookURL)
	fmt.Println(threshold)
	// Output:
	// https://hooks.example.com/tenant/11111111-1111-1111-1111-111111111111
	// 0.85
}
