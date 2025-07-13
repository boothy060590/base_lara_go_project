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
// CHANNEL MANAGER TESTS
// ============================================================================

func TestChannelManager_New(t *testing.T) {
	config := go_core.DefaultChannelConfig()
	manager := go_core.NewChannelManager(config)

	assert.NotNil(t, manager)
}

func TestChannelManager_NewWithNilConfig(t *testing.T) {
	manager := go_core.NewChannelManager(nil)

	assert.NotNil(t, manager)
}

// ============================================================================
// FAN OUT TESTS
// ============================================================================

func TestFanOut_Basic(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 10)

	// Send some data
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	// Create fan out
	outputs := go_core.FanOut(manager, input, 3)

	// Collect results
	var results [][]int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, output := range outputs {
		wg.Add(1)
		go func(index int, ch <-chan int) {
			defer wg.Done()
			var result []int
			for item := range ch {
				result = append(result, item)
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(i, output)
	}

	wg.Wait()

	// With round-robin distribution, each output gets a subset of items
	// Total items should be distributed across all outputs
	totalItems := 0
	for _, result := range results {
		totalItems += len(result)
	}
	assert.Equal(t, 5, totalItems)

	// Each output should have at least one item
	for _, result := range results {
		assert.Greater(t, len(result), 0)
	}
}

func TestFanOut_ZeroOutputs(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int)

	outputs := go_core.FanOut(manager, input, 0)
	assert.Nil(t, outputs)
}

func TestFanOut_EmptyInput(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int)

	go func() {
		close(input)
	}()

	outputs := go_core.FanOut(manager, input, 2)

	// All outputs should be closed
	for _, output := range outputs {
		_, ok := <-output
		assert.False(t, ok)
	}
}

// ============================================================================
// FAN IN TESTS
// ============================================================================

func TestFanIn_Basic(t *testing.T) {
	manager := go_core.NewChannelManager(nil)

	// Create input channels
	input1 := make(chan int, 5)
	input2 := make(chan int, 5)

	// Send data to inputs
	go func() {
		defer close(input1)
		for i := 0; i < 3; i++ {
			input1 <- i
		}
	}()

	go func() {
		defer close(input2)
		for i := 3; i < 6; i++ {
			input2 <- i
		}
	}()

	// Create fan in
	output := go_core.FanIn(manager, []<-chan int{input1, input2})

	// Collect results
	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Should have all items from both inputs
	assert.Len(t, results, 6)
	for i := 0; i < 6; i++ {
		assert.Contains(t, results, i)
	}
}

func TestFanIn_EmptyInputs(t *testing.T) {
	manager := go_core.NewChannelManager(nil)

	output := go_core.FanIn(manager, []<-chan int{})

	// Output should be closed immediately
	_, ok := <-output
	assert.False(t, ok)
}

func TestFanIn_SingleInput(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 3)

	go func() {
		defer close(input)
		for i := 0; i < 3; i++ {
			input <- i
		}
	}()

	output := go_core.FanIn(manager, []<-chan int{input})

	var results []int
	for item := range output {
		results = append(results, item)
	}

	assert.Len(t, results, 3)
	for i := 0; i < 3; i++ {
		assert.Contains(t, results, i)
	}
}

// ============================================================================
// PIPELINE TESTS
// ============================================================================

func TestCreatePipeline_Basic(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 5)

	// Send data
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	// Create pipeline stages
	stage1 := func(item int) int {
		return item * 2
	}

	stage2 := func(item int) int {
		return item + 1
	}

	// Create pipeline
	output := go_core.CreatePipeline(manager, input, stage1, stage2)

	// Collect results
	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Verify transformation: (i * 2) + 1
	expected := []int{1, 3, 5, 7, 9}
	assert.Equal(t, expected, results)
}

func TestCreatePipeline_EmptyStages(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 3)

	go func() {
		defer close(input)
		for i := 0; i < 3; i++ {
			input <- i
		}
	}()

	output := go_core.CreatePipeline(manager, input)

	var results []int
	for item := range output {
		results = append(results, item)
	}

	assert.Len(t, results, 3)
	for i := 0; i < 3; i++ {
		assert.Contains(t, results, i)
	}
}

func TestCreatePipeline_ErrorHandling(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 3)

	go func() {
		defer close(input)
		for i := 0; i < 3; i++ {
			input <- i
		}
	}()

	// Stage that might panic
	stage := func(item int) int {
		if item == 1 {
			panic("test panic")
		}
		return item * 2
	}

	output := go_core.CreatePipeline(manager, input, stage)

	// Should handle panic gracefully
	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Should have processed items that didn't cause panic
	assert.Len(t, results, 2)
	assert.Contains(t, results, 0) // 0 * 2 = 0
	assert.Contains(t, results, 4) // 2 * 2 = 4
}

// ============================================================================
// BATCH TESTS
// ============================================================================

func TestBatch_Basic(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 10)

	// Send data
	go func() {
		defer close(input)
		for i := 0; i < 7; i++ {
			input <- i
		}
	}()

	// Create batch processor
	output := go_core.Batch(manager, input, 3)

	// Collect results
	var results [][]int
	for batch := range output {
		results = append(results, batch)
	}

	// Should have 3 batches: [0,1,2], [3,4,5], [6]
	assert.Len(t, results, 3)
	assert.Len(t, results[0], 3)
	assert.Len(t, results[1], 3)
	assert.Len(t, results[2], 1)

	// Verify content
	assert.Equal(t, []int{0, 1, 2}, results[0])
	assert.Equal(t, []int{3, 4, 5}, results[1])
	assert.Equal(t, []int{6}, results[2])
}

func TestBatch_ExactMultiple(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 6)

	go func() {
		defer close(input)
		for i := 0; i < 6; i++ {
			input <- i
		}
	}()

	output := go_core.Batch(manager, input, 2)

	var results [][]int
	for batch := range output {
		results = append(results, batch)
	}

	assert.Len(t, results, 3)
	for _, batch := range results {
		assert.Len(t, batch, 2)
	}
}

func TestBatch_EmptyInput(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int)

	go func() {
		close(input)
	}()

	output := go_core.Batch(manager, input, 3)

	// Should be closed immediately
	_, ok := <-output
	assert.False(t, ok)
}

// ============================================================================
// RATE LIMIT TESTS
// ============================================================================

func TestRateLimit_Basic(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 5)

	go func() {
		defer close(input)
		for i := 0; i < 3; i++ {
			input <- i
		}
	}()

	// Rate limit to 1 item per 100ms
	output := go_core.RateLimit(manager, input, 100*time.Millisecond)

	start := time.Now()
	var results []int
	for item := range output {
		results = append(results, item)
	}
	duration := time.Since(start)

	// Should have all items
	assert.Len(t, results, 3)
	for i := 0; i < 3; i++ {
		assert.Contains(t, results, i)
	}

	// Should have taken at least 200ms (2 intervals between 3 items)
	assert.GreaterOrEqual(t, duration, 200*time.Millisecond)
}

// ============================================================================
// RETRY WITH BACKOFF TESTS
// ============================================================================

func TestRetryWithBackoff_Basic(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 3)

	go func() {
		defer close(input)
		for i := 0; i < 3; i++ {
			input <- i
		}
	}()

	attempts := make(map[int]int)
	operation := func(item int) error {
		attempts[item]++
		if attempts[item] < 2 {
			return fmt.Errorf("temporary error")
		}
		return nil
	}

	output := go_core.RetryWithBackoff(manager, input, operation)

	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Should have all items
	assert.Len(t, results, 3)
	for i := 0; i < 3; i++ {
		assert.Contains(t, results, i)
		assert.Equal(t, 2, attempts[i]) // Each item should have been attempted twice
	}
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 2)

	go func() {
		defer close(input)
		input <- 1
		input <- 2
	}()

	operation := func(item int) error {
		return fmt.Errorf("permanent error")
	}

	output := go_core.RetryWithBackoff(manager, input, operation)

	var results []int
	for item := range output {
		results = append(results, item)
	}

	// Should have no results due to permanent errors
	assert.Len(t, results, 0)
}

// ============================================================================
// CHANNEL UTILS TESTS
// ============================================================================

func TestChannelUtils_Collect(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	utils := go_core.NewChannelUtils(manager)

	input := make(chan int, 5)
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	results := go_core.Collect(utils, input)

	assert.Len(t, results, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, results, i)
	}
}

func TestChannelUtils_CollectWithTimeout(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	utils := go_core.NewChannelUtils(manager)

	input := make(chan int, 3)
	go func() {
		defer close(input)
		for i := 0; i < 3; i++ {
			input <- i
		}
	}()

	results := go_core.CollectWithTimeout(utils, input, 1*time.Second)

	assert.Len(t, results, 3)
	for i := 0; i < 3; i++ {
		assert.Contains(t, results, i)
	}
}

func TestChannelUtils_CollectN(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	utils := go_core.NewChannelUtils(manager)

	input := make(chan int, 10)
	go func() {
		defer close(input)
		for i := 0; i < 10; i++ {
			input <- i
		}
	}()

	results := go_core.CollectN(utils, input, 5)

	assert.Len(t, results, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, results, i)
	}
}

func TestChannelUtils_Filter(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	utils := go_core.NewChannelUtils(manager)

	input := make(chan int, 10)
	go func() {
		defer close(input)
		for i := 0; i < 10; i++ {
			input <- i
		}
	}()

	// Filter even numbers
	predicate := func(item int) bool {
		return item%2 == 0
	}

	output := go_core.Filter(utils, input, predicate)

	var results []int
	for item := range output {
		results = append(results, item)
	}

	assert.Len(t, results, 5)
	for _, item := range results {
		assert.Equal(t, 0, item%2)
	}
}

func TestChannelUtils_Map(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	utils := go_core.NewChannelUtils(manager)

	input := make(chan int, 5)
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	// Map to strings
	mapper := func(item int) string {
		return fmt.Sprintf("item_%d", item)
	}

	output := go_core.Map(utils, input, mapper)

	var results []string
	for item := range output {
		results = append(results, item)
	}

	assert.Len(t, results, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, results, fmt.Sprintf("item_%d", i))
	}
}

func TestChannelUtils_Reduce(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	utils := go_core.NewChannelUtils(manager)

	input := make(chan int, 5)
	go func() {
		defer close(input)
		for i := 1; i <= 5; i++ {
			input <- i
		}
	}()

	// Sum all numbers
	reducer := func(acc, item int) int {
		return acc + item
	}

	result := go_core.Reduce(utils, input, 0, reducer)

	assert.Equal(t, 15, result) // 1 + 2 + 3 + 4 + 5 = 15
}

// ============================================================================
// CONTEXT AWARE CHANNEL TESTS
// ============================================================================

func TestContextAwareChannel_ProcessWithContext(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	ctx := context.Background()
	cac := go_core.NewContextAwareChannel[int](ctx, manager)

	input := make(chan int, 5)
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	processed := make([]int, 0)
	processor := func(ctx context.Context, item int) error {
		processed = append(processed, item*2)
		return nil
	}

	output := cac.ProcessWithContext(input, processor)

	var results []int
	for item := range output {
		results = append(results, item)
	}

	assert.Len(t, results, 5)
	assert.Len(t, processed, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, results, i)
		assert.Contains(t, processed, i*2)
	}
}

func TestContextAwareChannel_FanOutWithContext(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	ctx := context.Background()
	cac := go_core.NewContextAwareChannel[int](ctx, manager)

	input := make(chan int, 5)
	go func() {
		defer close(input)
		for i := 0; i < 5; i++ {
			input <- i
		}
	}()

	outputs := cac.FanOutWithContext(input, 2)

	var results [][]int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, output := range outputs {
		wg.Add(1)
		go func(index int, ch <-chan int) {
			defer wg.Done()
			var result []int
			for item := range ch {
				result = append(result, item)
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(i, output)
	}

	wg.Wait()

	assert.Len(t, results, 2)
	// With round-robin distribution, items should be distributed across outputs
	// Total items should equal input (5 items)
	totalItems := 0
	for _, result := range results {
		totalItems += len(result)
	}
	assert.Equal(t, 5, totalItems)
	// Each output should have some items (not necessarily equal)
	for _, result := range results {
		assert.Greater(t, len(result), 0)
	}
}

// ============================================================================
// GLOBAL FUNCTION TESTS
// ============================================================================

func TestGlobalFunctions(t *testing.T) {
	input := make(chan int, 3)
	go func() {
		defer close(input)
		for i := 0; i < 3; i++ {
			input <- i
		}
	}()

	// Test global FanOut
	outputs := go_core.FanOutGlobal(input, 2)
	assert.Len(t, outputs, 2)

	// Test global FanIn
	output := go_core.FanInGlobal(outputs)
	var results []int
	for item := range output {
		results = append(results, item)
	}
	assert.Len(t, results, 3)
}

// ============================================================================
// PERFORMANCE TESTS
// ============================================================================

func TestChannelPerformance(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 1000)

	// Send data
	go func() {
		defer close(input)
		for i := 0; i < 1000; i++ {
			input <- i
		}
	}()

	// Create a simple pipeline
	stage := func(item int) int {
		return item * 2
	}

	output := go_core.CreatePipeline(manager, input, stage)

	start := time.Now()
	var results []int
	for item := range output {
		results = append(results, item)
	}
	duration := time.Since(start)

	t.Logf("Processed %d items in %v (%d ops/sec)",
		len(results), duration, int(float64(len(results))/duration.Seconds()))

	assert.Len(t, results, 1000)
}

func TestChannelConcurrency(t *testing.T) {
	manager := go_core.NewChannelManager(nil)
	input := make(chan int, 100)

	// Send data concurrently
	go func() {
		defer close(input)
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					input <- id*10 + j
				}
			}(i)
		}
		wg.Wait()
	}()

	// Fan out to multiple workers
	outputs := go_core.FanOut(manager, input, 5)

	var results [][]int
	var wg sync.WaitGroup
	var mu sync.Mutex

	start := time.Now()
	for i, output := range outputs {
		wg.Add(1)
		go func(index int, ch <-chan int) {
			defer wg.Done()
			var result []int
			for item := range ch {
				result = append(result, item)
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(i, output)
	}

	wg.Wait()
	duration := time.Since(start)

	// Each output should have received some items
	totalItems := 0
	for i, result := range results {
		totalItems += len(result)
		t.Logf("Worker %d processed %d items", i, len(result))
	}

	t.Logf("Processed %d items in %v (%.2f ops/sec)", totalItems, duration, float64(totalItems)/duration.Seconds())

	if totalItems != 100 {
		t.Errorf("Expected 100 items processed, got %d (missing %d)", totalItems, 100-totalItems)
	}
	assert.Equal(t, 100, totalItems)
}

// ============================================================================
// DATA LOSS DETECTION TESTS
// ============================================================================

func TestDataLossDetection(t *testing.T) {
	t.Run("HighLoadNoDataLoss", func(t *testing.T) {
		manager := go_core.NewChannelManager(nil)
		input := make(chan int, 1000)

		// Send large amount of data
		go func() {
			defer close(input)
			for i := 0; i < 1000; i++ {
				input <- i
			}
		}()

		// Fan out to multiple workers
		outputs := go_core.FanOut(manager, input, 10)

		var results [][]int
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i, output := range outputs {
			wg.Add(1)
			go func(index int, ch <-chan int) {
				defer wg.Done()
				var result []int
				for item := range ch {
					result = append(result, item)
				}
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}(i, output)
		}

		wg.Wait()

		// Verify no data loss
		totalItems := 0
		for i, result := range results {
			totalItems += len(result)
			t.Logf("Worker %d processed %d items", i, len(result))
		}

		assert.Equal(t, 1000, totalItems, "Data loss detected: expected 1000 items, got %d", totalItems)
	})

	t.Run("BackpressureNoDataLoss", func(t *testing.T) {
		manager := go_core.NewChannelManager(nil)
		input := make(chan int, 10) // Small buffer to create backpressure

		// Send data faster than it can be processed
		go func() {
			defer close(input)
			for i := 0; i < 100; i++ {
				input <- i
			}
		}()

		// Slow processing stage
		slowStage := func(item int) int {
			time.Sleep(1 * time.Millisecond)
			return item * 2
		}

		output := go_core.CreatePipeline(manager, input, slowStage)

		var results []int
		for item := range output {
			results = append(results, item)
		}

		// Verify no data loss despite backpressure
		assert.Equal(t, 100, len(results), "Data loss detected under backpressure: expected 100 items, got %d", len(results))

		// Verify all items were processed correctly
		for i := 0; i < 100; i++ {
			assert.Contains(t, results, i*2)
		}
	})

	t.Run("ConcurrentFanOutNoDataLoss", func(t *testing.T) {
		manager := go_core.NewChannelManager(nil)
		input := make(chan int, 500)

		// Send data concurrently
		go func() {
			defer close(input)
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < 50; j++ {
						input <- id*50 + j
					}
				}(i)
			}
			wg.Wait()
		}()

		// Fan out to multiple workers
		outputs := go_core.FanOut(manager, input, 5)

		var results [][]int
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i, output := range outputs {
			wg.Add(1)
			go func(index int, ch <-chan int) {
				defer wg.Done()
				var result []int
				for item := range ch {
					result = append(result, item)
				}
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}(i, output)
		}

		wg.Wait()

		// Verify no data loss
		totalItems := 0
		for i, result := range results {
			totalItems += len(result)
			t.Logf("Worker %d processed %d items", i, len(result))
		}

		assert.Equal(t, 500, totalItems, "Data loss detected in concurrent fan-out: expected 500 items, got %d", totalItems)
	})

	t.Run("RetryNoDataLoss", func(t *testing.T) {
		manager := go_core.NewChannelManager(nil)
		input := make(chan int, 10)

		go func() {
			defer close(input)
			for i := 0; i < 10; i++ {
				input <- i
			}
		}()

		// Operation that fails initially then succeeds
		attempts := make(map[int]int)
		operation := func(item int) error {
			attempts[item]++
			if attempts[item] < 3 {
				return fmt.Errorf("temporary error")
			}
			return nil
		}

		retryOutput := go_core.RetryWithBackoff(manager, input, operation)

		var results []int
		for item := range retryOutput {
			results = append(results, item)
		}

		// Verify no data loss despite retries
		assert.Equal(t, 10, len(results), "Data loss detected in retry scenario: expected 10 items, got %d", len(results))

		// Verify all items were attempted the correct number of times
		for i := 0; i < 10; i++ {
			assert.Equal(t, 3, attempts[i], "Item %d was not retried correctly", i)
		}
	})
}
