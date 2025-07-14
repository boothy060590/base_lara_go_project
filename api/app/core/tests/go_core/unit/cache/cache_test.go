package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheInterface tests the basic cache interface functionality
func TestCacheInterface(t *testing.T) {
	// Test that we can create a local cache
	cache := go_core.NewLocalCache[string]()
	require.NotNil(t, cache, "Cache should not be nil")

	// Test basic operations
	key := "test_key"
	value := "test_value"

	// Test Set
	err := cache.Set(key, &value, 1*time.Hour)
	require.NoError(t, err, "Set should not return error")

	// Test Get
	retrieved, err := cache.Get(key)
	require.NoError(t, err, "Get should not return error")
	require.NotNil(t, retrieved, "Retrieved value should not be nil")
	require.Equal(t, value, *retrieved, "Retrieved value should match original")

	// Test Has
	exists, err := cache.Has(key)
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Key should exist")

	// Test Delete
	err = cache.Delete(key)
	require.NoError(t, err, "Delete should not return error")

	// Test that key no longer exists
	exists, err = cache.Has(key)
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "Key should not exist after deletion")
}

// TestCacheExpiration tests cache expiration functionality
func TestCacheExpiration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	key := "expire_key"
	value := "expire_value"

	// Set with short TTL
	err := cache.Set(key, &value, 10*time.Millisecond)
	require.NoError(t, err, "Set should not return error")

	// Should exist immediately
	exists, err := cache.Has(key)
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Key should exist immediately")

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should not exist after expiration
	exists, err = cache.Has(key)
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "Key should not exist after expiration")

	// Get should return nil
	retrieved, err := cache.Get(key)
	require.NoError(t, err, "Get should not return error")
	require.Nil(t, retrieved, "Retrieved value should be nil after expiration")
}

// TestCacheGetOrSet tests the GetOrSet functionality
func TestCacheGetOrSet(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	key := "get_or_set_key"
	value := "get_or_set_value"

	// Test GetOrSet when key doesn't exist
	factoryCalled := false
	factory := func() (*string, error) {
		factoryCalled = true
		return &value, nil
	}

	result, err := cache.GetOrSet(key, factory, 1*time.Hour)
	require.NoError(t, err, "GetOrSet should not return error")
	require.NotNil(t, result, "Result should not be nil")
	require.Equal(t, value, *result, "Result should match factory value")
	require.True(t, factoryCalled, "Factory should be called")

	// Test GetOrSet when key exists
	factoryCalled = false
	result2, err := cache.GetOrSet(key, factory, 1*time.Hour)
	require.NoError(t, err, "GetOrSet should not return error")
	require.NotNil(t, result2, "Result should not be nil")
	require.Equal(t, value, *result2, "Result should match cached value")
	require.False(t, factoryCalled, "Factory should not be called again")
}

// TestCacheBatchOperations tests batch operations
func TestCacheBatchOperations(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test SetMany
	values := map[string]*string{
		"key1": stringPtr("value1"),
		"key2": stringPtr("value2"),
		"key3": stringPtr("value3"),
	}

	err := cache.SetMany(values, 1*time.Hour)
	require.NoError(t, err, "SetMany should not return error")

	// Test GetMany
	keys := []string{"key1", "key2", "key3", "nonexistent"}
	results, err := cache.GetMany(keys)
	require.NoError(t, err, "GetMany should not return error")
	require.Len(t, results, 3, "Should return 3 results")
	require.Equal(t, "value1", *results["key1"], "key1 should have correct value")
	require.Equal(t, "value2", *results["key2"], "key2 should have correct value")
	require.Equal(t, "value3", *results["key3"], "key3 should have correct value")
	require.Nil(t, results["nonexistent"], "nonexistent key should be nil")

	// Test DeleteMany
	err = cache.DeleteMany([]string{"key1", "key2"})
	require.NoError(t, err, "DeleteMany should not return error")

	// Verify deletion
	exists, err := cache.Has("key1")
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "key1 should not exist after deletion")

	exists, err = cache.Has("key2")
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "key2 should not exist after deletion")

	exists, err = cache.Has("key3")
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "key3 should still exist")
}

// TestCacheDeletePattern tests pattern-based deletion
func TestCacheDeletePattern(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Set up test data
	testData := map[string]*string{
		"user:1:profile": stringPtr("profile1"),
		"user:2:profile": stringPtr("profile2"),
		"user:3:profile": stringPtr("profile3"),
		"session:1":      stringPtr("session1"),
		"session:2":      stringPtr("session2"),
	}

	err := cache.SetMany(testData, 1*time.Hour)
	require.NoError(t, err, "SetMany should not return error")

	// Delete all user profiles
	err = cache.DeletePattern("user:*:profile")
	require.NoError(t, err, "DeletePattern should not return error")

	// Verify user profiles are deleted
	for key := range testData {
		if key[:4] == "user" {
			exists, err := cache.Has(key)
			require.NoError(t, err, "Has should not return error")
			require.False(t, exists, "User profile should be deleted: %s", key)
		}
	}

	// Verify sessions still exist
	exists, err := cache.Has("session:1")
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Session should still exist")

	exists, err = cache.Has("session:2")
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Session should still exist")
}

// TestCacheFlush tests the flush functionality
func TestCacheFlush(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Add some data
	values := map[string]*string{
		"key1": stringPtr("value1"),
		"key2": stringPtr("value2"),
		"key3": stringPtr("value3"),
	}

	err := cache.SetMany(values, 1*time.Hour)
	require.NoError(t, err, "SetMany should not return error")

	// Verify data exists
	for key := range values {
		exists, err := cache.Has(key)
		require.NoError(t, err, "Has should not return error")
		require.True(t, exists, "Key should exist before flush: %s", key)
	}

	// Flush cache
	err = cache.Flush()
	require.NoError(t, err, "Flush should not return error")

	// Verify all data is gone
	for key := range values {
		exists, err := cache.Has(key)
		require.NoError(t, err, "Has should not return error")
		require.False(t, exists, "Key should not exist after flush: %s", key)
	}
}

// TestCacheWithContext tests context-aware operations
func TestCacheWithContext(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test WithContext
	ctx := context.Background()
	contextCache := cache.WithContext(ctx)
	require.NotNil(t, contextCache, "Context cache should not be nil")

	// Test basic operations with context
	key := "context_key"
	value := "context_value"

	err := contextCache.Set(key, &value, 1*time.Hour)
	require.NoError(t, err, "Set should not return error")

	retrieved, err := contextCache.Get(key)
	require.NoError(t, err, "Get should not return error")
	require.NotNil(t, retrieved, "Retrieved value should not be nil")
	require.Equal(t, value, *retrieved, "Retrieved value should match original")
}

// TestCachePerformanceStats tests performance statistics
func TestCachePerformanceStats(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Perform some operations
	key := "stats_key"
	value := "stats_value"

	err := cache.Set(key, &value, 1*time.Hour)
	require.NoError(t, err, "Set should not return error")

	_, err = cache.Get(key)
	require.NoError(t, err, "Get should not return error")

	_, err = cache.Has(key)
	require.NoError(t, err, "Has should not return error")

	// Get performance stats
	stats := cache.GetPerformanceStats()
	require.NotNil(t, stats, "Performance stats should not be nil")
	require.IsType(t, map[string]interface{}{}, stats, "Stats should be a map")

	// Get optimization stats
	optStats := cache.GetOptimizationStats()
	require.NotNil(t, optStats, "Optimization stats should not be nil")
	require.IsType(t, map[string]interface{}{}, optStats, "Optimization stats should be a map")
}

// TestCacheConcurrency tests concurrent access to cache
func TestCacheConcurrency(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test concurrent Set operations
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
				value := fmt.Sprintf("concurrent_value_%d_%d", id, j)

				err := cache.Set(key, &value, 1*time.Hour)
				if err != nil {
					errors <- err
				}

				// Also test Get
				retrieved, err := cache.Get(key)
				if err != nil {
					errors <- err
				} else if retrieved == nil {
					errors <- fmt.Errorf("retrieved value is nil for key: %s", key)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		require.NoError(t, err, "Concurrent operation should not return error")
	}
}

// TestCacheEdgeCases tests edge cases and error conditions
func TestCacheEdgeCases(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test Get with non-existent key
	result, err := cache.Get("nonexistent")
	require.NoError(t, err, "Get should not return error for non-existent key")
	require.Nil(t, result, "Result should be nil for non-existent key")

	// Test GetMany with empty keys
	results, err := cache.GetMany([]string{})
	require.NoError(t, err, "GetMany should not return error for empty keys")
	require.Empty(t, results, "Results should be empty for empty keys")

	// Test SetMany with empty values
	err = cache.SetMany(map[string]*string{}, 1*time.Hour)
	require.NoError(t, err, "SetMany should not return error for empty values")

	// Test DeleteMany with empty keys
	err = cache.DeleteMany([]string{})
	require.NoError(t, err, "DeleteMany should not return error for empty keys")

	// Test DeletePattern with non-existent pattern
	err = cache.DeletePattern("nonexistent:*")
	require.NoError(t, err, "DeletePattern should not return error for non-existent pattern")
}

// TestCacheNumericOperations tests numeric operations (Increment/Decrement)
func TestCacheNumericOperations(t *testing.T) {
	cache := go_core.NewLocalCache[int64]()

	key := "numeric_key"

	// Test Increment
	result, err := cache.Increment(key, 5)
	require.NoError(t, err, "Increment should not return error")
	require.Equal(t, int64(5), result, "Increment result should be 5")

	// Test Increment again
	result, err = cache.Increment(key, 3)
	require.NoError(t, err, "Increment should not return error")
	require.Equal(t, int64(8), result, "Increment result should be 8")

	// Test Decrement
	result, err = cache.Decrement(key, 2)
	require.NoError(t, err, "Decrement should not return error")
	require.Equal(t, int64(6), result, "Decrement result should be 6")

	// Test Get to verify final value
	retrieved, err := cache.Get(key)
	require.NoError(t, err, "Get should not return error")
	require.NotNil(t, retrieved, "Retrieved value should not be nil")
	require.Equal(t, int64(6), *retrieved, "Final value should be 6")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// TestCacheTypeSafety tests type safety with different types
func TestCacheTypeSafety(t *testing.T) {
	// Test with string type
	stringCache := go_core.NewLocalCache[string]()
	require.NotNil(t, stringCache, "String cache should not be nil")

	// Test with int type
	intCache := go_core.NewLocalCache[int]()
	require.NotNil(t, intCache, "Int cache should not be nil")

	// Test with struct type
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	structCache := go_core.NewLocalCache[TestStruct]()
	require.NotNil(t, structCache, "Struct cache should not be nil")

	// Test operations with different types
	testStruct := TestStruct{ID: 1, Name: "test"}

	err := structCache.Set("struct_key", &testStruct, 1*time.Hour)
	require.NoError(t, err, "Set should not return error for struct")

	retrieved, err := structCache.Get("struct_key")
	require.NoError(t, err, "Get should not return error for struct")
	require.NotNil(t, retrieved, "Retrieved struct should not be nil")
	require.Equal(t, testStruct.ID, retrieved.ID, "Struct ID should match")
	require.Equal(t, testStruct.Name, retrieved.Name, "Struct Name should match")
}

// TestCacheMemoryLeaks tests for potential memory leaks
func TestCacheMemoryLeaks(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Add many items
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("memory_key_%d", i)
		value := fmt.Sprintf("memory_value_%d", i)
		err := cache.Set(key, &value, 1*time.Hour)
		require.NoError(t, err, "Set should not return error")
	}

	// Delete all items
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("memory_key_%d", i)
		err := cache.Delete(key)
		require.NoError(t, err, "Delete should not return error")
	}

	// Flush to ensure cleanup
	err := cache.Flush()
	require.NoError(t, err, "Flush should not return error")

	// Verify cache is empty
	results, err := cache.GetMany([]string{"memory_key_0", "memory_key_1", "memory_key_2"})
	require.NoError(t, err, "GetMany should not return error")
	require.Empty(t, results, "Cache should be empty after deletion and flush")
}

// TestCacheContextCancellation tests context cancellation for all operations
func TestCacheContextCancellation(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test GetWithContext cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := cache.GetWithContext(ctx, "test")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test SetWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = cache.SetWithContext(ctx, "test", stringPtr("value"), time.Minute)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test DeleteWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = cache.DeleteWithContext(ctx, "test")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test HasWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = cache.HasWithContext(ctx, "test")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test GetOrSetWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = cache.GetOrSetWithContext(ctx, "test", func() (*string, error) {
		return stringPtr("value"), nil
	}, time.Minute)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test IncrementWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = cache.IncrementWithContext(ctx, "test", 1)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test DecrementWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = cache.DecrementWithContext(ctx, "test", 1)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test GetManyWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = cache.GetManyWithContext(ctx, []string{"test1", "test2"})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test SetManyWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = cache.SetManyWithContext(ctx, map[string]*string{"test": stringPtr("value")}, time.Minute)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test DeleteManyWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = cache.DeleteManyWithContext(ctx, []string{"test1", "test2"})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test DeletePatternWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = cache.DeletePatternWithContext(ctx, "test*")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test FlushWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = cache.FlushWithContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestCacheContextTimeout tests context timeout for all operations
func TestCacheContextTimeout(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test GetWithContext timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout

	_, err := cache.GetWithContext(ctx, "test")
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)

	// Test SetWithContext timeout
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout

	err = cache.SetWithContext(ctx, "test", stringPtr("value"), time.Minute)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

// TestCacheContextRaceConditions tests race conditions with context-aware operations
func TestCacheContextRaceConditions(t *testing.T) {
	cache := go_core.NewLocalCache[string]()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent GetWithContext operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			key := fmt.Sprintf("key-%d", id)
			_, _ = cache.GetWithContext(ctx, key)
		}(i)
	}
	wg.Wait()

	// Test concurrent SetWithContext operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			key := fmt.Sprintf("key-%d", id)
			_ = cache.SetWithContext(ctx, key, stringPtr(fmt.Sprintf("value-%d", id)), time.Minute)
		}(i)
	}
	wg.Wait()

	// Test concurrent GetWithContext and SetWithContext operations
	wg.Add(numGoroutines * 2)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			key := fmt.Sprintf("race-key-%d", id)
			_, _ = cache.GetWithContext(ctx, key)
		}(i)
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			key := fmt.Sprintf("race-key-%d", id)
			_ = cache.SetWithContext(ctx, key, stringPtr(fmt.Sprintf("race-value-%d", id)), time.Minute)
		}(i)
	}
	wg.Wait()

	// Verify no data races occurred
	ctx := context.Background()
	for i := 0; i < numGoroutines; i++ {
		key := fmt.Sprintf("race-key-%d", i)
		value, err := cache.GetWithContext(ctx, key)
		assert.NoError(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, fmt.Sprintf("race-value-%d", i), *value)
	}
}

// TestCacheContextMixedOperations tests mixing context-aware and non-context operations
func TestCacheContextMixedOperations(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Set using non-context method
	err := cache.Set("test", stringPtr("value"), time.Minute)
	assert.NoError(t, err)

	// Get using context method
	ctx := context.Background()
	value, err := cache.GetWithContext(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "value", *value)

	// Set using context method
	err = cache.SetWithContext(ctx, "test2", stringPtr("value2"), time.Minute)
	assert.NoError(t, err)

	// Get using non-context method
	value, err = cache.Get("test2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", *value)
}

// TestCacheContextBatchOperations tests context-aware batch operations
func TestCacheContextBatchOperations(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test GetManyWithContext
	ctx := context.Background()
	keys := []string{"batch1", "batch2", "batch3"}
	values := map[string]*string{
		"batch1": stringPtr("value1"),
		"batch2": stringPtr("value2"),
		"batch3": stringPtr("value3"),
	}

	// Set values
	err := cache.SetManyWithContext(ctx, values, time.Minute)
	assert.NoError(t, err)

	// Get values
	result, err := cache.GetManyWithContext(ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "value1", *result["batch1"])
	assert.Equal(t, "value2", *result["batch2"])
	assert.Equal(t, "value3", *result["batch3"])

	// Test DeleteManyWithContext
	err = cache.DeleteManyWithContext(ctx, keys)
	assert.NoError(t, err)

	// Verify deletion
	result, err = cache.GetManyWithContext(ctx, keys)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestCacheContextNumericOperations tests context-aware numeric operations
func TestCacheContextNumericOperations(t *testing.T) {
	cache := go_core.NewLocalCache[int64]()

	ctx := context.Background()

	// Test IncrementWithContext
	value, err := cache.IncrementWithContext(ctx, "counter", 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), value)

	value, err = cache.IncrementWithContext(ctx, "counter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), value)

	// Test DecrementWithContext
	value, err = cache.DecrementWithContext(ctx, "counter", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(6), value)

	// Verify final value
	valuePtr, err := cache.GetWithContext(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(6), *valuePtr)
}

// TestCacheContextPatternOperations tests context-aware pattern operations
func TestCacheContextPatternOperations(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	ctx := context.Background()

	// Set values with different patterns
	values := map[string]*string{
		"user:1": stringPtr("user1"),
		"user:2": stringPtr("user2"),
		"post:1": stringPtr("post1"),
		"post:2": stringPtr("post2"),
		"other":  stringPtr("other"),
	}

	err := cache.SetManyWithContext(ctx, values, time.Minute)
	assert.NoError(t, err)

	// Delete pattern "user:*"
	err = cache.DeletePatternWithContext(ctx, "user:*")
	assert.NoError(t, err)

	// Verify only user keys were deleted
	result, err := cache.GetManyWithContext(ctx, []string{"user:1", "user:2", "post:1", "post:2", "other"})
	assert.NoError(t, err)
	assert.Len(t, result, 3) // post:1, post:2, other should remain
	assert.NotNil(t, result["post:1"])
	assert.NotNil(t, result["post:2"])
	assert.NotNil(t, result["other"])
	assert.Nil(t, result["user:1"])
	assert.Nil(t, result["user:2"])
}

// TestCacheContextFlush tests context-aware flush operation
func TestCacheContextFlush(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	ctx := context.Background()

	// Set some values
	values := map[string]*string{
		"key1": stringPtr("value1"),
		"key2": stringPtr("value2"),
		"key3": stringPtr("value3"),
	}

	err := cache.SetManyWithContext(ctx, values, time.Minute)
	assert.NoError(t, err)

	// Verify values exist
	result, err := cache.GetManyWithContext(ctx, []string{"key1", "key2", "key3"})
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Flush with context
	err = cache.FlushWithContext(ctx)
	assert.NoError(t, err)

	// Verify all values are gone
	result, err = cache.GetManyWithContext(ctx, []string{"key1", "key2", "key3"})
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestCacheContextWithCancellationDuringOperation tests context cancellation during long operations
func TestCacheContextWithCancellationDuringOperation(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Set a value
	err := cache.Set("test", stringPtr("value"), time.Minute)
	assert.NoError(t, err)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	// Try to get the value - should fail due to cancellation
	_, err = cache.GetWithContext(ctx, "test")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Try to set a value - should fail due to cancellation
	err = cache.SetWithContext(ctx, "test2", stringPtr("value2"), time.Minute)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
