package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/require"
)

// TestCacheSystemIntegration tests integration between different cache components
func TestCacheSystemIntegration(t *testing.T) {
	// Test local cache integration
	localCache := go_core.NewLocalCache[string]()
	require.NotNil(t, localCache, "Local cache should not be nil")

	// Test context-aware cache integration
	contextManager := go_core.NewContextManager(go_core.DefaultContextConfig())
	contextCache := go_core.NewContextAwareCache(localCache, contextManager)
	require.NotNil(t, contextCache, "Context-aware cache should not be nil")

	// Test basic operations through context-aware cache
	ctx := context.Background()
	key := "integration_key"
	value := "integration_value"

	// Set through context-aware cache
	err := contextCache.Set(ctx, key, &value, 1*time.Hour)
	require.NoError(t, err, "Context-aware Set should not return error")

	// Get through context-aware cache
	retrieved, err := contextCache.Get(ctx, key)
	require.NoError(t, err, "Context-aware Get should not return error")
	require.NotNil(t, retrieved, "Retrieved value should not be nil")
	require.Equal(t, value, *retrieved, "Retrieved value should match original")

	// Verify through underlying cache
	underlyingRetrieved, err := localCache.Get(key)
	require.NoError(t, err, "Underlying cache Get should not return error")
	require.NotNil(t, underlyingRetrieved, "Underlying cache value should not be nil")
	require.Equal(t, value, *underlyingRetrieved, "Underlying cache value should match")
}

// TestCachePerformanceIntegration tests performance integration
func TestCachePerformanceIntegration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Perform operations to generate performance data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("perf_key_%d", i)
		value := fmt.Sprintf("perf_value_%d", i)
		err := cache.Set(key, &value, 1*time.Hour)
		require.NoError(t, err, "Set should not return error")

		_, err = cache.Get(key)
		require.NoError(t, err, "Get should not return error")
	}

	// Get performance stats
	stats := cache.GetPerformanceStats()
	require.NotNil(t, stats, "Performance stats should not be nil")

	// Verify stats contain expected data
	require.Contains(t, stats, "cache", "Stats should contain cache data")
	cacheStats, ok := stats["cache"].(map[string]interface{})
	require.True(t, ok, "Cache stats should be a map")
	require.Contains(t, cacheStats, "operations_count", "Cache stats should contain operations count")
	require.Contains(t, cacheStats, "cache_size", "Cache stats should contain cache size")

	// Get optimization stats
	optStats := cache.GetOptimizationStats()
	require.NotNil(t, optStats, "Optimization stats should not be nil")

	// Verify optimization stats contain expected data
	require.Contains(t, optStats, "atomic_operations", "Optimization stats should contain atomic operations")
	require.Contains(t, optStats, "cache_size", "Optimization stats should contain cache size")
}

// TestCacheContextIntegration tests context integration scenarios
func TestCacheContextIntegration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()
	contextManager := go_core.NewContextManager(go_core.DefaultContextConfig())
	contextCache := go_core.NewContextAwareCache(cache, contextManager)

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	key := "timeout_key"
	value := "timeout_value"

	// Set with timeout context
	err := contextCache.Set(ctx, key, &value, 1*time.Hour)
	require.NoError(t, err, "Set with timeout context should not return error")

	// Get with timeout context
	retrieved, err := contextCache.Get(ctx, key)
	require.NoError(t, err, "Get with timeout context should not return error")
	require.NotNil(t, retrieved, "Retrieved value should not be nil")
	require.Equal(t, value, *retrieved, "Retrieved value should match original")

	// Test with cancelled context
	cancelledCtx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc() // Cancel immediately

	// Operations with cancelled context should fail (context-aware cache respects context cancellation)
	err = contextCache.Set(cancelledCtx, "cancelled_key", &value, 1*time.Hour)
	require.Error(t, err, "Set with cancelled context should return error")
	require.Contains(t, err.Error(), "cancelled", "Error should indicate context was cancelled")

	retrieved, err = contextCache.Get(cancelledCtx, "cancelled_key")
	require.Error(t, err, "Get with cancelled context should return error")
	require.Contains(t, err.Error(), "cancelled", "Error should indicate context was cancelled")
}

// TestCacheBatchIntegration tests batch operation integration
func TestCacheBatchIntegration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test large batch operations
	const batchSize = 1000
	values := make(map[string]*string, batchSize)

	for i := 0; i < batchSize; i++ {
		key := fmt.Sprintf("batch_key_%d", i)
		value := fmt.Sprintf("batch_value_%d", i)
		values[key] = &value
	}

	// Set many values
	err := cache.SetMany(values, 1*time.Hour)
	require.NoError(t, err, "SetMany should not return error")

	// Get many values
	keys := make([]string, 0, batchSize)
	for key := range values {
		keys = append(keys, key)
	}

	results, err := cache.GetMany(keys)
	require.NoError(t, err, "GetMany should not return error")
	require.Len(t, results, batchSize, "Should return all batch results")

	// Verify all values match
	for key, expectedValue := range values {
		actualValue, exists := results[key]
		require.True(t, exists, "Key should exist in results: %s", key)
		require.Equal(t, *expectedValue, *actualValue, "Value should match for key: %s", key)
	}

	// Delete many values
	keysToDelete := make([]string, 0, batchSize/2)
	for i := 0; i < batchSize/2; i++ {
		keysToDelete = append(keysToDelete, fmt.Sprintf("batch_key_%d", i))
	}
	err = cache.DeleteMany(keysToDelete)
	require.NoError(t, err, "DeleteMany should not return error")

	// Verify deletion
	for _, key := range keysToDelete {
		exists, err := cache.Has(key)
		require.NoError(t, err, "Has should not return error")
		require.False(t, exists, "Key should not exist after deletion: %s", key)
	}

	// Verify remaining keys still exist
	for i := batchSize / 2; i < batchSize; i++ {
		key := fmt.Sprintf("batch_key_%d", i)
		exists, err := cache.Has(key)
		require.NoError(t, err, "Has should not return error")
		require.True(t, exists, "Key should still exist: %s", key)
	}
}

// TestCachePatternIntegration tests pattern-based operations integration
func TestCachePatternIntegration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Set up test data with different patterns
	testData := map[string]*string{
		"user:1:profile":  stringPtr("profile1"),
		"user:2:profile":  stringPtr("profile2"),
		"user:3:profile":  stringPtr("profile3"),
		"user:1:settings": stringPtr("settings1"),
		"user:2:settings": stringPtr("settings2"),
		"session:1":       stringPtr("session1"),
		"session:2":       stringPtr("session2"),
		"cache:stats":     stringPtr("stats"),
		"cache:config":    stringPtr("config"),
	}

	err := cache.SetMany(testData, 1*time.Hour)
	require.NoError(t, err, "SetMany should not return error")

	// Test pattern deletion for user profiles only
	err = cache.DeletePattern("user:*:profile")
	require.NoError(t, err, "DeletePattern should not return error")

	// Verify user profiles are deleted
	for key := range testData {
		if key[:4] == "user" && key[len(key)-7:] == ":profile" {
			exists, err := cache.Has(key)
			require.NoError(t, err, "Has should not return error")
			require.False(t, exists, "User profile should be deleted: %s", key)
		}
	}

	// Verify user settings still exist
	for key := range testData {
		if key[:4] == "user" && key[len(key)-9:] == ":settings" {
			exists, err := cache.Has(key)
			require.NoError(t, err, "Has should not return error")
			require.True(t, exists, "User settings should still exist: %s", key)
		}
	}

	// Test pattern deletion for all cache keys
	err = cache.DeletePattern("cache:*")
	require.NoError(t, err, "DeletePattern should not return error")

	// Verify cache keys are deleted
	exists, err := cache.Has("cache:stats")
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "Cache stats should be deleted")

	exists, err = cache.Has("cache:config")
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "Cache config should be deleted")

	// Verify sessions still exist
	exists, err = cache.Has("session:1")
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Session should still exist")
}

// TestCacheExpirationIntegration tests expiration integration scenarios
func TestCacheExpirationIntegration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test mixed expiration times
	testData := map[string]*string{
		"short":  stringPtr("short_value"),
		"medium": stringPtr("medium_value"),
		"long":   stringPtr("long_value"),
	}

	// Set with different TTLs
	err := cache.Set("short", testData["short"], 50*time.Millisecond)
	require.NoError(t, err, "Set short TTL should not return error")

	err = cache.Set("medium", testData["medium"], 100*time.Millisecond)
	require.NoError(t, err, "Set medium TTL should not return error")

	err = cache.Set("long", testData["long"], 1*time.Hour)
	require.NoError(t, err, "Set long TTL should not return error")

	// Verify all exist immediately
	for key := range testData {
		exists, err := cache.Has(key)
		require.NoError(t, err, "Has should not return error")
		require.True(t, exists, "Key should exist immediately: %s", key)
	}

	// Wait for short expiration
	time.Sleep(75 * time.Millisecond)

	// Verify short is expired, others still exist
	exists, err := cache.Has("short")
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "Short TTL key should be expired")

	exists, err = cache.Has("medium")
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Medium TTL key should still exist")

	exists, err = cache.Has("long")
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Long TTL key should still exist")

	// Wait for medium expiration
	time.Sleep(50 * time.Millisecond)

	// Verify medium is expired, long still exists
	exists, err = cache.Has("medium")
	require.NoError(t, err, "Has should not return error")
	require.False(t, exists, "Medium TTL key should be expired")

	exists, err = cache.Has("long")
	require.NoError(t, err, "Has should not return error")
	require.True(t, exists, "Long TTL key should still exist")
}

// TestCacheConcurrencyIntegration tests concurrency integration scenarios
func TestCacheConcurrencyIntegration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test concurrent Set/Get operations
	const numGoroutines = 20
	const numOperations = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations*2) // Set + Get operations

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("concurrent_integration_%d_%d", id, j)
				value := fmt.Sprintf("concurrent_integration_value_%d_%d", id, j)

				// Set operation
				err := cache.Set(key, &value, 1*time.Hour)
				if err != nil {
					errors <- fmt.Errorf("Set error: %w", err)
					continue
				}

				// Get operation
				retrieved, err := cache.Get(key)
				if err != nil {
					errors <- fmt.Errorf("Get error: %w", err)
					continue
				}

				if retrieved == nil {
					errors <- fmt.Errorf("retrieved value is nil for key: %s", key)
					continue
				}

				if *retrieved != value {
					errors <- fmt.Errorf("value mismatch for key %s: expected %s, got %s", key, value, *retrieved)
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

	// Test concurrent batch operations
	var wg2 sync.WaitGroup
	batchErrors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()

			// Create batch data
			batchData := make(map[string]*string, 10)
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("batch_concurrent_%d_%d", id, j)
				value := fmt.Sprintf("batch_concurrent_value_%d_%d", id, j)
				batchData[key] = &value
			}

			// SetMany operation
			err := cache.SetMany(batchData, 1*time.Hour)
			if err != nil {
				batchErrors <- fmt.Errorf("SetMany error: %w", err)
				return
			}

			// GetMany operation
			keys := make([]string, 0, len(batchData))
			for key := range batchData {
				keys = append(keys, key)
			}

			results, err := cache.GetMany(keys)
			if err != nil {
				batchErrors <- fmt.Errorf("GetMany error: %w", err)
				return
			}

			// Verify results
			for key, expectedValue := range batchData {
				actualValue, exists := results[key]
				if !exists {
					batchErrors <- fmt.Errorf("key not found in batch results: %s", key)
					return
				}
				if *actualValue != *expectedValue {
					batchErrors <- fmt.Errorf("value mismatch in batch for key %s: expected %s, got %s", key, *expectedValue, *actualValue)
					return
				}
			}
		}(i)
	}

	wg2.Wait()
	close(batchErrors)

	// Check for batch errors
	for err := range batchErrors {
		require.NoError(t, err, "Concurrent batch operation should not return error")
	}
}

// TestCacheMemoryIntegration tests memory management integration
func TestCacheMemoryIntegration(t *testing.T) {
	cache := go_core.NewLocalCache[string]()

	// Test memory usage with large data
	const numItems = 10000
	largeData := make(map[string]*string, numItems)

	for i := 0; i < numItems; i++ {
		key := fmt.Sprintf("memory_integration_key_%d", i)
		value := fmt.Sprintf("memory_integration_value_%d_with_some_additional_data_to_make_it_larger_%d", i, i)
		largeData[key] = &value
	}

	// Set large batch
	err := cache.SetMany(largeData, 1*time.Hour)
	require.NoError(t, err, "SetMany with large data should not return error")

	// Verify all data is accessible
	keys := make([]string, 0, numItems)
	for key := range largeData {
		keys = append(keys, key)
	}

	results, err := cache.GetMany(keys)
	require.NoError(t, err, "GetMany with large data should not return error")
	require.Len(t, results, numItems, "Should return all large batch results")

	// Verify data integrity
	for key, expectedValue := range largeData {
		actualValue, exists := results[key]
		require.True(t, exists, "Key should exist in large batch results: %s", key)
		require.Equal(t, *expectedValue, *actualValue, "Value should match for large batch key: %s", key)
	}

	// Test memory cleanup
	err = cache.Flush()
	require.NoError(t, err, "Flush should not return error")

	// Verify cache is empty
	emptyResults, err := cache.GetMany(keys[:10]) // Test first 10 keys
	require.NoError(t, err, "GetMany after flush should not return error")
	require.Empty(t, emptyResults, "Cache should be empty after flush")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
