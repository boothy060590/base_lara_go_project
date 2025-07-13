package pipeline

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	go_core "base_lara_go_project/app/core/go_core"
)

// ============================================================================
// LARAVEL PIPELINE TESTS
// ============================================================================

func TestLaravelPipeline_BasicFlow(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	// Create test stages
	stage1 := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		// Modify data
		modifiedData := "stage1_" + data
		return next(modifiedData)
	})

	stage2 := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		// Modify data again
		modifiedData := "stage2_" + data
		return next(modifiedData)
	})

	// Add stages to pipeline
	pipeline.Through(stage1, stage2)

	// Test data flow
	var result string
	err := pipeline.Then(context.Background(), "test", func(data string) error {
		result = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "stage2_stage1_test", result)
}

func TestLaravelPipeline_EmptyPipeline(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	var result string
	err := pipeline.Then(context.Background(), "test", func(data string) error {
		result = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestLaravelPipeline_StageError(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	errorStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return fmt.Errorf("stage error")
	})

	pipeline.Through(errorStage)

	err := pipeline.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stage error")
}

func TestLaravelPipeline_ContextCancellation(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	slowStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		select {
		case <-time.After(100 * time.Millisecond):
			return next(data)
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	pipeline.Through(slowStage)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := pipeline.Send(ctx, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestLaravelPipeline_ConcurrentAccess(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next(data)
	})

	pipeline.Through(stage)

	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := pipeline.Send(context.Background(), fmt.Sprintf("test_%d", id))
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
}

func TestLaravelPipeline_ViaMethod(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	stage1 := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next("stage1_" + data)
	})

	stage2 := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next("stage2_" + data)
	})

	// Use Via to set stages
	pipeline.Via(stage1, stage2)

	var result string
	err := pipeline.Then(context.Background(), "test", func(data string) error {
		result = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "stage2_stage1_test", result)
}

// ============================================================================
// PIPELINE DECORATOR TESTS
// ============================================================================

func TestWithCacheDecorator(t *testing.T) {
	cache := go_core.NewLocalCache[string]()
	pipeline := go_core.NewLaravelPipeline[string]()

	// Create a stage that modifies data
	stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next("processed_" + data)
	})

	// Add cache decorator
	cacheDecorator := go_core.WithCache(cache, "test_key", 1*time.Hour)
	pipeline.Through(cacheDecorator, stage)

	// First call - should process and cache
	var result1 string
	err := pipeline.Then(context.Background(), "test", func(data string) error {
		result1 = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "processed_test", result1)

	// Second call - should use cache
	var result2 string
	err = pipeline.Then(context.Background(), "test", func(data string) error {
		result2 = data
		return nil
	})

	assert.NoError(t, err)
	// Should return processed data since cache decorator calls next with original data
	assert.Equal(t, "processed_test", result2)
}

func TestWithAnalyticsDecorator(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	analyticsCalls := make([]string, 0)
	tracker := func(event string, data string) error {
		analyticsCalls = append(analyticsCalls, event)
		return nil
	}

	analyticsDecorator := go_core.WithAnalytics(tracker)
	pipeline.Through(analyticsDecorator)

	err := pipeline.Send(context.Background(), "test")
	assert.NoError(t, err)
	assert.Contains(t, analyticsCalls, "pipeline_start")
	assert.Contains(t, analyticsCalls, "pipeline_success")
}

func TestWithLoggingDecorator(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	logCalls := make([]string, 0)
	logger := func(event string, data string) {
		logCalls = append(logCalls, event)
	}

	loggingDecorator := go_core.WithLogging(logger)
	pipeline.Through(loggingDecorator)

	err := pipeline.Send(context.Background(), "test")
	assert.NoError(t, err)
	assert.Contains(t, logCalls, "pipeline_processing")
	assert.Contains(t, logCalls, "pipeline_completed")
}

func TestWithRetryDecorator(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	attempts := 0
	failingStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("temporary error")
		}
		return next(data)
	})

	retryDecorator := go_core.WithRetry[string](3, 10*time.Millisecond)
	pipeline.Through(retryDecorator, failingStage)

	err := pipeline.Send(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestWithTimeoutDecorator(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	slowStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		time.Sleep(100 * time.Millisecond)
		return next(data)
	})

	timeoutDecorator := go_core.WithTimeout[string](50 * time.Millisecond)
	pipeline.Through(timeoutDecorator, slowStage)

	err := pipeline.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestWithValidationDecorator(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	validator := func(data string) error {
		if len(data) < 5 {
			return fmt.Errorf("data too short")
		}
		return nil
	}

	validationDecorator := go_core.WithValidation(validator)
	pipeline.Through(validationDecorator)

	// Test with valid data
	err := pipeline.Send(context.Background(), "valid_data")
	assert.NoError(t, err)

	// Test with invalid data
	err = pipeline.Send(context.Background(), "abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestWithRateLimitDecorator(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	limiter := go_core.NewRateLimiter(100*time.Millisecond, 1)
	rateLimitDecorator := go_core.WithRateLimit[string](limiter)
	pipeline.Through(rateLimitDecorator)

	// First call should succeed
	err := pipeline.Send(context.Background(), "test1")
	assert.NoError(t, err)

	// Second call should fail due to rate limiting
	err = pipeline.Send(context.Background(), "test2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

// ============================================================================
// RATE LIMITER TESTS
// ============================================================================

func TestRateLimiter_Basic(t *testing.T) {
	limiter := go_core.NewRateLimiter(100*time.Millisecond, 2)

	// First two calls should succeed
	assert.True(t, limiter.Allow())
	assert.True(t, limiter.Allow())

	// Third call should fail
	assert.False(t, limiter.Allow())
}

func TestRateLimiter_Refill(t *testing.T) {
	limiter := go_core.NewRateLimiter(50*time.Millisecond, 1)

	// First call should succeed
	assert.True(t, limiter.Allow())

	// Second call should fail immediately
	assert.False(t, limiter.Allow())

	// Wait for refill
	time.Sleep(60 * time.Millisecond)

	// Should succeed again
	assert.True(t, limiter.Allow())
}

// ============================================================================
// PIPELINE BUILDER TESTS
// ============================================================================

func TestPipelineBuilder_Basic(t *testing.T) {
	builder := go_core.NewPipelineBuilder[string]()

	// Add decorators
	cache := go_core.NewLocalCache[string]()
	builder.WithCache(cache, "test_key", 1*time.Hour)
	builder.WithLogging(func(event string, data string) {})

	pipeline := builder.Build()
	assert.NotNil(t, pipeline)

	// Test the pipeline
	var result string
	err := pipeline.Then(context.Background(), "test", func(data string) error {
		result = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestPipelineBuilder_Chained(t *testing.T) {
	builder := go_core.NewPipelineBuilder[string]()

	cache := go_core.NewLocalCache[string]()
	pipeline := builder.
		WithCache(cache, "test_key", 1*time.Hour).
		WithLogging(func(event string, data string) {}).
		WithValidation(func(data string) error { return nil }).
		Build()

	assert.NotNil(t, pipeline)

	var result string
	err := pipeline.Then(context.Background(), "test", func(data string) error {
		result = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "test", result)
}

// ============================================================================
// PERFORMANCE TESTS
// ============================================================================

func TestPipelinePerformance(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	// Add multiple stages
	for i := 0; i < 10; i++ {
		stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
			return next(data)
		})
		pipeline.Through(stage)
	}

	start := time.Now()
	iterations := 1000

	for i := 0; i < iterations; i++ {
		err := pipeline.Send(context.Background(), "test")
		assert.NoError(t, err)
	}

	duration := time.Since(start)
	t.Logf("Processed %d iterations in %v (%d ops/sec)",
		iterations, duration, int(float64(iterations)/duration.Seconds()))
}

func TestPipelineConcurrency(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next(data)
	})

	pipeline.Through(stage)

	concurrency := 100
	iterations := 10

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := pipeline.Send(context.Background(), fmt.Sprintf("test_%d_%d", id, j))
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	totalOps := concurrency * iterations

	t.Logf("Processed %d concurrent operations in %v (%d ops/sec)",
		totalOps, duration, int(float64(totalOps)/duration.Seconds()))
}

// ============================================================================
// EDGE CASES AND ERROR HANDLING
// ============================================================================

func TestPipelineNilStages(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	// Test with nil stages
	pipeline.Through(nil)

	_ = pipeline.Send(context.Background(), "test")
	// Should not panic, but behavior depends on implementation
	// For now, we'll just ensure it doesn't crash
}

func TestPipelineEmptyData(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next(data)
	})

	pipeline.Through(stage)

	err := pipeline.Send(context.Background(), "")
	assert.NoError(t, err)
}

func TestPipelineContextValues(t *testing.T) {
	pipeline := go_core.NewLaravelPipeline[string]()

	stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		// Check if context values are preserved
		if value := ctx.Value("test_key"); value != "test_value" {
			return fmt.Errorf("context value not preserved")
		}
		return next(data)
	})

	pipeline.Through(stage)

	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	err := pipeline.Send(ctx, "test")
	assert.NoError(t, err)
}
