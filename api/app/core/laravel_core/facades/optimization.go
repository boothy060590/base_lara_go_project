package facades_core

import (
	"context"
	"sync"
	"time"

	app_core "base_lara_go_project/app/core/go_core"
)

// ============================================================================
// OPTIMIZATION FACADES
// ============================================================================

// WorkStealingFacade provides Laravel-style access to work stealing pools
type WorkStealingFacade struct {
	pool *app_core.WorkStealingPool[any]
	mu   sync.RWMutex
}

var (
	workStealingInstance *WorkStealingFacade
	workStealingOnce     sync.Once
)

// WorkStealing returns the singleton work stealing facade instance
func WorkStealing() *WorkStealingFacade {
	workStealingOnce.Do(func() {
		workStealingInstance = &WorkStealingFacade{}
	})
	return workStealingInstance
}

// SetPool sets the work stealing pool
func (wsf *WorkStealingFacade) SetPool(pool *app_core.WorkStealingPool[any]) {
	wsf.mu.Lock()
	defer wsf.mu.Unlock()
	wsf.pool = pool
}

// GetPool returns the work stealing pool
func (wsf *WorkStealingFacade) GetPool() *app_core.WorkStealingPool[any] {
	wsf.mu.RLock()
	defer wsf.mu.RUnlock()
	return wsf.pool
}

// Submit submits a work item to the pool
func (wsf *WorkStealingFacade) Submit(id string, data any, handler func(context.Context, any) error, timeout time.Duration) error {
	pool := wsf.GetPool()
	if pool == nil {
		return nil // Return nil if pool not available
	}

	item := app_core.WorkItem[any]{
		ID:      id,
		Data:    data,
		Handler: handler,
		Timeout: timeout,
	}

	return pool.Submit(item)
}

// SubmitAsync submits a work item asynchronously
func (wsf *WorkStealingFacade) SubmitAsync(id string, data any, handler func(context.Context, any) error, timeout time.Duration) <-chan error {
	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		resultChan <- wsf.Submit(id, data, handler, timeout)
	}()

	return resultChan
}

// GetMetrics returns the work stealing pool metrics
func (wsf *WorkStealingFacade) GetMetrics() *app_core.WorkStealingMetrics {
	pool := wsf.GetPool()
	if pool == nil {
		return nil
	}
	return pool.GetMetrics()
}

// Shutdown gracefully shuts down the work stealing pool
func (wsf *WorkStealingFacade) Shutdown() {
	pool := wsf.GetPool()
	if pool != nil {
		pool.Shutdown()
	}
}

// ProfileGuidedFacade provides Laravel-style access to profile-guided optimization
type ProfileGuidedFacade struct {
	optimizer *app_core.ProfileGuidedOptimizer[any]
	mu        sync.RWMutex
}

var (
	profileGuidedInstance *ProfileGuidedFacade
	profileGuidedOnce     sync.Once
)

// ProfileGuided returns the singleton profile-guided facade instance
func ProfileGuided() *ProfileGuidedFacade {
	profileGuidedOnce.Do(func() {
		profileGuidedInstance = &ProfileGuidedFacade{}
	})
	return profileGuidedInstance
}

// SetOptimizer sets the profile-guided optimizer
func (pgf *ProfileGuidedFacade) SetOptimizer(optimizer *app_core.ProfileGuidedOptimizer[any]) {
	pgf.mu.Lock()
	defer pgf.mu.Unlock()
	pgf.optimizer = optimizer
}

// GetOptimizer returns the profile-guided optimizer
func (pgf *ProfileGuidedFacade) GetOptimizer() *app_core.ProfileGuidedOptimizer[any] {
	pgf.mu.RLock()
	defer pgf.mu.RUnlock()
	return pgf.optimizer
}

// GetMetrics returns the profile-guided optimization metrics
func (pgf *ProfileGuidedFacade) GetMetrics() *app_core.ProfileGuidedMetrics {
	optimizer := pgf.GetOptimizer()
	if optimizer == nil {
		return nil
	}
	return optimizer.GetMetrics()
}

// IsEnabled returns whether profile-guided optimization is enabled
func (pgf *ProfileGuidedFacade) IsEnabled() bool {
	optimizer := pgf.GetOptimizer()
	return optimizer != nil
}

// CustomAllocatorFacade provides Laravel-style access to custom allocators
type CustomAllocatorFacade struct {
	allocator *app_core.CustomAllocator[any]
	mu        sync.RWMutex
}

var (
	customAllocatorInstance *CustomAllocatorFacade
	customAllocatorOnce     sync.Once
)

// CustomAllocator returns the singleton custom allocator facade instance
func CustomAllocator() *CustomAllocatorFacade {
	customAllocatorOnce.Do(func() {
		customAllocatorInstance = &CustomAllocatorFacade{}
	})
	return customAllocatorInstance
}

// SetAllocator sets the custom allocator
func (caf *CustomAllocatorFacade) SetAllocator(allocator *app_core.CustomAllocator[any]) {
	caf.mu.Lock()
	defer caf.mu.Unlock()
	caf.allocator = allocator
}

// GetAllocator returns the custom allocator
func (caf *CustomAllocatorFacade) GetAllocator() *app_core.CustomAllocator[any] {
	caf.mu.RLock()
	defer caf.mu.RUnlock()
	return caf.allocator
}

// Allocate allocates memory using the custom allocator
func (caf *CustomAllocatorFacade) Allocate(size int) (any, error) {
	allocator := caf.GetAllocator()
	if allocator == nil {
		return nil, nil // Return nil if allocator not available
	}
	return allocator.Allocate(size)
}

// Deallocate returns memory to the custom allocator
func (caf *CustomAllocatorFacade) Deallocate(obj any, size int) error {
	allocator := caf.GetAllocator()
	if allocator == nil {
		return nil // Return nil if allocator not available
	}
	return allocator.Deallocate(obj, size)
}

// GetMetrics returns the custom allocator metrics
func (caf *CustomAllocatorFacade) GetMetrics() *app_core.CustomAllocatorMetrics {
	allocator := caf.GetAllocator()
	if allocator == nil {
		return nil
	}
	return allocator.GetMetrics()
}

// IsEnabled returns whether custom allocators are enabled
func (caf *CustomAllocatorFacade) IsEnabled() bool {
	allocator := caf.GetAllocator()
	return allocator != nil
}

// OptimizationFacade provides a unified facade for all optimization services
type OptimizationFacade struct {
	mu sync.RWMutex
}

var (
	optimizationInstance *OptimizationFacade
	optimizationOnce     sync.Once
)

// Optimization returns the singleton optimization facade instance
func Optimization() *OptimizationFacade {
	optimizationOnce.Do(func() {
		optimizationInstance = &OptimizationFacade{}
	})
	return optimizationInstance
}

// WorkStealing returns the work stealing facade
func (of *OptimizationFacade) WorkStealing() *WorkStealingFacade {
	return WorkStealing()
}

// ProfileGuided returns the profile-guided facade
func (of *OptimizationFacade) ProfileGuided() *ProfileGuidedFacade {
	return ProfileGuided()
}

// CustomAllocator returns the custom allocator facade
func (of *OptimizationFacade) CustomAllocator() *CustomAllocatorFacade {
	return CustomAllocator()
}

// GetMetrics returns metrics from all optimization services
func (of *OptimizationFacade) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Work stealing metrics
	if workStealingMetrics := WorkStealing().GetMetrics(); workStealingMetrics != nil {
		metrics["work_stealing"] = workStealingMetrics
	}

	// Profile-guided metrics
	if profileGuidedMetrics := ProfileGuided().GetMetrics(); profileGuidedMetrics != nil {
		metrics["profile_guided"] = profileGuidedMetrics
	}

	// Custom allocator metrics
	if customAllocatorMetrics := CustomAllocator().GetMetrics(); customAllocatorMetrics != nil {
		metrics["custom_allocator"] = customAllocatorMetrics
	}

	return metrics
}

// IsEnabled returns whether any optimization service is enabled
func (of *OptimizationFacade) IsEnabled() bool {
	return WorkStealing().GetPool() != nil ||
		ProfileGuided().IsEnabled() ||
		CustomAllocator().IsEnabled()
}

// Shutdown gracefully shuts down all optimization services
func (of *OptimizationFacade) Shutdown() {
	WorkStealing().Shutdown()
	// Profile-guided and custom allocators don't need explicit shutdown
}
