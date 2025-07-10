package facades_core

import (
	"context"
	"sync"
	"time"

	app_core "base_lara_go_project/app/core/go_core"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// CONTEXT FACADE
// ============================================================================

// ContextFacade provides Laravel-style access to context management features
type ContextFacade struct {
	manager *app_core.ContextManager
	mu      sync.RWMutex
}

var (
	contextInstance *ContextFacade
	contextOnce     sync.Once
)

// Context returns the singleton context facade instance
func Context() *ContextFacade {
	contextOnce.Do(func() {
		contextInstance = &ContextFacade{
			manager: app_core.NewContextManager(app_core.DefaultContextConfig()),
		}
	})
	return contextInstance
}

// SetManager sets a custom context manager
func (cf *ContextFacade) SetManager(manager *app_core.ContextManager) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.manager = manager
}

// GetManager returns the current context manager
func (cf *ContextFacade) GetManager() *app_core.ContextManager {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.manager
}

// WithTimeout creates a context with automatic timeout
func (cf *ContextFacade) WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return cf.manager.WithTimeout(ctx, timeout)
}

// WithDeadline creates a context with automatic deadline
func (cf *ContextFacade) WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return cf.manager.WithDeadline(ctx, deadline)
}

// WithValues creates a context with propagated values
func (cf *ContextFacade) WithValues(ctx context.Context, values map[string]interface{}) context.Context {
	return cf.manager.WithValues(ctx, values)
}

// ExecuteWithTimeout executes a function with automatic timeout
func (cf *ContextFacade) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	return cf.manager.ExecuteWithTimeout(ctx, timeout, fn)
}

// ExecuteWithDeadline executes a function with automatic deadline
func (cf *ContextFacade) ExecuteWithDeadline(ctx context.Context, deadline time.Time, fn func(context.Context) error) error {
	return cf.manager.ExecuteWithDeadline(ctx, deadline, fn)
}

// NewOperation creates a new context-aware operation
func (cf *ContextFacade) NewOperation(operation func(context.Context) (interface{}, error)) *app_core.ContextAwareOperation[interface{}] {
	return app_core.NewContextAwareOperation(operation, cf.manager)
}

// WithDecorator decorates an operation with context awareness
func (cf *ContextFacade) WithDecorator(operation func(context.Context) (interface{}, error)) func(context.Context) (interface{}, error) {
	return app_core.WithContextDecorator(operation)
}

// WithTimeoutDecorator decorates an operation with timeout
func (cf *ContextFacade) WithTimeoutDecorator(timeout time.Duration) func(func(context.Context) (interface{}, error)) func(context.Context) (interface{}, error) {
	return app_core.WithTimeoutDecorator[interface{}](timeout)
}

// WithRetryDecorator decorates an operation with retry logic
func (cf *ContextFacade) WithRetryDecorator(maxAttempts int, delay time.Duration) func(func(context.Context) (interface{}, error)) func(context.Context) (interface{}, error) {
	return app_core.WithRetryDecorator[interface{}](maxAttempts, delay)
}

// ============================================================================
// CONTEXT-AWARE CONTROLLER
// ============================================================================

// ContextAwareController provides automatic context optimization for controllers
type ContextAwareController struct {
	contextManager *app_core.ContextManager
}

// NewContextAwareController creates a new context-aware controller
func NewContextAwareController() *ContextAwareController {
	return &ContextAwareController{
		contextManager: app_core.NewContextManager(app_core.DefaultContextConfig()),
	}
}

// ExecuteWithContext executes a controller action with automatic context optimization
func (cac *ContextAwareController) ExecuteWithContext(ctx *gin.Context, action func(*gin.Context) error) error {
	// Convert gin context to standard context
	goCtx := ctx.Request.Context()

	// Add request-specific values
	goCtx = context.WithValue(goCtx, "request_id", ctx.GetString("request_id"))
	goCtx = context.WithValue(goCtx, "user_id", ctx.GetString("user_id"))

	// Execute with automatic timeout
	return cac.contextManager.ExecuteWithTimeout(goCtx, 30*time.Second, func(goCtx context.Context) error {
		// Create a new gin context with the updated context
		ctx.Request = ctx.Request.WithContext(goCtx)
		return action(ctx)
	})
}

// ============================================================================
// CONTEXT-AWARE SERVICE
// ============================================================================

// ContextAwareService provides automatic context optimization for services
type ContextAwareService struct {
	contextManager *app_core.ContextManager
}

// NewContextAwareService creates a new context-aware service
func NewContextAwareService() *ContextAwareService {
	return &ContextAwareService{
		contextManager: app_core.NewContextManager(app_core.DefaultContextConfig()),
	}
}

// ExecuteWithContext executes a service method with automatic context optimization
func (cas *ContextAwareService) ExecuteWithContext(ctx context.Context, action func(context.Context) error) error {
	return cas.contextManager.ExecuteWithTimeout(ctx, 30*time.Second, action)
}

// ExecuteWithTimeout executes a service method with custom timeout
func (cas *ContextAwareService) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, action func(context.Context) error) error {
	return cas.contextManager.ExecuteWithTimeout(ctx, timeout, action)
}

// ============================================================================
// CONTEXT UTILITIES
// ============================================================================

// ContextUtils provides utility functions for context management
type ContextUtils struct {
	utils *app_core.ContextUtils
}

// NewContextUtils creates new context utilities
func NewContextUtils() *ContextUtils {
	manager := app_core.NewContextManager(app_core.DefaultContextConfig())
	return &ContextUtils{
		utils: app_core.NewContextUtils(manager),
	}
}

// MergeContexts merges multiple contexts into one
func (cu *ContextUtils) MergeContexts(ctxs ...context.Context) context.Context {
	return cu.utils.MergeContexts(ctxs...)
}

// IsContextExpired checks if a context has expired
func (cu *ContextUtils) IsContextExpired(ctx context.Context) bool {
	return cu.utils.IsContextExpired(ctx)
}

// GetContextTimeout returns the timeout from a context
func (cu *ContextUtils) GetContextTimeout(ctx context.Context) (time.Duration, bool) {
	return cu.utils.GetContextTimeout(ctx)
}

// ============================================================================
// GLOBAL CONTEXT FUNCTIONS (Laravel-style static access)
// ============================================================================

// ExecuteWithGlobalTimeout executes a function with global timeout settings
func ExecuteWithGlobalTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	return app_core.ExecuteWithGlobalTimeout(ctx, timeout, fn)
}

// NewContextAwareOperation creates a new context-aware operation with global settings
func NewContextAwareOperation(operation func(context.Context) (interface{}, error)) *app_core.ContextAwareOperation[interface{}] {
	return app_core.NewContextAwareOperation(operation, app_core.GlobalContextManager)
}
