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

// MockRepository for testing context-aware repository
type MockRepository[T any] struct {
	findResult  *T
	findError   error
	createError error
	updateError error
	deleteError error
}

func (m *MockRepository[T]) Find(id uint) (*T, error) {
	return m.findResult, m.findError
}

func (m *MockRepository[T]) FindBy(field string, value any) (*T, error) {
	return m.findResult, m.findError
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

// Context-aware methods
func (m *MockRepository[T]) FindWithContext(ctx context.Context, id uint) (*T, error) {
	return m.Find(id)
}

func (m *MockRepository[T]) FindByWithContext(ctx context.Context, field string, value any) (*T, error) {
	return m.FindBy(field, value)
}

func (m *MockRepository[T]) FindAllWithContext(ctx context.Context) ([]T, error) {
	return m.FindAll()
}

func (m *MockRepository[T]) CreateWithContext(ctx context.Context, model *T) error {
	return m.Create(model)
}

func (m *MockRepository[T]) UpdateWithContext(ctx context.Context, model *T) error {
	return m.Update(model)
}

func (m *MockRepository[T]) DeleteWithContext(ctx context.Context, id uint) error {
	return m.Delete(id)
}

// Query methods
func (m *MockRepository[T]) Where(conditions map[string]any) go_core.Query[T] {
	return &MockQuery[T]{}
}

func (m *MockRepository[T]) WhereRaw(query string, args ...any) go_core.Query[T] {
	return &MockQuery[T]{}
}

func (m *MockRepository[T]) WhereWithContext(ctx context.Context, conditions map[string]any) go_core.Query[T] {
	return &MockQuery[T]{}
}

func (m *MockRepository[T]) WhereRawWithContext(ctx context.Context, query string, args ...any) go_core.Query[T] {
	return &MockQuery[T]{}
}

// Transaction methods
func (m *MockRepository[T]) Transaction(fn func(go_core.Repository[T]) error) error {
	return fn(m)
}

func (m *MockRepository[T]) TransactionWithContext(ctx context.Context, fn func(go_core.Repository[T]) error) error {
	return fn(m)
}

func (m *MockRepository[T]) WithContext(ctx context.Context) go_core.Repository[T] {
	return m
}

// Utility methods
func (m *MockRepository[T]) Exists(id uint) (bool, error) {
	return m.findResult != nil, m.findError
}

func (m *MockRepository[T]) Count() (int64, error) {
	return 0, nil
}

func (m *MockRepository[T]) CountWhere(conditions map[string]any) (int64, error) {
	return 0, nil
}

func (m *MockRepository[T]) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	return m.Exists(id)
}

func (m *MockRepository[T]) CountWithContext(ctx context.Context) (int64, error) {
	return m.Count()
}

func (m *MockRepository[T]) CountWhereWithContext(ctx context.Context, conditions map[string]any) (int64, error) {
	return m.CountWhere(conditions)
}

// Performance methods
func (m *MockRepository[T]) GetPerformanceStats() map[string]interface{} {
	return map[string]interface{}{}
}

func (m *MockRepository[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{}
}

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

// MockEventManager for testing context-aware event dispatcher
type MockEventManager[T any] struct {
	dispatchError error
	listeners     map[string][]go_core.EventListener[T]
}

func (m *MockEventManager[T]) Dispatch(event *go_core.Event[T]) error {
	// Invoke listeners for this event
	if listeners, exists := m.listeners[event.Name]; exists {
		for _, listener := range listeners {
			// Create a background context for the listener
			ctx := context.Background()
			if err := listener(ctx, event); err != nil {
				return err
			}
		}
	}
	return m.dispatchError
}

func (m *MockEventManager[T]) DispatchAsync(event *go_core.Event[T]) error {
	return m.dispatchError
}

func (m *MockEventManager[T]) Listen(eventName string, listener go_core.EventListener[T]) error {
	if m.listeners == nil {
		m.listeners = make(map[string][]go_core.EventListener[T])
	}
	m.listeners[eventName] = append(m.listeners[eventName], listener)
	return nil
}

func (m *MockEventManager[T]) GetEvent(eventID string) (*go_core.Event[T], error) {
	return nil, nil
}

func (m *MockEventManager[T]) GetEventsByName(name string, limit int) ([]*go_core.Event[T], error) {
	return []*go_core.Event[T]{}, nil
}

func (m *MockEventManager[T]) GetEventsByTimeRange(start, end time.Time) ([]*go_core.Event[T], error) {
	return []*go_core.Event[T]{}, nil
}

func (m *MockEventManager[T]) HasListeners(eventName string) bool {
	return len(m.listeners[eventName]) > 0
}

func (m *MockEventManager[T]) GetListenerCount(eventName string) int {
	return len(m.listeners[eventName])
}

func (m *MockEventManager[T]) RemoveListener(eventName string, listener go_core.EventListener[T]) error {
	return nil
}

// MockJobDispatcher for testing context-aware job dispatcher
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

// TestContextSystemIntegration tests complete context system integration
func TestContextSystemIntegration(t *testing.T) {
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Create context manager with custom config
		config := &go_core.ContextConfig{
			DefaultTimeout:     1 * time.Second,
			MaxTimeout:         5 * time.Second,
			EnableDeadline:     true,
			EnableCancellation: true,
			PropagateValues:    true,
		}
		manager := go_core.NewContextManager(config)

		// Create context-aware components
		mockRepo := &MockRepository[string]{
			findResult: stringPtr("test-data"),
		}
		contextRepo := go_core.NewContextAwareRepository(mockRepo, manager)

		mockEventManager := &MockEventManager[string]{}
		contextEventDispatcher := go_core.NewContextAwareEventDispatcher(mockEventManager)

		mockJobDispatcher := &MockJobDispatcher[string]{}
		contextJobDispatcher := go_core.NewContextAwareJobDispatcher(mockJobDispatcher)

		// Test complete workflow with context
		ctx := context.WithValue(context.Background(), "request_id", "test-123")
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		// Repository operation
		result, err := contextRepo.Find(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "test-data", *result)

		// Event dispatch
		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err = contextEventDispatcher.Dispatch(ctx, event)
		require.NoError(t, err)

		// Job dispatch
		err = contextJobDispatcher.Dispatch(ctx, "test-job")
		require.NoError(t, err)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Create context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Test repository with cancelled context
		mockRepo := &MockRepository[string]{}
		contextRepo := go_core.NewContextAwareRepository(mockRepo, manager)

		_, err := contextRepo.Find(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canceled")

		// Test event dispatcher with cancelled context
		mockEventManager := &MockEventManager[string]{}
		contextEventDispatcher := go_core.NewContextAwareEventDispatcher(mockEventManager)

		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err = contextEventDispatcher.Dispatch(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canceled")

		// Test job dispatcher with cancelled context
		mockJobDispatcher := &MockJobDispatcher[string]{}
		contextJobDispatcher := go_core.NewContextAwareJobDispatcher(mockJobDispatcher)

		err = contextJobDispatcher.Dispatch(ctx, "test-job")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canceled")
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Create context with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Test repository with timeout
		mockRepo := &MockRepository[string]{
			findError: fmt.Errorf("database timeout"),
		}
		contextRepo := go_core.NewContextAwareRepository(mockRepo, manager)

		_, err := contextRepo.Find(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database timeout")

		// Test event dispatcher with timeout
		mockEventManager := &MockEventManager[string]{
			dispatchError: fmt.Errorf("event timeout"),
		}
		contextEventDispatcher := go_core.NewContextAwareEventDispatcher(mockEventManager)

		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err = contextEventDispatcher.Dispatch(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event timeout")
	})
}

// TestContextAwareComponentsIntegration tests integration between context-aware components
func TestContextAwareComponentsIntegration(t *testing.T) {
	t.Run("RepositoryIntegration", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		mockRepo := &MockRepository[string]{
			findResult: stringPtr("test-data"),
		}
		contextRepo := go_core.NewContextAwareRepository(mockRepo, manager)

		ctx := context.Background()

		// Test Find
		result, err := contextRepo.Find(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "test-data", *result)

		// Test FindAll
		results, err := contextRepo.FindAll(ctx)
		require.NoError(t, err)
		assert.NotNil(t, results)

		// Test Create
		testData := "new-data"
		err = contextRepo.Create(ctx, &testData)
		require.NoError(t, err)

		// Test Update
		err = contextRepo.Update(ctx, &testData)
		require.NoError(t, err)

		// Test Delete
		err = contextRepo.Delete(ctx, 1)
		require.NoError(t, err)
	})

	t.Run("EventDispatcherIntegration", func(t *testing.T) {
		mockEventManager := &MockEventManager[string]{}
		contextEventDispatcher := go_core.NewContextAwareEventDispatcher(mockEventManager)

		ctx := context.Background()

		// Test event dispatch
		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err := contextEventDispatcher.Dispatch(ctx, event)
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
		err = contextEventDispatcher.Dispatch(ctx, event)
		require.NoError(t, err)

		wg.Wait()

		assert.NotNil(t, receivedEvent)
		assert.Equal(t, event.ID, receivedEvent.ID)
		assert.NotNil(t, receivedContext)
	})

	t.Run("JobDispatcherIntegration", func(t *testing.T) {
		mockJobDispatcher := &MockJobDispatcher[string]{}
		contextJobDispatcher := go_core.NewContextAwareJobDispatcher(mockJobDispatcher)

		ctx := context.Background()

		// Test async dispatch
		err := contextJobDispatcher.Dispatch(ctx, "test-job")
		require.NoError(t, err)

		// Test sync dispatch
		err = contextJobDispatcher.DispatchSync(ctx, "test-job-sync")
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

		// Test timed out context
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond)
		assert.True(t, utils.IsContextExpired(ctx))
	})

	t.Run("ContextTimeoutDetection", func(t *testing.T) {
		utils := go_core.NewContextUtils(go_core.NewContextManager(nil))

		// Test context without timeout
		ctx := context.Background()
		timeout, ok := utils.GetContextTimeout(ctx)
		assert.False(t, ok)
		assert.Equal(t, time.Duration(0), timeout)

		// Test context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		timeout, ok = utils.GetContextTimeout(ctx)
		assert.True(t, ok)
		assert.True(t, timeout > 0)
		assert.True(t, timeout <= 5*time.Second)
	})
}

// TestContextPerformanceIntegration tests context performance under load
func TestContextPerformanceIntegration(t *testing.T) {
	t.Run("ConcurrentContextOperations", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		var wg sync.WaitGroup
		results := make([]string, 100)
		errors := make([]error, 100)

		start := time.Now()

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				operation := func(ctx context.Context) (string, error) {
					time.Sleep(1 * time.Millisecond)
					return fmt.Sprintf("result-%d", index), nil
				}

				cao := go_core.NewContextAwareOperation(operation, manager)
				cao = cao.WithTimeout(100 * time.Millisecond)

				result, err := cao.Execute()
				results[index] = result
				errors[index] = err
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// All operations should succeed
		successCount := 0
		for i := 0; i < 100; i++ {
			if errors[i] == nil {
				successCount++
			}
		}

		assert.Equal(t, 100, successCount)
		assert.True(t, duration < 5*time.Second, "Operations should complete within 5 seconds")
	})

	t.Run("ContextAwareCachePerformance", func(t *testing.T) {
		cache := go_core.NewLocalCache[string]()
		manager := go_core.NewContextManager(nil)
		contextCache := go_core.NewContextAwareCache(cache, manager)

		ctx := context.Background()
		start := time.Now()

		// Perform many cache operations
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)

			err := contextCache.Set(ctx, key, &value, time.Minute)
			require.NoError(t, err)

			retrieved, err := contextCache.Get(ctx, key)
			require.NoError(t, err)
			assert.Equal(t, value, *retrieved)
		}

		duration := time.Since(start)
		assert.True(t, duration < 2*time.Second, "Cache operations should complete within 2 seconds")
	})
}

// TestContextErrorHandlingIntegration tests error handling in context scenarios
func TestContextErrorHandlingIntegration(t *testing.T) {
	t.Run("RepositoryErrorHandling", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		mockRepo := &MockRepository[string]{
			findError: fmt.Errorf("database connection failed"),
		}
		contextRepo := go_core.NewContextAwareRepository(mockRepo, manager)

		ctx := context.Background()
		_, err := contextRepo.Find(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection failed")
	})

	t.Run("EventDispatcherErrorHandling", func(t *testing.T) {
		mockEventManager := &MockEventManager[string]{
			dispatchError: fmt.Errorf("event bus unavailable"),
		}
		contextEventDispatcher := go_core.NewContextAwareEventDispatcher(mockEventManager)

		ctx := context.Background()
		event := &go_core.Event[string]{
			ID:        "test-event",
			Name:      "test.event",
			Data:      "test-data",
			Timestamp: time.Now(),
			Source:    "test",
		}
		err := contextEventDispatcher.Dispatch(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event bus unavailable")
	})

	t.Run("JobDispatcherErrorHandling", func(t *testing.T) {
		mockJobDispatcher := &MockJobDispatcher[string]{
			dispatchError: fmt.Errorf("job queue full"),
		}
		contextJobDispatcher := go_core.NewContextAwareJobDispatcher(mockJobDispatcher)

		ctx := context.Background()
		err := contextJobDispatcher.Dispatch(ctx, "test-job")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job queue full")
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
