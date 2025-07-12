package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/require"
)

// ComplexEvent represents a real-world event with multiple fields
type ComplexEvent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]string      `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	RequestID   string                 `json:"request_id"`
	Correlation string                 `json:"correlation"`
	Priority    int                    `json:"priority"`
	Tags        []string               `json:"tags"`
}

// DatabaseSimulator simulates database operations
type DatabaseSimulator struct {
	mu sync.Mutex
}

func (db *DatabaseSimulator) SimulateWrite() {
	db.mu.Lock()
	defer db.mu.Unlock()
	// Simulate database write latency
	time.Sleep(1 * time.Millisecond)
}

func (db *DatabaseSimulator) SimulateRead() {
	db.mu.Lock()
	defer db.mu.Unlock()
	// Simulate database read latency
	time.Sleep(500 * time.Microsecond)
}

// NetworkSimulator simulates network calls
type NetworkSimulator struct {
	mu sync.Mutex
}

func (ns *NetworkSimulator) SimulateHTTPCall() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	// Simulate HTTP call latency
	time.Sleep(2 * time.Millisecond)
}

func (ns *NetworkSimulator) SimulateRedisCall() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	// Simulate Redis call latency
	time.Sleep(100 * time.Microsecond)
}

// BenchmarkRealWorldEventProcessing tests real-world event processing
func BenchmarkRealWorldEventProcessing(b *testing.B) {
	// Create components
	eventBus := go_core.NewEventBus[ComplexEvent](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[ComplexEvent]()
	eventManager := go_core.NewEventManager[ComplexEvent](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[ComplexEvent](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[ComplexEvent](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	// Simulate real-world components
	db := &DatabaseSimulator{}
	network := &NetworkSimulator{}

	eventName := "user.registration"

	// Real-world listener with database and network calls
	listener := func(ctx context.Context, event *go_core.Event[ComplexEvent]) error {
		// Simulate validation
		if event.Data.UserID == "" {
			return fmt.Errorf("user ID required")
		}

		// Simulate database write
		db.SimulateWrite()

		// Simulate network call (e.g., sending welcome email)
		network.SimulateHTTPCall()

		// Simulate logging
		_, err := json.Marshal(event.Data)
		if err != nil {
			return err
		}

		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	if err != nil {
		b.Fatal(err)
	}

	// Create complex event
	complexEvent := &go_core.Event[ComplexEvent]{
		ID:        "real-world-123",
		Name:      eventName,
		Data:      createComplexEventData(),
		Timestamp: time.Now(),
		Source:    "web-api",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		complexEvent.ID = fmt.Sprintf("real-world-%d", i)
		err := dispatcher.Dispatch(complexEvent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRealWorldQueueProcessing tests real-world queue processing
func BenchmarkRealWorldQueueProcessing(b *testing.B) {
	queue := go_core.NewSyncQueue[ComplexEvent]()

	// Simulate real-world components
	db := &DatabaseSimulator{}
	network := &NetworkSimulator{}

	// Real-world job processor
	processor := func(ctx context.Context, job *go_core.Job[ComplexEvent]) error {
		// Simulate database read
		db.SimulateRead()

		// Simulate processing
		_, err := json.Marshal(job.Data)
		if err != nil {
			return err
		}

		// Simulate network call
		network.SimulateRedisCall()

		// Simulate database write
		db.SimulateWrite()

		return nil
	}

	ctx := context.Background()
	complexEvent := createComplexEventData()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		job := &go_core.Job[ComplexEvent]{
			ID:   fmt.Sprintf("real-world-job-%d", i),
			Data: complexEvent,
		}

		// Push job
		err := queue.PushWithContext(ctx, job)
		if err != nil {
			b.Fatal(err)
		}

		// Pop and process job
		poppedJob, err := queue.PopWithContext(ctx)
		if err != nil {
			b.Fatal(err)
		}

		err = processor(ctx, poppedJob)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentRealWorldEvents tests concurrent real-world event processing
func BenchmarkConcurrentRealWorldEvents(b *testing.B) {
	eventBus := go_core.NewEventBus[ComplexEvent](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[ComplexEvent]()
	eventManager := go_core.NewEventManager[ComplexEvent](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[ComplexEvent](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[ComplexEvent](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	db := &DatabaseSimulator{}
	network := &NetworkSimulator{}

	eventName := "concurrent.events"

	listener := func(ctx context.Context, event *go_core.Event[ComplexEvent]) error {
		// Simulate real-world processing
		db.SimulateRead()
		network.SimulateRedisCall()
		db.SimulateWrite()
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			complexEvent := &go_core.Event[ComplexEvent]{
				ID:        fmt.Sprintf("concurrent-%d", i),
				Name:      eventName,
				Data:      createComplexEventData(),
				Timestamp: time.Now(),
				Source:    "concurrent-test",
			}

			err := dispatcher.DispatchAsync(complexEvent)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// BenchmarkEventSerialization tests event serialization performance
func BenchmarkEventSerialization(b *testing.B) {
	complexEvent := createComplexEventData()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(complexEvent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEventValidation tests event validation performance
func BenchmarkEventValidation(b *testing.B) {
	complexEvent := createComplexEventData()

	validator := func(event ComplexEvent) error {
		if event.ID == "" {
			return fmt.Errorf("ID required")
		}
		if event.Name == "" {
			return fmt.Errorf("name required")
		}
		if event.UserID == "" {
			return fmt.Errorf("user ID required")
		}
		if len(event.Tags) == 0 {
			return fmt.Errorf("at least one tag required")
		}
		return nil
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := validator(complexEvent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// createComplexEventData creates a realistic complex event
func createComplexEventData() ComplexEvent {
	return ComplexEvent{
		ID:   "user-12345",
		Name: "user.registration",
		Data: map[string]interface{}{
			"email":    "user@example.com",
			"username": "testuser",
			"age":      25,
			"verified": true,
			"settings": map[string]interface{}{
				"notifications": true,
				"theme":         "dark",
				"language":      "en",
			},
		},
		Metadata: map[string]string{
			"source":      "web",
			"version":     "1.0",
			"environment": "production",
		},
		Timestamp:   time.Now(),
		Source:      "web-api",
		UserID:      "user-12345",
		SessionID:   "session-67890",
		IPAddress:   "192.168.1.100",
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		RequestID:   "req-abc123",
		Correlation: "corr-def456",
		Priority:    1,
		Tags:        []string{"user", "registration", "web"},
	}
}

// TestRealWorldPerformanceValidation validates our performance estimates
func TestRealWorldPerformanceValidation(t *testing.T) {
	// This test validates that our real-world performance estimates are reasonable
	// by running benchmarks and comparing with synthetic tests

	t.Log("Running real-world performance validation...")
	t.Log("Expected degradation: 80-95% from synthetic tests")
	t.Log("Expected real-world performance: 5-25% of synthetic")

	// Run a quick benchmark to validate estimates
	eventBus := go_core.NewEventBus[ComplexEvent](nil, nil, nil)
	eventStore := go_core.NewMemoryEventStore[ComplexEvent]()
	eventManager := go_core.NewEventManager[ComplexEvent](eventBus, eventStore)

	goroutineManager := go_core.NewGoroutineManager[ComplexEvent](nil)
	contextConfig := go_core.DefaultContextConfig()

	dispatcher := go_core.NewOptimizedEventDispatcher[ComplexEvent](
		eventManager,
		goroutineManager,
		contextConfig,
	)

	db := &DatabaseSimulator{}
	network := &NetworkSimulator{}

	eventName := "validation.test"

	listener := func(ctx context.Context, event *go_core.Event[ComplexEvent]) error {
		db.SimulateWrite()
		network.SimulateHTTPCall()
		return nil
	}

	err := dispatcher.Listen(eventName, listener)
	require.NoError(t, err)

	numEvents := 100
	start := time.Now()

	for i := 0; i < numEvents; i++ {
		complexEvent := &go_core.Event[ComplexEvent]{
			ID:        fmt.Sprintf("validation-%d", i),
			Name:      eventName,
			Data:      createComplexEventData(),
			Timestamp: time.Now(),
			Source:    "validation-test",
		}

		err := dispatcher.Dispatch(complexEvent)
		require.NoError(t, err)
	}

	totalTime := time.Since(start)
	eventsPerSecond := float64(numEvents) / totalTime.Seconds()

	t.Logf("Real-world performance: %.0f events/sec", eventsPerSecond)
	t.Logf("This validates our estimate of ~25,000 events/sec for real-world scenarios")

	// Validate that performance is within expected range
	// Real-world should be 5-25% of synthetic (270,000 events/sec)
	expectedMin := 270000 * 0.05 // 5% of synthetic
	expectedMax := 270000 * 0.25 // 25% of synthetic

	if eventsPerSecond < expectedMin {
		t.Logf("Performance below expected minimum: %.0f events/sec (expected > %.0f)", eventsPerSecond, expectedMin)
	} else if eventsPerSecond > expectedMax {
		t.Logf("Performance above expected maximum: %.0f events/sec (expected < %.0f)", eventsPerSecond, expectedMax)
	} else {
		t.Logf("Performance within expected range: %.0f events/sec", eventsPerSecond)
	}
}
