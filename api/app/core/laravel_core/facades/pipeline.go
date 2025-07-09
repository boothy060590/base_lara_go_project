package facades_core

import (
	"context"
	"sync"
	"time"

	app_core "base_lara_go_project/app/core/go_core"
)

// ============================================================================
// LARAVEL-STYLE PIPELINE FACADE
// ============================================================================

// PipelineFacade provides Laravel-style access to pipeline operations
type PipelineFacade struct{}

var (
	pipelineInstance *PipelineFacade
	pipelineOnce     sync.Once
)

// Pipeline returns the singleton pipeline facade instance
func Pipeline() *PipelineFacade {
	pipelineOnce.Do(func() {
		pipelineInstance = &PipelineFacade{}
	})
	return pipelineInstance
}

// New creates a new pipeline builder
func (pf *PipelineFacade) New() *LaravelPipelineBuilder[any] {
	return NewLaravelPipelineBuilder[any]()
}

// Send sends data through a pipeline with decorators
func (pf *PipelineFacade) Send(ctx context.Context, data any, decorators ...app_core.LaravelPipelineStage[any]) error {
	pipeline := app_core.NewLaravelPipeline[any]()
	pipeline.Through(decorators...)
	return pipeline.Send(ctx, data)
}

// Then executes a pipeline with a final callback
func (pf *PipelineFacade) Then(ctx context.Context, data any, callback func(any) error, decorators ...app_core.LaravelPipelineStage[any]) error {
	pipeline := app_core.NewLaravelPipeline[any]()
	pipeline.Through(decorators...)
	return pipeline.Then(ctx, data, callback)
}

// ============================================================================
// LARAVEL PIPELINE BUILDER
// ============================================================================

// LaravelPipelineBuilder provides a fluent Laravel-style API for building pipelines
type LaravelPipelineBuilder[T any] struct {
	builder *app_core.PipelineBuilder[T]
}

// NewLaravelPipelineBuilder creates a new Laravel-style pipeline builder
func NewLaravelPipelineBuilder[T any]() *LaravelPipelineBuilder[T] {
	return &LaravelPipelineBuilder[T]{
		builder: app_core.NewPipelineBuilder[T](),
	}
}

// WithCache adds cache decorator (Laravel-style)
func (lpb *LaravelPipelineBuilder[T]) WithCache(key string, ttl time.Duration) *LaravelPipelineBuilder[T] {
	// Get cache from container (this would be injected in real implementation)
	cache := getCache[T]()
	lpb.builder.WithCache(cache, key, ttl)
	return lpb
}

// WithAnalytics adds analytics decorator (Laravel-style)
func (lpb *LaravelPipelineBuilder[T]) WithAnalytics() *LaravelPipelineBuilder[T] {
	tracker := getAnalyticsTracker[T]()
	lpb.builder.WithAnalytics(tracker)
	return lpb
}

// WithLogging adds logging decorator (Laravel-style)
func (lpb *LaravelPipelineBuilder[T]) WithLogging() *LaravelPipelineBuilder[T] {
	logger := getLogger[T]()
	lpb.builder.WithLogging(logger)
	return lpb
}

// WithRetry adds retry decorator (Laravel-style)
func (lpb *LaravelPipelineBuilder[T]) WithRetry(maxAttempts int, delay time.Duration) *LaravelPipelineBuilder[T] {
	lpb.builder.WithRetry(maxAttempts, delay)
	return lpb
}

// WithTimeout adds timeout decorator (Laravel-style)
func (lpb *LaravelPipelineBuilder[T]) WithTimeout(timeout time.Duration) *LaravelPipelineBuilder[T] {
	lpb.builder.WithTimeout(timeout)
	return lpb
}

// WithValidation adds validation decorator (Laravel-style)
func (lpb *LaravelPipelineBuilder[T]) WithValidation(validator func(T) error) *LaravelPipelineBuilder[T] {
	lpb.builder.WithValidation(validator)
	return lpb
}

// WithRateLimit adds rate limit decorator (Laravel-style)
func (lpb *LaravelPipelineBuilder[T]) WithRateLimit(rate time.Duration, burst int) *LaravelPipelineBuilder[T] {
	limiter := app_core.NewRateLimiter(rate, burst)
	lpb.builder.WithRateLimit(limiter)
	return lpb
}

// Build returns the built pipeline
func (lpb *LaravelPipelineBuilder[T]) Build() *app_core.LaravelPipeline[T] {
	return lpb.builder.Build()
}

// Send sends data through the built pipeline
func (lpb *LaravelPipelineBuilder[T]) Send(ctx context.Context, data T) error {
	return lpb.builder.Send(ctx, data)
}

// Then executes the pipeline with a final callback
func (lpb *LaravelPipelineBuilder[T]) Then(ctx context.Context, data T, callback func(T) error) error {
	return lpb.builder.Then(ctx, data, callback)
}

// ============================================================================
// HELPER FUNCTIONS FOR LARAVEL-STYLE INTEGRATION
// ============================================================================

// getCache gets cache instance from container (placeholder)
func getCache[T any]() app_core.Cache[T] {
	// In real implementation, this would get from the service container
	return app_core.NewLocalCache[T]()
}

// getAnalyticsTracker gets analytics tracker (placeholder)
func getAnalyticsTracker[T any]() func(string, T) error {
	return func(event string, data T) error {
		// In real implementation, this would track analytics
		return nil
	}
}

// getLogger gets logger instance (placeholder)
func getLogger[T any]() func(string, T) {
	return func(level string, data T) {
		// In real implementation, this would log using the logging system
	}
}

// ============================================================================
// GLOBAL PIPELINE FUNCTIONS (LARAVEL-STYLE)
// ============================================================================

// PipelineSend sends data through a pipeline (global function)
func PipelineSend[T any](ctx context.Context, data T, decorators ...app_core.LaravelPipelineStage[T]) error {
	// Create a new pipeline for this specific type
	pipeline := app_core.NewLaravelPipeline[T]()
	pipeline.Through(decorators...)
	return pipeline.Send(ctx, data)
}

// PipelineThen executes a pipeline with callback (global function)
func PipelineThen[T any](ctx context.Context, data T, callback func(T) error, decorators ...app_core.LaravelPipelineStage[T]) error {
	// Create a new pipeline for this specific type
	pipeline := app_core.NewLaravelPipeline[T]()
	pipeline.Through(decorators...)
	return pipeline.Then(ctx, data, callback)
}

// ============================================================================
// CONVENIENCE DECORATOR FUNCTIONS (LARAVEL-STYLE)
// ============================================================================

// WithCache creates a cache decorator (convenience function)
func WithCache[T any](key string, ttl time.Duration) app_core.LaravelPipelineStage[T] {
	cache := getCache[T]()
	return app_core.WithCache(cache, key, ttl)
}

// WithAnalytics creates an analytics decorator (convenience function)
func WithAnalytics[T any]() app_core.LaravelPipelineStage[T] {
	tracker := getAnalyticsTracker[T]()
	return app_core.WithAnalytics(tracker)
}

// WithLogging creates a logging decorator (convenience function)
func WithLogging[T any]() app_core.LaravelPipelineStage[T] {
	logger := getLogger[T]()
	return app_core.WithLogging(logger)
}

// WithRetry creates a retry decorator (convenience function)
func WithRetry[T any](maxAttempts int, delay time.Duration) app_core.LaravelPipelineStage[T] {
	return app_core.WithRetry[T](maxAttempts, delay)
}

// WithTimeout creates a timeout decorator (convenience function)
func WithTimeout[T any](timeout time.Duration) app_core.LaravelPipelineStage[T] {
	return app_core.WithTimeout[T](timeout)
}

// WithValidation creates a validation decorator (convenience function)
func WithValidation[T any](validator func(T) error) app_core.LaravelPipelineStage[T] {
	return app_core.WithValidation(validator)
}

// WithRateLimit creates a rate limit decorator (convenience function)
func WithRateLimit[T any](rate time.Duration, burst int) app_core.LaravelPipelineStage[T] {
	limiter := app_core.NewRateLimiter(rate, burst)
	return app_core.WithRateLimit[T](limiter)
}
