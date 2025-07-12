package events

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventStore_BasicOperations tests basic event store operations
func TestEventStore_BasicOperations(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	// Test data
	event := &go_core.Event[string]{
		ID:        "test-123",
		Name:      "test.event",
		Data:      "test data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Store event
	err := store.Store(event)
	require.NoError(t, err)

	// Retrieve event
	retrieved, err := store.Get(event.ID)
	require.NoError(t, err)
	assert.Equal(t, event.ID, retrieved.ID)
	assert.Equal(t, event.Name, retrieved.Name)
	assert.Equal(t, event.Data, retrieved.Data)

	// Verify count
	count, err := store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// TestEventStore_StoreMany tests storing multiple events
func TestEventStore_StoreMany(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	events := []*go_core.Event[string]{
		{
			ID:        "test-1",
			Name:      "test.event1",
			Data:      "data 1",
			Timestamp: time.Now(),
			Source:    "test",
		},
		{
			ID:        "test-2",
			Name:      "test.event2",
			Data:      "data 2",
			Timestamp: time.Now(),
			Source:    "test",
		},
		{
			ID:        "test-3",
			Name:      "test.event1", // Same name as first
			Data:      "data 3",
			Timestamp: time.Now(),
			Source:    "test",
		},
	}

	// Store multiple events
	err := store.StoreMany(events)
	require.NoError(t, err)

	// Verify all events were stored
	count, err := store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Verify individual events
	for _, event := range events {
		retrieved, err := store.Get(event.ID)
		require.NoError(t, err)
		assert.Equal(t, event.ID, retrieved.ID)
		assert.Equal(t, event.Data, retrieved.Data)
	}
}

// TestEventStore_GetByEventName tests retrieving events by name
func TestEventStore_GetByEventName(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	// Store events with different names
	events := []*go_core.Event[string]{
		{
			ID:        "test-1",
			Name:      "user.created",
			Data:      "user 1",
			Timestamp: time.Now(),
			Source:    "test",
		},
		{
			ID:        "test-2",
			Name:      "user.updated",
			Data:      "user 2",
			Timestamp: time.Now(),
			Source:    "test",
		},
		{
			ID:        "test-3",
			Name:      "user.created",
			Data:      "user 3",
			Timestamp: time.Now(),
			Source:    "test",
		},
	}

	err := store.StoreMany(events)
	require.NoError(t, err)

	// Get events by name
	userCreated, err := store.GetByEventName("user.created", 10)
	require.NoError(t, err)
	assert.Len(t, userCreated, 2)

	userUpdated, err := store.GetByEventName("user.updated", 10)
	require.NoError(t, err)
	assert.Len(t, userUpdated, 1)

	// Test limit
	limited, err := store.GetByEventName("user.created", 1)
	require.NoError(t, err)
	assert.Len(t, limited, 1)

	// Test non-existent event name
	nonexistent, err := store.GetByEventName("nonexistent", 10)
	require.NoError(t, err)
	assert.Len(t, nonexistent, 0)
}

// TestEventStore_GetByTimeRange tests retrieving events by time range
func TestEventStore_GetByTimeRange(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	now := time.Now()

	events := []*go_core.Event[string]{
		{
			ID:        "test-1",
			Name:      "test.event",
			Data:      "old event",
			Timestamp: now.Add(-2 * time.Hour),
			Source:    "test",
		},
		{
			ID:        "test-2",
			Name:      "test.event",
			Data:      "recent event",
			Timestamp: now.Add(-30 * time.Minute),
			Source:    "test",
		},
		{
			ID:        "test-3",
			Name:      "test.event",
			Data:      "very recent event",
			Timestamp: now,
			Source:    "test",
		},
	}

	err := store.StoreMany(events)
	require.NoError(t, err)

	// Get events in last hour
	recent, err := store.GetByTimeRange(now.Add(-1*time.Hour), now)
	require.NoError(t, err)
	assert.Len(t, recent, 2)

	// Get events in last 3 hours
	all, err := store.GetByTimeRange(now.Add(-3*time.Hour), now)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Get events in future (should be empty)
	future, err := store.GetByTimeRange(now.Add(1*time.Hour), now.Add(2*time.Hour))
	require.NoError(t, err)
	assert.Len(t, future, 0)
}

// TestEventStore_Delete tests event deletion
func TestEventStore_Delete(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	event := &go_core.Event[string]{
		ID:        "test-123",
		Name:      "test.event",
		Data:      "test data",
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Store event
	err := store.Store(event)
	require.NoError(t, err)

	// Verify event exists
	count, err := store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete event
	err = store.Delete(event.ID)
	require.NoError(t, err)

	// Verify event was deleted
	count, err = store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Try to get deleted event
	_, err = store.Get(event.ID)
	require.Error(t, err)
}

// TestEventStore_Clear tests clearing all events
func TestEventStore_Clear(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	events := []*go_core.Event[string]{
		{
			ID:        "test-1",
			Name:      "test.event1",
			Data:      "data 1",
			Timestamp: time.Now(),
			Source:    "test",
		},
		{
			ID:        "test-2",
			Name:      "test.event2",
			Data:      "data 2",
			Timestamp: time.Now(),
			Source:    "test",
		},
	}

	// Store events
	err := store.StoreMany(events)
	require.NoError(t, err)

	// Verify events were stored
	count, err := store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Clear all events
	err = store.Clear()
	require.NoError(t, err)

	// Verify all events were cleared
	count, err = store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Verify individual events are gone
	for _, event := range events {
		_, err := store.Get(event.ID)
		require.Error(t, err)
	}
}

// TestEventStore_CountByEventName tests counting events by name
func TestEventStore_CountByEventName(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	events := []*go_core.Event[string]{
		{
			ID:        "test-1",
			Name:      "user.created",
			Data:      "user 1",
			Timestamp: time.Now(),
			Source:    "test",
		},
		{
			ID:        "test-2",
			Name:      "user.updated",
			Data:      "user 2",
			Timestamp: time.Now(),
			Source:    "test",
		},
		{
			ID:        "test-3",
			Name:      "user.created",
			Data:      "user 3",
			Timestamp: time.Now(),
			Source:    "test",
		},
	}

	err := store.StoreMany(events)
	require.NoError(t, err)

	// Count by event name
	userCreatedCount, err := store.CountByEventName("user.created")
	require.NoError(t, err)
	assert.Equal(t, int64(2), userCreatedCount)

	userUpdatedCount, err := store.CountByEventName("user.updated")
	require.NoError(t, err)
	assert.Equal(t, int64(1), userUpdatedCount)

	// Count non-existent event name
	nonexistentCount, err := store.CountByEventName("nonexistent")
	require.NoError(t, err)
	assert.Equal(t, int64(0), nonexistentCount)
}

// TestEventStore_EdgeCases tests various edge cases
func TestEventStore_EdgeCases(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	t.Run("StoreNilEvent", func(t *testing.T) {
		err := store.Store(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event is nil")
	})

	t.Run("StoreEmptyID", func(t *testing.T) {
		event := &go_core.Event[string]{
			ID:        "",
			Name:      "test.event",
			Data:      "test data",
			Timestamp: time.Now(),
			Source:    "test",
		}

		err := store.Store(event)
		require.NoError(t, err) // Should allow empty ID
	})

	t.Run("GetNonExistentEvent", func(t *testing.T) {
		_, err := store.Get("nonexistent")
		require.Error(t, err)
	})

	t.Run("DeleteNonExistentEvent", func(t *testing.T) {
		err := store.Delete("nonexistent")
		require.NoError(t, err) // Should not error
	})

	t.Run("StoreManyWithNil", func(t *testing.T) {
		// Create a fresh store for this test
		freshStore := go_core.NewMemoryEventStore[string]()

		events := []*go_core.Event[string]{
			{
				ID:        "test-1",
				Name:      "test.event",
				Data:      "data 1",
				Timestamp: time.Now(),
				Source:    "test",
			},
			nil, // Nil event
			{
				ID:        "test-2",
				Name:      "test.event",
				Data:      "data 2",
				Timestamp: time.Now(),
				Source:    "test",
			},
		}

		err := freshStore.StoreMany(events)
		require.NoError(t, err) // Should skip nil events silently

		// Verify only non-nil events were stored
		count, err := freshStore.Count()
		require.NoError(t, err)
		assert.Equal(t, int64(2), count) // Only 2 non-nil events should be stored
	})

	t.Run("EmptyStoreOperations", func(t *testing.T) {
		// Create a fresh store for this test
		freshStore := go_core.NewMemoryEventStore[string]()

		// Test operations on empty store
		count, err := freshStore.Count()
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		events, err := freshStore.GetByEventName("test", 10)
		require.NoError(t, err)
		assert.Len(t, events, 0)

		events, err = freshStore.GetByTimeRange(time.Now(), time.Now().Add(time.Hour))
		require.NoError(t, err)
		assert.Len(t, events, 0)
	})
}

// TestEventStore_Concurrency tests concurrent operations
func TestEventStore_Concurrency(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	// Concurrent store operations
	var wg sync.WaitGroup
	numEvents := 100

	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := &go_core.Event[string]{
				ID:        fmt.Sprintf("concurrent-%d", id),
				Name:      "test.concurrent",
				Data:      fmt.Sprintf("data-%d", id),
				Timestamp: time.Now(),
				Source:    "test",
			}

			err := store.Store(event)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all events were stored
	count, err := store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(numEvents), count)

	// Concurrent read operations
	var readWg sync.WaitGroup
	for i := 0; i < numEvents; i++ {
		readWg.Add(1)
		go func(id int) {
			defer readWg.Done()

			eventID := fmt.Sprintf("concurrent-%d", id)
			event, err := store.Get(eventID)
			require.NoError(t, err)
			assert.Equal(t, eventID, event.ID)
		}(i)
	}

	readWg.Wait()
}

// TestEventStore_DuplicateID tests handling of duplicate event IDs
func TestEventStore_DuplicateID(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	event1 := &go_core.Event[string]{
		ID:        "duplicate-123",
		Name:      "test.event1",
		Data:      "data 1",
		Timestamp: time.Now(),
		Source:    "test",
	}

	event2 := &go_core.Event[string]{
		ID:        "duplicate-123", // Same ID
		Name:      "test.event2",
		Data:      "data 2",
		Timestamp: time.Now(),
		Source:    "test",
	}

	// Store first event
	err := store.Store(event1)
	require.NoError(t, err)

	// Store second event with same ID (should overwrite)
	err = store.Store(event2)
	require.NoError(t, err)

	// Verify only one event exists
	count, err := store.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify the second event overwrote the first
	retrieved, err := store.Get("duplicate-123")
	require.NoError(t, err)
	assert.Equal(t, event2.Name, retrieved.Name)
	assert.Equal(t, event2.Data, retrieved.Data)
}

// TestEventStore_Performance tests basic performance characteristics
func TestEventStore_Performance(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	// Store many events
	numEvents := 1000
	events := make([]*go_core.Event[string], numEvents)

	for i := 0; i < numEvents; i++ {
		events[i] = &go_core.Event[string]{
			ID:        fmt.Sprintf("perf-%d", i),
			Name:      fmt.Sprintf("test.event.%d", i%10), // 10 different event names
			Data:      fmt.Sprintf("data-%d", i),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Source:    "test",
		}
	}

	// Measure store performance
	start := time.Now()
	err := store.StoreMany(events)
	storeTime := time.Since(start)

	require.NoError(t, err)
	t.Logf("Stored %d events in %v", numEvents, storeTime)

	// Measure retrieval performance
	start = time.Now()
	for i := 0; i < numEvents; i++ {
		_, err := store.Get(fmt.Sprintf("perf-%d", i))
		require.NoError(t, err)
	}
	retrieveTime := time.Since(start)

	t.Logf("Retrieved %d events in %v", numEvents, retrieveTime)

	// Measure count performance
	start = time.Now()
	count, err := store.Count()
	countTime := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, int64(numEvents), count)
	t.Logf("Counted %d events in %v", numEvents, countTime)
}

// TestEventStore_TimeRangeEdgeCases tests time range edge cases
func TestEventStore_TimeRangeEdgeCases(t *testing.T) {
	store := go_core.NewMemoryEventStore[string]()

	now := time.Now()

	events := []*go_core.Event[string]{
		{
			ID:        "test-1",
			Name:      "test.event",
			Data:      "exact start",
			Timestamp: now,
			Source:    "test",
		},
		{
			ID:        "test-2",
			Name:      "test.event",
			Data:      "exact end",
			Timestamp: now.Add(time.Hour),
			Source:    "test",
		},
		{
			ID:        "test-3",
			Name:      "test.event",
			Data:      "between",
			Timestamp: now.Add(30 * time.Minute),
			Source:    "test",
		},
	}

	err := store.StoreMany(events)
	require.NoError(t, err)

	// Test exact time range
	exact, err := store.GetByTimeRange(now, now.Add(time.Hour))
	require.NoError(t, err)
	assert.Len(t, exact, 3)

	// Test reversed time range (should be empty)
	reversed, err := store.GetByTimeRange(now.Add(time.Hour), now)
	require.NoError(t, err)
	assert.Len(t, reversed, 0)

	// Test zero time range
	zero, err := store.GetByTimeRange(now, now)
	require.NoError(t, err)
	assert.Len(t, zero, 1) // Should include exact match
}
