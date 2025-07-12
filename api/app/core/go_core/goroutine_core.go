package go_core

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// GoroutineConfig defines configuration for goroutine management
type GoroutineConfig struct {
	MaxWorkers        int           `json:"max_workers"`
	WorkerTimeout     time.Duration `json:"worker_timeout"`
	QueueBufferSize   int           `json:"queue_buffer_size"`
	EnableAutoScaling bool          `json:"enable_auto_scaling"`
	MinWorkers        int           `json:"min_workers"`
	MaxWorkersPerCPU  int           `json:"max_workers_per_cpu"`
}

// DefaultGoroutineConfig returns sensible defaults for goroutine management
func DefaultGoroutineConfig() *GoroutineConfig {
	return &GoroutineConfig{
		MaxWorkers:        runtime.NumCPU() * 2,
		WorkerTimeout:     30 * time.Second,
		QueueBufferSize:   1000,
		EnableAutoScaling: true,
		MinWorkers:        2,
		MaxWorkersPerCPU:  4,
	}
}

// GoroutineManager manages goroutine pools and automatic scaling
type GoroutineManager[T any] struct {
	config     *GoroutineConfig
	workerPool *WorkerPool[T]
	metrics    *GoroutineMetrics
	mu         sync.RWMutex
}

// NewGoroutineManager creates a new goroutine manager
func NewGoroutineManager[T any](config *GoroutineConfig) *GoroutineManager[T] {
	if config == nil {
		config = DefaultGoroutineConfig()
	}

	gm := &GoroutineManager[T]{
		config:  config,
		metrics: NewGoroutineMetrics(),
	}

	gm.workerPool = NewWorkerPool[T](config.MaxWorkers, config.QueueBufferSize, gm.metrics)

	if config.EnableAutoScaling {
		go gm.autoScale()
	}

	return gm
}

// GoroutineJob represents a job for the goroutine worker pool
// It embeds Job[T] and adds Timeout and Handler fields
// Handler is a function that processes the job

type GoroutineJob[T any] struct {
	Job     Job[T]
	Timeout time.Duration
	Handler func(ctx context.Context, job *GoroutineJob[T]) error
}

// WorkerPool manages a pool of worker goroutines
type WorkerPool[T any] struct {
	maxWorkers      int
	queueBufferSize int
	jobQueue        chan GoroutineJob[T]
	workers         []*Worker[T]
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	metrics         *GoroutineMetrics
}

// Worker represents a single worker goroutine
type Worker[T any] struct {
	id       int
	jobQueue chan GoroutineJob[T]
	ctx      context.Context
	cancel   context.CancelFunc
	active   bool
	mu       sync.RWMutex
	metrics  *GoroutineMetrics
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool[T any](maxWorkers, queueBufferSize int, metrics *GoroutineMetrics) *WorkerPool[T] {
	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool[T]{
		maxWorkers:      maxWorkers,
		queueBufferSize: queueBufferSize,
		jobQueue:        make(chan GoroutineJob[T], queueBufferSize),
		workers:         make([]*Worker[T], 0, maxWorkers),
		ctx:             ctx,
		cancel:          cancel,
		metrics:         metrics,
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		worker := wp.newWorker(i)
		wp.workers = append(wp.workers, worker)
		go worker.start()
	}

	return wp
}

// newWorker creates a new worker
func (wp *WorkerPool[T]) newWorker(id int) *Worker[T] {
	ctx, cancel := context.WithCancel(wp.ctx)
	return &Worker[T]{
		id:       id,
		jobQueue: wp.jobQueue,
		ctx:      ctx,
		cancel:   cancel,
		active:   true,
		metrics:  wp.metrics,
	}
}

// start begins the worker's job processing loop
func (w *Worker[T]) start() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case job := <-w.jobQueue:
			w.processJob(job)
		}
	}
}

// processJob processes a single job
func (w *Worker[T]) processJob(job GoroutineJob[T]) {
	w.mu.Lock()
	w.active = true
	w.mu.Unlock()

	start := time.Now()
	defer func() {
		w.mu.Lock()
		w.active = false
		w.mu.Unlock()

		// Update metrics after job is processed
		if w.metrics != nil {
			processingTime := time.Since(start)
			activeWorkers := 1 // We'll use 1 for this worker; for more accuracy, could aggregate from pool
			queueLength := len(w.jobQueue)
			w.metrics.UpdateMetrics(activeWorkers, queueLength, processingTime)
		}
	}()

	// Create context with timeout if specified
	ctx := w.ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(w.ctx, job.Timeout)
		defer cancel()
	}

	// Execute job handler
	if job.Handler != nil {
		if err := job.Handler(ctx, &job); err != nil {
			// Log error or handle failure
			fmt.Printf("Job %+v failed: %v\n", job.Job.ID, err)
		}
	}
}

// Submit submits a job to the worker pool
func (wp *WorkerPool[T]) Submit(job GoroutineJob[T]) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool[T]) Shutdown() {
	wp.cancel()
	close(wp.jobQueue)
}

// GetActiveWorkerCount returns the number of active workers
func (wp *WorkerPool[T]) GetActiveWorkerCount() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	active := 0
	for _, worker := range wp.workers {
		worker.mu.RLock()
		if worker.active {
			active++
		}
		worker.mu.RUnlock()
	}
	return active
}

// GetTotalWorkerCount returns the total number of workers
func (wp *WorkerPool[T]) GetTotalWorkerCount() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return len(wp.workers)
}

// QueueLength returns the current length of the job queue
func (wp *WorkerPool[T]) QueueLength() int {
	return len(wp.jobQueue)
}

// GoroutineMetrics tracks goroutine performance metrics
type GoroutineMetrics struct {
	TotalJobsProcessed    int64         `json:"total_jobs_processed"`
	ActiveWorkers         int           `json:"active_workers"`
	QueueLength           int           `json:"queue_length"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	LastUpdated           time.Time     `json:"last_updated"`
	mu                    sync.RWMutex
}

// NewGoroutineMetrics creates new metrics tracker
func NewGoroutineMetrics() *GoroutineMetrics {
	return &GoroutineMetrics{
		LastUpdated: time.Now(),
	}
}

// UpdateMetrics updates the metrics with current values
func (gm *GoroutineMetrics) UpdateMetrics(activeWorkers, queueLength int, processingTime time.Duration) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.ActiveWorkers = activeWorkers
	gm.QueueLength = queueLength
	gm.TotalJobsProcessed++

	// Update average processing time
	if gm.AverageProcessingTime == 0 {
		gm.AverageProcessingTime = processingTime
	} else {
		gm.AverageProcessingTime = (gm.AverageProcessingTime + processingTime) / 2
	}

	gm.LastUpdated = time.Now()
}

// autoScale automatically scales the worker pool based on load
func (gm *GoroutineManager[T]) autoScale() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gm.scaleWorkers()
		}
	}
}

// scaleWorkers scales workers based on current load
func (gm *GoroutineManager[T]) scaleWorkers() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	currentWorkers := len(gm.workerPool.workers)
	activeWorkers := gm.workerPool.GetActiveWorkerCount()
	queueLength := len(gm.workerPool.jobQueue)

	// Calculate target workers based on load
	targetWorkers := gm.calculateTargetWorkers(activeWorkers, queueLength)

	// Scale up if needed
	if targetWorkers > currentWorkers {
		gm.scaleUp(targetWorkers - currentWorkers)
	} else if targetWorkers < currentWorkers {
		gm.scaleDown(currentWorkers - targetWorkers)
	}
}

// calculateTargetWorkers calculates the optimal number of workers
func (gm *GoroutineManager[T]) calculateTargetWorkers(activeWorkers, queueLength int) int {
	// Base calculation on CPU cores
	baseWorkers := runtime.NumCPU() * gm.config.MaxWorkersPerCPU

	// Adjust based on queue length
	if queueLength > gm.config.QueueBufferSize/2 {
		baseWorkers = int(float64(baseWorkers) * 1.5)
	}

	// Ensure within bounds
	if baseWorkers < gm.config.MinWorkers {
		baseWorkers = gm.config.MinWorkers
	}
	if baseWorkers > gm.config.MaxWorkers {
		baseWorkers = gm.config.MaxWorkers
	}

	return baseWorkers
}

// scaleUp adds more workers to the pool
func (gm *GoroutineManager[T]) scaleUp(count int) {
	currentCount := len(gm.workerPool.workers)

	for i := 0; i < count; i++ {
		worker := gm.workerPool.newWorker(currentCount + i)
		gm.workerPool.workers = append(gm.workerPool.workers, worker)
		go worker.start()
	}
}

// scaleDown reduces the number of workers
func (gm *GoroutineManager[T]) scaleDown(count int) {
	// For now, we'll just mark workers as inactive
	// In a real implementation, you'd want to gracefully shut down workers
	currentCount := len(gm.workerPool.workers)
	if count > currentCount {
		count = currentCount
	}

	// Mark the last 'count' workers as inactive
	for i := currentCount - count; i < currentCount; i++ {
		if i < len(gm.workerPool.workers) {
			gm.workerPool.workers[i].cancel()
		}
	}
}

// GoroutineAwareRepository extends Repository with automatic goroutine optimization
type GoroutineAwareRepository[T any] struct {
	repository Repository[T]
	manager    *GoroutineManager[T]
}

// NewGoroutineAwareRepository creates a new goroutine-aware repository
func NewGoroutineAwareRepository[T any](repo Repository[T], manager *GoroutineManager[T]) *GoroutineAwareRepository[T] {
	return &GoroutineAwareRepository[T]{
		repository: repo,
		manager:    manager,
	}
}

// FindAsync finds a model by ID asynchronously
func (gar *GoroutineAwareRepository[T]) FindAsync(id uint) <-chan RepositoryResult[T] {
	resultChan := make(chan RepositoryResult[T], 1)

	go func() {
		defer close(resultChan)

		start := time.Now()
		result, err := gar.repository.Find(id)
		processingTime := time.Since(start)

		// Update metrics
		gar.manager.metrics.UpdateMetrics(
			gar.manager.workerPool.GetActiveWorkerCount(),
			len(gar.manager.workerPool.jobQueue),
			processingTime,
		)

		var data T
		if result != nil {
			data = *result
		}
		resultChan <- RepositoryResult[T]{
			Data:  data,
			Error: err,
		}
	}()

	return resultChan
}

// FindManyAsync finds multiple models asynchronously
func (gar *GoroutineAwareRepository[T]) FindManyAsync(ids []uint) <-chan RepositoryResult[[]T] {
	resultChan := make(chan RepositoryResult[[]T], 1)

	go func() {
		defer close(resultChan)

		// Use worker pool for parallel processing
		results := make([]T, 0, len(ids))
		var mu sync.Mutex
		var wg sync.WaitGroup

		for _, id := range ids {
			wg.Add(1)
			go func(id uint) {
				defer wg.Done()
				if result, err := gar.repository.Find(id); err == nil && result != nil {
					mu.Lock()
					results = append(results, *result) // *result is T
					mu.Unlock()
				}
			}(id)
		}

		wg.Wait()

		resultChan <- RepositoryResult[[]T]{
			Data:  results,
			Error: nil,
		}
	}()

	return resultChan
}

// RepositoryResult represents the result of an async repository operation
type RepositoryResult[T any] struct {
	Data  T
	Error error
}

// GoroutineAwareEventDispatcher extends EventDispatcher with automatic goroutine optimization
type GoroutineAwareEventDispatcher[T any] struct {
	dispatcher EventDispatcher[T]
	manager    *GoroutineManager[T]
}

// NewGoroutineAwareEventDispatcher creates a new goroutine-aware event dispatcher
func NewGoroutineAwareEventDispatcher[T any](dispatcher EventDispatcher[T], manager *GoroutineManager[T]) *GoroutineAwareEventDispatcher[T] {
	return &GoroutineAwareEventDispatcher[T]{
		dispatcher: dispatcher,
		manager:    manager,
	}
}

// DispatchAsync dispatches an event asynchronously using the worker pool
func (gaed *GoroutineAwareEventDispatcher[T]) DispatchAsync(event *Event[T]) error {
	job := GoroutineJob[T]{
		Job: Job[T]{
			ID: fmt.Sprintf("event_%s_%d", event.Name, time.Now().UnixNano()),
		},
		Timeout: 30 * time.Second,
		Handler: func(ctx context.Context, job *GoroutineJob[T]) error {
			return gaed.dispatcher.Dispatch(event)
		},
	}

	return gaed.manager.workerPool.Submit(job)
}

// GoroutineAwareJobDispatcher extends JobDispatcher with automatic goroutine optimization
type GoroutineAwareJobDispatcher[T any] struct {
	dispatcher JobDispatcher[T]
	manager    *GoroutineManager[T]
}

// NewGoroutineAwareJobDispatcher creates a new goroutine-aware job dispatcher
func NewGoroutineAwareJobDispatcher[T any](dispatcher JobDispatcher[T], manager *GoroutineManager[T]) *GoroutineAwareJobDispatcher[T] {
	return &GoroutineAwareJobDispatcher[T]{
		dispatcher: dispatcher,
		manager:    manager,
	}
}

// DispatchAsync dispatches a job asynchronously using the worker pool
func (gajd *GoroutineAwareJobDispatcher[T]) DispatchAsync(job T) error {
	// Create a job for the worker pool
	poolJob := GoroutineJob[T]{
		Job: Job[T]{
			ID: fmt.Sprintf("job_%d", time.Now().UnixNano()),
		},
		Timeout: 30 * time.Second,
		Handler: func(ctx context.Context, poolJob *GoroutineJob[T]) error {
			return gajd.dispatcher.Dispatch(job)
		},
	}

	return gajd.manager.workerPool.Submit(poolJob)
}

// GetMetrics returns current goroutine metrics
func (gm *GoroutineManager[T]) GetMetrics() *GoroutineMetrics {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.metrics
}

// GetWorkerPool returns the underlying worker pool
func (gm *GoroutineManager[T]) GetWorkerPool() *WorkerPool[T] {
	return gm.workerPool
}
