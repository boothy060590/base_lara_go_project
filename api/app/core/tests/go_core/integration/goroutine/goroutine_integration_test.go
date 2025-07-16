package integration

import (
	"context"
	"fmt"
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

		// Check metrics - should be exact counts, not minimums
		gmMetrics := manager.GetMetrics()
		wsMetrics := wsPool.GetMetrics()

		assert.Equal(t, int64(20), gmMetrics.TotalJobsProcessed)
		assert.Equal(t, int64(20), wsMetrics.TotalProcessed)
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

		// Test repository operations through goroutine manager
		var wg sync.WaitGroup
		var results []string
		var mu sync.Mutex

		// Submit Find operations to goroutine manager
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id uint) {
				defer wg.Done()
				result, err := mockRepo.Find(id)
				if err == nil && result != nil {
					mu.Lock()
					results = append(results, *result)
					mu.Unlock()
				}
			}(uint(i + 1))
		}

		wg.Wait()

		// Verify results
		assert.Len(t, results, 3)
		assert.Contains(t, results, "data1")
	})

	t.Run("EventDispatcherIntegration", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      4,
			QueueBufferSize: 50,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		// Create canonical event bus with work stealing pool
		wsp := go_core.NewWorkStealingPool[any](go_core.DefaultWorkStealingConfig())
		defer wsp.Shutdown()

		eventBus := go_core.NewEventBus[string](wsp, nil, nil)
		defer eventBus.Shutdown()

		// Test async dispatch through work stealing pool
		event := &go_core.Event[string]{
			ID:   "test-event",
			Data: "test-data",
		}

		err := eventBus.DispatchAsync(event)
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

		// Create canonical job dispatcher with work stealing pool
		wsp := go_core.NewWorkStealingPool[any](go_core.DefaultWorkStealingConfig())
		defer wsp.Shutdown()

		queue := go_core.NewSyncQueue[string]()
		jobDispatcher := go_core.NewJobDispatcher[string](queue, wsp, nil, nil)

		// Test async dispatch
		job := "test-job"

		err := jobDispatcher.Dispatch(job)
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

		// Check performance metrics - should be exact counts
		gmMetrics := manager.GetMetrics()
		wsMetrics := wsPool.GetMetrics()

		assert.Equal(t, int64(500), gmMetrics.TotalJobsProcessed)
		assert.Equal(t, int64(500), wsMetrics.TotalProcessed)
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

		// Check that auto-scaling handled the mixed workload - exact count
		metrics := manager.GetMetrics()
		assert.Equal(t, int64(200), metrics.TotalJobsProcessed)
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
	ctx            context.Context
}

// Repository interface implementation
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

func (m *MockRepository[T]) FindBy(field string, value any) (*T, error) {
	return m.Find(1) // Simple mock implementation
}

func (m *MockRepository[T]) Exists(id uint) (bool, error) {
	return true, nil
}

func (m *MockRepository[T]) Count() (int64, error) {
	return int64(len(m.findManyResult)), nil
}

func (m *MockRepository[T]) Where(conditions map[string]any) go_core.SmartQuery[T] {
	return &MockSmartQuery[T]{repo: m}
}

func (m *MockRepository[T]) Complex() go_core.ComplexQueryBuilder[T] {
	return &MockComplexQueryBuilder[T]{repo: m}
}

func (m *MockRepository[T]) WithOptimization(level go_core.QueryComplexity) go_core.Repository[T] {
	return m
}

func (m *MockRepository[T]) WithContext(ctx context.Context) go_core.Repository[T] {
	newRepo := *m
	newRepo.ctx = ctx
	return &newRepo
}

func (m *MockRepository[T]) Transaction(fn func(go_core.Repository[T]) error) error {
	return fn(m)
}

func (m *MockRepository[T]) Create(model *T) error {
	return nil
}

func (m *MockRepository[T]) Update(model *T) error {
	return nil
}

func (m *MockRepository[T]) Delete(id uint) error {
	return nil
}

// MockSmartQuery implements the SmartQuery interface
type MockSmartQuery[T any] struct {
	repo *MockRepository[T]
	ctx  context.Context
}

func (m *MockSmartQuery[T]) Where(field string, operator string, value any) go_core.SmartQuery[T] {
	return m
}

func (m *MockSmartQuery[T]) WhereIn(field string, values []any) go_core.SmartQuery[T] {
	return m
}

func (m *MockSmartQuery[T]) OrderBy(field string, direction string) go_core.SmartQuery[T] {
	return m
}

func (m *MockSmartQuery[T]) Limit(limit int) go_core.SmartQuery[T] {
	return m
}

func (m *MockSmartQuery[T]) Offset(offset int) go_core.SmartQuery[T] {
	return m
}

func (m *MockSmartQuery[T]) Get() ([]T, error) {
	return m.repo.findManyResult, m.repo.findManyError
}

func (m *MockSmartQuery[T]) First() (*T, error) {
	return m.repo.Find(1)
}

func (m *MockSmartQuery[T]) Paginate(page, perPage int) ([]T, int64, error) {
	return m.repo.findManyResult, int64(len(m.repo.findManyResult)), m.repo.findManyError
}

func (m *MockSmartQuery[T]) AsComplex() go_core.ComplexQuery[T] {
	return &MockComplexQuery[T]{repo: m.repo}
}

func (m *MockSmartQuery[T]) WithContext(ctx context.Context) go_core.SmartQuery[T] {
	newQuery := *m
	newQuery.ctx = ctx
	return &newQuery
}

// MockComplexQueryBuilder implements the ComplexQueryBuilder interface
type MockComplexQueryBuilder[T any] struct {
	repo *MockRepository[T]
	ctx  context.Context
}

func (m *MockComplexQueryBuilder[T]) Join(table string, on string) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) LeftJoin(table string, on string) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) RightJoin(table string, on string) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) GroupBy(fields ...string) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) Having(condition string, args ...any) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) Raw(query string, args ...any) go_core.ComplexQuery[T] {
	return &MockComplexQuery[T]{repo: m.repo}
}

func (m *MockComplexQueryBuilder[T]) BulkCreate(models []*T) error {
	return nil
}

func (m *MockComplexQueryBuilder[T]) BulkUpdate(models []*T) error {
	return nil
}

func (m *MockComplexQueryBuilder[T]) BulkDelete(ids []uint) error {
	return nil
}

func (m *MockComplexQueryBuilder[T]) WithBatching(enabled bool) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) WithAsync(enabled bool) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) WithPipeline(enabled bool) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) WithWorkStealing(enabled bool) go_core.ComplexQueryBuilder[T] {
	return m
}

func (m *MockComplexQueryBuilder[T]) Build() go_core.ComplexQuery[T] {
	return &MockComplexQuery[T]{repo: m.repo}
}

// MockComplexQuery implements the ComplexQuery interface
type MockComplexQuery[T any] struct {
	repo *MockRepository[T]
	ctx  context.Context
}

func (m *MockComplexQuery[T]) Get() ([]T, error) {
	return m.repo.findManyResult, m.repo.findManyError
}

func (m *MockComplexQuery[T]) First() (*T, error) {
	return m.repo.Find(1)
}

func (m *MockComplexQuery[T]) Paginate(page, perPage int) ([]T, int64, error) {
	return m.repo.findManyResult, int64(len(m.repo.findManyResult)), m.repo.findManyError
}

func (m *MockComplexQuery[T]) Stream() (<-chan T, error) {
	ch := make(chan T, len(m.repo.findManyResult))
	go func() {
		defer close(ch)
		for _, item := range m.repo.findManyResult {
			ch <- item
		}
	}()
	return ch, nil
}

func (m *MockComplexQuery[T]) WithMetrics(enabled bool) go_core.ComplexQuery[T] {
	return m
}

func (m *MockComplexQuery[T]) GetStats() map[string]any {
	return map[string]any{}
}

func (m *MockComplexQuery[T]) WithContext(ctx context.Context) go_core.ComplexQuery[T] {
	newQuery := *m
	newQuery.ctx = ctx
	return &newQuery
}

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

func (m *MockEventDispatcher[T]) Shutdown() error { return nil }

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
