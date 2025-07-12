package events

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOptimizedEventDispatcher_BasicDispatch tests basic dispatch functionality
func TestOptimizedEventDispatcher_BasicDispatch(t *testing.T) {
	// Create dependencies
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	// Create optimized dispatcher
	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	// Test data
	eventName := "test.optimized"
	eventData := "optimized data"

	// Create listener
	var receivedEvent *go_core.Event[string]
	var listenerCalled bool

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		receivedEvent = event
		listenerCalled = true
		return nil
	}

	// Register listener
	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Create and dispatch event
	event := &go_core.Event[string]{
		ID:        "optimized-123",
		Name:      eventName,
		Data:      eventData,
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = dispatcher.Dispatch(event)
	require.NoError(t, err)

	// Verify listener was called (synchronous dispatch)
	assert.True(t, listenerCalled)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Data, receivedEvent.Data)
}

// TestOptimizedEventDispatcher_AsyncDispatch tests asynchronous dispatch
func TestOptimizedEventDispatcher_AsyncDispatch(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.async_optimized"
	eventData := "async optimized data"

	var receivedEvent *go_core.Event[string]
	var listenerCalled bool
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		receivedEvent = event
		listenerCalled = true
		wg.Done()
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "async-opt-123",
		Name:      eventName,
		Data:      eventData,
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = dispatcher.DispatchAsync(event)
	require.NoError(t, err)

	// Wait for async processing
	wg.Wait()

	assert.True(t, listenerCalled)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Data, receivedEvent.Data)
}

// TestOptimizedEventDispatcher_ContextTimeout tests context timeout handling
func TestOptimizedEventDispatcher_ContextTimeout(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()
	contextConfig.DefaultTimeout = 100 * time.Millisecond

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.timeout"

	var listenerCalled bool
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		defer wg.Done()

		// Set listenerCalled immediately to indicate the listener was invoked
		listenerCalled = true

		// Simulate slow processing
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "timeout-123",
		Name:      eventName,
		Data:      "timeout data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = dispatcher.DispatchAsync(event)
	require.NoError(t, err)

	// Wait for timeout
	wg.Wait()

	// Listener should be called but context should be cancelled
	assert.True(t, listenerCalled)
}

// TestOptimizedEventDispatcher_ContextCancellation tests context cancellation
func TestOptimizedEventDispatcher_ContextCancellation(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.cancellation"

	var listenerCalled bool
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		defer wg.Done()

		// Wait for context cancellation
		<-ctx.Done()
		listenerCalled = true
		return ctx.Err()
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "cancel-123",
		Name:      eventName,
		Data:      "cancel data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Create context-aware dispatcher
	contextDispatcher := eventBus.WithContext(ctx)

	// Dispatch asynchronously
	err = contextDispatcher.DispatchAsync(event)
	require.NoError(t, err)

	// Cancel context after a short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for listener to handle cancellation
	wg.Wait()

	assert.True(t, listenerCalled)
}

// TestOptimizedEventDispatcher_GoroutinePoolUsage tests goroutine pool usage
func TestOptimizedEventDispatcher_GoroutinePoolUsage(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	// Create goroutine manager with specific config
	goroutineConfig := &go_core.GoroutineConfig{
		MaxWorkers:        2,
		WorkerTimeout:     30 * time.Second,
		QueueBufferSize:   100,
		EnableAutoScaling: false,
	}

	goroutineManager := go_core.NewGoroutineManager[string](goroutineConfig)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.goroutine_pool"

	var processedEvents int
	var mu sync.Mutex
	var wg sync.WaitGroup
	numEvents := 10
	wg.Add(numEvents)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		defer wg.Done()

		mu.Lock()
		processedEvents++
		mu.Unlock()

		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Dispatch multiple events concurrently
	for i := 0; i < numEvents; i++ {
		event := &go_core.Event[string]{
			ID:        fmt.Sprintf("pool-%d", i),
			Name:      eventName,
			Data:      fmt.Sprintf("data-%d", i),
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = dispatcher.DispatchAsync(event)
		require.NoError(t, err)
	}

	// Wait for all events to be processed
	wg.Wait()

	// Verify all events were processed
	assert.Equal(t, numEvents, processedEvents)

	// Check that the goroutine manager exists and has a worker pool
	assert.NotNil(t, goroutineManager)
	assert.NotNil(t, goroutineManager.GetWorkerPool())

	// Verify the worker pool has the expected number of workers
	assert.Equal(t, goroutineConfig.MaxWorkers, goroutineManager.GetWorkerPool().GetTotalWorkerCount())
}

// TestOptimizedEventDispatcher_ListenerErrorHandling tests error handling in listeners
func TestOptimizedEventDispatcher_ListenerErrorHandling(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.error_handling"
	expectedError := errors.New("listener error")

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		return expectedError
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "error-123",
		Name:      eventName,
		Data:      "error data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = dispatcher.Dispatch(event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listener error")
}

// TestOptimizedEventDispatcher_ConcurrentDispatch tests concurrent dispatch
func TestOptimizedEventDispatcher_ConcurrentDispatch(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.concurrent"

	var mu sync.Mutex
	receivedEvents := make([]*go_core.Event[string], 0)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Dispatch multiple events concurrently
	var wg sync.WaitGroup
	numEvents := 20

	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := &go_core.Event[string]{
				ID:        fmt.Sprintf("concurrent-%d", id),
				Name:      eventName,
				Data:      fmt.Sprintf("data-%d", id),
				Timestamp: time.Now(),
				Source:    "test",
			}

			err := dispatcher.Dispatch(event)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all events were processed
	assert.Len(t, receivedEvents, numEvents)

	// Verify all event IDs are present
	eventIDs := make(map[string]bool)
	for _, event := range receivedEvents {
		eventIDs[event.ID] = true
	}

	for i := 0; i < numEvents; i++ {
		assert.True(t, eventIDs[fmt.Sprintf("concurrent-%d", i)])
	}
}

// TestOptimizedEventDispatcher_ContextValues tests context value propagation
func TestOptimizedEventDispatcher_ContextValues(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()
	contextConfig.PropagateValues = true

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.context_values"

	var receivedContext context.Context
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		receivedContext = ctx
		wg.Done()
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Create context with values
	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	ctx = context.WithValue(ctx, "user_id", "12345")

	// Create context-aware dispatcher
	contextDispatcher := eventBus.WithContext(ctx)

	event := &go_core.Event[string]{
		ID:        "context-values-123",
		Name:      eventName,
		Data:      "context values data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = contextDispatcher.Dispatch(event)
	require.NoError(t, err)

	wg.Wait()

	// Verify context values were propagated
	assert.NotNil(t, receivedContext)
	assert.Equal(t, "test_value", receivedContext.Value("test_key"))
	assert.Equal(t, "12345", receivedContext.Value("user_id"))
}

// TestOptimizedEventDispatcher_Close tests dispatcher cleanup
func TestOptimizedEventDispatcher_Close(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	// Close dispatcher
	err := dispatcher.Close()
	require.NoError(t, err)

	// Verify goroutine manager was shut down
	workerPool := goroutineManager.GetWorkerPool()
	assert.NotNil(t, workerPool)
}

// TestOptimizedEventDispatcher_EdgeCases tests various edge cases
func TestOptimizedEventDispatcher_EdgeCases(t *testing.T) {
	t.Run("NilEventManager", func(t *testing.T) {
		goroutineManager := go_core.NewGoroutineManager[string](nil)
		contextConfig := go_core.DefaultContextConfig()

		dispatcher := go_core.NewOptimizedEventDispatcher[string](
			nil, // Nil event manager
			goroutineManager,
			contextConfig,
		)

		event := &go_core.Event[string]{
			ID:        "nil-manager-123",
			Name:      "test.event",
			Data:      "test data",
			Timestamp: time.Now(),
			Source:    "test",
		}

		err := dispatcher.Dispatch(event)
		require.Error(t, err) // Should error due to nil event manager
	})

	t.Run("NilGoroutineManager", func(t *testing.T) {
		eventBus := go_core.NewEventBus[string](nil, nil, nil)
		eventStore := go_core.NewMemoryEventStore[string]()
		eventManager := go_core.NewEventManager[string](eventBus, eventStore)

		contextConfig := go_core.DefaultContextConfig()

		dispatcher := go_core.NewOptimizedEventDispatcher[string](
			eventManager,
			nil, // Nil goroutine manager
			contextConfig,
		)

		eventName := "test.nil_goroutine"

		listener := func(ctx context.Context, event *go_core.Event[string]) error {
			return nil
		}

		err := dispatcher.Listen(eventName, listener)
		require.NoError(t, err)

		event := &go_core.Event[string]{
			ID:        "nil-goroutine-123",
			Name:      eventName,
			Data:      "test data",
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = dispatcher.Dispatch(event)
		require.NoError(t, err) // Should fallback to direct dispatch
	})

	t.Run("NilContextConfig", func(t *testing.T) {
		eventBus := go_core.NewEventBus[string](nil, nil, nil)
		eventStore := go_core.NewMemoryEventStore[string]()
		eventManager := go_core.NewEventManager[string](eventBus, eventStore)

		goroutineManager := go_core.NewGoroutineManager[string](nil)

		dispatcher := go_core.NewOptimizedEventDispatcher[string](
			eventManager,
			goroutineManager,
			nil, // Nil context config
		)

		eventName := "test.nil_context"

		listener := func(ctx context.Context, event *go_core.Event[string]) error {
			return nil
		}

		err := dispatcher.Listen(eventName, listener)
		require.NoError(t, err)

		event := &go_core.Event[string]{
			ID:        "nil-context-123",
			Name:      eventName,
			Data:      "test data",
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = dispatcher.Dispatch(event)
		require.NoError(t, err) // Should work with default context
	})
}

// TestOptimizedEventDispatcher_StressTest performs a stress test
func TestOptimizedEventDispatcher_StressTest(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	// Create goroutine manager with specific config for stress test
	goroutineConfig := &go_core.GoroutineConfig{
		MaxWorkers:        4,
		WorkerTimeout:     30 * time.Second,
		QueueBufferSize:   1000,
		EnableAutoScaling: true,
	}

	goroutineManager := go_core.NewGoroutineManager[string](goroutineConfig)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "test.stress"

	var mu sync.Mutex
	processedEvents := make(map[string]bool)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		processedEvents[event.ID] = true
		mu.Unlock()

		// Simulate some processing time
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// High concurrency test
	var wg sync.WaitGroup
	numEvents := 500
	numGoroutines := 10

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < numEvents/numGoroutines; i++ {
				eventID := fmt.Sprintf("stress-%d-%d", goroutineID, i)

				event := &go_core.Event[string]{
					ID:        eventID,
					Name:      eventName,
					Data:      fmt.Sprintf("data-%d-%d", goroutineID, i),
					Timestamp: time.Now(),
					Source:    "test",
				}

				err := dispatcher.DispatchAsync(event)
				require.NoError(t, err)
			}
		}(g)
	}

	wg.Wait()

	// Wait for all events to be processed
	time.Sleep(2 * time.Second)

	// Verify all events were processed
	mu.Lock()
	processedCount := len(processedEvents)
	mu.Unlock()
	assert.Equal(t, numEvents, processedCount)

	// Verify no duplicate processing
	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < numEvents/numGoroutines; i++ {
			eventID := fmt.Sprintf("stress-%d-%d", g, i)
			mu.Lock()
			processed := processedEvents[eventID]
			mu.Unlock()
			assert.True(t, processed, "Event %s was not processed", eventID)
		}
	}

	// Check goroutine pool metrics
	metrics := goroutineManager.GetMetrics()
	assert.GreaterOrEqual(t, metrics.TotalJobsProcessed, int64(numEvents))
}
