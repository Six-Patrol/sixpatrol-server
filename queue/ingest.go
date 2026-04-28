package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const defaultProxyVideoQueueSize = 100

// ProxyVideoMessage describes a video chunk that needs async processing.
type ProxyVideoMessage struct {
	TenantID  uuid.UUID
	FilePath  string
	StreamID  string
	Timestamp time.Time
}

var (
	proxyVideoQueue   chan ProxyVideoMessage
	proxyVideoQueueMu sync.RWMutex
)

func init() {
	proxyVideoQueue = make(chan ProxyVideoMessage, defaultProxyVideoQueueSize)
}

// SetProxyVideoQueue replaces the proxy video queue and returns the previous queue.
func SetProxyVideoQueue(ch chan ProxyVideoMessage) chan ProxyVideoMessage {
	if ch == nil {
		ch = make(chan ProxyVideoMessage, defaultProxyVideoQueueSize)
	}

	proxyVideoQueueMu.Lock()
	defer proxyVideoQueueMu.Unlock()
	previous := proxyVideoQueue
	proxyVideoQueue = ch
	return previous
}

// GetProxyVideoQueue returns the current proxy video queue.
func GetProxyVideoQueue() <-chan ProxyVideoMessage {
	proxyVideoQueueMu.RLock()
	defer proxyVideoQueueMu.RUnlock()
	return proxyVideoQueue
}

// PublishProxyVideo enqueues a proxy video message without blocking.
func PublishProxyVideo(msg ProxyVideoMessage) error {
	proxyVideoQueueMu.RLock()
	queue := proxyVideoQueue
	proxyVideoQueueMu.RUnlock()

	if queue == nil {
		return fmt.Errorf("proxy video queue not initialized")
	}

	select {
	case queue <- msg:
		return nil
	default:
		return fmt.Errorf("proxy video queue is full")
	}
}
