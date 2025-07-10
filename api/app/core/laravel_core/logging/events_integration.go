package logging_core

import (
	"context"
	"fmt"
	"sync"
	"time"

	go_core "base_lara_go_project/app/core/go_core"
	config_core "base_lara_go_project/app/core/laravel_core/config"
	"sync/atomic"
)

// EventLogHandler provides high-performance event-based logging
type EventLogHandler[T any] struct {
	eventDispatcher go_core.EventDispatcher[T]
	config          *config_core.ConfigFacade
	level           go_core.LogLevel
	metrics         *go_core.LoggingMetrics
	mu              sync.RWMutex
}

// NewEventLogHandler creates a new event-based log handler
func NewEventLogHandler[T any](eventDispatcher go_core.EventDispatcher[T], config *config_core.ConfigFacade) *EventLogHandler[T] {
	return &EventLogHandler[T]{
		eventDispatcher: eventDispatcher,
		config:          config,
		level:           go_core.LogLevelDebug,
		metrics:         &go_core.LoggingMetrics{},
	}
}

// Handle implements LogHandler interface with event integration
func (h *EventLogHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	start := time.Now()
	defer func() {
		h.recordMetrics("event_handler", time.Since(start))
	}()

	// Create log event
	logEvent := &go_core.Event[T]{
		ID:        fmt.Sprintf("log_%d", time.Now().UnixNano()),
		Name:      h.getEventTypeForLevel(entry.Level),
		Data:      entry.Context,
		Timestamp: time.Now(),
		Source:    "logging_system",
	}

	// Dispatch event
	err := h.eventDispatcher.Dispatch(logEvent)
	if err != nil {
		return fmt.Errorf("failed to dispatch log event: %w", err)
	}

	return nil
}

// ShouldHandle determines if the handler should handle the given level
func (h *EventLogHandler[T]) ShouldHandle(level go_core.LogLevel) bool {
	return level >= h.level
}

// GetLevel returns the handler's level
func (h *EventLogHandler[T]) GetLevel() go_core.LogLevel {
	return h.level
}

// Close closes the event handler
func (h *EventLogHandler[T]) Close() error {
	// Event dispatcher doesn't have Close method, just return success
	return nil
}

// Flush flushes event logs
func (h *EventLogHandler[T]) Flush() error {
	// Event dispatcher doesn't have Flush method, just return success
	return nil
}

// getEventTypeForLevel returns event type based on log level
func (h *EventLogHandler[T]) getEventTypeForLevel(level go_core.LogLevel) string {
	switch level {
	case go_core.LogLevelFatal:
		return "log.fatal"
	case go_core.LogLevelError:
		return "log.error"
	case go_core.LogLevelWarning:
		return "log.warning"
	case go_core.LogLevelInfo:
		return "log.info"
	case go_core.LogLevelDebug:
		return "log.debug"
	default:
		return "log.info"
	}
}

// recordMetrics records performance metrics
func (h *EventLogHandler[T]) recordMetrics(operation string, duration time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	atomic.AddInt64(&h.metrics.EntriesLogged, 1)
	atomic.StoreInt64(&h.metrics.AverageLatency, int64(duration))
}

// GetMetrics returns handler metrics
func (h *EventLogHandler[T]) GetMetrics() *go_core.LoggingMetrics {
	return h.metrics
}

// LogEventListener listens for log events and processes them
type LogEventListener[T any] struct {
	eventDispatcher go_core.EventDispatcher[T]
	handlers        map[string]go_core.LogHandler[T]
	config          *config_core.ConfigFacade
	metrics         *go_core.LoggingMetrics
	mu              sync.RWMutex
	listening       int32
}

// NewLogEventListener creates a new log event listener
func NewLogEventListener[T any](eventDispatcher go_core.EventDispatcher[T], config *config_core.ConfigFacade) *LogEventListener[T] {
	return &LogEventListener[T]{
		eventDispatcher: eventDispatcher,
		handlers:        make(map[string]go_core.LogHandler[T]),
		config:          config,
		metrics:         &go_core.LoggingMetrics{},
	}
}

// Start starts listening for log events
func (l *LogEventListener[T]) Start(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&l.listening, 0, 1) {
		// Register event listeners
		l.registerEventListeners()
		return nil
	}
	return fmt.Errorf("listener already running")
}

// Stop stops listening for log events
func (l *LogEventListener[T]) Stop() error {
	atomic.StoreInt32(&l.listening, 0)
	return nil
}

// AddHandler adds a log handler to the listener
func (l *LogEventListener[T]) AddHandler(name string, handler go_core.LogHandler[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers[name] = handler
}

// RemoveHandler removes a log handler from the listener
func (l *LogEventListener[T]) RemoveHandler(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.handlers, name)
}

// registerEventListeners registers listeners for different log event types
func (l *LogEventListener[T]) registerEventListeners() {
	// Register listeners for each log level
	eventTypes := []string{"log.fatal", "log.error", "log.warning", "log.info", "log.debug"}

	for _, eventType := range eventTypes {
		l.eventDispatcher.Listen(eventType, func(ctx context.Context, event *go_core.Event[T]) error {
			return l.handleLogEvent(event)
		})
	}
}

// handleLogEvent handles a log event
func (l *LogEventListener[T]) handleLogEvent(event *go_core.Event[T]) error {
	// Create log entry from event data
	logEntry := go_core.LogEntry[T]{
		Level:     l.getLevelFromEventName(event.Name),
		Message:   fmt.Sprintf("Event log: %s", event.Name),
		Context:   event.Data,
		Timestamp: event.Timestamp,
		TraceID:   event.ID,
		SpanID:    "",
	}

	// Process with all handlers
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, handler := range l.handlers {
		if handler.ShouldHandle(logEntry.Level) {
			err := handler.Handle(logEntry)
			if err != nil {
				atomic.AddInt64(&l.metrics.HandlerErrors, 1)
				// Continue with other handlers
			}
		}
	}

	atomic.AddInt64(&l.metrics.EntriesLogged, 1)
	return nil
}

// getLevelFromEventName converts event name to log level
func (l *LogEventListener[T]) getLevelFromEventName(eventName string) go_core.LogLevel {
	switch eventName {
	case "log.fatal":
		return go_core.LogLevelFatal
	case "log.error":
		return go_core.LogLevelError
	case "log.warning":
		return go_core.LogLevelWarning
	case "log.info":
		return go_core.LogLevelInfo
	case "log.debug":
		return go_core.LogLevelDebug
	default:
		return go_core.LogLevelInfo
	}
}

// GetMetrics returns listener metrics
func (l *LogEventListener[T]) GetMetrics() *go_core.LoggingMetrics {
	return l.metrics
}

// LogEventProcessor processes log events with advanced features
type LogEventProcessor[T any] struct {
	eventDispatcher go_core.EventDispatcher[T]
	handlers        map[string]go_core.LogHandler[T]
	config          *config_core.ConfigFacade
	batchSize       int
	batchTTL        time.Duration
	metrics         *go_core.LoggingMetrics
	mu              sync.RWMutex
	processing      int32
}

// NewLogEventProcessor creates a new log event processor
func NewLogEventProcessor[T any](eventDispatcher go_core.EventDispatcher[T], config *config_core.ConfigFacade) *LogEventProcessor[T] {
	return &LogEventProcessor[T]{
		eventDispatcher: eventDispatcher,
		handlers:        make(map[string]go_core.LogHandler[T]),
		config:          config,
		batchSize:       100,
		batchTTL:        5 * time.Second,
		metrics:         &go_core.LoggingMetrics{},
	}
}

// Start starts event processing
func (p *LogEventProcessor[T]) Start(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&p.processing, 0, 1) {
		go p.processLoop(ctx)
		return nil
	}
	return fmt.Errorf("processor already running")
}

// Stop stops event processing
func (p *LogEventProcessor[T]) Stop() error {
	atomic.StoreInt32(&p.processing, 0)
	return nil
}

// AddHandler adds a log handler to the processor
func (p *LogEventProcessor[T]) AddHandler(name string, handler go_core.LogHandler[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[name] = handler
}

// RemoveHandler removes a log handler from the processor
func (p *LogEventProcessor[T]) RemoveHandler(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.handlers, name)
}

// processLoop continuously processes log events
func (p *LogEventProcessor[T]) processLoop(ctx context.Context) {
	// Register event listeners for processing
	p.registerEventListeners()

	for atomic.LoadInt32(&p.processing) == 1 {
		select {
		case <-ctx.Done():
			return
		default:
			// Small delay to prevent busy waiting
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// registerEventListeners registers listeners for different log event types
func (p *LogEventProcessor[T]) registerEventListeners() {
	// Register listeners for each log level
	eventTypes := []string{"log.fatal", "log.error", "log.warning", "log.info", "log.debug"}

	for _, eventType := range eventTypes {
		p.eventDispatcher.Listen(eventType, func(ctx context.Context, event *go_core.Event[T]) error {
			return p.handleLogEvent(event)
		})
	}
}

// handleLogEvent handles a log event
func (p *LogEventProcessor[T]) handleLogEvent(event *go_core.Event[T]) error {
	// Create log entry from event data
	logEntry := go_core.LogEntry[T]{
		Level:     p.getLevelFromEventName(event.Name),
		Message:   fmt.Sprintf("Event log: %s", event.Name),
		Context:   event.Data,
		Timestamp: event.Timestamp,
		TraceID:   event.ID,
		SpanID:    "",
	}

	// Process with all handlers
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, handler := range p.handlers {
		if handler.ShouldHandle(logEntry.Level) {
			err := handler.Handle(logEntry)
			if err != nil {
				atomic.AddInt64(&p.metrics.HandlerErrors, 1)
				// Continue with other handlers
			}
		}
	}

	atomic.AddInt64(&p.metrics.EntriesLogged, 1)
	return nil
}

// getLevelFromEventName converts event name to log level
func (p *LogEventProcessor[T]) getLevelFromEventName(eventName string) go_core.LogLevel {
	switch eventName {
	case "log.fatal":
		return go_core.LogLevelFatal
	case "log.error":
		return go_core.LogLevelError
	case "log.warning":
		return go_core.LogLevelWarning
	case "log.info":
		return go_core.LogLevelInfo
	case "log.debug":
		return go_core.LogLevelDebug
	default:
		return go_core.LogLevelInfo
	}
}

// GetMetrics returns processor metrics
func (p *LogEventProcessor[T]) GetMetrics() *go_core.LoggingMetrics {
	return p.metrics
}
