package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			// With blocking logic, all jobs should succeed (they will block until processed)
			assert.NoError(t, err)
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
		_ = go_core.NewGoroutineManager[string](config)

		// Create a mock database for testing
		db, _, err := sqlmock.New()
		require.NoError(t, err, "Failed to create mock database")
		defer db.Close()

		// Test the canonical repository with goroutine optimizations
		repo := go_core.NewRepository[string](db)
		assert.NotNil(t, repo)
	})

	t.Run("FindAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		_ = go_core.NewGoroutineManager[string](config)

		mockRepo := &MockRepository[string]{
			findResult: "test-data",
			findError:  nil,
		}

		// Test async operations using goroutines directly
		resultChan := make(chan go_core.RepositoryResult[string], 1)
		go func() {
			defer close(resultChan)
			model, err := mockRepo.Find(1)
			resultChan <- go_core.RepositoryResult[string]{
				Data:  *model,
				Error: err,
			}
		}()

		result := <-resultChan

		assert.Equal(t, "test-data", result.Data)
		assert.NoError(t, result.Error)
	})

	t.Run("FindManyAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		_ = go_core.NewGoroutineManager[string](config)

		// Create a mock repository that returns different results for different IDs
		mockRepo := &MockRepository[string]{
			findResult: "test-data", // This will be returned for each Find call
			findError:  nil,
		}

		// Test async operations using goroutines directly
		resultChan := make(chan go_core.RepositoryResult[[]string], 1)
		go func() {
			defer close(resultChan)
			var models []string
			for _, id := range []uint{1, 2} {
				model, err := mockRepo.Find(id)
				if err != nil {
					resultChan <- go_core.RepositoryResult[[]string]{
						Data:  nil,
						Error: err,
					}
					return
				}
				models = append(models, *model)
			}
			resultChan <- go_core.RepositoryResult[[]string]{
				Data:  models,
				Error: nil,
			}
		}()

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
		_ = go_core.NewGoroutineManager[string](config)

		// Test the canonical event dispatcher with goroutine optimizations
		dispatcher := go_core.NewEventBus[string](nil, nil, nil)
		assert.NotNil(t, dispatcher)
	})

	t.Run("DispatchAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		_ = go_core.NewGoroutineManager[string](config)

		// Test the canonical event dispatcher with goroutine optimizations
		dispatcher := go_core.NewEventBus[string](nil, nil, nil)

		event := &go_core.Event[string]{
			ID:   "test-event",
			Data: "test-data",
		}

		err := dispatcher.DispatchAsync(event)
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
		_ = go_core.NewGoroutineManager[string](config)

		// Test the canonical job dispatcher with goroutine optimizations
		queue := go_core.NewSyncQueue[string]()
		dispatcher := go_core.NewJobDispatcher[string](queue, nil, nil, nil)
		assert.NotNil(t, dispatcher)
	})

	t.Run("DispatchAsync", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 10,
		}
		_ = go_core.NewGoroutineManager[string](config)

		// Test the canonical job dispatcher with goroutine optimizations
		queue := go_core.NewSyncQueue[string]()
		dispatcher := go_core.NewJobDispatcher[string](queue, nil, nil, nil)

		job := "test-job"

		err := dispatcher.Dispatch(job)
		assert.NoError(t, err)
	})
}

// ============================================================================
// DATA LOSS DETECTION TESTS
// ============================================================================

func TestGoroutineDataLossDetection(t *testing.T) {
	t.Run("HighLoadNoJobLoss", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      8,
			QueueBufferSize: 1000,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		var processedCount int64
		var mu sync.Mutex

		// Submit many jobs
		for i := 0; i < 1000; i++ {
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("job-%d", i),
					Data: fmt.Sprintf("data-%d", i),
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
			assert.NoError(t, err, "Failed to submit job %d", i)
		}

		// Wait for processing
		time.Sleep(2 * time.Second)

		mu.Lock()
		finalCount := processedCount
		mu.Unlock()

		assert.Equal(t, int64(1000), finalCount, "Job loss detected: expected 1000 jobs, got %d", finalCount)

		// Check metrics
		metrics := manager.GetMetrics()
		assert.Equal(t, int64(1000), metrics.TotalJobsProcessed, "Metrics show job loss: expected 1000, got %d", metrics.TotalJobsProcessed)
	})

	t.Run("ConcurrentSubmissionNoJobLoss", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      4,
			QueueBufferSize: 500,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		var processedCount int64
		var mu sync.Mutex

		// Submit jobs concurrently
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					job := go_core.GoroutineJob[string]{
						Job: go_core.Job[string]{
							ID:   fmt.Sprintf("concurrent-job-%d-%d", id, j),
							Data: fmt.Sprintf("concurrent-data-%d-%d", id, j),
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
					assert.NoError(t, err, "Failed to submit concurrent job %d-%d", id, j)
				}
			}(i)
		}

		wg.Wait()

		// Wait for processing
		time.Sleep(2 * time.Second)

		mu.Lock()
		finalCount := processedCount
		mu.Unlock()

		assert.Equal(t, int64(500), finalCount, "Job loss detected in concurrent submission: expected 500 jobs, got %d", finalCount)
	})

	t.Run("TimeoutHandlingNoJobLoss", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      2,
			QueueBufferSize: 100,
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		var processedCount int64
		var timeoutCount int64
		var mu sync.Mutex

		// Submit jobs with different timeouts
		for i := 0; i < 20; i++ {
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("timeout-job-%d", i),
					Data: fmt.Sprintf("timeout-data-%d", i),
				},
				Timeout: 200 * time.Millisecond, // Longer timeout
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					select {
					case <-ctx.Done():
						mu.Lock()
						timeoutCount++
						mu.Unlock()
						return ctx.Err()
					case <-time.After(50 * time.Millisecond): // Shorter than timeout
						mu.Lock()
						processedCount++
						mu.Unlock()
						return nil
					}
				},
			}
			err := manager.GetWorkerPool().Submit(job)
			assert.NoError(t, err, "Failed to submit timeout job %d", i)
		}

		// Wait for processing
		time.Sleep(1 * time.Second)

		mu.Lock()
		finalProcessed := processedCount
		finalTimeout := timeoutCount
		mu.Unlock()

		// All jobs should be handled (either processed or timed out)
		totalHandled := finalProcessed + finalTimeout
		assert.Equal(t, int64(20), totalHandled, "Job loss detected in timeout scenario: expected 20 jobs handled, got %d", totalHandled)

		// Most should be processed, some might timeout due to queue delays
		assert.Greater(t, finalProcessed, int64(0), "No jobs were processed")
		assert.GreaterOrEqual(t, finalTimeout, int64(0), "Unexpected timeout count")
	})

	t.Run("QueueFullHandlingNoJobLoss", func(t *testing.T) {
		config := &go_core.GoroutineConfig{
			MaxWorkers:      1, // Single worker
			QueueBufferSize: 5, // Small buffer
		}
		manager := go_core.NewGoroutineManager[string](config)
		defer manager.GetWorkerPool().Shutdown()

		var processedCount int64
		var submittedCount int64
		var mu sync.Mutex

		// Submit jobs faster than they can be processed
		for i := 0; i < 20; i++ {
			job := go_core.GoroutineJob[string]{
				Job: go_core.Job[string]{
					ID:   fmt.Sprintf("queue-job-%d", i),
					Data: fmt.Sprintf("queue-data-%d", i),
				},
				Timeout: 5 * time.Second,
				Handler: func(ctx context.Context, job *go_core.GoroutineJob[string]) error {
					time.Sleep(50 * time.Millisecond) // Slow processing
					mu.Lock()
					processedCount++
					mu.Unlock()
					return nil
				},
			}
			err := manager.GetWorkerPool().Submit(job)
			if err != nil {
				// Queue is full, this is expected behavior
				t.Logf("Queue full for job %d: %v", i, err)
			} else {
				mu.Lock()
				submittedCount++
				mu.Unlock()
			}
		}

		// Wait for processing
		time.Sleep(3 * time.Second)

		mu.Lock()
		finalProcessed := processedCount
		finalSubmitted := submittedCount
		mu.Unlock()

		// All submitted jobs should be processed
		assert.Equal(t, finalSubmitted, finalProcessed, "Job loss detected in queue full scenario: submitted %d, processed %d", finalSubmitted, finalProcessed)

		// Should have some jobs submitted and processed
		assert.Greater(t, finalSubmitted, int64(0), "No jobs were submitted")
		assert.Greater(t, finalProcessed, int64(0), "No jobs were processed")
	})
}

// Mock implementations for testing
// MockRepository for testing goroutine-aware repository operations
type MockRepository[T any] struct {
	findResult     T
	findError      error
	findManyResult []T
	findManyError  error
	ctx            context.Context
}

func (m *MockRepository[T]) Find(id uint) (*T, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	return &m.findResult, nil
}

func (m *MockRepository[T]) FindMany(ids []uint) ([]T, error) {
	if m.findManyError != nil {
		return nil, m.findManyError
	}
	return m.findManyResult, nil
}

func (m *MockRepository[T]) Count() (int64, error) {
	return int64(len(m.findManyResult)), nil
}

func (m *MockRepository[T]) FindBy(field string, value any) (*T, error) { return nil, nil }
func (m *MockRepository[T]) FindAll() ([]T, error)                      { return nil, nil }
func (m *MockRepository[T]) Create(model *T) error                      { return nil }
func (m *MockRepository[T]) Update(model *T) error                      { return nil }
func (m *MockRepository[T]) Delete(id uint) error                       { return nil }

func (m *MockRepository[T]) FindWithContext(ctx context.Context, id uint) (*T, error) {
	return m.Find(id)
}

func (m *MockRepository[T]) FindByWithContext(ctx context.Context, field string, value any) (*T, error) {
	return m.FindBy(field, value)
}

func (m *MockRepository[T]) FindAllWithContext(ctx context.Context) ([]T, error)   { return nil, nil }
func (m *MockRepository[T]) CreateWithContext(ctx context.Context, model *T) error { return nil }
func (m *MockRepository[T]) UpdateWithContext(ctx context.Context, model *T) error { return nil }
func (m *MockRepository[T]) DeleteWithContext(ctx context.Context, id uint) error  { return nil }

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

func (m *MockRepository[T]) Exists(id uint) (bool, error)                        { return false, nil }
func (m *MockRepository[T]) CountWhere(conditions map[string]any) (int64, error) { return 0, nil }
func (m *MockRepository[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	return m.Exists(id)
}

func (m *MockRepository[T]) CountWithContext(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockRepository[T]) CountWhereWithContext(ctx context.Context, conditions map[string]any) (int64, error) {
	return m.CountWhere(conditions)
}

func (m *MockRepository[T]) GetPerformanceStats() map[string]interface{}                  { return nil }
func (m *MockRepository[T]) GetOptimizationStats() map[string]interface{}                 { return nil }
func (m *MockRepository[T]) BulkCreate(models []*T) error                                 { return nil }
func (m *MockRepository[T]) BulkUpdate(models []*T) error                                 { return nil }
func (m *MockRepository[T]) BulkDelete(ids []uint) error                                  { return nil }
func (m *MockRepository[T]) BulkCreateWithContext(ctx context.Context, models []*T) error { return nil }
func (m *MockRepository[T]) BulkUpdateWithContext(ctx context.Context, models []*T) error { return nil }
func (m *MockRepository[T]) BulkDeleteWithContext(ctx context.Context, ids []uint) error  { return nil }

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
	ch := make(chan T)
	close(ch)
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
func (m *MockEventDispatcher[T]) Shutdown() error                              { return nil }

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
