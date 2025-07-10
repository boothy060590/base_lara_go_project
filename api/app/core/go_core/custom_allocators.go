package go_core

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// CUSTOM ALLOCATORS FOR MEMORY OPTIMIZATION
// ============================================================================

// CustomAllocatorConfig defines configuration for custom allocators
type CustomAllocatorConfig struct {
	Enabled            bool          `json:"enabled"`
	PoolSize           int           `json:"pool_size"`
	MaxObjectSize      int           `json:"max_object_size"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	EnableMetrics      bool          `json:"enable_metrics"`
	EnableProfiling    bool          `json:"enable_profiling"`
	AllocationStrategy string        `json:"allocation_strategy"`
}

// DefaultCustomAllocatorConfig returns sensible defaults for custom allocators
func DefaultCustomAllocatorConfig() *CustomAllocatorConfig {
	return &CustomAllocatorConfig{
		Enabled:            true,
		PoolSize:           1000,
		MaxObjectSize:      1024 * 1024, // 1MB
		CleanupInterval:    5 * time.Minute,
		EnableMetrics:      true,
		EnableProfiling:    true,
		AllocationStrategy: "pool", // pool, slab, or custom
	}
}

// CustomAllocator implements memory optimization for specific workloads
type CustomAllocator[T any] struct {
	config  *CustomAllocatorConfig
	pools   map[int]*CustomObjectPool[T]
	slabs   map[int]*SlabAllocator[T]
	metrics *CustomAllocatorMetrics
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
}

// CustomObjectPool represents a pool of reusable objects
type CustomObjectPool[T any] struct {
	objects   chan T
	size      int
	allocated int64
	returned  int64
	mu        sync.RWMutex
}

// SlabAllocator represents a slab-based memory allocator
type SlabAllocator[T any] struct {
	slabs      []*Slab[T]
	slabSize   int
	objectSize int
	allocated  int64
	returned   int64
	mu         sync.RWMutex
}

// Slab represents a memory slab
type Slab[T any] struct {
	data      []T
	used      []bool
	freeCount int
	mu        sync.RWMutex
}

// CustomAllocatorMetrics tracks allocation metrics
type CustomAllocatorMetrics struct {
	TotalAllocations   int64         `json:"total_allocations"`
	TotalDeallocations int64         `json:"total_deallocations"`
	PoolHits           int64         `json:"pool_hits"`
	PoolMisses         int64         `json:"pool_misses"`
	AverageAllocTime   time.Duration `json:"average_alloc_time"`
	MemoryEfficiency   float64       `json:"memory_efficiency"`
	LastUpdated        time.Time     `json:"last_updated"`
	mu                 sync.RWMutex
}

// NewCustomAllocator creates a new custom allocator
func NewCustomAllocator[T any](config *CustomAllocatorConfig) *CustomAllocator[T] {
	if config == nil {
		config = DefaultCustomAllocatorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ca := &CustomAllocator[T]{
		config:  config,
		pools:   make(map[int]*CustomObjectPool[T]),
		slabs:   make(map[int]*SlabAllocator[T]),
		metrics: NewCustomAllocatorMetrics(),
		ctx:     ctx,
		cancel:  cancel,
	}

	if config.Enabled {
		go ca.startCleanup()
	}

	return ca
}

// Allocate allocates memory using the custom allocator
func (ca *CustomAllocator[T]) Allocate(size int) (T, error) {
	start := time.Now()
	defer func() {
		ca.metrics.UpdateAllocTime(time.Since(start))
	}()

	switch ca.config.AllocationStrategy {
	case "pool":
		return ca.allocateFromPool(size)
	case "slab":
		return ca.allocateFromSlab(size)
	default:
		return ca.allocateDefault(size)
	}
}

// Deallocate returns memory to the custom allocator
func (ca *CustomAllocator[T]) Deallocate(obj T, size int) error {
	switch ca.config.AllocationStrategy {
	case "pool":
		return ca.deallocateToPool(obj, size)
	case "slab":
		return ca.deallocateToSlab(obj, size)
	default:
		return ca.deallocateDefault(obj, size)
	}
}

// allocateFromPool allocates from object pool
func (ca *CustomAllocator[T]) allocateFromPool(size int) (T, error) {
	ca.mu.RLock()
	pool, exists := ca.pools[size]
	ca.mu.RUnlock()

	if !exists {
		ca.mu.Lock()
		pool = NewCustomObjectPool[T](ca.config.PoolSize)
		ca.pools[size] = pool
		ca.mu.Unlock()
	}

	obj, err := pool.Get()
	if err != nil {
		ca.metrics.RecordPoolMiss()
		// Fall back to default allocation
		return ca.allocateDefault(size)
	}

	ca.metrics.RecordPoolHit()
	return obj, nil
}

// allocateFromSlab allocates from slab allocator
func (ca *CustomAllocator[T]) allocateFromSlab(size int) (T, error) {
	ca.mu.RLock()
	slab, exists := ca.slabs[size]
	ca.mu.RUnlock()

	if !exists {
		ca.mu.Lock()
		slab = NewSlabAllocator[T](size, ca.config.PoolSize)
		ca.slabs[size] = slab
		ca.mu.Unlock()
	}

	obj, err := slab.Allocate()
	if err != nil {
		// Fall back to default allocation
		return ca.allocateDefault(size)
	}

	return obj, nil
}

// allocateDefault performs default allocation
func (ca *CustomAllocator[T]) allocateDefault(size int) (T, error) {
	var obj T
	// Default allocation - this would be replaced with actual allocation logic
	ca.metrics.RecordAllocation()
	return obj, nil
}

// deallocateToPool returns object to pool
func (ca *CustomAllocator[T]) deallocateToPool(obj T, size int) error {
	ca.mu.RLock()
	pool, exists := ca.pools[size]
	ca.mu.RUnlock()

	if !exists {
		return fmt.Errorf("pool not found for size %d", size)
	}

	return pool.Put(obj)
}

// deallocateToSlab returns object to slab
func (ca *CustomAllocator[T]) deallocateToSlab(obj T, size int) error {
	ca.mu.RLock()
	slab, exists := ca.slabs[size]
	ca.mu.RUnlock()

	if !exists {
		return fmt.Errorf("slab not found for size %d", size)
	}

	return slab.Deallocate(obj)
}

// deallocateDefault performs default deallocation
func (ca *CustomAllocator[T]) deallocateDefault(obj T, size int) error {
	ca.metrics.RecordDeallocation()
	return nil
}

// startCleanup begins the cleanup loop
func (ca *CustomAllocator[T]) startCleanup() {
	ticker := time.NewTicker(ca.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ca.ctx.Done():
			return
		case <-ticker.C:
			ca.cleanup()
		}
	}
}

// cleanup performs periodic cleanup of unused objects
func (ca *CustomAllocator[T]) cleanup() {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	// Clean up pools
	for size, pool := range ca.pools {
		if pool.GetUsage() < 0.1 { // Less than 10% usage
			delete(ca.pools, size)
		}
	}

	// Clean up slabs
	for size, slab := range ca.slabs {
		if slab.GetUsage() < 0.1 { // Less than 10% usage
			delete(ca.slabs, size)
		}
	}
}

// GetMetrics returns the current metrics
func (ca *CustomAllocator[T]) GetMetrics() *CustomAllocatorMetrics {
	return ca.metrics
}

// NewCustomObjectPool creates a new object pool
func NewCustomObjectPool[T any](size int) *CustomObjectPool[T] {
	return &CustomObjectPool[T]{
		objects: make(chan T, size),
		size:    size,
	}
}

// Get retrieves an object from the pool
func (op *CustomObjectPool[T]) Get() (T, error) {
	select {
	case obj := <-op.objects:
		atomic.AddInt64(&op.allocated, 1)
		return obj, nil
	default:
		var zero T
		return zero, fmt.Errorf("pool is empty")
	}
}

// Put returns an object to the pool
func (op *CustomObjectPool[T]) Put(obj T) error {
	select {
	case op.objects <- obj:
		atomic.AddInt64(&op.returned, 1)
		return nil
	default:
		return fmt.Errorf("pool is full")
	}
}

// GetUsage returns the usage percentage of the pool
func (op *CustomObjectPool[T]) GetUsage() float64 {
	op.mu.RLock()
	defer op.mu.RUnlock()

	allocated := atomic.LoadInt64(&op.allocated)
	returned := atomic.LoadInt64(&op.returned)

	if allocated == 0 {
		return 0
	}

	return float64(returned) / float64(allocated)
}

// NewSlabAllocator creates a new slab allocator
func NewSlabAllocator[T any](objectSize, slabSize int) *SlabAllocator[T] {
	sa := &SlabAllocator[T]{
		slabSize:   slabSize,
		objectSize: objectSize,
		slabs:      make([]*Slab[T], 0),
	}

	// Create initial slab
	sa.addSlab()

	return sa
}

// addSlab adds a new slab to the allocator
func (sa *SlabAllocator[T]) addSlab() {
	slab := &Slab[T]{
		data:      make([]T, sa.slabSize),
		used:      make([]bool, sa.slabSize),
		freeCount: sa.slabSize,
	}

	sa.mu.Lock()
	sa.slabs = append(sa.slabs, slab)
	sa.mu.Unlock()
}

// Allocate allocates an object from a slab
func (sa *SlabAllocator[T]) Allocate() (T, error) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	// Try to find a free slot in existing slabs
	for _, slab := range sa.slabs {
		if obj, err := slab.Allocate(); err == nil {
			atomic.AddInt64(&sa.allocated, 1)
			return obj, nil
		}
	}

	// No free slots, need to add a new slab
	sa.mu.RUnlock()
	sa.mu.Lock()
	sa.addSlab()
	sa.mu.Unlock()
	sa.mu.RLock()

	// Try again with the new slab
	lastSlab := sa.slabs[len(sa.slabs)-1]
	obj, err := lastSlab.Allocate()
	if err != nil {
		var zero T
		return zero, err
	}

	atomic.AddInt64(&sa.allocated, 1)
	return obj, nil
}

// Deallocate returns an object to a slab
func (sa *SlabAllocator[T]) Deallocate(obj T) error {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	// Find the slab containing this object
	for _, slab := range sa.slabs {
		if err := slab.Deallocate(obj); err == nil {
			atomic.AddInt64(&sa.returned, 1)
			return nil
		}
	}

	return fmt.Errorf("object not found in any slab")
}

// GetUsage returns the usage percentage of the slab allocator
func (sa *SlabAllocator[T]) GetUsage() float64 {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	allocated := atomic.LoadInt64(&sa.allocated)
	returned := atomic.LoadInt64(&sa.returned)

	if allocated == 0 {
		return 0
	}

	return float64(returned) / float64(allocated)
}

// Allocate allocates an object from a slab
func (s *Slab[T]) Allocate() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.freeCount == 0 {
		var zero T
		return zero, fmt.Errorf("slab is full")
	}

	// Find first free slot
	for i, used := range s.used {
		if !used {
			s.used[i] = true
			s.freeCount--
			return s.data[i], nil
		}
	}

	var zero T
	return zero, fmt.Errorf("no free slots available")
}

// Deallocate returns an object to a slab
func (s *Slab[T]) Deallocate(obj T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the object in the slab
	for i, data := range s.data {
		if &data == &obj {
			if s.used[i] {
				s.used[i] = false
				s.freeCount++
				return nil
			}
		}
	}

	return fmt.Errorf("object not found in slab")
}

// NewCustomAllocatorMetrics creates new custom allocator metrics
func NewCustomAllocatorMetrics() *CustomAllocatorMetrics {
	return &CustomAllocatorMetrics{
		LastUpdated: time.Now(),
	}
}

// RecordAllocation records an allocation
func (cam *CustomAllocatorMetrics) RecordAllocation() {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	cam.TotalAllocations++
	cam.LastUpdated = time.Now()
}

// RecordDeallocation records a deallocation
func (cam *CustomAllocatorMetrics) RecordDeallocation() {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	cam.TotalDeallocations++
	cam.LastUpdated = time.Now()
}

// RecordPoolHit records a pool hit
func (cam *CustomAllocatorMetrics) RecordPoolHit() {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	cam.PoolHits++
	cam.LastUpdated = time.Now()
}

// RecordPoolMiss records a pool miss
func (cam *CustomAllocatorMetrics) RecordPoolMiss() {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	cam.PoolMisses++
	cam.LastUpdated = time.Now()
}

// UpdateAllocTime updates the average allocation time
func (cam *CustomAllocatorMetrics) UpdateAllocTime(duration time.Duration) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	if cam.TotalAllocations > 0 {
		total := cam.AverageAllocTime * time.Duration(cam.TotalAllocations-1)
		cam.AverageAllocTime = (total + duration) / time.Duration(cam.TotalAllocations)
	} else {
		cam.AverageAllocTime = duration
	}
}
