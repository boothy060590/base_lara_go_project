package observers_core

import (
	"log"
)

// CacheInterface defines the interface for cache operations
type CacheInterface interface {
	Delete(key string) error
}

// CacheableModel defines the interface for cacheable models
type CacheableModel interface {
	GetCacheKey() string
}

// ModelObserver defines the interface for model observers
type ModelObserver interface {
	Created(model interface{}) error
	Updated(model interface{}) error
	Deleted(model interface{}) error
	Saved(model interface{}) error
}

// CacheableModelObserver provides automatic cache invalidation for cacheable models
type CacheableModelObserver struct {
	cacheService CacheInterface
}

// NewCacheableModelObserver creates a new cacheable model observer
func NewCacheableModelObserver(cacheService CacheInterface) *CacheableModelObserver {
	return &CacheableModelObserver{
		cacheService: cacheService,
	}
}

// Created handles model creation events
func (o *CacheableModelObserver) Created(model interface{}) error {
	if cacheable, ok := model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// Updated handles model update events
func (o *CacheableModelObserver) Updated(model interface{}) error {
	if cacheable, ok := model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// Deleted handles model deletion events
func (o *CacheableModelObserver) Deleted(model interface{}) error {
	if cacheable, ok := model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// Saved handles model save events
func (o *CacheableModelObserver) Saved(model interface{}) error {
	if cacheable, ok := model.(CacheableModel); ok {
		return o.invalidateCache(cacheable)
	}
	return nil
}

// invalidateCache invalidates cache for a cacheable model
func (o *CacheableModelObserver) invalidateCache(cacheable CacheableModel) error {
	// Invalidate by cache key
	cacheKey := cacheable.GetCacheKey()
	if cacheKey != "" {
		return o.cacheService.Delete(cacheKey)
	}
	return nil
}

// ModelEventType represents the type of model event
type ModelEventType string

const (
	ModelEventCreated ModelEventType = "created"
	ModelEventUpdated ModelEventType = "updated"
	ModelEventDeleted ModelEventType = "deleted"
	ModelEventSaved   ModelEventType = "saved"
)

// ModelEvent represents a model event
type ModelEvent struct {
	Type  ModelEventType
	Model interface{}
}

// ModelEventBus defines the interface for model event bus
type ModelEventBus interface {
	Publish(event ModelEvent) error
	Subscribe(eventType ModelEventType, observer ModelObserver)
}

// SimpleModelEventBus provides a simple implementation of ModelEventBus
type SimpleModelEventBus struct {
	observers map[ModelEventType][]ModelObserver
}

// NewSimpleModelEventBus creates a new simple model event bus
func NewSimpleModelEventBus() *SimpleModelEventBus {
	return &SimpleModelEventBus{
		observers: make(map[ModelEventType][]ModelObserver),
	}
}

// Publish publishes a model event to all subscribed observers
func (bus *SimpleModelEventBus) Publish(event ModelEvent) error {
	observers, exists := bus.observers[event.Type]
	if !exists {
		return nil
	}

	for _, observer := range observers {
		var err error
		switch event.Type {
		case ModelEventCreated:
			err = observer.Created(event.Model)
		case ModelEventUpdated:
			err = observer.Updated(event.Model)
		case ModelEventDeleted:
			err = observer.Deleted(event.Model)
		case ModelEventSaved:
			err = observer.Saved(event.Model)
		}

		if err != nil {
			log.Printf("Error in model observer %s: %v", event.Type, err)
		}
	}

	return nil
}

// Subscribe subscribes an observer to a specific event type
func (bus *SimpleModelEventBus) Subscribe(eventType ModelEventType, observer ModelObserver) {
	bus.observers[eventType] = append(bus.observers[eventType], observer)
}

// RegisterModelObserver registers a model observer with the event bus
func RegisterModelObserver(eventBus ModelEventBus, observer ModelObserver) {
	eventBus.Subscribe(ModelEventCreated, observer)
	eventBus.Subscribe(ModelEventUpdated, observer)
	eventBus.Subscribe(ModelEventDeleted, observer)
	eventBus.Subscribe(ModelEventSaved, observer)
}

// RegisterCacheableModel registers a cacheable model with automatic cache invalidation
func RegisterCacheableModel(eventBus ModelEventBus, cacheService CacheInterface) {
	observer := NewCacheableModelObserver(cacheService)
	RegisterModelObserver(eventBus, observer)
}
