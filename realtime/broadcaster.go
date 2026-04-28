package realtime

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

const defaultBufferSize = 16

// Broadcaster manages tenant-scoped subscriptions for SSE updates.
type Broadcaster struct {
	mu   sync.RWMutex
	subs map[uuid.UUID]map[chan string]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subs: make(map[uuid.UUID]map[chan string]struct{}),
	}
}

func (b *Broadcaster) Subscribe(tenantID uuid.UUID) (chan string, func()) {
	ch := make(chan string, defaultBufferSize)
	var once sync.Once

	b.mu.Lock()
	if _, ok := b.subs[tenantID]; !ok {
		b.subs[tenantID] = make(map[chan string]struct{})
	}
	b.subs[tenantID][ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			if subscribers, ok := b.subs[tenantID]; ok {
				delete(subscribers, ch)
				if len(subscribers) == 0 {
					delete(b.subs, tenantID)
				}
			}
			b.mu.Unlock()
			close(ch)
		})
	}

	return ch, unsubscribe
}

func (b *Broadcaster) Publish(tenantID uuid.UUID, payload string) error {
	b.mu.RLock()
	subscribers, ok := b.subs[tenantID]
	b.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no subscribers")
	}

	for ch := range subscribers {
		select {
		case ch <- payload:
		default:
		}
	}
	return nil
}

var defaultBroadcaster = NewBroadcaster()

// Subscribe registers a subscriber for a tenant.
func Subscribe(tenantID uuid.UUID) (chan string, func()) {
	return defaultBroadcaster.Subscribe(tenantID)
}

// Publish sends a payload to subscribers for a tenant.
func Publish(tenantID uuid.UUID, payload string) error {
	return defaultBroadcaster.Publish(tenantID, payload)
}

// SetBroadcaster swaps the default broadcaster and returns the previous one.
func SetBroadcaster(b *Broadcaster) *Broadcaster {
	if b == nil {
		b = NewBroadcaster()
	}
	previous := defaultBroadcaster
	defaultBroadcaster = b
	return previous
}
