package go_core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// LARAVEL-STYLE PIPELINE SYSTEM
// ============================================================================

// LaravelPipelineStage represents a stage in a Laravel-style pipeline
type LaravelPipelineStage[T any] interface {
	Handle(ctx context.Context, data T, next func(T) error) error
}

// LaravelPipelineStageFunc is a function-based pipeline stage
type LaravelPipelineStageFunc[T any] func(ctx context.Context, data T, next func(T) error) error

// Handle implements LaravelPipelineStage interface
func (f LaravelPipelineStageFunc[T]) Handle(ctx context.Context, data T, next func(T) error) error {
	return f(ctx, data, next)
}

// LaravelPipeline represents a Laravel-style pipeline with middleware
type LaravelPipeline[T any] struct {
	stages []LaravelPipelineStage[T]
	mu     sync.RWMutex
}

// NewLaravelPipeline creates a new Laravel-style pipeline
func NewLaravelPipeline[T any]() *LaravelPipeline[T] {
	return &LaravelPipeline[T]{
		stages: make([]LaravelPipelineStage[T], 0),
	}
}

// Send sends data through the pipeline
func (p *LaravelPipeline[T]) Send(ctx context.Context, data T) error {
	p.mu.RLock()
	stages := make([]LaravelPipelineStage[T], len(p.stages))
	copy(stages, p.stages)
	p.mu.RUnlock()

	return p.process(ctx, data, stages, 0)
}

// process recursively processes stages in the pipeline
func (p *LaravelPipeline[T]) process(ctx context.Context, data T, stages []LaravelPipelineStage[T], index int) error {
	if index >= len(stages) {
		// End of pipeline - no more stages
		return nil
	}

	stage := stages[index]
	return stage.Handle(ctx, data, func(data T) error {
		return p.process(ctx, data, stages, index+1)
	})
}

// Through adds stages to the pipeline
func (p *LaravelPipeline[T]) Through(stages ...LaravelPipelineStage[T]) *LaravelPipeline[T] {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stages = append(p.stages, stages...)
	return p
}

// Via sets the stages for the pipeline (replaces existing stages)
func (p *LaravelPipeline[T]) Via(stages ...LaravelPipelineStage[T]) *LaravelPipeline[T] {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stages = stages
	return p
}

// Then executes the pipeline with a final callback
func (p *LaravelPipeline[T]) Then(ctx context.Context, data T, callback func(T) error) error {
	// Create a final stage that calls the callback
	finalStage := LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		return callback(data)
	})

	// Add the final stage temporarily
	p.mu.Lock()
	originalStages := make([]LaravelPipelineStage[T], len(p.stages))
	copy(originalStages, p.stages)
	p.stages = append(p.stages, finalStage)
	p.mu.Unlock()

	// Process the pipeline
	err := p.Send(ctx, data)

	// Restore original stages
	p.mu.Lock()
	p.stages = originalStages
	p.mu.Unlock()

	return err
}

// ============================================================================
// PIPELINE DECORATORS (LARAVEL-STYLE)
// ============================================================================

// WithCache decorator that caches the result
func WithCache[T any](cache Cache[T], key string, ttl time.Duration) LaravelPipelineStage[T] {
	return LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		// Try to get from cache first
		if cached, err := cache.Get(key); err == nil && cached != nil {
			// Return cached data without calling next
			return nil
		}

		// Not in cache, process normally
		err := next(data)
		if err != nil {
			return err
		}

		// Cache the result
		cache.Set(key, &data, ttl)
		return nil
	})
}

// WithAnalytics decorator that tracks analytics
func WithAnalytics[T any](tracker func(string, T) error) LaravelPipelineStage[T] {
	return LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		// Track before processing
		if err := tracker("pipeline_start", data); err != nil {
			// Log error but continue processing
		}

		// Process normally
		err := next(data)
		if err != nil {
			// Track error
			tracker("pipeline_error", data)
			return err
		}

		// Track success
		tracker("pipeline_success", data)
		return nil
	})
}

// WithLogging decorator that adds logging
func WithLogging[T any](logger func(string, T)) LaravelPipelineStage[T] {
	return LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		logger("pipeline_processing", data)

		err := next(data)
		if err != nil {
			logger("pipeline_error", data)
		} else {
			logger("pipeline_completed", data)
		}

		return err
	})
}

// WithRetry decorator that retries on failure
func WithRetry[T any](maxAttempts int, delay time.Duration) LaravelPipelineStage[T] {
	return LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		var lastErr error

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err := next(data)
			if err == nil {
				return nil
			}

			lastErr = err
			if attempt < maxAttempts {
				time.Sleep(delay)
			}
		}

		return fmt.Errorf("pipeline failed after %d attempts: %w", maxAttempts, lastErr)
	})
}

// WithTimeout decorator that adds timeout
func WithTimeout[T any](timeout time.Duration) LaravelPipelineStage[T] {
	return LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Create a channel to receive the result
		resultChan := make(chan error, 1)

		go func() {
			resultChan <- next(data)
		}()

		select {
		case err := <-resultChan:
			return err
		case <-ctx.Done():
			return fmt.Errorf("pipeline timeout after %v", timeout)
		}
	})
}

// WithValidation decorator that validates data
func WithValidation[T any](validator func(T) error) LaravelPipelineStage[T] {
	return LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		if err := validator(data); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		return next(data)
	})
}

// WithRateLimit decorator that adds rate limiting
func WithRateLimit[T any](limiter *RateLimiter) LaravelPipelineStage[T] {
	return LaravelPipelineStageFunc[T](func(ctx context.Context, data T, next func(T) error) error {
		if !limiter.Allow() {
			return fmt.Errorf("rate limit exceeded")
		}

		return next(data)
	})
}

// ============================================================================
// RATE LIMITER FOR PIPELINE DECORATORS
// ============================================================================

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	tokens    chan struct{}
	rate      time.Duration
	mu        sync.Mutex
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, burst),
		rate:   rate,
	}

	// Fill the token bucket
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token refill goroutine
	go rl.refillTokens()

	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// refillTokens refills the token bucket
func (rl *RateLimiter) refillTokens() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
			// Token added successfully
		default:
			// Bucket is full, skip
		}
	}
}

// ============================================================================
// PIPELINE BUILDER FOR FLUENT API
// ============================================================================

// PipelineBuilder provides a fluent API for building pipelines
type PipelineBuilder[T any] struct {
	pipeline *LaravelPipeline[T]
}

// NewPipelineBuilder creates a new pipeline builder
func NewPipelineBuilder[T any]() *PipelineBuilder[T] {
	return &PipelineBuilder[T]{
		pipeline: NewLaravelPipeline[T](),
	}
}

// WithCache adds cache decorator
func (pb *PipelineBuilder[T]) WithCache(cache Cache[T], key string, ttl time.Duration) *PipelineBuilder[T] {
	pb.pipeline.Through(WithCache(cache, key, ttl))
	return pb
}

// WithAnalytics adds analytics decorator
func (pb *PipelineBuilder[T]) WithAnalytics(tracker func(string, T) error) *PipelineBuilder[T] {
	pb.pipeline.Through(WithAnalytics(tracker))
	return pb
}

// WithLogging adds logging decorator
func (pb *PipelineBuilder[T]) WithLogging(logger func(string, T)) *PipelineBuilder[T] {
	pb.pipeline.Through(WithLogging(logger))
	return pb
}

// WithRetry adds retry decorator
func (pb *PipelineBuilder[T]) WithRetry(maxAttempts int, delay time.Duration) *PipelineBuilder[T] {
	pb.pipeline.Through(WithRetry[T](maxAttempts, delay))
	return pb
}

// WithTimeout adds timeout decorator
func (pb *PipelineBuilder[T]) WithTimeout(timeout time.Duration) *PipelineBuilder[T] {
	pb.pipeline.Through(WithTimeout[T](timeout))
	return pb
}

// WithValidation adds validation decorator
func (pb *PipelineBuilder[T]) WithValidation(validator func(T) error) *PipelineBuilder[T] {
	pb.pipeline.Through(WithValidation(validator))
	return pb
}

// WithRateLimit adds rate limit decorator
func (pb *PipelineBuilder[T]) WithRateLimit(limiter *RateLimiter) *PipelineBuilder[T] {
	pb.pipeline.Through(WithRateLimit[T](limiter))
	return pb
}

// Build returns the built pipeline
func (pb *PipelineBuilder[T]) Build() *LaravelPipeline[T] {
	return pb.pipeline
}

// Send sends data through the built pipeline
func (pb *PipelineBuilder[T]) Send(ctx context.Context, data T) error {
	return pb.pipeline.Send(ctx, data)
}

// Then executes the pipeline with a final callback
func (pb *PipelineBuilder[T]) Then(ctx context.Context, data T, callback func(T) error) error {
	return pb.pipeline.Then(ctx, data, callback)
}
