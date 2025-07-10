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

// QueueLogHandler provides high-performance queue-based logging
type QueueLogHandler[T any] struct {
	queue   go_core.Queue[T]
	config  *config_core.ConfigFacade
	level   go_core.LogLevel
	metrics *go_core.LoggingMetrics
	mu      sync.RWMutex
}

// NewQueueLogHandler creates a new queue-based log handler
func NewQueueLogHandler[T any](queue go_core.Queue[T], config *config_core.ConfigFacade) *QueueLogHandler[T] {
	return &QueueLogHandler[T]{
		queue:   queue,
		config:  config,
		level:   go_core.LogLevelDebug,
		metrics: &go_core.LoggingMetrics{},
	}
}

// Handle implements LogHandler interface with queue integration
func (h *QueueLogHandler[T]) Handle(entry go_core.LogEntry[T]) error {
	start := time.Now()
	defer func() {
		h.recordMetrics("queue_handler", time.Since(start))
	}()

	// Create log job for queue processing
	logJob := &go_core.Job[T]{
		ID:         fmt.Sprintf("log_%d", time.Now().UnixNano()),
		Data:       entry.Context,
		Attempts:   0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	// Push to queue
	err := h.queue.Push(logJob)
	if err != nil {
		return fmt.Errorf("failed to queue log entry: %w", err)
	}

	return nil
}

// ShouldHandle determines if the handler should handle the given level
func (h *QueueLogHandler[T]) ShouldHandle(level go_core.LogLevel) bool {
	return level >= h.level
}

// GetLevel returns the handler's level
func (h *QueueLogHandler[T]) GetLevel() go_core.LogLevel {
	return h.level
}

// Close closes the queue handler
func (h *QueueLogHandler[T]) Close() error {
	// Queue doesn't have Close method, just return success
	return nil
}

// Flush flushes queued logs
func (h *QueueLogHandler[T]) Flush() error {
	// Queue doesn't have Flush method, just return success
	return nil
}

// getPriorityForLevel returns queue priority based on log level
func (h *QueueLogHandler[T]) getPriorityForLevel(level go_core.LogLevel) int {
	switch level {
	case go_core.LogLevelFatal:
		return 1 // Highest priority
	case go_core.LogLevelError:
		return 2
	case go_core.LogLevelWarning:
		return 3
	case go_core.LogLevelInfo:
		return 4
	case go_core.LogLevelDebug:
		return 5 // Lowest priority
	default:
		return 4
	}
}

// getHandlerID returns a unique handler identifier
func (h *QueueLogHandler[T]) getHandlerID() string {
	return fmt.Sprintf("queue_handler_%d", time.Now().UnixNano())
}

// recordMetrics records performance metrics
func (h *QueueLogHandler[T]) recordMetrics(operation string, duration time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	atomic.AddInt64(&h.metrics.EntriesLogged, 1)
	atomic.StoreInt64(&h.metrics.AverageLatency, int64(duration))
}

// GetMetrics returns handler metrics
func (h *QueueLogHandler[T]) GetMetrics() *go_core.LoggingMetrics {
	return h.metrics
}

// LogJob represents a log entry for queue processing
type LogJob[T any] struct {
	Entry     go_core.LogEntry[T] `json:"entry"`
	Timestamp time.Time           `json:"timestamp"`
	HandlerID string              `json:"handler_id"`
}

// LogQueueProcessor processes log jobs from the queue
type LogQueueProcessor[T any] struct {
	queue      go_core.Queue[T]
	handlers   map[string]go_core.LogHandler[T]
	config     *config_core.ConfigFacade
	metrics    *go_core.LoggingMetrics
	mu         sync.RWMutex
	processing int32
}

// NewLogQueueProcessor creates a new log queue processor
func NewLogQueueProcessor[T any](queue go_core.Queue[T], config *config_core.ConfigFacade) *LogQueueProcessor[T] {
	return &LogQueueProcessor[T]{
		queue:    queue,
		handlers: make(map[string]go_core.LogHandler[T]),
		config:   config,
		metrics:  &go_core.LoggingMetrics{},
	}
}

// Start starts processing log jobs from the queue
func (p *LogQueueProcessor[T]) Start(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&p.processing, 0, 1) {
		go p.processLoop(ctx)
		return nil
	}
	return fmt.Errorf("processor already running")
}

// Stop stops processing log jobs
func (h *LogQueueProcessor[T]) Stop() error {
	atomic.StoreInt32(&h.processing, 0)
	return nil
}

// AddHandler adds a log handler to the processor
func (p *LogQueueProcessor[T]) AddHandler(name string, handler go_core.LogHandler[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[name] = handler
}

// RemoveHandler removes a log handler from the processor
func (p *LogQueueProcessor[T]) RemoveHandler(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.handlers, name)
}

// processLoop continuously processes log jobs from the queue
func (p *LogQueueProcessor[T]) processLoop(ctx context.Context) {
	for atomic.LoadInt32(&p.processing) == 1 {
		select {
		case <-ctx.Done():
			return
		default:
			// Process next job with timeout
			jobCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := p.processNextJob(jobCtx)
			cancel()

			if err != nil {
				// Log error but continue processing
				atomic.AddInt64(&p.metrics.HandlerErrors, 1)
			}

			// Small delay to prevent busy waiting
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// processNextJob processes a single log job from the queue
func (p *LogQueueProcessor[T]) processNextJob(ctx context.Context) error {
	// Pop job from queue
	job, err := p.queue.Pop()
	if err != nil {
		return err
	}

	// Process the job
	return p.processLogJob(job)
}

// processLogJob processes a log job
func (p *LogQueueProcessor[T]) processLogJob(job *go_core.Job[T]) error {
	// Create log entry from job data
	logEntry := go_core.LogEntry[T]{
		Level:     go_core.LogLevelInfo, // Default level
		Message:   fmt.Sprintf("Queued log entry: %v", job.Data),
		Context:   job.Data,
		Timestamp: job.CreatedAt,
		TraceID:   job.ID,
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

// GetMetrics returns processor metrics
func (p *LogQueueProcessor[T]) GetMetrics() *go_core.LoggingMetrics {
	return p.metrics
}

// BatchLogProcessor processes log jobs in batches for better performance
type BatchLogProcessor[T any] struct {
	queue      go_core.Queue[T]
	handlers   map[string]go_core.LogHandler[T]
	config     *config_core.ConfigFacade
	batchSize  int
	batchTTL   time.Duration
	metrics    *go_core.LoggingMetrics
	mu         sync.RWMutex
	processing int32
}

// NewBatchLogProcessor creates a new batch log processor
func NewBatchLogProcessor[T any](queue go_core.Queue[T], config *config_core.ConfigFacade) *BatchLogProcessor[T] {
	return &BatchLogProcessor[T]{
		queue:     queue,
		handlers:  make(map[string]go_core.LogHandler[T]),
		config:    config,
		batchSize: 100,
		batchTTL:  5 * time.Second,
		metrics:   &go_core.LoggingMetrics{},
	}
}

// Start starts batch processing
func (p *BatchLogProcessor[T]) Start(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&p.processing, 0, 1) {
		go p.batchProcessLoop(ctx)
		return nil
	}
	return fmt.Errorf("processor already running")
}

// Stop stops batch processing
func (p *BatchLogProcessor[T]) Stop() error {
	atomic.StoreInt32(&p.processing, 0)
	return nil
}

// batchProcessLoop processes log jobs in batches
func (p *BatchLogProcessor[T]) batchProcessLoop(ctx context.Context) {
	ticker := time.NewTicker(p.batchTTL)
	defer ticker.Stop()

	var batch []*go_core.Job[T]

	for atomic.LoadInt32(&p.processing) == 1 {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if len(batch) > 0 {
				p.processBatch(batch)
				batch = batch[:0]
			}
		default:
			// Try to get more jobs for the batch
			if len(batch) < p.batchSize {
				if job, err := p.queue.Pop(); err == nil {
					batch = append(batch, job)
				}
			} else {
				// Process full batch
				p.processBatch(batch)
				batch = batch[:0]
			}

			time.Sleep(10 * time.Millisecond)
		}
	}

	// Process remaining batch
	if len(batch) > 0 {
		p.processBatch(batch)
	}
}

// processBatch processes a batch of log jobs
func (p *BatchLogProcessor[T]) processBatch(batch []*go_core.Job[T]) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, job := range batch {
		// Create log entry from job data
		logEntry := go_core.LogEntry[T]{
			Level:     go_core.LogLevelInfo, // Default level
			Message:   fmt.Sprintf("Queued log entry: %v", job.Data),
			Context:   job.Data,
			Timestamp: job.CreatedAt,
			TraceID:   job.ID,
			SpanID:    "",
		}

		for _, handler := range p.handlers {
			if handler.ShouldHandle(logEntry.Level) {
				err := handler.Handle(logEntry)
				if err != nil {
					atomic.AddInt64(&p.metrics.HandlerErrors, 1)
				}
			}
		}

		atomic.AddInt64(&p.metrics.EntriesLogged, 1)
	}
}

// GetMetrics returns processor metrics
func (p *BatchLogProcessor[T]) GetMetrics() *go_core.LoggingMetrics {
	return p.metrics
}
