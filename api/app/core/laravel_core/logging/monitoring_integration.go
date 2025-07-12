package logging_core

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	go_core "base_lara_go_project/app/core/go_core"
)

// LoggingMonitor provides monitoring and metrics for logging systems
type LoggingMonitor[T any] struct {
	handlers   map[string]go_core.LogHandler[T]
	metrics    *go_core.LoggingMetrics
	mu         sync.RWMutex
	monitoring int32
}

// NewLoggingMonitor creates a new logging monitor
func NewLoggingMonitor[T any]() *LoggingMonitor[T] {
	return &LoggingMonitor[T]{
		handlers: make(map[string]go_core.LogHandler[T]),
		metrics:  &go_core.LoggingMetrics{},
	}
}

// Start starts monitoring
func (m *LoggingMonitor[T]) Start(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&m.monitoring, 0, 1) {
		go m.monitorLoop(ctx)
		return nil
	}
	return nil
}

// Stop stops monitoring
func (m *LoggingMonitor[T]) Stop() error {
	atomic.StoreInt32(&m.monitoring, 0)
	return nil
}

// AddHandler adds a handler to monitor
func (m *LoggingMonitor[T]) AddHandler(name string, handler go_core.LogHandler[T]) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[name] = handler
}

// RemoveHandler removes a handler from monitoring
func (m *LoggingMonitor[T]) RemoveHandler(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.handlers, name)
}

// GetMetrics returns monitoring metrics
func (m *LoggingMonitor[T]) GetMetrics() *go_core.LoggingMetrics {
	return m.metrics
}

// monitorLoop runs the monitoring loop
func (m *LoggingMonitor[T]) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for atomic.LoadInt32(&m.monitoring) == 1 {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.updateMetrics()
		}
	}
}

// updateMetrics updates monitoring metrics
func (m *LoggingMonitor[T]) updateMetrics() {
	// Basic metrics update
	// This can be extended as needed
}
