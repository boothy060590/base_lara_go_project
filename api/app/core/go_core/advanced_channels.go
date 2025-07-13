package go_core

import (
	"context"
	"sync"
	"time"
)

// ============================================================================
// ADVANCED CHANNEL PATTERNS
// ============================================================================

// ChannelConfig defines configuration for channel operations
type ChannelConfig struct {
	BufferSize    int           `json:"buffer_size"`
	Timeout       time.Duration `json:"timeout"`
	MaxWorkers    int           `json:"max_workers"`
	EnableBackoff bool          `json:"enable_backoff"`
	BackoffDelay  time.Duration `json:"backoff_delay"`
	MaxRetries    int           `json:"max_retries"`
}

// DefaultChannelConfig returns sensible defaults for channel operations
func DefaultChannelConfig() *ChannelConfig {
	return &ChannelConfig{
		BufferSize:    1000,
		Timeout:       5 * time.Second, // Shorter timeout for operations
		MaxWorkers:    10,
		EnableBackoff: true,
		BackoffDelay:  10 * time.Millisecond, // Shorter backoff for retries
		MaxRetries:    5,                     // More retries with shorter delays
	}
}

// ChannelManager manages advanced channel operations
type ChannelManager struct {
	config *ChannelConfig
	mu     sync.RWMutex
}

// NewChannelManager creates a new channel manager
func NewChannelManager(config *ChannelConfig) *ChannelManager {
	if config == nil {
		config = DefaultChannelConfig()
	}

	return &ChannelManager{
		config: config,
	}
}

// ============================================================================
// ADVANCED CHANNEL PATTERNS
// ============================================================================

// FanOut splits a single input channel into multiple output channels
func FanOut[T any](cm *ChannelManager, input <-chan T, outputs int) []<-chan T {
	if outputs <= 0 {
		return nil
	}

	outputChannels := make([]chan T, outputs)
	for i := 0; i < outputs; i++ {
		outputChannels[i] = make(chan T, cm.config.BufferSize)
	}

	go func() {
		defer func() {
			for _, ch := range outputChannels {
				close(ch)
			}
		}()

		index := 0
		for item := range input {
			// Round-robin distribution with blocking/retry logic
			sent := false
			attempts := 0
			maxAttempts := outputs * cm.config.MaxRetries // Try each channel up to MaxRetries times

			for !sent && attempts < maxAttempts {
				ch := outputChannels[index%outputs]

				// Try non-blocking send first
				select {
				case ch <- item:
					// Item sent successfully
					sent = true
				default:
					// Channel is full, try next channel
					index++
					attempts++

					// Small delay to prevent busy waiting
					if attempts < maxAttempts {
						time.Sleep(cm.config.BackoffDelay)
					}
				}
			}

			if !sent {
				// If all channels are full, use blocking send with retry logic
				ch := outputChannels[index%outputs]
				retryCount := 0
				for !sent && retryCount < cm.config.MaxRetries {
					select {
					case ch <- item:
						sent = true
					case <-time.After(cm.config.Timeout):
						// Timeout reached, try next channel
						index++
						retryCount++
						ch = outputChannels[index%outputs]
					}
				}

				if !sent {
					// Final attempt - use blocking send without timeout
					ch <- item
				}
			}

			index++
		}
	}()

	// Convert to read-only channels
	result := make([]<-chan T, outputs)
	for i, ch := range outputChannels {
		result[i] = ch
	}

	return result
}

// FanIn merges multiple input channels into a single output channel
func FanIn[T any](cm *ChannelManager, inputs []<-chan T) <-chan T {
	output := make(chan T, cm.config.BufferSize)

	go func() {
		defer close(output)

		var wg sync.WaitGroup
		for _, input := range inputs {
			wg.Add(1)
			go func(ch <-chan T) {
				defer wg.Done()
				for item := range ch {
					// Use blocking/retry logic to prevent data loss
					sent := false
					retryCount := 0

					for !sent && retryCount < cm.config.MaxRetries {
						// Try non-blocking send first
						select {
						case output <- item:
							sent = true
						default:
							// Channel is full, use blocking send with timeout
							select {
							case output <- item:
								sent = true
							case <-time.After(cm.config.Timeout):
								retryCount++
								// Small delay before retry
								time.Sleep(cm.config.BackoffDelay)
							}
						}
					}

					if !sent {
						// Final attempt - use blocking send without timeout
						output <- item
					}
				}
			}(input)
		}

		wg.Wait()
	}()

	return output
}

// CreatePipeline creates a processing pipeline with multiple stages
func CreatePipeline[T any](cm *ChannelManager, input <-chan T, stages ...func(T) T) <-chan T {
	if len(stages) == 0 {
		return input
	}

	current := input
	for _, stage := range stages {
		current = applyStage(cm, current, stage)
	}

	return current
}

// applyStage applies a single processing stage
func applyStage[T any](cm *ChannelManager, input <-chan T, stage func(T) T) <-chan T {
	output := make(chan T, cm.config.BufferSize)

	go func() {
		defer close(output)

		for item := range input {
			// Add panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log panic but continue processing other items
						// In a real implementation, you might want to log this
					}
				}()
				processed := stage(item)
				// Use blocking/retry logic to prevent data loss
				sent := false
				retryCount := 0

				for !sent && retryCount < cm.config.MaxRetries {
					// Try non-blocking send first
					select {
					case output <- processed:
						sent = true
					default:
						// Channel is full, use blocking send with timeout
						select {
						case output <- processed:
							sent = true
						case <-time.After(cm.config.Timeout):
							retryCount++
							// Small delay before retry
							time.Sleep(cm.config.BackoffDelay)
						}
					}
				}

				if !sent {
					// Final attempt - use blocking send without timeout
					output <- processed
				}
			}()
		}
	}()

	return output
}

// Batch processes items in batches
func Batch[T any](cm *ChannelManager, input <-chan T, batchSize int) <-chan []T {
	output := make(chan []T, cm.config.BufferSize)

	go func() {
		defer close(output)

		batch := make([]T, 0, batchSize)
		for item := range input {
			batch = append(batch, item)

			if len(batch) >= batchSize {
				// Use blocking/retry logic to prevent data loss
				sent := false
				retryCount := 0

				for !sent && retryCount < cm.config.MaxRetries {
					// Try non-blocking send first
					select {
					case output <- batch:
						batch = make([]T, 0, batchSize)
						sent = true
					default:
						// Channel is full, use blocking send with timeout
						select {
						case output <- batch:
							batch = make([]T, 0, batchSize)
							sent = true
						case <-time.After(cm.config.Timeout):
							retryCount++
							// Small delay before retry
							time.Sleep(cm.config.BackoffDelay)
						}
					}
				}

				if !sent {
					// Final attempt - use blocking send without timeout
					output <- batch
					batch = make([]T, 0, batchSize)
				}
			}
		}

		// Send remaining items
		if len(batch) > 0 {
			// Use blocking/retry logic to prevent data loss
			sent := false
			retryCount := 0

			for !sent && retryCount < cm.config.MaxRetries {
				// Try non-blocking send first
				select {
				case output <- batch:
					sent = true
				default:
					// Channel is full, use blocking send with timeout
					select {
					case output <- batch:
						sent = true
					case <-time.After(cm.config.Timeout):
						retryCount++
						// Small delay before retry
						time.Sleep(cm.config.BackoffDelay)
					}
				}
			}

			if !sent {
				// Final attempt - use blocking send without timeout
				output <- batch
			}
		}
	}()

	return output
}

// RateLimit limits the rate of items processed
func RateLimit[T any](cm *ChannelManager, input <-chan T, rate time.Duration) <-chan T {
	output := make(chan T, cm.config.BufferSize)

	go func() {
		defer close(output)

		ticker := time.NewTicker(rate)
		defer ticker.Stop()

		for item := range input {
			<-ticker.C
			// Use blocking/retry logic to prevent data loss
			sent := false
			retryCount := 0

			for !sent && retryCount < cm.config.MaxRetries {
				// Try non-blocking send first
				select {
				case output <- item:
					sent = true
				default:
					// Channel is full, use blocking send with timeout
					select {
					case output <- item:
						sent = true
					case <-time.After(cm.config.Timeout):
						retryCount++
						// Small delay before retry
						time.Sleep(cm.config.BackoffDelay)
					}
				}
			}

			if !sent {
				// Final attempt - use blocking send without timeout
				output <- item
			}
		}
	}()

	return output
}

// RetryWithBackoff retries failed operations with exponential backoff
func RetryWithBackoff[T any](cm *ChannelManager, input <-chan T, operation func(T) error) <-chan T {
	output := make(chan T, cm.config.BufferSize)

	go func() {
		defer close(output)

		for item := range input {
			success := false
			attempts := 0

			for !success && attempts < cm.config.MaxRetries {
				err := operation(item)
				if err == nil {
					success = true
					// Use blocking/retry logic to prevent data loss
					sent := false
					retryCount := 0

					for !sent && retryCount < cm.config.MaxRetries {
						// Try non-blocking send first
						select {
						case output <- item:
							sent = true
						default:
							// Channel is full, use blocking send with timeout
							select {
							case output <- item:
								sent = true
							case <-time.After(cm.config.Timeout):
								retryCount++
								// Small delay before retry
								time.Sleep(cm.config.BackoffDelay)
							}
						}
					}

					if !sent {
						// Final attempt - use blocking send without timeout
						output <- item
					}
				} else {
					attempts++
					if attempts < cm.config.MaxRetries {
						// Exponential backoff
						delay := cm.config.BackoffDelay * time.Duration(attempts)
						time.Sleep(delay)
					}
				}
			}
		}
	}()

	return output
}

// ============================================================================
// ADVANCED CHANNEL UTILITIES
// ============================================================================

// ChannelUtils provides utility functions for channel operations
type ChannelUtils struct {
	manager *ChannelManager
}

// NewChannelUtils creates new channel utilities
func NewChannelUtils(manager *ChannelManager) *ChannelUtils {
	return &ChannelUtils{
		manager: manager,
	}
}

// Collect collects all items from a channel into a slice
func Collect[T any](cu *ChannelUtils, input <-chan T) []T {
	var result []T
	for item := range input {
		result = append(result, item)
	}
	return result
}

// CollectWithTimeout collects items with a timeout
func CollectWithTimeout[T any](cu *ChannelUtils, input <-chan T, timeout time.Duration) []T {
	var result []T
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case item, ok := <-input:
			if !ok {
				return result
			}
			result = append(result, item)
		case <-timer.C:
			return result
		}
	}
}

// CollectN collects up to n items from a channel
func CollectN[T any](cu *ChannelUtils, input <-chan T, n int) []T {
	result := make([]T, 0, n)
	for i := 0; i < n; i++ {
		select {
		case item, ok := <-input:
			if !ok {
				return result
			}
			result = append(result, item)
		case <-time.After(cu.manager.config.Timeout):
			return result
		}
	}
	return result
}

// Filter filters items based on a predicate
func Filter[T any](cu *ChannelUtils, input <-chan T, predicate func(T) bool) <-chan T {
	output := make(chan T, cu.manager.config.BufferSize)

	go func() {
		defer close(output)

		for item := range input {
			if predicate(item) {
				// Use blocking/retry logic to prevent data loss
				sent := false
				retryCount := 0

				for !sent && retryCount < cu.manager.config.MaxRetries {
					// Try non-blocking send first
					select {
					case output <- item:
						sent = true
					default:
						// Channel is full, use blocking send with timeout
						select {
						case output <- item:
							sent = true
						case <-time.After(cu.manager.config.Timeout):
							retryCount++
							// Small delay before retry
							time.Sleep(cu.manager.config.BackoffDelay)
						}
					}
				}

				if !sent {
					// Final attempt - use blocking send without timeout
					output <- item
				}
			}
		}
	}()

	return output
}

// Map transforms items using a mapping function
func Map[T any, U any](cu *ChannelUtils, input <-chan T, mapper func(T) U) <-chan U {
	output := make(chan U, cu.manager.config.BufferSize)

	go func() {
		defer close(output)

		for item := range input {
			mapped := mapper(item)
			// Use blocking/retry logic to prevent data loss
			sent := false
			retryCount := 0

			for !sent && retryCount < cu.manager.config.MaxRetries {
				// Try non-blocking send first
				select {
				case output <- mapped:
					sent = true
				default:
					// Channel is full, use blocking send with timeout
					select {
					case output <- mapped:
						sent = true
					case <-time.After(cu.manager.config.Timeout):
						retryCount++
						// Small delay before retry
						time.Sleep(cu.manager.config.BackoffDelay)
					}
				}
			}

			if !sent {
				// Final attempt - use blocking send without timeout
				output <- mapped
			}
		}
	}()

	return output
}

// Reduce reduces items using a reducer function
func Reduce[T any](cu *ChannelUtils, input <-chan T, initial T, reducer func(T, T) T) T {
	result := initial
	for item := range input {
		result = reducer(result, item)
	}
	return result
}

// ============================================================================
// CONTEXT-AWARE CHANNEL OPERATIONS
// ============================================================================

// ContextAwareChannel provides context-aware channel operations
type ContextAwareChannel[T any] struct {
	ctx     context.Context
	manager *ChannelManager
}

// NewContextAwareChannel creates a new context-aware channel
func NewContextAwareChannel[T any](ctx context.Context, manager *ChannelManager) *ContextAwareChannel[T] {
	return &ContextAwareChannel[T]{
		ctx:     ctx,
		manager: manager,
	}
}

// ProcessWithContext processes items with context awareness
func (cac *ContextAwareChannel[T]) ProcessWithContext(input <-chan T, processor func(context.Context, T) error) <-chan T {
	output := make(chan T, cac.manager.config.BufferSize)

	go func() {
		defer close(output)

		for {
			select {
			case item, ok := <-input:
				if !ok {
					return
				}

				err := processor(cac.ctx, item)
				if err == nil {
					// Use blocking/retry logic to prevent data loss
					sent := false
					retryCount := 0

					for !sent && retryCount < cac.manager.config.MaxRetries {
						// Try non-blocking send first
						select {
						case output <- item:
							sent = true
						case <-cac.ctx.Done():
							return
						default:
							// Channel is full, use blocking send with timeout
							select {
							case output <- item:
								sent = true
							case <-cac.ctx.Done():
								return
							case <-time.After(cac.manager.config.Timeout):
								retryCount++
								// Small delay before retry
								time.Sleep(cac.manager.config.BackoffDelay)
							}
						}
					}

					if !sent {
						// Final attempt - use blocking send without timeout
						select {
						case output <- item:
							// Item sent successfully
						case <-cac.ctx.Done():
							return
						}
					}
				}
			case <-cac.ctx.Done():
				return
			}
		}
	}()

	return output
}

// FanOutWithContext performs fan-out with context awareness
func (cac *ContextAwareChannel[T]) FanOutWithContext(input <-chan T, outputs int) []<-chan T {
	if outputs <= 0 {
		return nil
	}

	outputChannels := make([]chan T, outputs)
	for i := 0; i < outputs; i++ {
		outputChannels[i] = make(chan T, cac.manager.config.BufferSize)
	}

	go func() {
		defer func() {
			for _, ch := range outputChannels {
				close(ch)
			}
		}()

		index := 0
		for {
			select {
			case item, ok := <-input:
				if !ok {
					return
				}

				// Round-robin distribution with blocking/retry logic
				sent := false
				attempts := 0
				maxAttempts := len(outputChannels) * cac.manager.config.MaxRetries // Try each channel up to MaxRetries times

				for !sent && attempts < maxAttempts {
					ch := outputChannels[index%len(outputChannels)]

					// Try non-blocking send first
					select {
					case ch <- item:
						// Item sent successfully
						sent = true
					case <-cac.ctx.Done():
						return
					default:
						// Channel is full, try next channel
						index++
						attempts++

						// Small delay to prevent busy waiting
						if attempts < maxAttempts {
							time.Sleep(cac.manager.config.BackoffDelay)
						}
					}
				}

				if !sent {
					// If all channels are full, use blocking send with retry logic
					ch := outputChannels[index%len(outputChannels)]
					retryCount := 0
					for !sent && retryCount < cac.manager.config.MaxRetries {
						select {
						case ch <- item:
							sent = true
						case <-cac.ctx.Done():
							return
						case <-time.After(cac.manager.config.Timeout):
							// Timeout reached, try next channel
							index++
							retryCount++
							ch = outputChannels[index%len(outputChannels)]
						}
					}

					if !sent {
						// Final attempt - use blocking send without timeout
						select {
						case ch <- item:
							sent = true
						case <-cac.ctx.Done():
							return
						}
					}
				}
				index++
			case <-cac.ctx.Done():
				return
			}
		}
	}()

	// Convert to read-only channels
	result := make([]<-chan T, outputs)
	for i, ch := range outputChannels {
		result[i] = ch
	}

	return result
}

// ============================================================================
// GLOBAL CHANNEL MANAGER
// ============================================================================

// Global channel manager instance
var GlobalChannelManager = NewChannelManager(DefaultChannelConfig())

// FanOutGlobal performs fan-out using the global channel manager
func FanOutGlobal[T any](input <-chan T, outputs int) []<-chan T {
	return FanOut(GlobalChannelManager, input, outputs)
}

// FanInGlobal performs fan-in using the global channel manager
func FanInGlobal[T any](inputs []<-chan T) <-chan T {
	return FanIn(GlobalChannelManager, inputs)
}

// PipelineGlobal creates a pipeline using the global channel manager
func PipelineGlobal[T any](input <-chan T, stages ...func(T) T) <-chan T {
	return CreatePipeline(GlobalChannelManager, input, stages...)
}

// BatchGlobal processes items in batches using the global channel manager
func BatchGlobal[T any](input <-chan T, batchSize int) <-chan []T {
	return Batch(GlobalChannelManager, input, batchSize)
}

// RateLimitGlobal limits the rate using the global channel manager
func RateLimitGlobal[T any](input <-chan T, rate time.Duration) <-chan T {
	return RateLimit(GlobalChannelManager, input, rate)
}

// RetryWithBackoffGlobal retries with backoff using the global channel manager
func RetryWithBackoffGlobal[T any](input <-chan T, operation func(T) error) <-chan T {
	return RetryWithBackoff(GlobalChannelManager, input, operation)
}
