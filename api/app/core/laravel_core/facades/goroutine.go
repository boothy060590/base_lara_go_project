package facades_core

import (
	"context"
	"sync"
	"time"

	"base_lara_go_project/app/core/go_core"
)

// GoroutineFacade provides Laravel-style access to goroutine-optimized operations
type GoroutineFacade struct {
	manager *go_core.GoroutineManager[any]
	mu      sync.RWMutex
}

var (
	goroutineInstance *GoroutineFacade
	goroutineOnce     sync.Once
)

// Goroutine returns the singleton goroutine facade instance
func Goroutine() *GoroutineFacade {
	goroutineOnce.Do(func() {
		goroutineInstance = &GoroutineFacade{
			manager: go_core.NewGoroutineManager[any](nil), // Use default config
		}
	})
	return goroutineInstance
}

// SetManager sets a custom goroutine manager
func (gf *GoroutineFacade) SetManager(manager *go_core.GoroutineManager[any]) {
	gf.mu.Lock()
	defer gf.mu.Unlock()
	gf.manager = manager
}

// GetManager returns the current goroutine manager
func (gf *GoroutineFacade) GetManager() *go_core.GoroutineManager[any] {
	gf.mu.RLock()
	defer gf.mu.RUnlock()
	return gf.manager
}

// Async executes a function asynchronously using the worker pool
func (gf *GoroutineFacade) Async(fn func() error) error {
	job := go_core.GoroutineJob[any]{
		Job: go_core.Job[any]{
			ID: generateJobID(),
		},
		Timeout: 30 * time.Second,
		Handler: func(ctx context.Context, job *go_core.GoroutineJob[any]) error {
			return fn()
		},
	}

	return gf.manager.GetWorkerPool().Submit(job)
}

// AsyncWithTimeout executes a function asynchronously with a custom timeout
func (gf *GoroutineFacade) AsyncWithTimeout(fn func() error, timeout time.Duration) error {
	job := go_core.GoroutineJob[any]{
		Job: go_core.Job[any]{
			ID: generateJobID(),
		},
		Timeout: timeout,
		Handler: func(ctx context.Context, job *go_core.GoroutineJob[any]) error {
			return fn()
		},
	}

	return gf.manager.GetWorkerPool().Submit(job)
}

// Parallel executes multiple functions in parallel and waits for all to complete
func (gf *GoroutineFacade) Parallel(functions ...func() error) []error {
	var wg sync.WaitGroup
	errors := make([]error, len(functions))

	for i, fn := range functions {
		wg.Add(1)
		go func(index int, function func() error) {
			defer wg.Done()
			errors[index] = function()
		}(i, fn)
	}

	wg.Wait()
	return errors
}

// ParallelWithContext executes multiple functions in parallel with context support
func (gf *GoroutineFacade) ParallelWithContext(ctx context.Context, functions ...func(context.Context) error) []error {
	var wg sync.WaitGroup
	errors := make([]error, len(functions))

	for i, fn := range functions {
		wg.Add(1)
		go func(index int, function func(context.Context) error) {
			defer wg.Done()
			errors[index] = function(ctx)
		}(i, fn)
	}

	wg.Wait()
	return errors
}

// Retry executes a function with automatic retry logic
func (gf *GoroutineFacade) Retry(fn func() error, maxAttempts int, delay time.Duration) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxAttempts {
				time.Sleep(delay)
			}
		}
	}

	return lastErr
}

// RetryAsync executes a function with automatic retry logic asynchronously
func (gf *GoroutineFacade) RetryAsync(fn func() error, maxAttempts int, delay time.Duration) error {
	return gf.Async(func() error {
		return gf.Retry(fn, maxAttempts, delay)
	})
}

// Batch processes items in batches using goroutines
func (gf *GoroutineFacade) Batch(items []interface{}, batchSize int, processor func([]interface{}) error) error {
	if len(items) == 0 {
		return nil
	}

	// Split items into batches
	batches := make([][]interface{}, 0, (len(items)+batchSize-1)/batchSize)
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	// Process batches in parallel
	return gf.Parallel(func() error {
		for _, batch := range batches {
			if err := processor(batch); err != nil {
				return err
			}
		}
		return nil
	})[0]
}

// Map applies a function to each item in a slice using goroutines
func (gf *GoroutineFacade) Map(items []interface{}, mapper func(interface{}) (interface{}, error)) ([]interface{}, error) {
	if len(items) == 0 {
		return []interface{}{}, nil
	}

	results := make([]interface{}, len(items))
	errors := make([]error, len(items))

	var wg sync.WaitGroup
	for i, item := range items {
		wg.Add(1)
		go func(index int, value interface{}) {
			defer wg.Done()
			result, err := mapper(value)
			results[index] = result
			errors[index] = err
		}(i, item)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// Filter filters items using goroutines
func (gf *GoroutineFacade) Filter(items []interface{}, predicate func(interface{}) (bool, error)) ([]interface{}, error) {
	if len(items) == 0 {
		return []interface{}{}, nil
	}

	results := make([]interface{}, 0, len(items))
	errors := make([]error, len(items))

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, item := range items {
		wg.Add(1)
		go func(index int, value interface{}) {
			defer wg.Done()
			keep, err := predicate(value)
			errors[index] = err
			if err == nil && keep {
				mu.Lock()
				results = append(results, value)
				mu.Unlock()
			}
		}(i, item)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// Reduce reduces items using goroutines (note: order is not guaranteed)
func (gf *GoroutineFacade) Reduce(items []interface{}, initial interface{}, reducer func(interface{}, interface{}) (interface{}, error)) (interface{}, error) {
	if len(items) == 0 {
		return initial, nil
	}

	result := initial
	var mu sync.Mutex

	// Process items in parallel
	errors := gf.Parallel(func() error {
		for _, item := range items {
			mu.Lock()
			newResult, err := reducer(result, item)
			if err != nil {
				mu.Unlock()
				return err
			}
			result = newResult
			mu.Unlock()
		}
		return nil
	})

	if len(errors) > 0 && errors[0] != nil {
		return result, errors[0]
	}

	return result, nil
}

// GetMetrics returns current goroutine metrics
func (gf *GoroutineFacade) GetMetrics() *go_core.GoroutineMetrics {
	return gf.manager.GetMetrics()
}

// GetActiveWorkerCount returns the number of active workers
func (gf *GoroutineFacade) GetActiveWorkerCount() int {
	return gf.manager.GetWorkerPool().GetActiveWorkerCount()
}

// GetQueueLength returns the current queue length
func (gf *GoroutineFacade) GetQueueLength() int {
	return gf.manager.GetWorkerPool().QueueLength()
}

// Shutdown gracefully shuts down the goroutine manager
func (gf *GoroutineFacade) Shutdown() {
	gf.manager.GetWorkerPool().Shutdown()
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return "job_" + time.Now().Format("20060102150405") + "_" + time.Now().Format("000000000")
}

// GoroutineAwareRepository provides Laravel-style repository operations with automatic goroutine optimization
type GoroutineAwareRepository[T any] struct {
	repository go_core.Repository[T]
	manager    *go_core.GoroutineManager[T]
}

// NewGoroutineAwareRepository creates a new goroutine-aware repository
func NewGoroutineAwareRepository[T any](repo go_core.Repository[T]) *GoroutineAwareRepository[T] {
	return &GoroutineAwareRepository[T]{
		repository: repo,
		manager:    go_core.NewGoroutineManager[T](nil),
	}
}

// FindAsync finds a model by ID asynchronously
func (gar *GoroutineAwareRepository[T]) FindAsync(id uint) <-chan go_core.RepositoryResult[T] {
	goroutineRepo := go_core.NewGoroutineAwareRepository(gar.repository, gar.manager)
	return goroutineRepo.FindAsync(id)
}

// FindManyAsync finds multiple models asynchronously
func (gar *GoroutineAwareRepository[T]) FindManyAsync(ids []uint) <-chan go_core.RepositoryResult[[]T] {
	goroutineRepo := go_core.NewGoroutineAwareRepository(gar.repository, gar.manager)
	return goroutineRepo.FindManyAsync(ids)
}

// CreateAsync creates a model asynchronously
func (gar *GoroutineAwareRepository[T]) CreateAsync(model *T) <-chan error {
	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		resultChan <- gar.repository.Create(model)
	}()

	return resultChan
}

// UpdateAsync updates a model asynchronously
func (gar *GoroutineAwareRepository[T]) UpdateAsync(model *T) <-chan error {
	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		resultChan <- gar.repository.Update(model)
	}()

	return resultChan
}

// DeleteAsync deletes a model asynchronously
func (gar *GoroutineAwareRepository[T]) DeleteAsync(id uint) <-chan error {
	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		resultChan <- gar.repository.Delete(id)
	}()

	return resultChan
}

// GoroutineAwareEventDispatcher provides Laravel-style event dispatching with automatic goroutine optimization
type GoroutineAwareEventDispatcher[T any] struct {
	dispatcher go_core.EventDispatcher[T]
	manager    *go_core.GoroutineManager[T]
}

// NewGoroutineAwareEventDispatcher creates a new goroutine-aware event dispatcher
func NewGoroutineAwareEventDispatcher[T any](dispatcher go_core.EventDispatcher[T]) *GoroutineAwareEventDispatcher[T] {
	return &GoroutineAwareEventDispatcher[T]{
		dispatcher: dispatcher,
		manager:    go_core.NewGoroutineManager[T](nil),
	}
}

// DispatchAsync dispatches an event asynchronously using the worker pool
func (gaed *GoroutineAwareEventDispatcher[T]) DispatchAsync(event *go_core.Event[T]) error {
	goroutineDispatcher := go_core.NewGoroutineAwareEventDispatcher(gaed.dispatcher, gaed.manager)
	return goroutineDispatcher.DispatchAsync(event)
}

// GoroutineAwareJobDispatcher provides Laravel-style job dispatching with automatic goroutine optimization
type GoroutineAwareJobDispatcher[T any] struct {
	dispatcher go_core.JobDispatcher[T]
	manager    *go_core.GoroutineManager[T]
}

// NewGoroutineAwareJobDispatcher creates a new goroutine-aware job dispatcher
func NewGoroutineAwareJobDispatcher[T any](dispatcher go_core.JobDispatcher[T]) *GoroutineAwareJobDispatcher[T] {
	return &GoroutineAwareJobDispatcher[T]{
		dispatcher: dispatcher,
		manager:    go_core.NewGoroutineManager[T](nil),
	}
}

// DispatchAsync dispatches a job asynchronously using the worker pool
func (gajd *GoroutineAwareJobDispatcher[T]) DispatchAsync(job T) error {
	goroutineDispatcher := go_core.NewGoroutineAwareJobDispatcher(gajd.dispatcher, gajd.manager)
	return goroutineDispatcher.DispatchAsync(job)
}
