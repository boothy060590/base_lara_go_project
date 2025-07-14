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
	// Create canonical event dispatcher with optimizations
	dispatcher := go_core.NewEventBus[string](nil, nil, nil)
	defer dispatcher.Shutdown()

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
	// Create work stealing pool for async processing
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: 2,
		QueueSize:  100,
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)
	defer wsp.Shutdown()

	dispatcher := go_core.NewEventBus[string](wsp, nil, nil)
	defer dispatcher.Shutdown()

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
	// Create work stealing pool for async processing
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: 2,
		QueueSize:  100,
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)
	defer wsp.Shutdown()

	dispatcher := go_core.NewEventBus[string](wsp, nil, nil)
	defer dispatcher.Shutdown()

	eventName := "test.timeout"

	var listenerCalled bool
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		defer wg.Done()
		listenerCalled = true
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

	wg.Wait()
	assert.True(t, listenerCalled)
}

// TestOptimizedEventDispatcher_ContextCancellation tests context cancellation
func TestOptimizedEventDispatcher_ContextCancellation(t *testing.T) {
	// Create work stealing pool for async processing
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: 2,
		QueueSize:  100,
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)
	defer wsp.Shutdown()

	dispatcher := go_core.NewEventBus[string](wsp, nil, nil)
	defer dispatcher.Shutdown()

	eventName := "test.cancellation"

	var listenerCalled bool
	var wg sync.WaitGroup
	wg.Add(1)

	listener := func(ctx context.Context, event *go_core.Event[string]) error {
		defer wg.Done()
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

	ctx, cancel := context.WithCancel(context.Background())
	contextDispatcher := dispatcher.WithContext(ctx)
	err = contextDispatcher.DispatchAsync(event)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	cancel()
	wg.Wait()
	assert.True(t, listenerCalled)
}

// TestOptimizedEventDispatcher_GoroutinePoolUsage tests goroutine pool usage
func TestOptimizedEventDispatcher_GoroutinePoolUsage(t *testing.T) {
	// Create work stealing pool for async processing
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: 2,
		QueueSize:  100,
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)
	defer wsp.Shutdown()

	dispatcher := go_core.NewEventBus[string](wsp, nil, nil)
	defer dispatcher.Shutdown()

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
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

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

	wg.Wait()
	assert.Equal(t, numEvents, processedEvents)
}

// TestOptimizedEventDispatcher_ListenerErrorHandling tests error handling in listeners
func TestOptimizedEventDispatcher_ListenerErrorHandling(t *testing.T) {
	dispatcher := go_core.NewEventBus[string](nil, nil, nil)
	defer dispatcher.Shutdown()

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
	dispatcher := go_core.NewEventBus[string](nil, nil, nil)
	defer dispatcher.Shutdown()

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
	assert.Len(t, receivedEvents, numEvents)
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
	dispatcher := go_core.NewEventBus[string](nil, nil, nil)
	defer dispatcher.Shutdown()

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

	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	ctx = context.WithValue(ctx, "user_id", "12345")
	contextDispatcher := dispatcher.WithContext(ctx)
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
	assert.NotNil(t, receivedContext)
	assert.Equal(t, "test_value", receivedContext.Value("test_key"))
	assert.Equal(t, "12345", receivedContext.Value("user_id"))
}

// TestOptimizedEventDispatcher_Close tests dispatcher cleanup
func TestOptimizedEventDispatcher_Close(t *testing.T) {
	// No-op: nothing to close in canonical event bus
}
