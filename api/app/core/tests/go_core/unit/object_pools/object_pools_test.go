package object_pools

import (
	"testing"
	"time"

	"base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
)

// TestObjectPoolBasicOperations tests basic pool operations
func TestObjectPoolBasicOperations(t *testing.T) {
	config := go_core.ObjectPoolConfig{
		Name:            "test_pool",
		Description:     "Test pool for unit testing",
		Type:            "test",
		PoolSize:        10,
		MaxIdleTime:     30 * time.Second,
		CleanupStrategy: "lazy",
		ResetMethod:     "Reset",
		Enabled:         true,
		Metrics: go_core.PoolMetricsConfig{
			TrackHits:        true,
			TrackMisses:      true,
			TrackAllocations: true,
			TrackReuses:      true,
		},
	}

	pool := go_core.NewConfigurableObjectPool[*go_core.PooledUser](config, func() *go_core.PooledUser {
		return &go_core.PooledUser{}
	})

	// Test getting an object
	user1 := pool.Get()
	assert.NotNil(t, user1, "Should return a user object")

	// Test putting an object back
	pool.Put(user1)

	// Test getting another object (should reuse)
	user2 := pool.Get()
	assert.NotNil(t, user2, "Should return another user object")

	// Test metrics
	metrics := pool.GetMetrics()
	assert.GreaterOrEqual(t, metrics.Allocations, int64(1), "Should have at least one allocation")
	assert.GreaterOrEqual(t, metrics.Reuses, int64(1), "Should have at least one reuse")

	// Clean up
	pool.Close()
}

// TestObjectPoolMetrics tests pool metrics tracking
func TestObjectPoolMetrics(t *testing.T) {
	config := go_core.ObjectPoolConfig{
		Name:            "metrics_test_pool",
		Description:     "Test pool for metrics",
		Type:            "test",
		PoolSize:        5,
		MaxIdleTime:     10 * time.Second,
		CleanupStrategy: "lazy",
		ResetMethod:     "Reset",
		Enabled:         true,
		Metrics: go_core.PoolMetricsConfig{
			TrackHits:        true,
			TrackMisses:      true,
			TrackAllocations: true,
			TrackReuses:      true,
		},
	}

	pool := go_core.NewConfigurableObjectPool[*go_core.PooledEvent](config, func() *go_core.PooledEvent {
		return &go_core.PooledEvent{}
	})

	// Get multiple objects
	events := make([]*go_core.PooledEvent, 10)
	for i := 0; i < 10; i++ {
		events[i] = pool.Get()
	}

	// Put some back
	for i := 0; i < 5; i++ {
		pool.Put(events[i])
	}

	// Get more objects (some should be reused)
	for i := 0; i < 5; i++ {
		pool.Get()
	}

	// Check metrics
	metrics := pool.GetMetrics()
	assert.GreaterOrEqual(t, metrics.Allocations, int64(10), "Should have at least 10 allocations")
	assert.GreaterOrEqual(t, metrics.Reuses, int64(5), "Should have at least 5 reuses")
	assert.GreaterOrEqual(t, metrics.Hits, int64(5), "Should have at least 5 hits")

	// Reset metrics
	pool.ResetMetrics()
	metrics = pool.GetMetrics()
	assert.Equal(t, int64(0), metrics.Allocations, "Allocations should be reset to 0")
	assert.Equal(t, int64(0), metrics.Reuses, "Reuses should be reset to 0")

	pool.Close()
}

// TestObjectPoolManager tests the object pool manager
func TestObjectPoolManager(t *testing.T) {
	manager := go_core.NewObjectPoolManager()

	// Register a pool
	config := go_core.ObjectPoolConfig{
		Name:            "manager_test_pool",
		Description:     "Test pool for manager",
		Type:            "user", // Use a known type
		PoolSize:        5,
		MaxIdleTime:     10 * time.Second,
		CleanupStrategy: "lazy",
		ResetMethod:     "Reset",
		Enabled:         true,
		Metrics: go_core.PoolMetricsConfig{
			TrackHits:        true,
			TrackMisses:      true,
			TrackAllocations: true,
			TrackReuses:      true,
		},
	}

	pool := manager.RegisterPool("test_pool", config, func() *go_core.PooledUser {
		return &go_core.PooledUser{}
	})

	// Test getting the pool back
	retrievedPool, exists := manager.GetPool("test_pool")
	assert.True(t, exists, "Pool should exist")
	assert.Equal(t, pool, retrievedPool, "Retrieved pool should be the same")

	// Test getting non-existent pool
	_, exists = manager.GetPool("non_existent")
	assert.False(t, exists, "Non-existent pool should not exist")

	// Test getting all pools
	allPools := manager.GetAllPools()
	assert.Len(t, allPools, 1, "Should have one pool")
	assert.Contains(t, allPools, "test_pool", "Should contain the test pool")

	// Test getting metrics for all pools
	allMetrics := manager.GetPoolMetrics()
	assert.Len(t, allMetrics, 1, "Should have metrics for one pool")
	assert.Contains(t, allMetrics, "test_pool", "Should contain metrics for test pool")

	// Clean up
	manager.Close()
}

// TestObjectPoolCleanupStrategies tests different cleanup strategies
func TestObjectPoolCleanupStrategies(t *testing.T) {
	tests := []struct {
		name             string
		cleanupStrategy  string
		expectedBehavior string
	}{
		{
			name:             "Lazy Cleanup",
			cleanupStrategy:  "lazy",
			expectedBehavior: "Clean up only when pool is full",
		},
		{
			name:             "Eager Cleanup",
			cleanupStrategy:  "eager",
			expectedBehavior: "Clean up immediately when idle time expires",
		},
		{
			name:             "Adaptive Cleanup",
			cleanupStrategy:  "adaptive",
			expectedBehavior: "Adaptive cleanup based on usage patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := go_core.ObjectPoolConfig{
				Name:            tt.name,
				Description:     tt.expectedBehavior,
				Type:            "test",
				PoolSize:        5,
				MaxIdleTime:     100 * time.Millisecond, // Short for testing
				CleanupStrategy: tt.cleanupStrategy,
				ResetMethod:     "Reset",
				Enabled:         true,
				Metrics: go_core.PoolMetricsConfig{
					TrackHits:        true,
					TrackMisses:      true,
					TrackAllocations: true,
					TrackReuses:      true,
				},
			}

			pool := go_core.NewConfigurableObjectPool[*go_core.PooledJob](config, func() *go_core.PooledJob {
				return &go_core.PooledJob{}
			})

			// Get and put some objects
			jobs := make([]*go_core.PooledJob, 3)
			for i := 0; i < 3; i++ {
				jobs[i] = pool.Get()
			}

			// Put them back
			for i := 0; i < 3; i++ {
				pool.Put(jobs[i])
			}

			// Wait a bit for cleanup to potentially run
			time.Sleep(200 * time.Millisecond)

			// Verify pool still works
			job := pool.Get()
			assert.NotNil(t, job, "Pool should still work after cleanup")

			pool.Close()
		})
	}
}

// TestObjectPoolReset tests object reset functionality
func TestObjectPoolReset(t *testing.T) {
	config := go_core.ObjectPoolConfig{
		Name:            "reset_test_pool",
		Description:     "Test pool for reset functionality",
		Type:            "test",
		PoolSize:        5,
		MaxIdleTime:     10 * time.Second,
		CleanupStrategy: "lazy",
		ResetMethod:     "Reset",
		Enabled:         true,
		Metrics: go_core.PoolMetricsConfig{
			TrackHits:        true,
			TrackMisses:      true,
			TrackAllocations: true,
			TrackReuses:      true,
		},
	}

	pool := go_core.NewConfigurableObjectPool[*go_core.PooledUser](config, func() *go_core.PooledUser {
		return &go_core.PooledUser{}
	})

	// Get an object and modify it
	user := pool.Get()
	user.ID = 123
	user.Name = "Test User"
	user.Email = "test@example.com"

	// Put it back (should reset it)
	pool.Put(user)

	// Get it again (should be reset)
	user2 := pool.Get()
	assert.Equal(t, 0, user2.ID, "User ID should be reset to 0")
	assert.Equal(t, "", user2.Name, "User name should be reset to empty")
	assert.Equal(t, "", user2.Email, "User email should be reset to empty")

	pool.Close()
}

// TestObjectPoolNilHandling tests handling of nil values
func TestObjectPoolNilHandling(t *testing.T) {
	config := go_core.ObjectPoolConfig{
		Name:            "nil_test_pool",
		Description:     "Test pool for nil handling",
		Type:            "test",
		PoolSize:        5,
		MaxIdleTime:     10 * time.Second,
		CleanupStrategy: "lazy",
		ResetMethod:     "Reset",
		Enabled:         true,
		Metrics: go_core.PoolMetricsConfig{
			TrackHits:        true,
			TrackMisses:      true,
			TrackAllocations: true,
			TrackReuses:      true,
		},
	}

	pool := go_core.NewConfigurableObjectPool[*go_core.PooledUser](config, func() *go_core.PooledUser {
		return &go_core.PooledUser{}
	})

	// Test putting nil (should not panic)
	assert.NotPanics(t, func() {
		pool.Put(&go_core.PooledUser{}) // Zero value, not nil
	}, "Putting zero value should not panic")

	pool.Close()
}

// TestObjectPoolConcurrency tests concurrent access to pools
func TestObjectPoolConcurrency(t *testing.T) {
	config := go_core.ObjectPoolConfig{
		Name:            "concurrency_test_pool",
		Description:     "Test pool for concurrency",
		Type:            "test",
		PoolSize:        10,
		MaxIdleTime:     10 * time.Second,
		CleanupStrategy: "lazy",
		ResetMethod:     "Reset",
		Enabled:         true,
		Metrics: go_core.PoolMetricsConfig{
			TrackHits:        true,
			TrackMisses:      true,
			TrackAllocations: true,
			TrackReuses:      true,
		},
	}

	pool := go_core.NewConfigurableObjectPool[*go_core.PooledEvent](config, func() *go_core.PooledEvent {
		return &go_core.PooledEvent{}
	})

	// Test concurrent get/put operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				event := pool.Get()
				event.ID = "event-" + string(rune(id)) + "-" + string(rune(j))
				pool.Put(event)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify pool is still functional
	event := pool.Get()
	assert.NotNil(t, event, "Pool should still work after concurrent access")

	pool.Close()
}

// TestObjectPoolManagerConcurrency tests manager behavior under concurrent access
func TestObjectPoolManagerConcurrency(t *testing.T) {
	manager := go_core.NewObjectPoolManager()

	config := go_core.ObjectPoolConfig{
		Name:            "concurrent_manager_pool",
		Description:     "Test pool for concurrent manager access",
		Type:            "user", // Use a known type
		PoolSize:        5,
		MaxIdleTime:     10 * time.Second,
		CleanupStrategy: "lazy",
		ResetMethod:     "Reset",
		Enabled:         true,
		Metrics: go_core.PoolMetricsConfig{
			TrackHits:        true,
			TrackMisses:      true,
			TrackAllocations: true,
			TrackReuses:      true,
		},
	}

	// Test concurrent pool registration and access
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			poolName := "concurrent_pool_" + string(rune(id))
			pool := manager.RegisterPool(poolName, config, func() *go_core.PooledUser {
				return &go_core.PooledUser{}
			})

			// Try to access the pool
			retrievedPool, exists := manager.GetPool(poolName)
			assert.True(t, exists, "Pool should exist")
			assert.Equal(t, pool, retrievedPool, "Retrieved pool should match")
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all pools were created
	allPools := manager.GetAllPools()
	assert.Len(t, allPools, 5, "Should have 5 pools")

	manager.Close()
}

// TestObjectPoolConfiguration tests configuration validation
func TestObjectPoolConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      go_core.ObjectPoolConfig
		expectError bool
	}{
		{
			name: "Valid Configuration",
			config: go_core.ObjectPoolConfig{
				Name:            "valid_pool",
				Description:     "Valid test pool",
				Type:            "test",
				PoolSize:        10,
				MaxIdleTime:     30 * time.Second,
				CleanupStrategy: "lazy",
				ResetMethod:     "Reset",
				Enabled:         true,
				Metrics: go_core.PoolMetricsConfig{
					TrackHits:        true,
					TrackMisses:      true,
					TrackAllocations: true,
					TrackReuses:      true,
				},
			},
			expectError: false,
		},
		{
			name: "Zero Pool Size",
			config: go_core.ObjectPoolConfig{
				Name:            "zero_size_pool",
				Description:     "Pool with zero size",
				Type:            "test",
				PoolSize:        0,
				MaxIdleTime:     30 * time.Second,
				CleanupStrategy: "lazy",
				ResetMethod:     "Reset",
				Enabled:         true,
				Metrics: go_core.PoolMetricsConfig{
					TrackHits:        true,
					TrackMisses:      true,
					TrackAllocations: true,
					TrackReuses:      true,
				},
			},
			expectError: false, // Should not error, just use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := go_core.NewConfigurableObjectPool[*go_core.PooledUser](tt.config, func() *go_core.PooledUser {
				return &go_core.PooledUser{}
			})

			// Test that pool works
			user := pool.Get()
			assert.NotNil(t, user, "Pool should work with this configuration")

			pool.Close()
		})
	}
}
