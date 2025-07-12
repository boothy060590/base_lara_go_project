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

// TestEventBus_BasicDispatch tests basic event dispatch functionality
func TestEventBus_BasicDispatch(t *testing.T) {
	// Create event bus with minimal dependencies
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	// Test data
	eventName := "test.event"
	eventData := "test data"

	// Create listener
	var receivedEvent *go_core.Event[string]
	var listenerCalled bool

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		receivedEvent = event
		listenerCalled = true
		return nil
	}

	// Register listener
	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// Create and dispatch event
	event := &go_core.Event[string]{
		ID:        "test-123",
		Name:      eventName,
		Data:      eventData,
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = eventBus.Dispatch(event)
	require.NoError(t, err)

	// Verify listener was called
	assert.True(t, listenerCalled)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Name, receivedEvent.Name)
	assert.Equal(t, event.Data, receivedEvent.Data)
}

// TestEventBus_AsyncDispatch tests asynchronous event dispatch
func TestEventBus_AsyncDispatch(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.async"
	eventData := "async data"

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

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "async-123",
		Name:      eventName,
		Data:      eventData,
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = eventBus.DispatchAsync(event)
	require.NoError(t, err)

	// Wait for async processing
	wg.Wait()

	assert.True(t, listenerCalled)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Data, receivedEvent.Data)
}

// TestEventBus_MultipleListeners tests multiple listeners for the same event
func TestEventBus_MultipleListeners(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.multiple"
	eventData := "multiple data"

	var listener1Called, listener2Called bool
	var wg sync.WaitGroup
	wg.Add(2)

	listener1 := func(ctx context.Context, event *go_core.Event[string]) error {
		listener1Called = true
		wg.Done()
		return nil
	}

	listener2 := func(ctx context.Context, event *go_core.Event[string]) error {
		listener2Called = true
		wg.Done()
		return nil
	}

	// Register both listeners
	err := eventBus.Listen(eventName, listener1)
	require.NoError(t, err)

	err = eventBus.Listen(eventName, listener2)
	require.NoError(t, err)

	// Verify both listeners are registered
	assert.Equal(t, 2, eventBus.GetListenerCount(eventName))

	event := &go_core.Event[string]{
		ID:        "multiple-123",
		Name:      eventName,
		Data:      eventData,
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = eventBus.Dispatch(event)
	require.NoError(t, err)

	// Wait for both listeners
	wg.Wait()

	assert.True(t, listener1Called)
	assert.True(t, listener2Called)
}

// TestEventBus_NoListeners tests dispatch when no listeners are registered
func TestEventBus_NoListeners(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.none"

	// Verify no listeners
	assert.False(t, eventBus.HasListeners(eventName))
	assert.Equal(t, 0, eventBus.GetListenerCount(eventName))

	event := &go_core.Event[string]{
		ID:        "none-123",
		Name:      eventName,
		Data:      "no listeners",
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Dispatch should succeed even with no listeners
	err := eventBus.Dispatch(event)
	require.NoError(t, err)
}

// TestEventBus_ListenerError tests handling of listener errors
func TestEventBus_ListenerError(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.error"
	expectedError := errors.New("listener error")

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		return expectedError
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "error-123",
		Name:      eventName,
		Data:      "error data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = eventBus.Dispatch(event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listener error")
}

// TestEventBus_MultipleListenerErrors tests multiple listeners with errors
func TestEventBus_MultipleListenerErrors(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.multiple_errors"
	error1 := errors.New("error 1")
	error2 := errors.New("error 2")

	listener1 := func(ctx context.Context, event *go_core.Event[string]) error {
		return error1
	}

	listener2 := func(ctx context.Context, event *go_core.Event[string]) error {
		return error2
	}

	err := eventBus.Listen(eventName, listener1)
	require.NoError(t, err)

	err = eventBus.Listen(eventName, listener2)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "errors-123",
		Name:      eventName,
		Data:      "error data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = eventBus.Dispatch(event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error 1")
	assert.Contains(t, err.Error(), "error 2")
}

// TestEventBus_RemoveListener tests removing event listeners
func TestEventBus_RemoveListener(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.remove"

	var listener1Called, listener2Called bool

	listener1 := func(ctx context.Context, event *go_core.Event[string]) error {
		listener1Called = true
		return nil
	}

	listener2 := func(ctx context.Context, event *go_core.Event[string]) error {
		listener2Called = true
		return nil
	}

	// Register both listeners
	err := eventBus.Listen(eventName, listener1)
	require.NoError(t, err)

	err = eventBus.Listen(eventName, listener2)
	require.NoError(t, err)

	assert.Equal(t, 2, eventBus.GetListenerCount(eventName))

	// Remove first listener
	err = eventBus.RemoveListener(eventName, listener1)
	require.NoError(t, err)

	assert.Equal(t, 1, eventBus.GetListenerCount(eventName))

	// Dispatch event
	event := &go_core.Event[string]{
		ID:        "remove-123",
		Name:      eventName,
		Data:      "remove data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = eventBus.Dispatch(event)
	require.NoError(t, err)

	// Only second listener should be called
	assert.False(t, listener1Called)
	assert.True(t, listener2Called)
}

// TestEventBus_ConcurrentDispatch tests concurrent event dispatch
func TestEventBus_ConcurrentDispatch(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.concurrent"

	var mu sync.Mutex
	receivedEvents := make([]*go_core.Event[string], 0)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
		return nil
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// Dispatch multiple events concurrently
	var wg sync.WaitGroup
	numEvents := 10

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

			err := eventBus.Dispatch(event)
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

// TestEventBus_ConcurrentListeners tests concurrent listener registration
func TestEventBus_ConcurrentListeners(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.concurrent_listeners"

	var wg sync.WaitGroup
	numListeners := 10

	// Register listeners concurrently
	for i := 0; i < numListeners; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			listener := func(ctx context.Context, event *go_core.Event[string]) error {
				return nil
			}

			err := eventBus.Listen(eventName, listener)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all listeners were registered
	assert.Equal(t, numListeners, eventBus.GetListenerCount(eventName))
	assert.True(t, eventBus.HasListeners(eventName))
}

// TestEventBus_PerformanceStats tests performance statistics collection
func TestEventBus_PerformanceStats(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.stats"

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		return nil
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// Dispatch some events
	for i := 0; i < 5; i++ {
		event := &go_core.Event[string]{
			ID:        fmt.Sprintf("stats-%d", i),
			Name:      eventName,
			Data:      fmt.Sprintf("data-%d", i),
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = eventBus.Dispatch(event)
		require.NoError(t, err)
	}

	// Get performance stats
	stats := eventBus.GetPerformanceStats()
	require.NotNil(t, stats)

	// Verify stats contain expected fields
	events, ok := stats["events"].(map[string]interface{})
	require.True(t, ok)

	assert.NotNil(t, events["operations_count"])
	assert.NotNil(t, events["event_pool_size"])
	assert.NotNil(t, events["listener_count"])
}

// TestEventBus_OptimizationStats tests optimization statistics
func TestEventBus_OptimizationStats(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.optimization"

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		return nil
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// Dispatch some events
	for i := 0; i < 3; i++ {
		event := &go_core.Event[string]{
			ID:        fmt.Sprintf("opt-%d", i),
			Name:      eventName,
			Data:      fmt.Sprintf("data-%d", i),
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = eventBus.Dispatch(event)
		require.NoError(t, err)
	}

	// Get optimization stats
	stats := eventBus.GetOptimizationStats()
	require.NotNil(t, stats)

	assert.NotNil(t, stats["atomic_operations"])
	assert.NotNil(t, stats["event_pool_usage"])
	assert.NotNil(t, stats["listener_count"])
}

// TestEventBus_WithContext tests context-aware event dispatch
func TestEventBus_WithContext(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.context"

	var receivedContext context.Context

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		receivedContext = ctx
		return nil
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create context-aware dispatcher
	contextDispatcher := eventBus.WithContext(ctx)

	event := &go_core.Event[string]{
		ID:        "context-123",
		Name:      eventName,
		Data:      "context data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = contextDispatcher.Dispatch(event)
	require.NoError(t, err)

	// Verify context was passed to listener
	assert.NotNil(t, receivedContext)
	assert.Equal(t, ctx, receivedContext)
}

// TestEventBus_ContextCancellation tests context cancellation handling
func TestEventBus_ContextCancellation(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

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

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Create context-aware dispatcher
	contextDispatcher := eventBus.WithContext(ctx)

	event := &go_core.Event[string]{
		ID:        "cancel-123",
		Name:      eventName,
		Data:      "cancel data",
		Timestamp: time.Now(),
		Source:    "test",
	}

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

// TestEventBus_WorkStealingPool tests event dispatch with work stealing pool
func TestEventBus_WorkStealingPool(t *testing.T) {
	// Create work stealing pool
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: 2,
		QueueSize:  100,
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)
	defer wsp.Shutdown()

	eventBus := go_core.NewEventBus[string](wsp, nil, nil)

	eventName := "test.work_stealing"

	var receivedEvent *go_core.Event[string]
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		receivedEvent = event
		wg.Done()
		return nil
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	event := &go_core.Event[string]{
		ID:        "wsp-123",
		Name:      eventName,
		Data:      "wsp data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = eventBus.DispatchAsync(event)
	require.NoError(t, err)

	wg.Wait()

	assert.NotNil(t, receivedEvent)
	assert.Equal(t, event.ID, receivedEvent.ID)
}

// TestEventBus_EdgeCases tests various edge cases
func TestEventBus_EdgeCases(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	t.Run("EmptyEventName", func(t *testing.T) {
		listener := func(ctx context.Context, event *go_core.Event[string]) error {
			return nil
		}

		err := eventBus.Listen("", listener)
		require.NoError(t, err)

		event := &go_core.Event[string]{
			ID:        "empty-123",
			Name:      "",
			Data:      "empty name",
			Timestamp: time.Now(),
			Source:    "test",
		}

		err = eventBus.Dispatch(event)
		require.NoError(t, err)
	})

	t.Run("NilEvent", func(t *testing.T) {
		err := eventBus.Dispatch(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event is nil")
	})

	t.Run("NilListener", func(t *testing.T) {
		err := eventBus.Listen("test.nil", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "listener is nil")
	})

	t.Run("RemoveNonExistentListener", func(t *testing.T) {
		listener := func(ctx context.Context, event *go_core.Event[string]) error {
			return nil
		}

		err := eventBus.RemoveListener("test.nonexistent", listener)
		require.NoError(t, err) // Should not error
	})
}

// TestEventBus_RaceConditions tests for race conditions
func TestEventBus_RaceConditions(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.race"

	var mu sync.Mutex
	counter := 0

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		counter++
		mu.Unlock()
		return nil
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// Concurrent dispatch and listener management
	var wg sync.WaitGroup
	numOperations := 100

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Dispatch event
			event := &go_core.Event[string]{
				ID:        fmt.Sprintf("race-%d", id),
				Name:      eventName,
				Data:      fmt.Sprintf("data-%d", id),
				Timestamp: time.Now(),
				Source:    "test",
			}

			err := eventBus.Dispatch(event)
			require.NoError(t, err)

			// Check listener count
			count := eventBus.GetListenerCount(eventName)
			assert.Equal(t, 1, count)
		}(i)
	}

	wg.Wait()

	// Verify all events were processed
	assert.Equal(t, numOperations, counter)
}

// TestEventBus_StressTest performs a stress test with high concurrency
func TestEventBus_StressTest(t *testing.T) {
	eventBus := go_core.NewEventBus[string](nil, nil, nil)

	eventName := "test.stress"

	var mu sync.Mutex
	processedEvents := make(map[string]bool)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		mu.Lock()
		processedEvents[event.ID] = true
		mu.Unlock()
		return nil
	}

	err := eventBus.Listen(eventName, listener)
	require.NoError(t, err)

	// High concurrency test
	var wg sync.WaitGroup
	numEvents := 1000
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

				err := eventBus.Dispatch(event)
				require.NoError(t, err)
			}
		}(g)
	}

	wg.Wait()

	// Verify all events were processed
	assert.Len(t, processedEvents, numEvents)

	// Verify no duplicate processing
	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < numEvents/numGoroutines; i++ {
			eventID := fmt.Sprintf("stress-%d-%d", g, i)
			assert.True(t, processedEvents[eventID], "Event %s was not processed", eventID)
		}
	}
}
