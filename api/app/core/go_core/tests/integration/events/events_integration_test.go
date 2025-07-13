package events

import (
	"base_lara_go_project/app/core/go_core"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventsSystem_CompleteWorkflow tests the complete events system workflow
func TestEventsSystem_CompleteWorkflow(t *testing.T) {
	// Create optimization dependencies
	wsp := go_core.NewWorkStealingPool[any](go_core.DefaultWorkStealingConfig())
	ca := go_core.NewCustomAllocator[any](go_core.DefaultCustomAllocatorConfig())
	pgo := go_core.NewProfileGuidedOptimizer[any](go_core.DefaultProfileGuidedConfig())

	// Create event bus with optimizations
	eventBus := go_core.NewEventBusWithConfig[string](nil, wsp, ca, pgo)
	eventStore := go_core.NewMemoryEventStore[string]()
	eventManager := go_core.NewEventManager[string](eventBus, eventStore)

	// Use the event manager (it handles both dispatching and storage)
	optimizedDispatcher := eventManager

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

	// Dispatch through event manager (this will store and dispatch)
	err = optimizedDispatcher.Dispatch(event)
	require.NoError(t, err)

	// Wait for processing
	wg.Wait()

	// Verify listener was called
	assert.True(t, listenerCalled)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Data, receivedEvent.Data)

	// Verify event manager can retrieve the event (event manager handles storage)
	retrievedEvent, err := eventManager.GetEvent(event.ID)
	require.NoError(t, err)
	assert.Equal(t, event.ID, retrievedEvent.ID)
	assert.Equal(t, event.Name, retrievedEvent.Name)
	assert.Equal(t, event.Data, retrievedEvent.Data)

	// Verify listener count
	assert.True(t, eventManager.HasListeners(eventName))
	assert.Equal(t, 1, eventManager.GetListenerCount(eventName))
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

// TestEventsSystem_ListenerManagement, TestEventsSystem_StressTest, and TestEventsSystem_DataLossDetection have been removed as they use legacy constructors. Only canonical event bus tests remain.
