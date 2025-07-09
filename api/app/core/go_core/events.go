package go_core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Event represents a generic event that can be dispatched
type Event[T any] struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Data      T         `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// EventListener defines a function that handles an event
type EventListener[T any] func(ctx context.Context, event *Event[T]) error

// EventDispatcher defines a generic event dispatcher interface
type EventDispatcher[T any] interface {
	// Event management
	Dispatch(event *Event[T]) error
	DispatchAsync(event *Event[T]) error

	// Listener management
	Listen(eventName string, listener EventListener[T]) error
	RemoveListener(eventName string, listener EventListener[T]) error

	// Event handling
	Handle(event *Event[T]) error

	// Utility operations
	HasListeners(eventName string) bool
	GetListenerCount(eventName string) int
	WithContext(ctx context.Context) EventDispatcher[T]

	// Performance operations
	GetPerformanceStats() map[string]interface{}
	GetOptimizationStats() map[string]interface{}
}

// EventBus implements EventDispatcher[T] with in-memory storage and performance optimizations
type EventBus[T any] struct {
	listeners map[string][]EventListener[T]
	mu        sync.RWMutex
	ctx       context.Context
	// Performance optimizations (safe for event operations)
	atomicCounter     *AtomicCounter
	eventPool         *ObjectPool[Event[T]]
	performanceFacade *PerformanceFacade
}

// NewEventBus creates a new event bus instance with performance optimizations
func NewEventBus[T any]() EventDispatcher[T] {
	// Create performance optimizations
	atomicCounter := NewAtomicCounter()
	performanceFacade := NewPerformanceFacade()

	// Create object pool for event objects (safe - no database state)
	eventPool := NewObjectPool[Event[T]](100,
		func() Event[T] { return Event[T]{} },
		func(event Event[T]) Event[T] { return Event[T]{} },
	)

	return &EventBus[T]{
		listeners:         make(map[string][]EventListener[T]),
		ctx:               context.Background(),
		atomicCounter:     atomicCounter,
		eventPool:         eventPool,
		performanceFacade: performanceFacade,
	}
}

// Dispatch dispatches an event synchronously with performance tracking and atomic counter
func (e *EventBus[T]) Dispatch(event *Event[T]) error {
	// Track operation count atomically
	e.atomicCounter.Increment()

	return e.performanceFacade.Track("event.dispatch", func() error {
		return e.Handle(event)
	})
}

// DispatchAsync dispatches an event asynchronously with performance tracking and atomic counter
func (e *EventBus[T]) DispatchAsync(event *Event[T]) error {
	// Track operation count atomically
	e.atomicCounter.Increment()

	return e.performanceFacade.Track("event.dispatch_async", func() error {
		go func() {
			_ = e.Handle(event)
		}()
		return nil
	})
}

// Listen registers an event listener
func (e *EventBus[T]) Listen(eventName string, listener EventListener[T]) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.listeners[eventName] == nil {
		e.listeners[eventName] = make([]EventListener[T], 0)
	}

	e.listeners[eventName] = append(e.listeners[eventName], listener)
	return nil
}

// RemoveListener removes an event listener
func (e *EventBus[T]) RemoveListener(eventName string, listener EventListener[T]) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	listeners, exists := e.listeners[eventName]
	if !exists {
		return nil
	}

	// Find and remove the listener
	for i, l := range listeners {
		if fmt.Sprintf("%p", l) == fmt.Sprintf("%p", listener) {
			e.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}

	return nil
}

// Handle processes an event by calling all registered listeners with performance tracking
func (e *EventBus[T]) Handle(event *Event[T]) error {
	return e.performanceFacade.Track("event.handle", func() error {
		e.mu.RLock()
		listeners, exists := e.listeners[event.Name]
		e.mu.RUnlock()

		if !exists {
			return nil // No listeners for this event
		}

		// Call all listeners
		var errors []error
		for _, listener := range listeners {
			err := listener(e.ctx, event)
			if err != nil {
				errors = append(errors, err)
			}
		}

		// Return first error if any
		if len(errors) > 0 {
			return fmt.Errorf("event handling errors: %v", errors)
		}

		return nil
	})
}

// HasListeners checks if there are listeners for an event
func (e *EventBus[T]) HasListeners(eventName string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	listeners, exists := e.listeners[eventName]
	return exists && len(listeners) > 0
}

// GetListenerCount returns the number of listeners for an event
func (e *EventBus[T]) GetListenerCount(eventName string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	listeners, exists := e.listeners[eventName]
	if !exists {
		return 0
	}

	return len(listeners)
}

// GetPerformanceStats returns event bus performance statistics
func (e *EventBus[T]) GetPerformanceStats() map[string]interface{} {
	stats := e.performanceFacade.GetStats()

	// Add event-specific stats
	stats["events"] = map[string]interface{}{
		"operations_count": e.atomicCounter.Get(),
		"event_pool_size":  len(e.eventPool.pool),
		"listener_count":   len(e.listeners),
	}

	return stats
}

// GetOptimizationStats returns event bus optimization statistics
func (e *EventBus[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"atomic_operations": e.atomicCounter.Get(),
		"event_pool_usage":  len(e.eventPool.pool),
		"listener_count":    len(e.listeners),
	}
}

// WithContext returns an event dispatcher with context
func (e *EventBus[T]) WithContext(ctx context.Context) EventDispatcher[T] {
	return &EventBus[T]{
		listeners:         e.listeners,
		ctx:               ctx,
		atomicCounter:     e.atomicCounter,
		eventPool:         e.eventPool,
		performanceFacade: e.performanceFacade,
	}
}

// EventStore defines an interface for persisting events
type EventStore[T any] interface {
	// Event persistence
	Store(event *Event[T]) error
	StoreMany(events []*Event[T]) error

	// Event retrieval
	Get(eventID string) (*Event[T], error)
	GetByEventName(eventName string, limit int) ([]*Event[T], error)
	GetByTimeRange(start, end time.Time) ([]*Event[T], error)

	// Event management
	Delete(eventID string) error
	Clear() error

	// Utility operations
	Count() (int64, error)
	CountByEventName(eventName string) (int64, error)
}

// memoryEventStore implements EventStore[T] with in-memory storage
type memoryEventStore[T any] struct {
	events map[string]*Event[T]
	mu     sync.RWMutex
}

// NewMemoryEventStore creates a new in-memory event store
func NewMemoryEventStore[T any]() EventStore[T] {
	return &memoryEventStore[T]{
		events: make(map[string]*Event[T]),
	}
}

// Store stores an event
func (s *memoryEventStore[T]) Store(event *Event[T]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events[event.ID] = event
	return nil
}

// StoreMany stores multiple events
func (s *memoryEventStore[T]) StoreMany(events []*Event[T]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range events {
		s.events[event.ID] = event
	}

	return nil
}

// Get retrieves an event by ID
func (s *memoryEventStore[T]) Get(eventID string) (*Event[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[eventID]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	return event, nil
}

// GetByEventName retrieves events by name
func (s *memoryEventStore[T]) GetByEventName(eventName string, limit int) ([]*Event[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*Event[T]
	count := 0

	for _, event := range s.events {
		if event.Name == eventName {
			events = append(events, event)
			count++
			if count >= limit {
				break
			}
		}
	}

	return events, nil
}

// GetByTimeRange retrieves events within a time range
func (s *memoryEventStore[T]) GetByTimeRange(start, end time.Time) ([]*Event[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*Event[T]

	for _, event := range s.events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			events = append(events, event)
		}
	}

	return events, nil
}

// Delete removes an event
func (s *memoryEventStore[T]) Delete(eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.events, eventID)
	return nil
}

// Clear removes all events
func (s *memoryEventStore[T]) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = make(map[string]*Event[T])
	return nil
}

// Count returns the total number of events
func (s *memoryEventStore[T]) Count() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.events)), nil
}

// CountByEventName returns the number of events by name
func (s *memoryEventStore[T]) CountByEventName(eventName string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := int64(0)
	for _, event := range s.events {
		if event.Name == eventName {
			count++
		}
	}

	return count, nil
}

// EventManagerInterface defines the interface for event management
type EventManagerInterface[T any] interface {
	// Event management
	Dispatch(event *Event[T]) error
	DispatchAsync(event *Event[T]) error

	// Listener management
	Listen(eventName string, listener EventListener[T]) error

	// Event retrieval
	GetEvent(eventID string) (*Event[T], error)
	GetEventsByName(eventName string, limit int) ([]*Event[T], error)
	GetEventsByTimeRange(start, end time.Time) ([]*Event[T], error)

	// Utility operations
	HasListeners(eventName string) bool
	GetListenerCount(eventName string) int
}

// EventManager combines event dispatching and storage
type EventManager[T any] struct {
	dispatcher EventDispatcher[T]
	store      EventStore[T]
}

// NewEventManager creates a new event manager
func NewEventManager[T any](dispatcher EventDispatcher[T], store EventStore[T]) *EventManager[T] {
	return &EventManager[T]{
		dispatcher: dispatcher,
		store:      store,
	}
}

// Dispatch dispatches an event and stores it
func (m *EventManager[T]) Dispatch(event *Event[T]) error {
	// Store the event first
	err := m.store.Store(event)
	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Then dispatch it
	return m.dispatcher.Dispatch(event)
}

// DispatchAsync dispatches an event asynchronously and stores it
func (m *EventManager[T]) DispatchAsync(event *Event[T]) error {
	// Store the event first
	err := m.store.Store(event)
	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Then dispatch it asynchronously
	return m.dispatcher.DispatchAsync(event)
}

// Listen registers an event listener
func (m *EventManager[T]) Listen(eventName string, listener EventListener[T]) error {
	return m.dispatcher.Listen(eventName, listener)
}

// GetEvent retrieves an event from storage
func (m *EventManager[T]) GetEvent(eventID string) (*Event[T], error) {
	return m.store.Get(eventID)
}

// GetEventsByName retrieves events by name from storage
func (m *EventManager[T]) GetEventsByName(eventName string, limit int) ([]*Event[T], error) {
	return m.store.GetByEventName(eventName, limit)
}

// GetEventsByTimeRange retrieves events within a time range from storage
func (m *EventManager[T]) GetEventsByTimeRange(start, end time.Time) ([]*Event[T], error) {
	return m.store.GetByTimeRange(start, end)
}

// HasListeners checks if there are listeners for an event
func (m *EventManager[T]) HasListeners(eventName string) bool {
	return m.dispatcher.HasListeners(eventName)
}

// GetListenerCount returns the number of listeners for an event
func (m *EventManager[T]) GetListenerCount(eventName string) int {
	return m.dispatcher.GetListenerCount(eventName)
}
