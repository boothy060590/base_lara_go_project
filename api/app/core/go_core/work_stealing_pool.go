package go_core

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// WORK STEALING POOL IMPLEMENTATION
// ============================================================================

// WorkStealingConfig defines configuration for work stealing pools
type WorkStealingConfig struct {
	NumWorkers      int           `json:"num_workers"`
	QueueSize       int           `json:"queue_size"`
	StealThreshold  int           `json:"steal_threshold"`
	StealBatchSize  int           `json:"steal_batch_size"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	EnableMetrics   bool          `json:"enable_metrics"`
	EnableProfiling bool          `json:"enable_profiling"`
}

// DefaultWorkStealingConfig returns sensible defaults for work stealing
func DefaultWorkStealingConfig() *WorkStealingConfig {
	return &WorkStealingConfig{
		NumWorkers:      runtime.NumCPU(),
		QueueSize:       1024,
		StealThreshold:  2, // Steal when queue has 2 or fewer items
		StealBatchSize:  10,
		IdleTimeout:     100 * time.Millisecond,
		EnableMetrics:   true,
		EnableProfiling: true,
	}
}

// WorkStealingPool implements a work stealing thread pool for optimal CPU utilization
type WorkStealingPool[T any] struct {
	config      *WorkStealingConfig
	workers     []*WorkStealingWorker[T]
	globalQueue *WorkQueue[T]
	metrics     *WorkStealingMetrics
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

// WorkStealingWorker represents a worker in the work stealing pool
type WorkStealingWorker[T any] struct {
	id          int
	localQueue  *WorkQueue[T]
	globalQueue *WorkQueue[T]
	workers     []*WorkStealingWorker[T]
	config      *WorkStealingConfig
	ctx         context.Context
	cancel      context.CancelFunc
	active      int32
	steals      int64
	processed   int64
	mu          sync.RWMutex
}

// WorkQueue represents a thread-safe work queue
type WorkQueue[T any] struct {
	items    []WorkItem[T]
	head     int
	tail     int
	size     int
	mu       sync.Mutex
	notEmpty *sync.Cond
}

// WorkItem represents a work item in the queue
type WorkItem[T any] struct {
	ID       string
	Data     T
	Handler  func(context.Context, T) error
	Timeout  time.Duration
	Priority int
}

// WorkStealingMetrics tracks performance metrics for the work stealing pool
type WorkStealingMetrics struct {
	TotalProcessed   int64         `json:"total_processed"`
	TotalSteals      int64         `json:"total_steals"`
	ActiveWorkers    int           `json:"active_workers"`
	QueueUtilization float64       `json:"queue_utilization"`
	AverageWaitTime  time.Duration `json:"average_wait_time"`
	LastUpdated      time.Time     `json:"last_updated"`
	mu               sync.RWMutex
}

// NewWorkStealingPool creates a new work stealing pool
func NewWorkStealingPool[T any](config *WorkStealingConfig) *WorkStealingPool[T] {
	if config == nil {
		config = DefaultWorkStealingConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	wsp := &WorkStealingPool[T]{
		config:      config,
		globalQueue: NewWorkQueue[T](config.QueueSize),
		metrics:     NewWorkStealingMetrics(),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Create workers
	wsp.workers = make([]*WorkStealingWorker[T], config.NumWorkers)
	for i := 0; i < config.NumWorkers; i++ {
		worker := wsp.newWorker(i)
		wsp.workers[i] = worker
	}

	// Start workers
	for _, worker := range wsp.workers {
		go worker.start()
	}

	return wsp
}

// newWorker creates a new work stealing worker
func (wsp *WorkStealingPool[T]) newWorker(id int) *WorkStealingWorker[T] {
	ctx, cancel := context.WithCancel(wsp.ctx)

	return &WorkStealingWorker[T]{
		id:          id,
		localQueue:  NewWorkQueue[T](wsp.config.QueueSize),
		globalQueue: wsp.globalQueue,
		workers:     wsp.workers,
		config:      wsp.config,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Submit submits a work item to the pool
func (wsp *WorkStealingPool[T]) Submit(item WorkItem[T]) error {
	// Try to submit to a random worker's local queue first
	worker := wsp.workers[item.Priority%len(wsp.workers)]
	if worker.localQueue.TryPush(item) {
		return nil
	}

	// Fall back to global queue
	return wsp.globalQueue.Push(item)
}

// SubmitAsync submits a work item asynchronously
func (wsp *WorkStealingPool[T]) SubmitAsync(item WorkItem[T]) <-chan error {
	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		resultChan <- wsp.Submit(item)
	}()

	return resultChan
}

// Shutdown gracefully shuts down the work stealing pool
func (wsp *WorkStealingPool[T]) Shutdown() {
	wsp.cancel()

	// Wait for all workers to finish
	var wg sync.WaitGroup
	for _, worker := range wsp.workers {
		wg.Add(1)
		go func(w *WorkStealingWorker[T]) {
			defer wg.Done()
			w.shutdown()
		}(worker)
	}
	wg.Wait()
}

// GetMetrics returns the current metrics
func (wsp *WorkStealingPool[T]) GetMetrics() *WorkStealingMetrics {
	wsp.mu.RLock()
	defer wsp.mu.RUnlock()

	// Update metrics
	activeWorkers := 0
	totalProcessed := int64(0)
	totalSteals := int64(0)

	for _, worker := range wsp.workers {
		if atomic.LoadInt32(&worker.active) > 0 {
			activeWorkers++
		}
		totalProcessed += atomic.LoadInt64(&worker.processed)
		totalSteals += atomic.LoadInt64(&worker.steals)
	}

	wsp.metrics.UpdateMetrics(activeWorkers, totalProcessed, totalSteals)
	return wsp.metrics
}

// start begins the worker's main loop
func (w *WorkStealingWorker[T]) start() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.runWorkLoop()
		}
	}
}

// runWorkLoop is the main work processing loop with stealing
func (w *WorkStealingWorker[T]) runWorkLoop() {
	// Try to get work from local queue first
	if item, ok := w.localQueue.TryPop(); ok {
		atomic.StoreInt32(&w.active, 1)
		w.processWorkItem(item)
		atomic.AddInt64(&w.processed, 1)
		atomic.StoreInt32(&w.active, 0)
		return
	}

	// Try global queue
	if item, ok := w.globalQueue.TryPop(); ok {
		atomic.StoreInt32(&w.active, 1)
		w.processWorkItem(item)
		atomic.AddInt64(&w.processed, 1)
		atomic.StoreInt32(&w.active, 0)
		return
	}

	// Try to steal work from other workers
	if w.tryStealWork() {
		return
	}

	// No work available, sleep briefly
	time.Sleep(w.config.IdleTimeout)
}

// tryStealWork attempts to steal work from other workers
func (w *WorkStealingWorker[T]) tryStealWork() bool {
	// Try to steal from random workers
	for i := 0; i < len(w.workers); i++ {
		targetIndex := (w.id + i) % len(w.workers)
		target := w.workers[targetIndex]

		if target.id == w.id {
			continue
		}

		// Try to steal a batch of work
		stolen := target.localQueue.TrySteal(w.config.StealBatchSize)
		if len(stolen) > 0 {
			atomic.AddInt64(&w.steals, int64(len(stolen)))

			// Process stolen work
			for _, item := range stolen {
				atomic.StoreInt32(&w.active, 1)
				w.processWorkItem(item)
				atomic.AddInt64(&w.processed, 1)
				atomic.StoreInt32(&w.active, 0)
			}
			return true
		}
	}

	return false
}

// processWorkItem processes a single work item
func (w *WorkStealingWorker[T]) processWorkItem(item WorkItem[T]) {
	ctx := w.ctx
	if item.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(w.ctx, item.Timeout)
		defer cancel()
	}

	if item.Handler != nil {
		if err := item.Handler(ctx, item.Data); err != nil {
			// Log error or handle failure
			fmt.Printf("Work item %s failed: %v\n", item.ID, err)
		}
	}
}

// shutdown gracefully shuts down the worker
func (w *WorkStealingWorker[T]) shutdown() {
	w.cancel()
}

// NewWorkQueue creates a new work queue
func NewWorkQueue[T any](size int) *WorkQueue[T] {
	wq := &WorkQueue[T]{
		items: make([]WorkItem[T], size),
		size:  size,
	}
	wq.notEmpty = sync.NewCond(&wq.mu)
	return wq
}

// Push adds an item to the queue
func (wq *WorkQueue[T]) Push(item WorkItem[T]) error {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	if wq.isFull() {
		return fmt.Errorf("queue is full")
	}

	wq.items[wq.tail] = item
	wq.tail = (wq.tail + 1) % wq.size
	wq.notEmpty.Signal()
	return nil
}

// TryPush attempts to add an item to the queue without blocking
func (wq *WorkQueue[T]) TryPush(item WorkItem[T]) bool {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	if wq.isFull() {
		return false
	}

	wq.items[wq.tail] = item
	wq.tail = (wq.tail + 1) % wq.size
	wq.notEmpty.Signal()
	return true
}

// Pop removes and returns an item from the queue
func (wq *WorkQueue[T]) Pop() (WorkItem[T], bool) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	for wq.isEmpty() {
		wq.notEmpty.Wait()
	}

	item := wq.items[wq.head]
	wq.head = (wq.head + 1) % wq.size
	return item, true
}

// TryPop attempts to remove an item from the queue without blocking
func (wq *WorkQueue[T]) TryPop() (WorkItem[T], bool) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	if wq.isEmpty() {
		var zero WorkItem[T]
		return zero, false
	}

	item := wq.items[wq.head]
	wq.head = (wq.head + 1) % wq.size
	return item, true
}

// TrySteal steals a batch of items from the queue
func (wq *WorkQueue[T]) TrySteal(batchSize int) []WorkItem[T] {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	if wq.isEmpty() {
		return nil
	}

	stolen := make([]WorkItem[T], 0, batchSize)
	for i := 0; i < batchSize && !wq.isEmpty(); i++ {
		item := wq.items[wq.head]
		stolen = append(stolen, item)
		wq.head = (wq.head + 1) % wq.size
	}

	return stolen
}

// isEmpty checks if the queue is empty
func (wq *WorkQueue[T]) isEmpty() bool {
	return wq.head == wq.tail
}

// isFull checks if the queue is full
func (wq *WorkQueue[T]) isFull() bool {
	return (wq.tail+1)%wq.size == wq.head
}

// NewWorkStealingMetrics creates new work stealing metrics
func NewWorkStealingMetrics() *WorkStealingMetrics {
	return &WorkStealingMetrics{
		LastUpdated: time.Now(),
	}
}

// UpdateMetrics updates the metrics with current values
func (wsm *WorkStealingMetrics) UpdateMetrics(activeWorkers int, totalProcessed, totalSteals int64) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	wsm.ActiveWorkers = activeWorkers
	wsm.TotalProcessed = totalProcessed
	wsm.TotalSteals = totalSteals
	wsm.LastUpdated = time.Now()
}
