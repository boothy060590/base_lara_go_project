package unit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
)

// TestContextConfig tests context configuration
func TestContextConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := go_core.DefaultContextConfig()
		assert.NotNil(t, config)
		assert.Equal(t, 30*time.Second, config.DefaultTimeout)
		assert.Equal(t, 5*time.Minute, config.MaxTimeout)
		assert.True(t, config.EnableDeadline)
		assert.True(t, config.EnableCancellation)
		assert.True(t, config.PropagateValues)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		config := &go_core.ContextConfig{
			DefaultTimeout:     10 * time.Second,
			MaxTimeout:         2 * time.Minute,
			EnableDeadline:     false,
			EnableCancellation: true,
			PropagateValues:    false,
		}
		assert.Equal(t, 10*time.Second, config.DefaultTimeout)
		assert.Equal(t, 2*time.Minute, config.MaxTimeout)
		assert.False(t, config.EnableDeadline)
		assert.True(t, config.EnableCancellation)
		assert.False(t, config.PropagateValues)
	})

	t.Run("ConfigFromMap", func(t *testing.T) {
		configMap := map[string]interface{}{
			"defaults": map[string]interface{}{
				"timeout":             15,
				"max_timeout":         120,
				"enable_deadline":     false,
				"enable_cancellation": true,
				"propagate_values":    false,
			},
		}

		config := go_core.NewContextConfigFromConfig(configMap)
		assert.Equal(t, 15*time.Second, config.DefaultTimeout)
		assert.Equal(t, 120*time.Second, config.MaxTimeout)
		assert.False(t, config.EnableDeadline)
		assert.True(t, config.EnableCancellation)
		assert.False(t, config.PropagateValues)
	})
}

// TestContextManager tests the context manager
func TestContextManager(t *testing.T) {
	t.Run("NewContextManager", func(t *testing.T) {
		config := go_core.DefaultContextConfig()
		manager := go_core.NewContextManager(config)
		assert.NotNil(t, manager)
		// Test that manager was created successfully
		assert.NotNil(t, manager)
	})

	t.Run("WithTimeout", func(t *testing.T) {
		config := &go_core.ContextConfig{
			DefaultTimeout: 10 * time.Second,
			MaxTimeout:     30 * time.Second,
		}
		manager := go_core.NewContextManager(config)

		// Test normal timeout
		ctx := context.Background()
		timeoutCtx, cancel := manager.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		deadline, ok := timeoutCtx.Deadline()
		assert.True(t, ok)
		assert.True(t, deadline.After(time.Now()))

		// Test timeout exceeding max
		timeoutCtx2, cancel2 := manager.WithTimeout(ctx, 60*time.Second)
		defer cancel2()

		deadline2, ok2 := timeoutCtx2.Deadline()
		assert.True(t, ok2)
		assert.True(t, deadline2.After(time.Now()))
		// Should be capped at MaxTimeout
		assert.True(t, deadline2.Before(time.Now().Add(31*time.Second)))
	})

	t.Run("WithDeadline", func(t *testing.T) {
		config := &go_core.ContextConfig{
			EnableDeadline: true,
		}
		manager := go_core.NewContextManager(config)

		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Second)
		deadlineCtx, cancel := manager.WithDeadline(ctx, deadline)
		defer cancel()

		actualDeadline, ok := deadlineCtx.Deadline()
		assert.True(t, ok)
		assert.Equal(t, deadline.Unix(), actualDeadline.Unix())
	})

	t.Run("WithDeadlineDisabled", func(t *testing.T) {
		config := &go_core.ContextConfig{
			EnableDeadline: false,
		}
		manager := go_core.NewContextManager(config)

		ctx := context.Background()
		deadline := time.Now().Add(5 * time.Second)
		deadlineCtx, cancel := manager.WithDeadline(ctx, deadline)
		defer cancel()

		// Should return a cancellable context, not a deadline context
		_, ok := deadlineCtx.Deadline()
		assert.False(t, ok)
	})

	t.Run("WithValues", func(t *testing.T) {
		config := &go_core.ContextConfig{
			PropagateValues: true,
		}
		manager := go_core.NewContextManager(config)

		ctx := context.Background()
		values := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		valueCtx := manager.WithValues(ctx, values)
		assert.Equal(t, "value1", valueCtx.Value("key1"))
		assert.Equal(t, 123, valueCtx.Value("key2"))
	})

	t.Run("WithValuesDisabled", func(t *testing.T) {
		config := &go_core.ContextConfig{
			PropagateValues: false,
		}
		manager := go_core.NewContextManager(config)

		ctx := context.Background()
		values := map[string]interface{}{
			"key1": "value1",
		}

		valueCtx := manager.WithValues(ctx, values)
		assert.Nil(t, valueCtx.Value("key1"))
	})
}

// TestContextManagerExecution tests execution methods
func TestContextManagerExecution(t *testing.T) {
	t.Run("ExecuteWithTimeout", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Test successful execution
		err := manager.ExecuteWithTimeout(context.Background(), 100*time.Millisecond, func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)

		// Test timeout
		err = manager.ExecuteWithTimeout(context.Background(), 10*time.Millisecond, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
	})

	t.Run("ExecuteWithDeadline", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Test successful execution
		deadline := time.Now().Add(100 * time.Millisecond)
		err := manager.ExecuteWithDeadline(context.Background(), deadline, func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)

		// Test deadline exceeded
		deadline = time.Now().Add(10 * time.Millisecond)
		err = manager.ExecuteWithDeadline(context.Background(), deadline, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deadline exceeded")
	})

	t.Run("ExecuteWithContext", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)

		// Test successful execution
		err := manager.ExecuteWithContext(context.Background(), func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, err)

		// Test context cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err = manager.ExecuteWithContext(ctx, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
	})
}

// TestContextAwareOperation tests context-aware operations
func TestContextAwareOperation(t *testing.T) {
	t.Run("NewContextAwareOperation", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		operation := func(ctx context.Context) (string, error) {
			return "test", nil
		}

		cao := go_core.NewContextAwareOperation(operation, manager)
		assert.NotNil(t, cao)
		// Test that operation was created successfully
		assert.NotNil(t, cao)
	})

	t.Run("WithContext", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		operation := func(ctx context.Context) (string, error) {
			return ctx.Value("test_key").(string), nil
		}

		cao := go_core.NewContextAwareOperation(operation, manager)
		ctx := context.WithValue(context.Background(), "test_key", "test_value")
		cao = cao.WithContext(ctx)

		result, err := cao.Execute()
		assert.NoError(t, err)
		assert.Equal(t, "test_value", result)
	})

	t.Run("WithTimeout", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		operation := func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "test", nil
		}

		cao := go_core.NewContextAwareOperation(operation, manager)
		cao = cao.WithTimeout(10 * time.Millisecond)

		result, err := cao.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
		assert.Equal(t, "", result)
	})

	t.Run("ExecuteWithoutTimeout", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		operation := func(ctx context.Context) (string, error) {
			return "test", nil
		}

		cao := go_core.NewContextAwareOperation(operation, manager)
		result, err := cao.Execute()
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

// TestContextDecorators tests context decorators
func TestContextDecorators(t *testing.T) {
	t.Run("WithContextDecorator", func(t *testing.T) {
		operation := func(ctx context.Context) (string, error) {
			start := ctx.Value("operation_start")
			assert.NotNil(t, start)
			return "test", nil
		}

		decorated := go_core.WithContextDecorator(operation)
		result, err := decorated(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("WithTimeoutDecorator", func(t *testing.T) {
		operation := func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "test", nil
		}

		decorated := go_core.WithTimeoutDecorator[string](10 * time.Millisecond)(operation)
		result, err := decorated(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
		assert.Equal(t, "", result)
	})

	t.Run("WithRetryDecorator", func(t *testing.T) {
		attempts := 0
		operation := func(ctx context.Context) (string, error) {
			attempts++
			if attempts < 3 {
				return "", fmt.Errorf("attempt %d", attempts)
			}
			return "success", nil
		}

		decorated := go_core.WithRetryDecorator[string](3, 10*time.Millisecond)(operation)
		result, err := decorated(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 3, attempts)
	})

	t.Run("WithRetryDecoratorMaxAttempts", func(t *testing.T) {
		operation := func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("always fails")
		}

		decorated := go_core.WithRetryDecorator[string](2, 10*time.Millisecond)(operation)
		result, err := decorated(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed after 2 attempts")
		assert.Equal(t, "", result)
	})

	t.Run("WithRetryDecoratorContextCancellation", func(t *testing.T) {
		operation := func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("always fails")
		}

		decorated := go_core.WithRetryDecorator[string](5, 100*time.Millisecond)(operation)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result, err := decorated(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Equal(t, "", result)
	})
}

// TestContextUtils tests context utilities
func TestContextUtils(t *testing.T) {
	t.Run("NewContextUtils", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		utils := go_core.NewContextUtils(manager)
		assert.NotNil(t, utils)
		// Test that utils was created successfully
		assert.NotNil(t, utils)
	})

	t.Run("MergeContexts", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		utils := go_core.NewContextUtils(manager)

		ctx1 := context.WithValue(context.Background(), "key1", "value1")
		ctx2 := context.WithValue(context.Background(), "key2", "value2")

		merged := utils.MergeContexts(ctx1, ctx2)
		assert.NotNil(t, merged)
		assert.Equal(t, "value1", merged.Value("key1"))
	})

	t.Run("MergeContextsEmpty", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		utils := go_core.NewContextUtils(manager)

		merged := utils.MergeContexts()
		assert.NotNil(t, merged)
		assert.Equal(t, context.Background(), merged)
	})

	t.Run("IsContextExpired", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		utils := go_core.NewContextUtils(manager)

		// Test active context
		ctx := context.Background()
		assert.False(t, utils.IsContextExpired(ctx))

		// Test cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.True(t, utils.IsContextExpired(ctx))

		// Test timed out context
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond)
		assert.True(t, utils.IsContextExpired(ctx))
	})

	t.Run("GetContextTimeout", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		utils := go_core.NewContextUtils(manager)

		// Test context without timeout
		ctx := context.Background()
		timeout, ok := utils.GetContextTimeout(ctx)
		assert.False(t, ok)
		assert.Equal(t, time.Duration(0), timeout)

		// Test context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		timeout, ok = utils.GetContextTimeout(ctx)
		assert.True(t, ok)
		assert.True(t, timeout > 0)
		assert.True(t, timeout <= 5*time.Second)
	})
}

// TestGlobalContextManager tests global context manager
func TestGlobalContextManager(t *testing.T) {
	t.Run("WithGlobalTimeout", func(t *testing.T) {
		ctx := context.Background()
		timeoutCtx, cancel := go_core.WithGlobalTimeout(ctx, 5*time.Second)
		defer cancel()

		deadline, ok := timeoutCtx.Deadline()
		assert.True(t, ok)
		assert.True(t, deadline.After(time.Now()))
	})

	t.Run("ExecuteWithGlobalTimeout", func(t *testing.T) {
		// Test successful execution
		err := go_core.ExecuteWithGlobalTimeout(context.Background(), 100*time.Millisecond, func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)

		// Test timeout
		err = go_core.ExecuteWithGlobalTimeout(context.Background(), 10*time.Millisecond, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
	})
}

// TestContextAwareInterfaces tests context-aware interfaces
func TestContextAwareInterfaces(t *testing.T) {
	t.Run("ContextAwareOperation", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		operation := func(ctx context.Context) (string, error) {
			return "test", nil
		}

		cao := go_core.NewContextAwareOperation(operation, manager)

		// Test WithContext chaining
		ctx := context.WithValue(context.Background(), "test", "value")
		cao = cao.WithContext(ctx)

		// Test WithTimeout chaining
		cao = cao.WithTimeout(5 * time.Second)

		// Test Execute
		result, err := cao.Execute()
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

// TestContextConcurrency tests context handling under concurrency
func TestContextConcurrency(t *testing.T) {
	t.Run("ConcurrentContextOperations", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		var wg sync.WaitGroup
		results := make([]string, 10)
		errors := make([]error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				operation := func(ctx context.Context) (string, error) {
					time.Sleep(10 * time.Millisecond)
					return fmt.Sprintf("result-%d", index), nil
				}

				cao := go_core.NewContextAwareOperation(operation, manager)
				cao = cao.WithTimeout(100 * time.Millisecond)

				result, err := cao.Execute()
				results[index] = result
				errors[index] = err
			}(i)
		}

		wg.Wait()

		// All operations should succeed
		for i := 0; i < 10; i++ {
			assert.NoError(t, errors[i])
			assert.Equal(t, fmt.Sprintf("result-%d", i), results[i])
		}
	})

	t.Run("ConcurrentContextCancellation", func(t *testing.T) {
		manager := go_core.NewContextManager(nil)
		var wg sync.WaitGroup
		errors := make([]error, 5)

		ctx, cancel := context.WithCancel(context.Background())

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				operation := func(ctx context.Context) (string, error) {
					// Check for context cancellation in the operation
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(100 * time.Millisecond):
						return "result", nil
					}
				}

				cao := go_core.NewContextAwareOperation(operation, manager)
				cao = cao.WithContext(ctx)

				_, err := cao.Execute()
				errors[index] = err
			}(i)
		}

		// Cancel context after a short delay
		time.Sleep(10 * time.Millisecond)
		cancel()

		wg.Wait()

		// Check that at least some operations were cancelled
		cancelledCount := 0
		for i := 0; i < 5; i++ {
			if errors[i] != nil && (errors[i] == context.Canceled || strings.Contains(errors[i].Error(), "cancelled")) {
				cancelledCount++
			}
		}
		assert.Greater(t, cancelledCount, 0, "At least one operation should be cancelled")
	})
}
