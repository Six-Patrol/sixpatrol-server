package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sixpatrol/sixpatrol-server/realtime"
)

func TestBroadcasterPublish(t *testing.T) {
	broadcaster := realtime.NewBroadcaster()
	tenantID := uuid.New()

	ch, cancel := broadcaster.Subscribe(tenantID)
	defer cancel()

	if err := broadcaster.Publish(tenantID, "hello"); err != nil {
		t.Fatalf("expected publish to succeed: %v", err)
	}

	select {
	case msg := <-ch:
		if msg != "hello" {
			t.Fatalf("expected hello, got %s", msg)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected message")
	}

	cancel()
	if err := broadcaster.Publish(tenantID, "later"); err == nil {
		t.Fatalf("expected publish error after unsubscribe")
	}
}
