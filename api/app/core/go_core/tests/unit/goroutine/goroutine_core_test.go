package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
)

// TestGoroutineConfig tests the configuration system
func TestGoroutineConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := go_core.DefaultGoroutineConfig()
		assert.NotNil(t, config)
		assert.Greater(t, config.MaxWorkers, 0)
		assert.Greater(t, config.WorkerTimeout, time.Duration(0))
		assert.Greater(t, config.QueueBufferSize, 0)
		assert.True(t, config.EnableAutoScaling)
		assert.Greater(t, config.MinWorkers, 0)
		assert.Greater(t, config.MaxWorkersPerCPU, 0)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		customConfig := &go_core.GoroutineConfig{
			MaxWorkers:        10,
			WorkerTimeout:     60 * time.Second,
			QueueBufferSize:   500,
			EnableAutoScaling: false,
			MinWorkers:        5,
			MaxWorkersPerCPU:  2,
		}

		manager := go_core.NewGoroutineManager[string](customConfig)
		assert.NotNil(t, manager)
		assert.Equal(t, customConfig.MaxWorkers, manager.GetWorkerPool().GetTotalWorkerCount())
	})
}

// TestGoroutineManager tests the main goroutine manager
func TestGoroutineManager(t *testing.T) {
	t.Run("NewManager", func(t *testing.T) {
		manager := go_core.NewGoroutineManager[string](nil)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.GetWorkerPool())
		assert.NotNil(t, manager.GetMetrics())
	})

	t.Run("ManagerWithNilConfig", func(t *testing.T) {
		manager := go_core.NewGoroutineManager[string](nil)
		assert.NotNil(t, manager)
		// Should use default config
		assert.Greater(t, manager.GetWorkerPool().GetTotalWorkerCount(), 0)
	})
}

// TestWorkerPool tests the worker pool functionality
func TestWorkerPool(t *testing.T) {
	t.Run("NewWorkerPool", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      5,
			QueueBufferSize: 100,
		}
		manager := go_core.NewGoroutineManager[string](config)
		pool := manager.GetWorkerPool()

		assert.Equal(t, 5, pool.GetTotalWorkerCount())
		// Allow some time for workers to become inactive
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 0, pool.GetActiveWorkerCount())
		assert.Equal(t, 0, pool.QueueLength())
	})

	t.Run("SubmitJob", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)
		pool := manager.GetWorkerPool()

		job := go_core.GoroutineJob[string]{
			Job: go_core.Job[string]{
				ID:   "test-job",
				Data: "test-data",
			},
			Timeout: 5 * time.Second,
			Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
				return nil
			},
		}

		err := pool.Submit(job)
		assert.NoError(t, err)
	})

	t.Run("SubmitJobWithFullQueue", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      1,
			QueueBufferSize: 2, // Small buffer
		}
		manager := go_core.NewGoroutineManager[string](config)
		pool := manager.GetWorkerPool()

		// Fill the queue with slow jobs that won't be processed quickly
		for i := 0; i < 3; i++ { // Submit more jobs than buffer size
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("job-%d", i),
					Data: "test-data",
				},
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					time.Sleep(200 * time.Millisecond) // Very slow job
					return nil
				},
			}
			err := pool.Submit(job)
			if i < 2 {
				assert.NoError(t, err) // First 2 should succeed
			} else {
				assert.Error(t, err) // Third should fail
				assert.Contains(t, err.Error(), "job queue is full")
				return
			}
		}
	})

	t.Run("Shutdown", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)
		pool := manager.GetWorkerPool()

		// Submit a job
		job := go_core.GoroutineJob[string]{
			Job: go_core.Job[string]{
				ID:   "test-job",
				Data: "test-data",
			},
			Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
				return nil
			},
		}
		err := pool.Submit(job)
		assert.NoError(t, err)

		// Shutdown
		pool.Shutdown()

		// Wait a bit for shutdown to complete
		time.Sleep(50 * time.Millisecond)

		// Try to submit after shutdown - should fail gracefully
		err = pool.Submit(job)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "worker pool is shutting down")
	})
}

// TestWorker tests individual worker functionality
func TestWorker(t *testing.T) {
	t.Run("WorkerProcessing", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      1,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)
		pool := manager.GetWorkerPool()

		var processed bool
		var mu sync.Mutex

		job := go_core.GoroutineJob[string]{
			Job: go_core.Job[string]{
				ID:   "test-job",
				Data: "test-data",
			},
			Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
				mu.Lock()
				processed = true
				mu.Unlock()
				return nil
			},
		}

		err := pool.Submit(job)
		assert.NoError(t, err)

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.True(t, processed)
		mu.Unlock()
	})

	t.Run("WorkerTimeout", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      1,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)
		pool := manager.GetWorkerPool()

		var jobCompleted bool
		var mu sync.Mutex

		job := go_core.GoroutineJob[string]{
			Job: go_core.Job[string]{
				ID:   "timeout-job",
				Data: "test-data",
			},
			Timeout: 50 * time.Millisecond,
			Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(200 * time.Millisecond): // Longer than timeout
					mu.Lock()
					jobCompleted = true
					mu.Unlock()
					return nil
				}
			},
		}

		err := pool.Submit(job)
		assert.NoError(t, err)

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.False(t, jobCompleted) // Should not complete due to timeout
		mu.Unlock()
	})
}

// TestGoroutineMetrics tests the metrics system
func TestGoroutineMetrics(t *testing.T) {
	t.Run("NewMetrics", func(t *testing.T) {
		metrics := go_core.NewGoroutineMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(0), metrics.TotalJobsProcessed)
		assert.Equal(t, 0, metrics.ActiveWorkers)
		assert.Equal(t, 0, metrics.QueueLength)
	})

	t.Run("UpdateMetrics", func(t *testing.T) {
		metrics := go_core.NewGoroutineMetrics()

		metrics.UpdateMetrics(5, 10, 100*time.Millisecond)

		assert.Equal(t, 5, metrics.ActiveWorkers)
		assert.Equal(t, 10, metrics.QueueLength)
		assert.Equal(t, 100*time.Millisecond, metrics.AverageProcessingTime)
	})

	t.Run("ConcurrentMetricsUpdate", func(t *testing.T) {
		metrics := go_core.NewGoroutineMetrics()
		var wg sync.WaitGroup

		// Update metrics concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				metrics.UpdateMetrics(id, id*2, time.Duration(id)*time.Millisecond)
			}(i)
		}

		wg.Wait()
		// Should not panic and should have some reasonable values
		assert.GreaterOrEqual(t, metrics.ActiveWorkers, 0)
		assert.GreaterOrEqual(t, metrics.QueueLength, 0)
	})
}

// TestGoroutineAwareRepository tests the goroutine-aware repository
func TestGoroutineAwareRepository(t *testing.T) {
	t.Run("NewGoroutineAwareRepository", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)

		// Mock repository
		mockRepo := &MockRepository[string]{}

		gar := go_core.NewGoroutineAwareRepository[string](mockRepo, manager)
		assert.NotNil(t, gar)
	})

	t.Run("FindAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)

		mockRepo := &MockRepository[string]{
			findResult: "test-data",
			findError:  nil,
		}

		gar := go_core.NewGoroutineAwareRepository[string](mockRepo, manager)

		resultChan := gar.FindAsync(1)
		result := <-resultChan

		assert.Equal(t, "test-data", result.Data)
		assert.NoError(t, result.Error)
	})

	t.Run("FindManyAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)

		// Create a mock repository that returns different results for different IDs
		mockRepo := &MockRepository[string]{
			findResult: "test-data", // This will be returned for each Find call
			findError:  nil,
		}

		gar := go_core.NewGoroutineAwareRepository[string](mockRepo, manager)

		resultChan := gar.FindManyAsync([]uint{1, 2})
		result := <-resultChan

		// Since FindManyAsync calls Find for each ID, and Find returns the same result,
		// we expect two copies of the same data
		expected := []string{"test-data", "test-data"}
		assert.Equal(t, expected, result.Data)
		assert.NoError(t, result.Error)
	})
}

// TestGoroutineAwareEventDispatcher tests the goroutine-aware event dispatcher
func TestGoroutineAwareEventDispatcher(t *testing.T) {
	t.Run("NewGoroutineAwareEventDispatcher", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)

		mockDispatcher := &MockEventDispatcher[string]{}

		gaed := go_core.NewGoroutineAwareEventDispatcher[string](mockDispatcher, manager)
		assert.NotNil(t, gaed)
	})

	t.Run("DispatchAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)

		mockDispatcher := &MockEventDispatcher[string]{
			dispatchError: nil,
		}

		gaed := go_core.NewGoroutineAwareEventDispatcher[string](mockDispatcher, manager)

		event := &go_core.Event[string]{
			ID:   "test-event",
			Data: "test-data",
		}

		err := gaed.DispatchAsync(event)
		assert.NoError(t, err)
	})
}

// TestGoroutineAwareJobDispatcher tests the goroutine-aware job dispatcher
func TestGoroutineAwareJobDispatcher(t *testing.T) {
	t.Run("NewGoroutineAwareJobDispatcher", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)

		mockDispatcher := &MockJobDispatcher[string]{}

		gajd := go_core.NewGoroutineAwareJobDispatcher[string](mockDispatcher, manager)
		assert.NotNil(t, gajd)
	})

	t.Run("DispatchAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		manager := go_core.NewGoroutineManager[string](config)

		mockDispatcher := &MockJobDispatcher[string]{
			dispatchError: nil,
		}

		gajd := go_core.NewGoroutineAwareJobDispatcher[string](mockDispatcher, manager)

		job := "test-job"

		err := gajd.DispatchAsync(job)
		assert.NoError(t, err)
	})
}

// Mock implementations for testing
type MockRepository[T any] struct {
	findResult     T
	findError      error
	findManyResult []T
	findManyError  error
}

func (m *MockRepository[T]) Find(id uint) (*T, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	return &m.findResult, nil
}

func (m *MockRepository[T]) FindMany(ids []uint) ([]T, error) {
	return m.findManyResult, m.findManyError
}

func (m *MockRepository[T]) Count() (int64, error) {
	return int64(len(m.findManyResult)), nil
}

// Implement other required methods with minimal implementations
func (m *MockRepository[T]) FindBy(field string, value any) (*T, error) { return nil, nil }
func (m *MockRepository[T]) FindAll() ([]T, error)                      { return nil, nil }
func (m *MockRepository[T]) Create(model *T) error                      { return nil }
func (m *MockRepository[T]) Update(model *T) error                      { return nil }
func (m *MockRepository[T]) Delete(id uint) error                       { return nil }
func (m *MockRepository[T]) FindWithContext(ctx context.Context, id uint) (*T, error) {
	return nil, nil
}
func (m *MockRepository[T]) FindByWithContext(ctx context.Context, field string, value any) (*T, error) {
	return nil, nil
}
func (m *MockRepository[T]) FindAllWithContext(ctx context.Context) ([]T, error)   { return nil, nil }
func (m *MockRepository[T]) CreateWithContext(ctx context.Context, model *T) error { return nil }
func (m *MockRepository[T]) UpdateWithContext(ctx context.Context, model *T) error { return nil }
func (m *MockRepository[T]) DeleteWithContext(ctx context.Context, id uint) error  { return nil }
func (m *MockRepository[T]) Where(conditions map[string]any) go_core.Query[T]      { return nil }
func (m *MockRepository[T]) WhereRaw(query string, args ...any) go_core.Query[T]   { return nil }
func (m *MockRepository[T]) WhereWithContext(ctx context.Context, conditions map[string]any) go_core.Query[T] {
	return nil
}
func (m *MockRepository[T]) WhereRawWithContext(ctx context.Context, query string, args ...any) go_core.Query[T] {
	return nil
}
func (m *MockRepository[T]) Transaction(fn func(go_core.Repository[T]) error) error { return nil }
func (m *MockRepository[T]) TransactionWithContext(ctx context.Context, fn func(go_core.Repository[T]) error) error {
	return nil
}
func (m *MockRepository[T]) WithContext(ctx context.Context) go_core.Repository[T] { return m }
func (m *MockRepository[T]) Exists(id uint) (bool, error)                          { return false, nil }
func (m *MockRepository[T]) CountWhere(conditions map[string]any) (int64, error)   { return 0, nil }
func (m *MockRepository[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	return false, nil
}
func (m *MockRepository[T]) CountWithContext(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockRepository[T]) CountWhereWithContext(ctx context.Context, conditions map[string]any) (int64, error) {
	return 0, nil
}
func (m *MockRepository[T]) GetPerformanceStats() map[string]interface{}  { return nil }
func (m *MockRepository[T]) GetOptimizationStats() map[string]interface{} { return nil }

type MockEventDispatcher[T any] struct {
	dispatchError error
}

func (m *MockEventDispatcher[T]) Dispatch(event *go_core.Event[T]) error {
	return m.dispatchError
}

func (m *MockEventDispatcher[T]) DispatchAsync(event *go_core.Event[T]) error {
	return m.dispatchError
}

func (m *MockEventDispatcher[T]) Listen(eventName string, listener go_core.EventListener[T]) error {
	return nil
}
func (m *MockEventDispatcher[T]) RemoveListener(eventName string, listener go_core.EventListener[T]) error {
	return nil
}
func (m *MockEventDispatcher[T]) Handle(event *go_core.Event[T]) error  { return nil }
func (m *MockEventDispatcher[T]) HasListeners(eventName string) bool    { return false }
func (m *MockEventDispatcher[T]) GetListenerCount(eventName string) int { return 0 }
func (m *MockEventDispatcher[T]) WithContext(ctx context.Context) go_core.EventDispatcher[T] {
	return m
}
func (m *MockEventDispatcher[T]) GetPerformanceStats() map[string]interface{}  { return nil }
func (m *MockEventDispatcher[T]) GetOptimizationStats() map[string]interface{} { return nil }

type MockJobDispatcher[T any] struct {
	dispatchError error
}

func (m *MockJobDispatcher[T]) Dispatch(job T) error {
	return m.dispatchError
}

func (m *MockJobDispatcher[T]) DispatchSync(job T) error {
	return m.dispatchError
}

func (m *MockJobDispatcher[T]) GetQueue() go_core.Queue[T]                               { return nil }
func (m *MockJobDispatcher[T]) WithContext(ctx context.Context) go_core.JobDispatcher[T] { return m }
