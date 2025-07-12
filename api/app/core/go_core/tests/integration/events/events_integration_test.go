package events

import (
	"base_lara_go_project/app/core/go_core"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventsSystem_CompleteWorkflow tests the complete events system workflow
func TestEventsSystem_CompleteWorkflow(t *testing.T) {
	// Create all components
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	// Test data
	eventName := "user.created"
	eventData := "user data"

	// Create listener
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

	// Register listener
	err := optimizedDispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Create and dispatch event
	event := &go_core.Event[string]{
		ID:        "user-123",
		Name:      eventName,
		Data:      eventData,
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Dispatch through optimized dispatcher
	err = optimizedDispatcher.Dispatch(event)
	require.NoError(t, err)

	// Wait for processing
	wg.Wait()

	// Verify listener was called
	assert.True(t, listenerCalled)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Data, receivedEvent.Data)

	// Verify event was stored
	storedEvent, err := eventStore.Get(event.ID)
	require.NoError(t, err)
	assert.Equal(t, event.ID, storedEvent.ID)
	assert.Equal(t, event.Name, storedEvent.Name)
	assert.Equal(t, event.Data, storedEvent.Data)

	// Verify event manager can retrieve the event
	retrievedEvent, err := eventManager.GetEvent(event.ID)
	require.NoError(t, err)
	assert.Equal(t, event.ID, retrievedEvent.ID)

	// Verify listener count
	assert.True(t, eventManager.HasListeners(eventName))
	assert.Equal(t, 1, eventManager.GetListenerCount(eventName))
}

// TestEventsSystem_AsyncWorkflow tests asynchronous event processing
func TestEventsSystem_AsyncWorkflow(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "user.updated"

	var processedEvents []*go_core.Event[string]
	var mu sync.Mutex
	var wg sync.WaitGroup
	numEvents := 5
	wg.Add(numEvents)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		processedEvents = append(processedEvents, event)
		mu.Unlock()
		wg.Done()
		return nil
	}

	err := optimizedDispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Dispatch multiple events asynchronously
	for i := 0; i < numEvents; i++ {
		event := &go_core.Event[string]{
			ID:        fmt.Sprintf("user-%d", i),
			Name:      eventName,
			Data:      fmt.Sprintf("user data %d", i),
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = optimizedDispatcher.DispatchAsync(event)
		require.NoError(t, err)
	}

	// Wait for all events to be processed
	wg.Wait()

	// Verify all events were processed
	assert.Len(t, processedEvents, numEvents)

	// Verify all events were stored
	for i := 0; i < numEvents; i++ {
		eventID := fmt.Sprintf("user-%d", i)
		storedEvent, err := eventStore.Get(eventID)
		require.NoError(t, err)
		assert.Equal(t, eventID, storedEvent.ID)
	}

	// Verify events can be retrieved by name
	storedEvents, err := eventManager.GetEventsByName(eventName, 10)
	require.NoError(t, err)
	assert.Len(t, storedEvents, numEvents)
}

// TestEventsSystem_ContextIntegration tests context integration across components
func TestEventsSystem_ContextIntegration(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()
	contextConfig.DefaultTimeout = 500 * time.Millisecond

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "context.test"

	var receivedContext context.Context
	var listenerCalled bool
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		receivedContext = ctx
		listenerCalled = true
		wg.Done()
		return nil
	}

	err := optimizedDispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Create context-aware event bus
	contextBus := eventBus.WithContext(ctx)

	event := &go_core.Event[string]{
		ID:        "context-123",
		Name:      eventName,
		Data:      "context data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = contextBus.Dispatch(event)
	require.NoError(t, err)

	wg.Wait()

	// Verify context was propagated
	assert.True(t, listenerCalled)
	assert.NotNil(t, receivedContext)
	assert.Equal(t, ctx, receivedContext)
}

// TestEventsSystem_ErrorHandling tests error handling across the system
func TestEventsSystem_ErrorHandling(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "error.test"
	expectedError := errors.New("listener error")

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		return expectedError
	}

	err := optimizedDispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "error-123",
		Name:      eventName,
		Data:      "error data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Dispatch should return error
	err = optimizedDispatcher.Dispatch(event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listener error")

	// Event should still be stored despite listener error
	storedEvent, err := eventStore.Get(event.ID)
	require.NoError(t, err)
	assert.Equal(t, event.ID, storedEvent.ID)
}

// TestEventsSystem_ConcurrentOperations tests concurrent operations across components
func TestEventsSystem_ConcurrentOperations(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "concurrent.test"

	var mu sync.Mutex
	processedEvents := make(map[string]bool)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		processedEvents[event.ID] = true
		mu.Unlock()
		return nil
	}

	err := optimizedDispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Concurrent dispatch and retrieval
	var wg sync.WaitGroup
	numEvents := 50

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

			// Dispatch event
			err := optimizedDispatcher.Dispatch(event)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Wait a bit for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify all events were processed
	mu.Lock()
	processedCount := len(processedEvents)
	mu.Unlock()
	assert.Equal(t, numEvents, processedCount)

	// Verify all events are in store
	for i := 0; i < numEvents; i++ {
		eventID := fmt.Sprintf("concurrent-%d", i)
		mu.Lock()
		processed := processedEvents[eventID]
		mu.Unlock()
		assert.True(t, processed, "Event %s was not processed", eventID)

		stored, err := eventStore.Get(eventID)
		require.NoError(t, err)
		assert.Equal(t, eventID, stored.ID)
	}
}

// TestEventsSystem_PerformanceOptimizations tests performance optimizations
func TestEventsSystem_PerformanceOptimizations(t *testing.T) {
	// Create work stealing pool for performance
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: 4,
		QueueSize:  100,
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)
	defer wsp.Shutdown()

	eventBus := go_core.NewEventBus[string](wsp, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "performance.test"

	var processedCount int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	numEvents := 100
	wg.Add(numEvents)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		processedCount++
		mu.Unlock()
		wg.Done()
		return nil
	}

	err := optimizedDispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// Measure dispatch performance
	start := time.Now()

	for i := 0; i < numEvents; i++ {
		event := &go_core.Event[string]{
			ID:        fmt.Sprintf("perf-%d", i),
			Name:      eventName,
			Data:      fmt.Sprintf("data-%d", i),
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = optimizedDispatcher.DispatchAsync(event)
		require.NoError(t, err)
	}

	wg.Wait()
	dispatchTime := time.Since(start)

	t.Logf("Dispatched %d events in %v", numEvents, dispatchTime)

	// Verify all events were processed
	assert.Equal(t, int64(numEvents), processedCount)

	// Check performance stats
	stats := eventBus.GetPerformanceStats()
	assert.NotNil(t, stats)

	optimizationStats := eventBus.GetOptimizationStats()
	assert.NotNil(t, optimizationStats)
}

// TestEventsSystem_EventRetrieval tests event retrieval across components
func TestEventsSystem_EventRetrieval(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	// Create events with different names and timestamps
	now := time.Now()
	events := []*go_core.Event[string]{
		{
			ID:        "retrieval-1",
			Name:      "user.created",
			Data:      "user 1",
			Timestamp: now.Add(-2 * time.Hour),
			Source:    "test",
		},
		{
			ID:        "retrieval-2",
			Name:      "user.updated",
			Data:      "user 2",
			Timestamp: now.Add(-1 * time.Hour),
			Source:    "test",
		},
		{
			ID:        "retrieval-3",
			Name:      "user.created",
			Data:      "user 3",
			Timestamp: now,
			Source:    "test",
		},
	}

	// Store events directly
	for _, event := range events {
		err := eventStore.Store(event)
		require.NoError(t, err)
	}

	// Wait a bit for events to be available in store
	time.Sleep(50 * time.Millisecond)

	// Test retrieval by name
	userCreated, err := eventManager.GetEventsByName("user.created", 10)
	require.NoError(t, err)
	assert.Len(t, userCreated, 2)

	userUpdated, err := eventManager.GetEventsByName("user.updated", 10)
	require.NoError(t, err)
	assert.Len(t, userUpdated, 1)

	// Test retrieval by time range
	recent, err := eventManager.GetEventsByTimeRange(now.Add(-30*time.Minute), now)
	require.NoError(t, err)
	assert.Len(t, recent, 1)

	// Test individual event retrieval
	for _, event := range events {
		retrieved, err := eventManager.GetEvent(event.ID)
		require.NoError(t, err)
		assert.Equal(t, event.ID, retrieved.ID)
		assert.Equal(t, event.Name, retrieved.Name)
		assert.Equal(t, event.Data, retrieved.Data)
	}
}

// TestEventsSystem_ListenerManagement tests listener management across components
func TestEventsSystem_ListenerManagement(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[string](nil)
	contextConfig := go_core.DefaultContextConfig()

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "listener.test"

	var mu sync.Mutex
	var listener1Called, listener2Called bool

	listener1 := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		listener1Called = true
		mu.Unlock()
		return nil
	}

	listener2 := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		listener2Called = true
		mu.Unlock()
		return nil
	}

	// Register listeners
	err := optimizedDispatcher.Listen(eventName, listener1)
	require.NoError(t, err)

	err = optimizedDispatcher.Listen(eventName, listener2)
	require.NoError(t, err)

	// Verify listeners are registered
	assert.True(t, eventManager.HasListeners(eventName))
	assert.Equal(t, 2, eventManager.GetListenerCount(eventName))

	// Dispatch event
	event := &go_core.Event[string]{
		ID:        "listener-123",
		Name:      eventName,
		Data:      "listener data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = optimizedDispatcher.Dispatch(event)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Verify both listeners were called
	mu.Lock()
	l1Called := listener1Called
	l2Called := listener2Called
	mu.Unlock()

	assert.True(t, l1Called)
	assert.True(t, l2Called)

	// Remove one listener
	err = optimizedDispatcher.RemoveListener(eventName, listener1)
	require.NoError(t, err)

	// Wait a bit for removal to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify listener count decreased
	assert.Equal(t, 1, eventManager.GetListenerCount(eventName))

	// Reset flags
	mu.Lock()
	listener1Called = false
	listener2Called = false
	mu.Unlock()

	// Dispatch another event
	event2 := &go_core.Event[string]{
		ID:        "listener-456",
		Name:      eventName,
		Data:      "listener data 2",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = optimizedDispatcher.Dispatch(event2)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Only second listener should be called
	mu.Lock()
	l1Called = listener1Called
	l2Called = listener2Called
	mu.Unlock()

	assert.False(t, l1Called)
	assert.True(t, l2Called)
}

// TestEventsSystem_StressTest performs a comprehensive stress test
func TestEventsSystem_StressTest(t *testing.T) {
	// Create components with high-performance configuration
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: 8,
		QueueSize:  1000,
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)
	defer wsp.Shutdown()

	eventBus := go_core.NewEventBus[string](wsp, nil, nil)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	goroutineConfig := &go_core.GoroutineConfig{
		MaxWorkers:        8,
		WorkerTimeout:     30 * time.Second,
		QueueBufferSize:   1000,
		EnableAutoScaling: true,
	}

	goroutineManager := go_core.NewGoroutineManager[string](goroutineConfig)
	contextConfig := go_core.DefaultContextConfig()

	optimizedDispatcher := go_core.NewOptimizedEventDispatcher[string](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	eventName := "stress.test"

	var mu sync.Mutex
	processedEvents := make(map[string]bool)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		processedEvents[event.ID] = true
		mu.Unlock()

		// Simulate processing time
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	err := optimizedDispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	// High concurrency test
	var wg sync.WaitGroup
	numEvents := 1000
	numGoroutines := 20

	start := time.Now()

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

				err := optimizedDispatcher.DispatchAsync(event)
				require.NoError(t, err)
			}
		}(g)
	}

	wg.Wait()

	// Wait for all events to be processed
	time.Sleep(3 * time.Second)

	totalTime := time.Since(start)
	t.Logf("Processed %d events in %v", numEvents, totalTime)

	// Verify all events were processed
	mu.Lock()
	processedCount := len(processedEvents)
	mu.Unlock()
	assert.Equal(t, numEvents, processedCount)

	// Verify all events are in store
	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < numEvents/numGoroutines; i++ {
			eventID := fmt.Sprintf("stress-%d-%d", g, i)
			mu.Lock()
			processed := processedEvents[eventID]
			mu.Unlock()
			assert.True(t, processed, "Event %s was not processed", eventID)

			stored, err := eventStore.Get(eventID)
			require.NoError(t, err)
			assert.Equal(t, eventID, stored.ID)
		}
	}

	// Check performance metrics
	stats := eventBus.GetPerformanceStats()
	assert.NotNil(t, stats)

	metrics := goroutineManager.GetMetrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.TotalJobsProcessed, int64(numEvents))
}
