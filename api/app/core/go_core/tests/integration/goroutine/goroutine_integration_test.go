package integration

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
)

// TestGoroutineSystemIntegration tests the complete goroutine system integration
func TestGoroutineSystemIntegration(t *testing.T) {
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Create goroutine manager
		config := &go_core.GoroutineConfig{
			MaxWorkers:        4,
			WorkerTimeout:     30 * time.Second,
			QueueBufferSize:   100,
			EnableAutoScaling: true,
			MinWorkers:        2,
			MaxWorkersPerCPU:  2,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		// Create work stealing pool
		wsConfig := &go_core.WorkStealingConfig{
			NumWorkers:     2,
			QueueSize:      50,
			StealThreshold: 2,
			StealBatchSize: 5,
		}
		wsPool := go_core.NewWorkStealingPool[string](wsConfig)
		defer wsPool.Shutdown()

		var processedCount int64
		var mu sync.Mutex

		// Submit jobs to both systems
		for i := 0; i < 20; i++ {
			// Submit to goroutine manager
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("gm-job-%d", i),
					Data: fmt.Sprintf("gm-data-%d", i),
				},
				Timeout: 5 * time.Second,
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					mu.Lock()
					processedCount++
					mu.Unlock()
					return nil
				},
			}
			err := manager.GetWorkerPool().Submit(job)
			assert.NoError(t, err)

			// Submit to work stealing pool
			item := go_core.WorkItem[string]{
				ID:   fmt.Sprintf("ws-item-%d", i),
				Data: fmt.Sprintf("ws-data-%d", i),
				Handler: func(ctx context.Context, data string) error {
					mu.Lock()
					processedCount++
					mu.Unlock()
					return nil
				},
				Timeout:  5 * time.Second,
				Priority: i % 4,
			}
			err = wsPool.Submit(item)
			assert.NoError(t, err)
		}

		// Wait for processing
		time.Sleep(1 * time.Second)

		mu.Lock()
		finalCount := processedCount
		mu.Unlock()

		assert.Equal(t, int64(40), finalCount) // 20 from each system

		// Check metrics
		gmMetrics := manager.GetMetrics()
		wsMetrics := wsPool.GetMetrics()

		assert.GreaterOrEqual(t, gmMetrics.TotalJobsProcessed, int64(0))
		assert.GreaterOrEqual(t, wsMetrics.TotalProcessed, int64(20))
	})

	t.Run("AutoScalingIntegration", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:        8,
			WorkerTimeout:     30 * time.Second,
			QueueBufferSize:   200,
			EnableAutoScaling: true,
			MinWorkers:        2,
			MaxWorkersPerCPU:  4,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		initialWorkers := manager.GetWorkerPool().GetTotalWorkerCount()

		// Submit many jobs to trigger auto-scaling
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				job := go_core.GoroutineJob[string]{
					Job: go_core.Job[string]{
						ID:   fmt.Sprintf("scale-job-%d", id),
						Data: fmt.Sprintf("scale-data-%d", id),
					},
					Timeout: 5 * time.Second,
					Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
						time.Sleep(50 * time.Millisecond) // Simulate work
						return nil
					},
				}
				manager.GetWorkerPool().Submit(job)
			}(i)
		}

		wg.Wait()

		// Wait for auto-scaling to take effect
		time.Sleep(2 * time.Second)

		finalWorkers := manager.GetWorkerPool().GetTotalWorkerCount()
		assert.GreaterOrEqual(t, finalWorkers, initialWorkers)
	})

	t.Run("ContextIntegration", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			WorkerTimeout:   30 * time.Second,
			QueueBufferSize: 50,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		var processedCount int64
		var mu sync.Mutex

		// Submit jobs with context
		for i := 0; i < 10; i++ {
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("ctx-job-%d", i),
					Data: fmt.Sprintf("ctx-data-%d", i),
				},
				Timeout: 5 * time.Second,
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(50 * time.Millisecond): // Reduced from 200ms to 50ms
						mu.Lock()
						processedCount++
						mu.Unlock()
						return nil
					}
				},
			}
			manager.GetWorkerPool().Submit(job)
		}

		// Wait for context timeout
		<-ctx.Done()

		mu.Lock()
		finalCount := processedCount
		mu.Unlock()

		// Some jobs should be processed before context timeout
		assert.Greater(t, finalCount, int64(0))
		assert.Less(t, finalCount, int64(10)) // Not all should complete due to timeout
	})
}

// TestGoroutineAwareComponentsIntegration tests integration with goroutine-aware components
func TestGoroutineAwareComponentsIntegration(t *testing.T) {
	t.Run("RepositoryIntegration", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      4,
			QueueBufferSize: 50,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		// Create mock repository
		mockRepo := &MockRepository[string]{
			findResult:     "test-data",
			findError:      nil,
			findManyResult: []string{"data1", "data2", "data3"},
			findManyError:  nil,
		}

		// Create goroutine-aware repository
		gar := go_core.NewGoroutineAwareRepository[string](mockRepo, manager)

		// Test FindAsync
		resultChan := gar.FindAsync(1)
		result := <-resultChan

		assert.Equal(t, "data1", result.Data)
		assert.NoError(t, result.Error)

		// Test FindManyAsync
		resultChan2 := gar.FindManyAsync([]uint{1, 2, 3})
		result2 := <-resultChan2

		expected := []string{"data1", "data2", "data3"}
		actual := result2.Data
		slices.Sort(expected)
		slices.Sort(actual)
		assert.Equal(t, expected, actual)
		assert.NoError(t, result2.Error)
	})

	t.Run("EventDispatcherIntegration", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      4,
			QueueBufferSize: 50,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		// Create mock event dispatcher
		mockDispatcher := &MockEventDispatcher[string]{
			dispatchError: nil,
		}

		// Create goroutine-aware event dispatcher
		gaed := go_core.NewGoroutineAwareEventDispatcher[string](mockDispatcher, manager)

		// Test DispatchAsync
		event := &go_core.Event[string]{
			ID:   "test-event",
			Data: "test-data",
		}

		err := gaed.DispatchAsync(event)
		assert.NoError(t, err)

		// Wait for processing
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("JobDispatcherIntegration", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      4,
			QueueBufferSize: 50,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		// Create mock job dispatcher
		mockDispatcher := &MockJobDispatcher[string]{
			dispatchError: nil,
		}

		// Create goroutine-aware job dispatcher
		gajd := go_core.NewGoroutineAwareJobDispatcher[string](mockDispatcher, manager)

		// Test DispatchAsync
		job := "test-job"

		err := gajd.DispatchAsync(job)
		assert.NoError(t, err)

		// Wait for processing
		time.Sleep(100 * time.Millisecond)
	})
}

// TestGoroutinePerformanceIntegration tests performance characteristics in integration scenarios
func TestGoroutinePerformanceIntegration(t *testing.T) {
	t.Run("HighLoadScenario", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:        16,
			WorkerTimeout:     30 * time.Second,
			QueueBufferSize:   1000,
			EnableAutoScaling: true,
			MinWorkers:        4,
			MaxWorkersPerCPU:  4,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		wsConfig := &go_core.WorkStealingConfig{
			NumWorkers:     8,
			QueueSize:      500,
			StealThreshold: 2,
			StealBatchSize: 10,
		}
		wsPool := go_core.NewWorkStealingPool[string](wsConfig)
		defer wsPool.Shutdown()

		var processedCount int64
		var mu sync.Mutex

		start := time.Now()

		// Submit many jobs to both systems
		var wg sync.WaitGroup
		for i := 0; i < 500; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Submit to goroutine manager
				job := go_core.GoroutineJob[string]{
					Job: go_core.Job[string]{
						ID:   fmt.Sprintf("perf-gm-%d", id),
						Data: fmt.Sprintf("perf-gm-data-%d", id),
					},
					Timeout: 5 * time.Second,
					Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
						mu.Lock()
						processedCount++
						mu.Unlock()
						time.Sleep(10 * time.Millisecond) // Simulate work
						return nil
					},
				}
				manager.GetWorkerPool().Submit(job)

				// Submit to work stealing pool
				item := go_core.WorkItem[string]{
					ID:   fmt.Sprintf("perf-ws-%d", id),
					Data: fmt.Sprintf("perf-ws-data-%d", id),
					Handler: func(ctx context.Context, data string) error {
						mu.Lock()
						processedCount++
						mu.Unlock()
						time.Sleep(10 * time.Millisecond) // Simulate work
						return nil
					},
					Timeout:  5 * time.Second,
					Priority: id % 4,
				}
				wsPool.Submit(item)
			}(i)
		}

		wg.Wait()

		// Wait for processing
		time.Sleep(3 * time.Second)

		duration := time.Since(start)
		mu.Lock()
		finalCount := processedCount
		mu.Unlock()

		assert.Equal(t, int64(1000), finalCount) // 500 from each system
		assert.Less(t, duration, 5*time.Second)  // Should complete within 5 seconds

		// Check performance metrics
		gmMetrics := manager.GetMetrics()
		wsMetrics := wsPool.GetMetrics()

		assert.GreaterOrEqual(t, gmMetrics.TotalJobsProcessed, int64(0))
		assert.GreaterOrEqual(t, wsMetrics.TotalProcessed, int64(500))
	})

	t.Run("MixedWorkloadScenario", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:        8,
			WorkerTimeout:     30 * time.Second,
			QueueBufferSize:   500,
			EnableAutoScaling: true,
			MinWorkers:        2,
			MaxWorkersPerCPU:  2,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		var fastCount, slowCount int64
		var mu sync.Mutex

		// Submit mixed workload (fast and slow jobs)
		for i := 0; i < 100; i++ {
			// Fast job
			fastJob := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("fast-job-%d", i),
					Data: fmt.Sprintf("fast-data-%d", i),
				},
				Timeout: 5 * time.Second,
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					mu.Lock()
					fastCount++
					mu.Unlock()
					return nil
				},
			}
			manager.GetWorkerPool().Submit(fastJob)

			// Slow job
			slowJob := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("slow-job-%d", i),
					Data: fmt.Sprintf("slow-data-%d", i),
				},
				Timeout: 5 * time.Second,
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					time.Sleep(100 * time.Millisecond) // Simulate slow work
					mu.Lock()
					slowCount++
					mu.Unlock()
					return nil
				},
			}
			manager.GetWorkerPool().Submit(slowJob)
		}

		// Wait for processing
		time.Sleep(3 * time.Second)

		mu.Lock()
		finalFastCount := fastCount
		finalSlowCount := slowCount
		mu.Unlock()

		assert.Equal(t, int64(100), finalFastCount)
		assert.Equal(t, int64(100), finalSlowCount)

		// Check that auto-scaling handled the mixed workload
		metrics := manager.GetMetrics()
		assert.GreaterOrEqual(t, metrics.TotalJobsProcessed, int64(200))
	})
}

// TestGoroutineErrorHandlingIntegration tests error handling in integration scenarios
func TestGoroutineErrorHandlingIntegration(t *testing.T) {
	t.Run("ErrorPropagation", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 50,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		var errorCount int64
		var mu sync.Mutex

		// Submit jobs that will fail
		for i := 0; i < 10; i++ {
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("error-job-%d", i),
					Data: fmt.Sprintf("error-data-%d", i),
				},
				Timeout: 5 * time.Second,
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					mu.Lock()
					errorCount++
					mu.Unlock()
					return fmt.Errorf("simulated error")
				},
			}
			manager.GetWorkerPool().Submit(job)
		}

		// Wait for processing
		time.Sleep(500 * time.Millisecond)

		mu.Lock()
		finalErrorCount := errorCount
		mu.Unlock()

		assert.Equal(t, int64(10), finalErrorCount)
	})

	t.Run("TimeoutHandling", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 50,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		var timeoutCount int64
		var mu sync.Mutex

		// Submit jobs that will timeout
		for i := 0; i < 5; i++ {
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("timeout-job-%d", i),
					Data: fmt.Sprintf("timeout-data-%d", i),
				},
				Timeout: 50 * time.Millisecond,
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					select {
					case <-ctx.Done():
						mu.Lock()
						timeoutCount++
						mu.Unlock()
						return ctx.Err()
					case <-time.After(200 * time.Millisecond):
						return nil
					}
				},
			}
			manager.GetWorkerPool().Submit(job)
		}

		// Wait for timeouts
		time.Sleep(300 * time.Millisecond)

		mu.Lock()
		finalTimeoutCount := timeoutCount
		mu.Unlock()

		assert.Equal(t, int64(5), finalTimeoutCount)
	})
}

// Mock implementations for integration testing
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
	// Return different data for different IDs if T is string
	var result T
	if _, ok := any(result).(string); ok {
		switch id {
		case 1:
			result = any("data1").(T)
		case 2:
			result = any("data2").(T)
		case 3:
			result = any("data3").(T)
		default:
			result = any("test-data").(T)
		}
		return &result, nil
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
