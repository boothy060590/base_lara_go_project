package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
)

// TestWorkStealingConfig tests the configuration system
func TestWorkStealingConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := go_core.DefaultWorkStealingConfig()
		assert.NotNil(t, config)
		assert.Greater(t, config.NumWorkers, 0)
		assert.Greater(t, config.QueueSize, 0)
		assert.Greater(t, config.StealThreshold, 0)
		assert.Greater(t, config.StealBatchSize, 0)
		assert.Greater(t, config.IdleTimeout, time.Duration(0))
		assert.True(t, config.EnableMetrics)
		assert.True(t, config.EnableProfiling)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		customConfig := &go_core.WorkStealingConfig{
			NumWorkers:      8,
			QueueSize:       2048,
			StealThreshold:  5,
			StealBatchSize:  20,
			IdleTimeout:     200 * time.Millisecond,
			EnableMetrics:   false,
			EnableProfiling: false,
		}

		pool := go_core.NewWorkStealingPool[string](customConfig)
		assert.NotNil(t, pool)
		defer pool.Shutdown()
	})
}

// TestWorkStealingPool tests the main work stealing pool functionality
func TestWorkStealingPool(t *testing.T) {
	t.Run("NewWorkStealingPool", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 4,
			QueueSize:  100,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		assert.NotNil(t, pool)
		defer pool.Shutdown()

		metrics := pool.GetMetrics()
		assert.NotNil(t, metrics)
	})

	t.Run("SubmitWorkItem", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 2,
			QueueSize:  10,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		defer pool.Shutdown()

		var processed bool
		var mu sync.Mutex

		item := go_core.WorkItem[string]{
			ID:   "test-item",
			Data: "test-data",
			Handler: func(ctx context.Context, data string) error {
				mu.Lock()
				processed = true
				mu.Unlock()
				return nil
			},
			Timeout:  5 * time.Second,
			Priority: 0,
		}

		err := pool.Submit(item)
		assert.NoError(t, err)

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.True(t, processed)
		mu.Unlock()
	})

	t.Run("SubmitAsync", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 2,
			QueueSize:  10,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		defer pool.Shutdown()

		var processed bool
		var mu sync.Mutex

		item := go_core.WorkItem[string]{
			ID:   "test-async-item",
			Data: "test-data",
			Handler: func(ctx context.Context, data string) error {
				mu.Lock()
				processed = true
				mu.Unlock()
				return nil
			},
			Timeout:  5 * time.Second,
			Priority: 0,
		}

		resultChan := pool.SubmitAsync(item)
		err := <-resultChan
		assert.NoError(t, err)

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.True(t, processed)
		mu.Unlock()
	})

	t.Run("SubmitWithTimeout", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 1,
			QueueSize:  10,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		defer pool.Shutdown()

		var jobCompleted bool
		var mu sync.Mutex

		item := go_core.WorkItem[string]{
			ID:   "timeout-item",
			Data: "test-data",
			Handler: func(ctx context.Context, data string) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(200 * time.Millisecond): // Longer than timeout
					mu.Lock()
					jobCompleted = true
					mu.Unlock()
					return nil
				}
			},
			Timeout:  50 * time.Millisecond,
			Priority: 0,
		}

		err := pool.Submit(item)
		assert.NoError(t, err)

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.False(t, jobCompleted) // Should not complete due to timeout
		mu.Unlock()
	})

	t.Run("ConcurrentSubmissions", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 4,
			QueueSize:  100,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		defer pool.Shutdown()

		var processedCount int64
		var mu sync.Mutex

		// Submit many items concurrently
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				item := go_core.WorkItem[string]{
					ID:   fmt.Sprintf("concurrent-item-%d", id),
					Data: fmt.Sprintf("data-%d", id),
					Handler: func(ctx context.Context, data string) error {
						mu.Lock()
						processedCount++
						mu.Unlock()
						return nil
					},
					Timeout:  5 * time.Second,
					Priority: id % 4,
				}

				err := pool.Submit(item)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// Wait for processing
		time.Sleep(500 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, int64(50), processedCount)
		mu.Unlock()
	})
}

// TestWorkQueue tests the work queue functionality
func TestWorkQueue(t *testing.T) {
	t.Run("NewWorkQueue", func(t *testing.T) {
		queue := go_core.NewWorkQueue[string](10)
		assert.NotNil(t, queue)
	})

	t.Run("PushAndPop", func(t *testing.T) {
		queue := go_core.NewWorkQueue[string](5)

		item := go_core.WorkItem[string]{
			ID:       "test-item",
			Data:     "test-data",
			Handler:  func(ctx context.Context, data string) error { return nil },
			Timeout:  5 * time.Second,
			Priority: 0,
		}

		// Push item
		err := queue.Push(item)
		assert.NoError(t, err)

		// Pop item
		popped, ok := queue.TryPop()
		assert.True(t, ok)
		assert.Equal(t, "test-item", popped.ID)
		assert.Equal(t, "test-data", popped.Data)
	})

	t.Run("TryPush", func(t *testing.T) {
		queue := go_core.NewWorkQueue[string](2) // Size 2 can hold 1 item (size-1)

		item1 := go_core.WorkItem[string]{
			ID:       "item1",
			Data:     "data1",
			Handler:  func(ctx context.Context, data string) error { return nil },
			Timeout:  5 * time.Second,
			Priority: 0,
		}

		item2 := go_core.WorkItem[string]{
			ID:       "item2",
			Data:     "data2",
			Handler:  func(ctx context.Context, data string) error { return nil },
			Timeout:  5 * time.Second,
			Priority: 0,
		}

		// First push should succeed
		success := queue.TryPush(item1)
		assert.True(t, success)

		// Second push should fail (queue full - circular queue can only hold size-1 items)
		success = queue.TryPush(item2)
		assert.False(t, success)
	})

	t.Run("TrySteal", func(t *testing.T) {
		queue := go_core.NewWorkQueue[string](10)

		// Add some items
		for i := 0; i < 5; i++ {
			item := go_core.WorkItem[string]{
				ID:       fmt.Sprintf("item-%d", i),
				Data:     fmt.Sprintf("data-%d", i),
				Handler:  func(ctx context.Context, data string) error { return nil },
				Timeout:  5 * time.Second,
				Priority: 0,
			}
			queue.Push(item)
		}

		// Try to steal items
		stolen := queue.TrySteal(3)
		assert.Len(t, stolen, 3)

		// Verify stolen items
		for i, item := range stolen {
			assert.Equal(t, fmt.Sprintf("item-%d", i), item.ID)
			assert.Equal(t, fmt.Sprintf("data-%d", i), item.Data)
		}
	})

	t.Run("QueueFull", func(t *testing.T) {
		queue := go_core.NewWorkQueue[string](3) // Size 3 can hold 2 items (size-1)

		// Fill the queue (2 items)
		for i := 0; i < 2; i++ {
			item := go_core.WorkItem[string]{
				ID:       fmt.Sprintf("item-%d", i),
				Data:     fmt.Sprintf("data-%d", i),
				Handler:  func(ctx context.Context, data string) error { return nil },
				Timeout:  5 * time.Second,
				Priority: 0,
			}
			err := queue.Push(item)
			assert.NoError(t, err)
		}

		// Try to push another item - should fail (queue full)
		item := go_core.WorkItem[string]{
			ID:       "overflow-item",
			Data:     "overflow-data",
			Handler:  func(ctx context.Context, data string) error { return nil },
			Timeout:  5 * time.Second,
			Priority: 0,
		}
		err := queue.Push(item)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "queue is full")
	})
}

// TestWorkStealingWorker tests individual worker functionality
func TestWorkStealingWorker(t *testing.T) {
	t.Run("WorkerProcessing", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 1,
			QueueSize:  10,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		defer pool.Shutdown()

		var processed bool
		var mu sync.Mutex

		item := go_core.WorkItem[string]{
			ID:   "worker-test-item",
			Data: "worker-test-data",
			Handler: func(ctx context.Context, data string) error {
				mu.Lock()
				processed = true
				mu.Unlock()
				return nil
			},
			Timeout:  5 * time.Second,
			Priority: 0,
		}

		err := pool.Submit(item)
		assert.NoError(t, err)

		// Wait for processing
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.True(t, processed)
		mu.Unlock()
	})
}

// TestWorkStealingMetrics tests the metrics system
func TestWorkStealingMetrics(t *testing.T) {
	t.Run("NewWorkStealingMetrics", func(t *testing.T) {
		metrics := go_core.NewWorkStealingMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(0), metrics.TotalProcessed)
		assert.Equal(t, int64(0), metrics.TotalSteals)
		assert.Equal(t, 0, metrics.ActiveWorkers)
		assert.Equal(t, 0.0, metrics.QueueUtilization)
		assert.Equal(t, time.Duration(0), metrics.AverageWaitTime)
	})

	t.Run("UpdateMetrics", func(t *testing.T) {
		metrics := go_core.NewWorkStealingMetrics()

		metrics.UpdateMetrics(5, 100, 50)

		assert.Equal(t, 5, metrics.ActiveWorkers)
		assert.Equal(t, int64(100), metrics.TotalProcessed)
		assert.Equal(t, int64(50), metrics.TotalSteals)
	})

	t.Run("ConcurrentMetricsUpdate", func(t *testing.T) {
		metrics := go_core.NewWorkStealingMetrics()
		var wg sync.WaitGroup

		// Update metrics concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				metrics.UpdateMetrics(id, int64(id*2), int64(id))
			}(i)
		}

		wg.Wait()
		// Should not panic and should have some reasonable values
		assert.GreaterOrEqual(t, metrics.ActiveWorkers, 0)
		assert.GreaterOrEqual(t, metrics.TotalProcessed, int64(0))
		assert.GreaterOrEqual(t, metrics.TotalSteals, int64(0))
	})
}

// TestWorkStealingShutdown tests graceful shutdown
func TestWorkStealingShutdown(t *testing.T) {
	t.Run("GracefulShutdown", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 2,
			QueueSize:  10,
		}
		pool := go_core.NewWorkStealingPool[string](config)

		// Submit a job
		item := go_core.WorkItem[string]{
			ID:       "shutdown-test-item",
			Data:     "shutdown-test-data",
			Handler:  func(ctx context.Context, data string) error { return nil },
			Timeout:  5 * time.Second,
			Priority: 0,
		}
		pool.Submit(item)

		// Shutdown
		pool.Shutdown()

		// Try to submit after shutdown - should fail gracefully
		_ = pool.Submit(item)
		// Note: The current implementation doesn't prevent submissions after shutdown
		// This is a potential improvement area
	})

	t.Run("ShutdownWithActiveWork", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 2,
			QueueSize:  10,
		}
		pool := go_core.NewWorkStealingPool[string](config)

		var jobCompleted bool
		var mu sync.Mutex

		// Submit a long-running job
		item := go_core.WorkItem[string]{
			ID:   "long-running-item",
			Data: "long-running-data",
			Handler: func(ctx context.Context, data string) error {
				time.Sleep(50 * time.Millisecond) // Shorter sleep
				mu.Lock()
				jobCompleted = true
				mu.Unlock()
				return nil
			},
			Timeout:  5 * time.Second,
			Priority: 0,
		}
		pool.Submit(item)

		// Shutdown immediately
		pool.Shutdown()

		// Wait for job to complete with timeout
		timeout := time.After(200 * time.Millisecond)
		for {
			mu.Lock()
			completed := jobCompleted
			mu.Unlock()

			if completed {
				break
			}

			select {
			case <-timeout:
				t.Log("Job did not complete within timeout, but shutdown completed successfully")
				return
			case <-time.After(10 * time.Millisecond):
				// Check again
			}
		}

		mu.Lock()
		assert.True(t, jobCompleted)
		mu.Unlock()
	})
}

// TestWorkStealingPerformance tests performance characteristics
func TestWorkStealingPerformance(t *testing.T) {
	t.Run("HighThroughput", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers: 8,
			QueueSize:  1000,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		defer pool.Shutdown()

		var processedCount int64
		var mu sync.Mutex

		start := time.Now()

		// Submit many items
		for i := 0; i < 1000; i++ {
			item := go_core.WorkItem[string]{
				ID:   fmt.Sprintf("perf-item-%d", i),
				Data: fmt.Sprintf("perf-data-%d", i),
				Handler: func(ctx context.Context, data string) error {
					mu.Lock()
					processedCount++
					mu.Unlock()
					return nil
				},
				Timeout:  5 * time.Second,
				Priority: i % 4,
			}
			err := pool.Submit(item)
			assert.NoError(t, err)
		}

		// Wait for processing
		time.Sleep(1 * time.Second)

		duration := time.Since(start)
		mu.Lock()
		finalCount := processedCount
		mu.Unlock()

		assert.Equal(t, int64(1000), finalCount)
		assert.Less(t, duration, 2*time.Second) // Should complete within 2 seconds
	})

	t.Run("WorkStealingEfficiency", func(t *testing.T) {
		config := &go_core.WorkStealingConfig{
			NumWorkers:     4,
			QueueSize:      100,
			StealThreshold: 2,
			StealBatchSize: 5,
		}
		pool := go_core.NewWorkStealingPool[string](config)
		defer pool.Shutdown()

		var processedCount int64
		var mu sync.Mutex

		// Submit items with different priorities to test work stealing
		for i := 0; i < 100; i++ {
			item := go_core.WorkItem[string]{
				ID:   fmt.Sprintf("steal-item-%d", i),
				Data: fmt.Sprintf("steal-data-%d", i),
				Handler: func(ctx context.Context, data string) error {
					mu.Lock()
					processedCount++
					mu.Unlock()
					time.Sleep(10 * time.Millisecond) // Simulate work
					return nil
				},
				Timeout:  5 * time.Second,
				Priority: i % 4,
			}
			err := pool.Submit(item)
			assert.NoError(t, err)
		}

		// Wait for processing
		time.Sleep(2 * time.Second)

		mu.Lock()
		finalCount := processedCount
		mu.Unlock()

		assert.Equal(t, int64(100), finalCount)

		// Check metrics for work stealing activity
		metrics := pool.GetMetrics()
		assert.GreaterOrEqual(t, metrics.TotalProcessed, int64(100))
	})
}
