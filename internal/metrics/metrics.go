package metrics

import (
	"sync"
	"time"
)

// Metrics tracks transformation statistics
type Metrics struct {
	mu                  sync.RWMutex
	MessagesReceived    int64
	MessagesTransformed int64
	MessagesFailed      int64
	MessagesPublished   int64
	TotalProcessingTime time.Duration
}

// New creates a new metrics instance
func New() *Metrics {
	return &Metrics{}
}

// IncrementReceived increments the received message counter
func (m *Metrics) IncrementReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesReceived++
}

// IncrementTransformed increments the transformed message counter
func (m *Metrics) IncrementTransformed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesTransformed++
}

// IncrementFailed increments the failed message counter
func (m *Metrics) IncrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesFailed++
}

// IncrementPublished increments the published message counter
func (m *Metrics) IncrementPublished() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesPublished++
}

// AddProcessingTime adds to the total processing time
func (m *Metrics) AddProcessingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalProcessingTime += duration
}

// GetSnapshot returns a thread-safe snapshot of metrics
func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgTime := time.Duration(0)
	if m.MessagesTransformed > 0 {
		avgTime = m.TotalProcessingTime / time.Duration(m.MessagesTransformed)
	}

	return map[string]interface{}{
		"received":     m.MessagesReceived,
		"transformed":  m.MessagesTransformed,
		"published":    m.MessagesPublished,
		"failed":       m.MessagesFailed,
		"avg_time":     avgTime,
		"total_time":   m.TotalProcessingTime,
	}
}
