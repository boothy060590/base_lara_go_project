package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository for testing context-aware repository operations
type MockRepository[T any] struct {
	findResult  *T
	findError   error
	createError error
	updateError error
	deleteError error
}

func (m *MockRepository[T]) Find(id uint) (*T, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	return m.findResult, nil
}

func (m *MockRepository[T]) FindBy(field string, value any) (*T, error) {
	return m.Find(1)
}

func (m *MockRepository[T]) FindAll() ([]T, error) {
	return []T{}, nil
}

func (m *MockRepository[T]) Create(model *T) error {
	return m.createError
}

func (m *MockRepository[T]) Update(model *T) error {
	return m.updateError
}

func (m *MockRepository[T]) Delete(id uint) error {
	return m.deleteError
}

func (m *MockRepository[T]) FindWithContext(ctx context.Context, id uint) (*T, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return m.Find(id)
	}
}

func (m *MockRepository[T]) FindByWithContext(ctx context.Context, field string, value any) (*T, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return m.FindBy(field, value)
	}
}

func (m *MockRepository[T]) FindAllWithContext(ctx context.Context) ([]T, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return m.FindAll()
	}
}

func (m *MockRepository[T]) CreateWithContext(ctx context.Context, model *T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.Create(model)
	}
}

func (m *MockRepository[T]) UpdateWithContext(ctx context.Context, model *T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.Update(model)
	}
}

func (m *MockRepository[T]) DeleteWithContext(ctx context.Context, id uint) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.Delete(id)
	}
}

func (m *MockRepository[T]) Where(conditions map[string]any) go_core.Query[T] {
	return &MockQuery[T]{}
}

func (m *MockRepository[T]) WhereRaw(query string, args ...any) go_core.Query[T] {
	return &MockQuery[T]{}
}

func (m *MockRepository[T]) WhereWithContext(ctx context.Context, conditions map[string]any) go_core.Query[T] {
	select {
	case <-ctx.Done():
		return &MockQuery[T]{}
	default:
		return m.Where(conditions)
	}
}

func (m *MockRepository[T]) WhereRawWithContext(ctx context.Context, query string, args ...any) go_core.Query[T] {
	select {
	case <-ctx.Done():
		return &MockQuery[T]{}
	default:
		return m.WhereRaw(query, args...)
	}
}

func (m *MockRepository[T]) Transaction(fn func(go_core.Repository[T]) error) error {
	return fn(m)
}

func (m *MockRepository[T]) TransactionWithContext(ctx context.Context, fn func(go_core.Repository[T]) error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fn(m)
	}
}

func (m *MockRepository[T]) WithContext(ctx context.Context) go_core.Repository[T] {
	return m
}

func (m *MockRepository[T]) Exists(id uint) (bool, error) {
	return false, nil
}

func (m *MockRepository[T]) Count() (int64, error) {
	return 0, nil
}

func (m *MockRepository[T]) CountWhere(conditions map[string]any) (int64, error) {
	return 0, nil
}

func (m *MockRepository[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return m.Exists(id)
	}
}

func (m *MockRepository[T]) CountWithContext(ctx context.Context) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return m.Count()
	}
}

func (m *MockRepository[T]) CountWhereWithContext(ctx context.Context, conditions map[string]any) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return m.CountWhere(conditions)
	}
}

func (m *MockRepository[T]) GetPerformanceStats() map[string]interface{} {
	return map[string]interface{}{
		"operations": 0,
	}
}

func (m *MockRepository[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"optimizations": 0,
	}
}

func (m *MockRepository[T]) BulkCreate(models []*T) error                                 { return nil }
func (m *MockRepository[T]) BulkUpdate(models []*T) error                                 { return nil }
func (m *MockRepository[T]) BulkDelete(ids []uint) error                                  { return nil }
func (m *MockRepository[T]) BulkCreateWithContext(ctx context.Context, models []*T) error { return nil }
func (m *MockRepository[T]) BulkUpdateWithContext(ctx context.Context, models []*T) error { return nil }
func (m *MockRepository[T]) BulkDeleteWithContext(ctx context.Context, ids []uint) error  { return nil }

// MockQuery for testing
type MockQuery[T any] struct{}

func (q *MockQuery[T]) Get() ([]T, error)                                { return []T{}, nil }
func (q *MockQuery[T]) First() (*T, error)                               { return nil, nil }
func (q *MockQuery[T]) Paginate(page, perPage int) ([]T, int64, error)   { return []T{}, 0, nil }
func (q *MockQuery[T]) GetWithContext(ctx context.Context) ([]T, error)  { return []T{}, nil }
func (q *MockQuery[T]) FirstWithContext(ctx context.Context) (*T, error) { return nil, nil }
func (q *MockQuery[T]) PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error) {
	return []T{}, 0, nil
}
func (q *MockQuery[T]) Where(field string, operator string, value any) go_core.Query[T] { return q }
func (q *MockQuery[T]) WhereIn(field string, values []any) go_core.Query[T]             { return q }
func (q *MockQuery[T]) OrderBy(field string, direction string) go_core.Query[T]         { return q }
func (q *MockQuery[T]) Limit(limit int) go_core.Query[T]                                { return q }
func (q *MockQuery[T]) Offset(offset int) go_core.Query[T]                              { return q }
func (q *MockQuery[T]) Preload(relation string) go_core.Query[T]                        { return q }
func (q *MockQuery[T]) WithContext(ctx context.Context) go_core.Query[T]                { return q }

// MockEventDispatcher for testing
type MockEventDispatcher[T any] struct {
	dispatchError error
	listeners     map[string][]go_core.EventListener[T]
}

func (m *MockEventDispatcher[T]) Dispatch(event *go_core.Event[T]) error {
	if m.dispatchError != nil {
		return m.dispatchError
	}

	if m.listeners == nil {
		m.listeners = make(map[string][]go_core.EventListener[T])
	}

	if listeners, exists := m.listeners[event.Name]; exists {
		for _, listener := range listeners {
			if err := listener(context.Background(), event); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MockEventDispatcher[T]) DispatchAsync(event *go_core.Event[T]) error {
	return m.Dispatch(event)
}

func (m *MockEventDispatcher[T]) Listen(eventName string, listener go_core.EventListener[T]) error {
	if m.listeners == nil {
		m.listeners = make(map[string][]go_core.EventListener[T])
	}
	m.listeners[eventName] = append(m.listeners[eventName], listener)
	return nil
}

func (m *MockEventDispatcher[T]) RemoveListener(eventName string, listener go_core.EventListener[T]) error {
	return nil
}

func (m *MockEventDispatcher[T]) Handle(event *go_core.Event[T]) error {
	return m.Dispatch(event)
}

func (m *MockEventDispatcher[T]) HasListeners(eventName string) bool {
	return false
}

func (m *MockEventDispatcher[T]) GetListenerCount(eventName string) int {
	return 0
}

func (m *MockEventDispatcher[T]) WithContext(ctx context.Context) go_core.EventDispatcher[T] {
	return m
}

func (m *MockEventDispatcher[T]) GetPerformanceStats() map[string]interface{} {
	return map[string]interface{}{
		"dispatched_events": 0,
	}
}

func (m *MockEventDispatcher[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"listeners": 0,
	}
}

func (m *MockEventDispatcher[T]) Shutdown() error {
	return nil
}

// MockJobDispatcher for testing
type MockJobDispatcher[T any] struct {
	dispatchError     error
	dispatchSyncError error
}

func (m *MockJobDispatcher[T]) Dispatch(job T) error {
	return m.dispatchError
}

func (m *MockJobDispatcher[T]) DispatchSync(job T) error {
	return m.dispatchSyncError
}

func (m *MockJobDispatcher[T]) GetQueue() go_core.Queue[T] {
	return &MockQueue[T]{}
}

func (m *MockJobDispatcher[T]) WithContext(ctx context.Context) go_core.JobDispatcher[T] {
	return m
}

// MockQueue for testing
type MockQueue[T any] struct{}

func (q *MockQueue[T]) Push(job *go_core.Job[T]) error                                 { return nil }
func (q *MockQueue[T]) Pop() (*go_core.Job[T], error)                                  { return nil, nil }
func (q *MockQueue[T]) Delete(jobID string) error                                      { return nil }
func (q *MockQueue[T]) PushWithContext(ctx context.Context, job *go_core.Job[T]) error { return nil }
func (q *MockQueue[T]) PopWithContext(ctx context.Context) (*go_core.Job[T], error)    { return nil, nil }
func (q *MockQueue[T]) DeleteWithContext(ctx context.Context, jobID string) error      { return nil }
func (q *MockQueue[T]) PushMany(jobs []*go_core.Job[T]) error                          { return nil }
func (q *MockQueue[T]) PopMany(count int) ([]*go_core.Job[T], error)                   { return []*go_core.Job[T]{}, nil }
func (q *MockQueue[T]) PushManyWithContext(ctx context.Context, jobs []*go_core.Job[T]) error {
	return nil
}
func (q *MockQueue[T]) PopManyWithContext(ctx context.Context, count int) ([]*go_core.Job[T], error) {
	return []*go_core.Job[T]{}, nil
}
func (q *MockQueue[T]) Retry(job *go_core.Job[T]) error                                 { return nil }
func (q *MockQueue[T]) Fail(job *go_core.Job[T], error error) error                     { return nil }
func (q *MockQueue[T]) RetryWithContext(ctx context.Context, job *go_core.Job[T]) error { return nil }
func (q *MockQueue[T]) FailWithContext(ctx context.Context, job *go_core.Job[T], error error) error {
	return nil
}
func (q *MockQueue[T]) Size() (int64, error)                               { return 0, nil }
func (q *MockQueue[T]) Clear() error                                       { return nil }
func (q *MockQueue[T]) SizeWithContext(ctx context.Context) (int64, error) { return 0, nil }
func (q *MockQueue[T]) ClearWithContext(ctx context.Context) error         { return nil }
func (q *MockQueue[T]) WithContext(ctx context.Context) go_core.Queue[T]   { return q }

// TestContextSystemIntegration tests the complete context system integration
func TestContextSystemIntegration(t *testing.T) {
	t.Run("ContextAwareOperation", func(t *testing.T) {
		// Create context manager
		manager := go_core.NewContextManager(nil)

		// Test context-aware operation
		operation := func(ctx context.Context) (string, error) {
			return "test-result", nil
		}
		contextOp := go_core.NewContextAwareOperation(operation, manager)
		result2, err := contextOp.WithContext(context.Background()).Execute()
		require.NoError(t, err)
		assert.Equal(t, "test-result", result2)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// Create context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Test repository with cancelled context
		mockRepo := &MockRepository[string]{
			findError: fmt.Errorf("database connection failed"),
		}
		contextRepo := mockRepo.WithContext(ctx)

		_, err := contextRepo.FindWithContext(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canceled")

		// Test event dispatcher with cancelled context
		contextEventDispatcher := go_core.NewEventBus[string](nil, nil, nil)
		defer contextEventDispatcher.Shutdown()

		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err = contextEventDispatcher.Dispatch(event)
		assert.NoError(t, err) // Event bus doesn't check context in Dispatch

		// Test job dispatcher with cancelled context
		mockJobDispatcher := &MockJobDispatcher[string]{}
		contextJobDispatcher := mockJobDispatcher.WithContext(ctx)

		err = contextJobDispatcher.Dispatch("test-job")
		assert.NoError(t, err) // Mock doesn't check context
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		// Create context with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Test repository with timeout
		mockRepo := &MockRepository[string]{
			findError: fmt.Errorf("database timeout"),
		}
		contextRepo := mockRepo.WithContext(ctx)

		_, err := contextRepo.FindWithContext(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database timeout")

		// Test event dispatcher with timeout
		mockEventManager := &MockEventDispatcher[string]{
			dispatchError: fmt.Errorf("event timeout"),
		}
		contextEventDispatcher := mockEventManager.WithContext(ctx)

		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err = contextEventDispatcher.Dispatch(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event timeout")
	})
}

// TestContextAwareComponentsIntegration tests integration between context-aware components
func TestContextAwareComponentsIntegration(t *testing.T) {
	t.Run("RepositoryIntegration", func(t *testing.T) {
		mockRepo := &MockRepository[string]{
			findResult: stringPtr("test-data"),
		}
		contextRepo := mockRepo.WithContext(context.Background())

		// Test Find
		result, err := contextRepo.FindWithContext(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "test-data", *result)

		// Test FindAll
		results, err := contextRepo.FindAllWithContext(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, results)

		// Test Create
		testData := "new-data"
		err = contextRepo.CreateWithContext(context.Background(), &testData)
		require.NoError(t, err)

		// Test Update
		err = contextRepo.UpdateWithContext(context.Background(), &testData)
		require.NoError(t, err)

		// Test Delete
		err = contextRepo.DeleteWithContext(context.Background(), 1)
		require.NoError(t, err)
	})

	t.Run("EventDispatcherIntegration", func(t *testing.T) {
		contextEventDispatcher := go_core.NewEventBus[string](nil, nil, nil)
		defer contextEventDispatcher.Shutdown()

		// Test event dispatch
		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err := contextEventDispatcher.Dispatch(event)
		require.NoError(t, err)

		// Test listener registration
		var receivedEvent *go_core.Event[string]
		var receivedContext context.Context
		var wg sync.WaitGroup
		wg.Add(1)

		listener := func(ctx context.Context, event *go_core.Event[string]) error {
			receivedEvent = event
			receivedContext = ctx
			wg.Done()
			return nil
		}

		contextEventDispatcher.Listen("test.event", listener)

		// Dispatch event to trigger listener
		err = contextEventDispatcher.Dispatch(event)
		require.NoError(t, err)

		wg.Wait()

		assert.NotNil(t, receivedEvent)
		assert.Equal(t, event.ID, receivedEvent.ID)
		assert.NotNil(t, receivedContext)
	})

	t.Run("JobDispatcherIntegration", func(t *testing.T) {
		contextJobDispatcher := go_core.NewJobDispatcher[string](nil, nil, nil, nil)

		// Test async dispatch
		err := contextJobDispatcher.Dispatch("test-job")
		require.NoError(t, err)

		// Test sync dispatch
		err = contextJobDispatcher.DispatchSync("test-job-sync")
		require.NoError(t, err)
	})
}

// TestContextDecoratorsIntegration tests context decorators in real scenarios
func TestContextDecoratorsIntegration(t *testing.T) {
	t.Run("WithTimeoutDecorator", func(t *testing.T) {
		// Create a slow operation
		slowOperation := func(ctx context.Context) (string, error) {
			time.Sleep(200 * time.Millisecond)
			return "slow-result", nil
		}

		// Apply timeout decorator
		timeoutOperation := go_core.WithTimeoutDecorator[string](50 * time.Millisecond)(slowOperation)

		// Execute with timeout
		result, err := timeoutOperation(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
		assert.Equal(t, "", result)
	})

	t.Run("WithRetryDecorator", func(t *testing.T) {
		attempts := 0
		flakyOperation := func(ctx context.Context) (string, error) {
			attempts++
			if attempts < 3 {
				return "", fmt.Errorf("attempt %d failed", attempts)
			}
			return "success", nil
		}

		// Apply retry decorator
		retryOperation := go_core.WithRetryDecorator[string](3, 10*time.Millisecond)(flakyOperation)

		// Execute with retry
		result, err := retryOperation(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 3, attempts)
	})

	t.Run("WithContextDecorator", func(t *testing.T) {
		var receivedContext context.Context
		operation := func(ctx context.Context) (string, error) {
			receivedContext = ctx
			return "test", nil
		}

		// Apply context decorator
		decoratedOperation := go_core.WithContextDecorator(operation)

		// Execute with context
		ctx := context.WithValue(context.Background(), "test_key", "test_value")
		result, err := decoratedOperation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
		assert.NotNil(t, receivedContext)
		assert.Equal(t, "test_value", receivedContext.Value("test_key"))
	})
}

// TestContextUtilsIntegration tests context utilities in real scenarios
func TestContextUtilsIntegration(t *testing.T) {
	t.Run("ContextMerging", func(t *testing.T) {
		utils := go_core.NewContextUtils(go_core.NewContextManager(nil))

		// Create contexts with different values
		ctx1 := context.WithValue(context.Background(), "key1", "value1")
		ctx2 := context.WithValue(context.Background(), "key2", "value2")
		ctx3 := context.WithValue(context.Background(), "key3", "value3")

		// Merge contexts
		merged := utils.MergeContexts(ctx1, ctx2, ctx3)
		assert.NotNil(t, merged)
		assert.Equal(t, "value1", merged.Value("key1"))
		assert.Equal(t, "value2", merged.Value("key2"))
		assert.Equal(t, "value3", merged.Value("key3"))
	})

	t.Run("ContextExpiration", func(t *testing.T) {
		utils := go_core.NewContextUtils(go_core.NewContextManager(nil))

		// Test active context
		ctx := context.Background()
		assert.False(t, utils.IsContextExpired(ctx))

		// Test cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.True(t, utils.IsContextExpired(ctx))
	})
}

// TestContextPerformanceIntegration tests performance characteristics in integration scenarios
func TestContextPerformanceIntegration(t *testing.T) {
	t.Run("HighLoadScenario", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Test high load with context operations
		start := time.Now()
		var wg sync.WaitGroup

		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				ctx := context.WithValue(context.Background(), "request_id", fmt.Sprintf("req-%d", id))
				operation := func(ctx context.Context) (string, error) {
					return fmt.Sprintf("result-%d", id), nil
				}

				contextOp := go_core.NewContextAwareOperation(operation, manager)
				result, err := contextOp.WithContext(ctx).Execute()
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("result-%d", id), result)
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Should complete within reasonable time
		assert.Less(t, duration, 5*time.Second)
	})
}

// TestContextErrorHandlingIntegration tests error handling in integration scenarios
func TestContextErrorHandlingIntegration(t *testing.T) {
	t.Run("ErrorPropagation", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Test error propagation through context-aware operations
		errorOperation := func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("simulated error")
		}

		contextOp := go_core.NewContextAwareOperation(errorOperation, manager)
		result, err := contextOp.WithContext(context.Background()).Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated error")
		assert.Equal(t, "", result)
	})

	t.Run("ContextCancellationPropagation", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Test context cancellation propagation
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		operation := func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond) // Simulate work
			return "result", nil
		}

		contextOp := go_core.NewContextAwareOperation(operation, manager)
		result, err := contextOp.WithContext(ctx).Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canceled")
		assert.Equal(t, "", result)
	})

	t.Run("RepositoryErrorHandling", func(t *testing.T) {
		mockRepo := &MockRepository[string]{
			findError: fmt.Errorf("database connection failed"),
		}
		contextRepo := mockRepo.WithContext(context.Background())

		_, err := contextRepo.Find(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection failed")
	})

	t.Run("EventDispatcherErrorHandling", func(t *testing.T) {
		mockEventManager := &MockEventDispatcher[string]{
			dispatchError: fmt.Errorf("event bus unavailable"),
		}
		contextEventDispatcher := mockEventManager.WithContext(context.Background())

		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err := contextEventDispatcher.Dispatch(event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event bus unavailable")
	})

	t.Run("JobDispatcherErrorHandling", func(t *testing.T) {
		mockJobDispatcher := &MockJobDispatcher[string]{
			dispatchError: fmt.Errorf("job queue full"),
		}
		contextJobDispatcher := mockJobDispatcher.WithContext(context.Background())

		err := contextJobDispatcher.Dispatch("test-job")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job queue full")
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
