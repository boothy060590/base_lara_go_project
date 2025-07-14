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
// PIPELINE SYSTEM INTEGRATION TESTS
// ============================================================================

func TestPipelineSystem_CompleteWorkflow(t *testing.T) {
	// Test complete pipeline workflow with multiple stages
	pipeline := go_core.NewLaravelPipeline[string]()

	// Create stages that simulate real-world processing
	validationStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		if len(data) < 3 {
			return fmt.Errorf("data too short")
		}
		return next(data)
	})

	transformationStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		transformed := "processed_" + data
		return next(transformed)
	})

	enrichmentStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		enriched := data + "_enriched"
		return next(enriched)
	})

	// Build pipeline
	pipeline.Through(validationStage, transformationStage, enrichmentStage)

	// Test data flow
	var result string
	err := pipeline.Then(context.Background(), "test_data", func(data string) error {
		result = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "processed_test_data_enriched", result)
}

func TestPipelineSystem_ErrorHandling(t *testing.T) {
	// Test error handling and propagation through pipeline
	pipeline := go_core.NewLaravelPipeline[string]()

	// Stage that fails for specific data
	failingStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		if data == "fail" {
			return fmt.Errorf("intentional failure")
		}
		return next(data)
	})

	pipeline.Through(failingStage)

	// Test successful case
	err := pipeline.Send(context.Background(), "success")
	assert.NoError(t, err)

	// Test failing case
	err = pipeline.Send(context.Background(), "fail")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "intentional failure")
}

func TestPipelineSystem_ContextIntegration(t *testing.T) {
	// Test context integration and cancellation
	pipeline := go_core.NewLaravelPipeline[string]()

	// Stage that respects context cancellation
	contextAwareStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		select {
		case <-time.After(100 * time.Millisecond):
			return next(data)
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	pipeline.Through(contextAwareStage)

	// Test with cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := pipeline.Send(ctx, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestPipelineSystem_ConcurrentOperations(t *testing.T) {
	// Test concurrent pipeline operations
	pipeline := go_core.NewLaravelPipeline[string]()

	// Simple processing stage
	stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next("processed_" + data)
	})

	pipeline.Through(stage)

	// Run concurrent operations
	concurrency := 10
	iterations := 5
	var wg sync.WaitGroup
	errors := make(chan error, concurrency*iterations)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := pipeline.Send(context.Background(), fmt.Sprintf("data_%d_%d", id, j))
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(t, err)
	}
}

// ============================================================================
// ADVANCED CHANNELS INTEGRATION TESTS
// ============================================================================

func TestAdvancedChannels_DataFlow(t *testing.T) {
	// Test complete data flow through advanced channel patterns
	manager := go_core.NewChannelManager(nil)

	// Create input channel
	input := make(chan int, 100)

	// Send data
	go func() {
		defer close(input)
		for i := 0; i < 50; i++ {
			input <- i
		}
	}()

	// Create processing pipeline
	stage1 := func(item int) int {
		return item * 2
	}

	stage2 := func(item int) int {
		return item + 1
	}

	// Apply pipeline
	output := go_core.CreatePipeline(manager, input, stage1, stage2)

	// Collect results
	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Verify transformation: (i * 2) + 1
	assert.Len(t, results, 50)
	for i := 0; i < 50; i++ {
		expected := (i * 2) + 1
		assert.Contains(t, results, expected)
	}
}

func TestAdvancedChannels_BackpressureHandling(t *testing.T) {
	// Test backpressure handling with slow consumers
	manager := go_core.NewChannelManager(nil)

	// Create input with limited buffer
	input := make(chan int, 5)

	// Send data faster than processing
	go func() {
		defer close(input)
		for i := 0; i < 20; i++ {
			input <- i
		}
	}()

	// Slow processing stage
	slowStage := func(item int) int {
		time.Sleep(10 * time.Millisecond)
		return item * 2
	}

	output := go_core.CreatePipeline(manager, input, slowStage)

	// Collect results
	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Should have processed all items despite backpressure
	assert.Len(t, results, 20)
	for i := 0; i < 20; i++ {
		assert.Contains(t, results, i*2)
	}
}

func TestAdvancedChannels_FanOutFanIn(t *testing.T) {
	// Test fan-out and fan-in patterns
	manager := go_core.NewChannelManager(nil)

	// Create input
	input := make(chan int, 20)
	go func() {
		defer close(input)
		for i := 0; i < 20; i++ {
			input <- i
		}
	}()

	// Fan out to multiple workers
	outputs := go_core.FanOut(manager, input, 4)

	// Process in parallel
	var processedOutputs []<-chan int
	for _, output := range outputs {
		processed := go_core.CreatePipeline(manager, output, func(item int) int {
			return item * 2
		})
		processedOutputs = append(processedOutputs, processed)
	}

	// Fan in results
	finalOutput := go_core.FanIn(manager, processedOutputs)

	// Collect results
	var results []int
	for item := range finalOutput {
		results = append(results, item)
	}

	// Verify all items were processed
	assert.Len(t, results, 20)
	for i := 0; i < 20; i++ {
		assert.Contains(t, results, i*2)
	}
}

func TestAdvancedChannels_BatchProcessing(t *testing.T) {
	// Test batch processing with different batch sizes
	manager := go_core.NewChannelManager(nil)

	// Create input
	input := make(chan int, 25)
	go func() {
		defer close(input)
		for i := 0; i < 25; i++ {
			input <- i
		}
	}()

	// Create batches of 5
	batched := go_core.Batch(manager, input, 5)

	// Process batches
	var results [][]int
	for batch := range batched {
		results = append(results, batch)
	}

	// Should have 5 batches of 5 items each
	assert.Len(t, results, 5)
	for _, batch := range results {
		assert.Len(t, batch, 5)
	}
}

func TestAdvancedChannels_RateLimiting(t *testing.T) {
	// Test rate limiting functionality
	manager := go_core.NewChannelManager(nil)

	// Create input
	input := make(chan int, 10)
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	// Apply rate limiting (1 item per 50ms)
	rateLimited := go_core.RateLimit(manager, input, 50*time.Millisecond)

	start := time.Now()
	var results []int
	for item := range rateLimited {
		results = append(results, item)
	}
	duration := time.Since(start)

	// Should have all items
	assert.Len(t, results, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, results, i)
	}

	// Should have taken at least 200ms (4 intervals between 5 items)
	assert.GreaterOrEqual(t, duration, 200*time.Millisecond)
}

func TestAdvancedChannels_RetryWithBackoff(t *testing.T) {
	// Test retry with backoff functionality
	manager := go_core.NewChannelManager(nil)

	// Create input
	input := make(chan int, 5)
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	// Operation that fails initially then succeeds
	attempts := make(map[int]int)
	operation := func(item int) error {
		attempts[item]++
		if attempts[item] < 2 {
			return fmt.Errorf("temporary error")
		}
		return nil
	}

	retryOutput := go_core.RetryWithBackoff(manager, input, operation)

	var results []int
	for item := range retryOutput {
		results = append(results, item)
	}

	// Should have all items
	assert.Len(t, results, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, results, i)
		assert.Equal(t, 2, attempts[i]) // Each item should have been attempted twice
	}
}

// ============================================================================
// PIPELINE DECORATORS INTEGRATION TESTS
// ============================================================================

func TestPipelineDecorators_CompleteWorkflow(t *testing.T) {
	// Test pipeline with multiple decorators
	pipeline := go_core.NewLaravelPipeline[string]()

	// Create cache
	cache := go_core.NewLocalCache[string]()

	// Add decorators
	cacheDecorator := go_core.WithCache(cache, "test_key", 1*time.Hour)
	loggingDecorator := go_core.WithLogging(func(event string, data string) {
		// Log events
	})

	validationDecorator := go_core.WithValidation(func(data string) error {
		if len(data) < 3 {
			return fmt.Errorf("data too short")
		}
		return nil
	})

	// Build pipeline with decorators
	pipeline.Through(cacheDecorator, loggingDecorator, validationDecorator)

	// Test pipeline
	var result string
	err := pipeline.Then(context.Background(), "valid_data", func(data string) error {
		result = data
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "valid_data", result)
}

func TestPipelineDecorators_ErrorHandling(t *testing.T) {
	// Test error handling with decorators
	pipeline := go_core.NewLaravelPipeline[string]()

	// Add retry decorator
	retryDecorator := go_core.WithRetry[string](3, 10*time.Millisecond)

	// Stage that fails initially
	attempts := 0
	failingStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("temporary error")
		}
		return next(data)
	})

	pipeline.Through(retryDecorator, failingStage)

	err := pipeline.Send(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestPipelineDecorators_TimeoutHandling(t *testing.T) {
	// Test timeout handling with decorators
	pipeline := go_core.NewLaravelPipeline[string]()

	// Add timeout decorator
	timeoutDecorator := go_core.WithTimeout[string](50 * time.Millisecond)

	// Slow stage
	slowStage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		time.Sleep(100 * time.Millisecond)
		return next(data)
	})

	pipeline.Through(timeoutDecorator, slowStage)

	err := pipeline.Send(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// ============================================================================
// PERFORMANCE INTEGRATION TESTS
// ============================================================================

func TestPipelinePerformance_HighThroughput(t *testing.T) {
	// Test high throughput performance
	pipeline := go_core.NewLaravelPipeline[string]()

	// Add multiple stages
	for i := 0; i < 5; i++ {
		stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
			return next(data)
		})
		pipeline.Through(stage)
	}

	// Test performance
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		err := pipeline.Send(context.Background(), "test")
		assert.NoError(t, err)
	}

	duration := time.Since(start)
	t.Logf("Processed %d iterations in %v (%d ops/sec)",
		iterations, duration, int(float64(iterations)/duration.Seconds()))
}

func TestAdvancedChannelsPerformance_HighLoad(t *testing.T) {
	// Test advanced channels performance under high load
	manager := go_core.NewChannelManager(nil)

	// Create input
	input := make(chan int, 1000)

	// Send data
	go func() {
		defer close(input)
		for i := 0; i < 1000; i++ {
			input <- i
		}
	}()

	// Create processing pipeline
	stage1 := func(item int) int {
		return item * 2
	}

	stage2 := func(item int) int {
		return item + 1
	}

	stage3 := func(item int) int {
		return item * 3
	}

	// Apply pipeline
	output := go_core.CreatePipeline(manager, input, stage1, stage2, stage3)

	// Collect results
	start := time.Now()
	var results []int
	for item := range output {
		results = append(results, item)
	}
	duration := time.Since(start)

	t.Logf("Processed %d items through 3-stage pipeline in %v (%d ops/sec)",
		len(results), duration, int(float64(len(results))/duration.Seconds()))

	assert.Len(t, results, 1000)
}

func TestPipelineConcurrency_StressTest(t *testing.T) {
	// Test pipeline under concurrent stress
	pipeline := go_core.NewLaravelPipeline[string]()

	// Add processing stage
	stage := go_core.LaravelPipelineStageFunc[string](func(ctx context.Context, data string, next func(string) error) error {
		return next("processed_" + data)
	})

	pipeline.Through(stage)

	// Run concurrent operations
	concurrency := 50
	iterations := 20
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := pipeline.Send(context.Background(), fmt.Sprintf("data_%d_%d", id, j))
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
// BACKPRESSURE AND MEMORY TESTS
// ============================================================================

func TestBackpressureHandling_LargeDataSets(t *testing.T) {
	// Test backpressure handling with large datasets
	manager := go_core.NewChannelManager(nil)

	// Create input with limited buffer
	input := make(chan int, 10)

	// Send large amount of data
	go func() {
		defer close(input)
		for i := 0; i < 1000; i++ {
			input <- i
		}
	}()

	// Slow processing stage to create backpressure
	slowStage := func(item int) int {
		time.Sleep(1 * time.Millisecond)
		return item * 2
	}

	output := go_core.CreatePipeline(manager, input, slowStage)

	// Collect results
	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Should have processed all items despite backpressure
	assert.Len(t, results, 1000)
	for i := 0; i < 1000; i++ {
		assert.Contains(t, results, i*2)
	}
}

func TestMemoryEfficiency_ChannelReuse(t *testing.T) {
	// Test memory efficiency with channel reuse
	manager := go_core.NewChannelManager(nil)

	// Create multiple pipelines and test memory usage
	for run := 0; run < 10; run++ {
		input := make(chan int, 100)

		go func() {
			defer close(input)
			for i := 0; i < 100; i++ {
				input <- i
			}
		}()

		output := go_core.CreatePipeline(manager, input, func(item int) int {
			return item * 2
		})

		var results []int
		for item := range output {
			results = append(results, item)
		}

		assert.Len(t, results, 100)
	}
}

// ============================================================================
// ERROR RECOVERY AND RESILIENCE TESTS
// ============================================================================

func TestErrorRecovery_PartialFailures(t *testing.T) {
	// Test error recovery with partial failures
	manager := go_core.NewChannelManager(nil)

	// Create input
	input := make(chan int, 10)
	go func() {
		defer close(input)
		for i := 0; i < 10; i++ {
			input <- i
		}
	}()

	// Operation that fails for specific items
	operation := func(item int) error {
		if item%3 == 0 {
			return fmt.Errorf("failure for item %d", item)
		}
		return nil
	}

	retryOutput := go_core.RetryWithBackoff(manager, input, operation)

	var results []int
	for item := range retryOutput {
		results = append(results, item)
	}

	// Should have items that didn't fail
	assert.Len(t, results, 6) // Items not divisible by 3
	for _, item := range results {
		assert.NotEqual(t, 0, item%3)
	}
}

func TestResilience_ContextCancellation(t *testing.T) {
	// Test resilience under context cancellation
	manager := go_core.NewChannelManager(nil)

	// Create input
	input := make(chan int, 10)
	go func() {
		defer close(input)
		for i := 0; i < 10; i++ {
			input <- i
		}
	}()

	// Create context-aware channel with timeout that allows some items
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	cac := go_core.NewContextAwareChannel[int](ctx, manager)

	// Process with context - each item takes 7ms, so 1 or more items should be processed
	processor := func(ctx context.Context, item int) error {
		select {
		case <-time.After(7 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	output := cac.ProcessWithContext(input, processor)

	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Should process at least 1 item, but not all
	assert.GreaterOrEqual(t, len(results), 1)
	assert.Less(t, len(results), 10)
	// All processed items should be from the start of the input
	for i, item := range results {
		assert.Equal(t, i, item)
	}
}
