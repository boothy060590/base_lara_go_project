package facades_core

import (
	app_core "base_lara_go_project/app/core/go_core"
)

// JobInterface defines the interface for jobs
type JobInterface interface {
	Execute() error
	GetName() string
}

// JobDispatcher defines the interface for dispatching jobs
type JobDispatcher interface {
	Dispatch(job interface{}) error
	DispatchSync(job interface{}) error
}

// Global job dispatcher instance
var JobDispatcherInstance JobDispatcher

// SetJobDispatcher sets the global job dispatcher
func SetJobDispatcher(dispatcher JobDispatcher) {
	JobDispatcherInstance = dispatcher
}

// Dispatch dispatches a job (respects ShouldQueue trait)
func Dispatch(job interface{}) error {
	return JobDispatcherInstance.Dispatch(job)
}

// DispatchSync dispatches a job synchronously (ignores ShouldQueue trait)
func DispatchSync(job interface{}) error {
	return JobDispatcherInstance.DispatchSync(job)
}

// GenericJobDispatcher provides type-safe job dispatching
type GenericJobDispatcher[T any] struct {
	dispatcher app_core.JobDispatcher[T]
}

// NewGenericJobDispatcher creates a new generic job dispatcher
func NewGenericJobDispatcher[T any](dispatcher app_core.JobDispatcher[T]) *GenericJobDispatcher[T] {
	return &GenericJobDispatcher[T]{
		dispatcher: dispatcher,
	}
}

// Dispatch dispatches a job (respects ShouldQueue trait)
func (d *GenericJobDispatcher[T]) Dispatch(job T) error {
	return d.dispatcher.Dispatch(job)
}

// DispatchSync dispatches a job synchronously (ignores ShouldQueue trait)
func (d *GenericJobDispatcher[T]) DispatchSync(job T) error {
	return d.dispatcher.DispatchSync(job)
}
