package go_core

import (
	"sync"
)

// ResponsePoolManager manages pools for different response types
type ResponsePoolManager struct {
	loginPool      *sync.Pool
	healthPool     *sync.Pool
	errorPool      *sync.Pool
	validationPool *sync.Pool
	userPool       *sync.Pool
	genericPool    *sync.Pool
	metrics        *ResponsePoolMetrics
}

// ResponsePoolMetrics tracks pool usage metrics
type ResponsePoolMetrics struct {
	LoginPoolHits      int64
	HealthPoolHits     int64
	ErrorPoolHits      int64
	ValidationPoolHits int64
	UserPoolHits       int64
	GenericPoolHits    int64
	TotalAllocations   int64
	TotalReuses        int64
	mu                 sync.RWMutex
}

// NewResponsePoolManager creates a new response pool manager
func NewResponsePoolManager() *ResponsePoolManager {
	rpm := &ResponsePoolManager{
		metrics: &ResponsePoolMetrics{},
	}

	// Initialize login response pool
	rpm.loginPool = &sync.Pool{
		New: func() interface{} {
			rpm.metrics.mu.Lock()
			rpm.metrics.TotalAllocations++
			rpm.metrics.mu.Unlock()
			return &LoginResponse{}
		},
	}

	// Initialize health response pool
	rpm.healthPool = &sync.Pool{
		New: func() interface{} {
			rpm.metrics.mu.Lock()
			rpm.metrics.TotalAllocations++
			rpm.metrics.mu.Unlock()
			return &HealthResponse{}
		},
	}

	// Initialize error response pool
	rpm.errorPool = &sync.Pool{
		New: func() interface{} {
			rpm.metrics.mu.Lock()
			rpm.metrics.TotalAllocations++
			rpm.metrics.mu.Unlock()
			return &ErrorResponse{}
		},
	}

	// Initialize validation response pool
	rpm.validationPool = &sync.Pool{
		New: func() interface{} {
			rpm.metrics.mu.Lock()
			rpm.metrics.TotalAllocations++
			rpm.metrics.mu.Unlock()
			return &ValidationResponse{}
		},
	}

	// Initialize user response pool
	rpm.userPool = &sync.Pool{
		New: func() interface{} {
			rpm.metrics.mu.Lock()
			rpm.metrics.TotalAllocations++
			rpm.metrics.mu.Unlock()
			return &UserResponse{}
		},
	}

	// Initialize generic response pool
	rpm.genericPool = &sync.Pool{
		New: func() interface{} {
			rpm.metrics.mu.Lock()
			rpm.metrics.TotalAllocations++
			rpm.metrics.mu.Unlock()
			return &GenericResponse{}
		},
	}

	return rpm
}

// GetLoginResponse gets a pooled login response
func (rpm *ResponsePoolManager) GetLoginResponse() *LoginResponse {
	rpm.metrics.mu.Lock()
	rpm.metrics.LoginPoolHits++
	rpm.metrics.TotalReuses++
	rpm.metrics.mu.Unlock()
	
	resp := rpm.loginPool.Get().(*LoginResponse)
	resp.Reset()
	return resp
}

// PutLoginResponse returns a login response to the pool
func (rpm *ResponsePoolManager) PutLoginResponse(resp *LoginResponse) {
	rpm.loginPool.Put(resp)
}

// GetHealthResponse gets a pooled health response
func (rpm *ResponsePoolManager) GetHealthResponse() *HealthResponse {
	rpm.metrics.mu.Lock()
	rpm.metrics.HealthPoolHits++
	rpm.metrics.TotalReuses++
	rpm.metrics.mu.Unlock()
	
	resp := rpm.healthPool.Get().(*HealthResponse)
	resp.Reset()
	return resp
}

// PutHealthResponse returns a health response to the pool
func (rpm *ResponsePoolManager) PutHealthResponse(resp *HealthResponse) {
	rpm.healthPool.Put(resp)
}

// GetErrorResponse gets a pooled error response
func (rpm *ResponsePoolManager) GetErrorResponse() *ErrorResponse {
	rpm.metrics.mu.Lock()
	rpm.metrics.ErrorPoolHits++
	rpm.metrics.TotalReuses++
	rpm.metrics.mu.Unlock()
	
	resp := rpm.errorPool.Get().(*ErrorResponse)
	resp.Reset()
	return resp
}

// PutErrorResponse returns an error response to the pool
func (rpm *ResponsePoolManager) PutErrorResponse(resp *ErrorResponse) {
	rpm.errorPool.Put(resp)
}

// GetValidationResponse gets a pooled validation response
func (rpm *ResponsePoolManager) GetValidationResponse() *ValidationResponse {
	rpm.metrics.mu.Lock()
	rpm.metrics.ValidationPoolHits++
	rpm.metrics.TotalReuses++
	rpm.metrics.mu.Unlock()
	
	resp := rpm.validationPool.Get().(*ValidationResponse)
	resp.Reset()
	return resp
}

// PutValidationResponse returns a validation response to the pool
func (rpm *ResponsePoolManager) PutValidationResponse(resp *ValidationResponse) {
	rpm.validationPool.Put(resp)
}

// GetUserResponse gets a pooled user response
func (rpm *ResponsePoolManager) GetUserResponse() *UserResponse {
	rpm.metrics.mu.Lock()
	rpm.metrics.UserPoolHits++
	rpm.metrics.TotalReuses++
	rpm.metrics.mu.Unlock()
	
	resp := rpm.userPool.Get().(*UserResponse)
	resp.Reset()
	return resp
}

// PutUserResponse returns a user response to the pool
func (rpm *ResponsePoolManager) PutUserResponse(resp *UserResponse) {
	rpm.userPool.Put(resp)
}

// GetGenericResponse gets a pooled generic response
func (rpm *ResponsePoolManager) GetGenericResponse() *GenericResponse {
	rpm.metrics.mu.Lock()
	rpm.metrics.GenericPoolHits++
	rpm.metrics.TotalReuses++
	rpm.metrics.mu.Unlock()
	
	resp := rpm.genericPool.Get().(*GenericResponse)
	resp.Reset()
	return resp
}

// PutGenericResponse returns a generic response to the pool
func (rpm *ResponsePoolManager) PutGenericResponse(resp *GenericResponse) {
	rpm.genericPool.Put(resp)
}

// GetMetrics returns current pool metrics
func (rpm *ResponsePoolManager) GetMetrics() ResponsePoolMetrics {
	rpm.metrics.mu.RLock()
	defer rpm.metrics.mu.RUnlock()
	// Return a copy to avoid lock copying
	return ResponsePoolMetrics{
		LoginPoolHits:      rpm.metrics.LoginPoolHits,
		HealthPoolHits:     rpm.metrics.HealthPoolHits,
		ErrorPoolHits:      rpm.metrics.ErrorPoolHits,
		ValidationPoolHits: rpm.metrics.ValidationPoolHits,
		UserPoolHits:       rpm.metrics.UserPoolHits,
		GenericPoolHits:    rpm.metrics.GenericPoolHits,
		TotalAllocations:   rpm.metrics.TotalAllocations,
		TotalReuses:        rpm.metrics.TotalReuses,
	}
}

// ResetMetrics resets all metrics counters
func (rpm *ResponsePoolManager) ResetMetrics() {
	rpm.metrics.mu.Lock()
	defer rpm.metrics.mu.Unlock()
	rpm.metrics = &ResponsePoolMetrics{}
}

// Specialized response types

// LoginResponse represents a login response (extends base LoginResponse)
type LoginResponse struct {
	Message string      `json:"message"`
	User    interface{} `json:"user"`
	Token   string      `json:"token"`
}

// Reset resets the login response for reuse
func (r *LoginResponse) Reset() {
	r.Message = ""
	r.User = nil
	r.Token = ""
}

// HealthResponse represents a health check response (extends base HealthResponse)
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Service   string `json:"service"`
	Server    string `json:"server"`
}

// Reset resets the health response for reuse
func (r *HealthResponse) Reset() {
	r.Status = ""
	r.Timestamp = 0
	r.Service = ""
	r.Server = ""
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Reset resets the error response for reuse
func (r *ErrorResponse) Reset() {
	r.Success = false
	r.Message = ""
	r.Errors = nil
}

// ValidationResponse represents a validation error response
type ValidationResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors"`
}

// Reset resets the validation response for reuse
func (r *ValidationResponse) Reset() {
	r.Message = ""
	if r.Errors == nil {
		r.Errors = make(map[string]interface{})
	} else {
		// Clear the map without reallocating
		for k := range r.Errors {
			delete(r.Errors, k)
		}
	}
}

// UserResponse represents a user-related response
type UserResponse struct {
	Message string      `json:"message"`
	User    interface{} `json:"user"`
}

// Reset resets the user response for reuse
func (r *UserResponse) Reset() {
	r.Message = ""
	r.User = nil
}

// GenericResponse represents a generic API response
type GenericResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Reset resets the generic response for reuse
func (r *GenericResponse) Reset() {
	r.Success = false
	r.Message = ""
	r.Data = nil
	r.Errors = nil
}

// Global response pool manager
var GlobalResponsePoolManager *ResponsePoolManager

// InitializeGlobalResponsePoolManager initializes the global response pool manager
func InitializeGlobalResponsePoolManager() {
	GlobalResponsePoolManager = NewResponsePoolManager()
}

// Helper functions for global pool manager

// GetPooledLoginResponse gets a pooled login response from global manager
func GetPooledLoginResponse() *LoginResponse {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	return GlobalResponsePoolManager.GetLoginResponse()
}

// PutPooledLoginResponse returns a login response to global pool
func PutPooledLoginResponse(resp *LoginResponse) {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	GlobalResponsePoolManager.PutLoginResponse(resp)
}

// GetPooledHealthResponse gets a pooled health response from global manager
func GetPooledHealthResponse() *HealthResponse {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	return GlobalResponsePoolManager.GetHealthResponse()
}

// PutPooledHealthResponse returns a health response to global pool
func PutPooledHealthResponse(resp *HealthResponse) {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	GlobalResponsePoolManager.PutHealthResponse(resp)
}

// GetPooledErrorResponse gets a pooled error response from global manager
func GetPooledErrorResponse() *ErrorResponse {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	return GlobalResponsePoolManager.GetErrorResponse()
}

// PutPooledErrorResponse returns an error response to global pool
func PutPooledErrorResponse(resp *ErrorResponse) {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	GlobalResponsePoolManager.PutErrorResponse(resp)
}

// GetPooledValidationResponse gets a pooled validation response from global manager
func GetPooledValidationResponse() *ValidationResponse {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	return GlobalResponsePoolManager.GetValidationResponse()
}

// PutPooledValidationResponse returns a validation response to global pool
func PutPooledValidationResponse(resp *ValidationResponse) {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	GlobalResponsePoolManager.PutValidationResponse(resp)
}

// GetPooledUserResponse gets a pooled user response from global manager
func GetPooledUserResponse() *UserResponse {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	return GlobalResponsePoolManager.GetUserResponse()
}

// PutPooledUserResponse returns a user response to global pool
func PutPooledUserResponse(resp *UserResponse) {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	GlobalResponsePoolManager.PutUserResponse(resp)
}

// GetPooledGenericResponse gets a pooled generic response from global manager
func GetPooledGenericResponse() *GenericResponse {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	return GlobalResponsePoolManager.GetGenericResponse()
}

// PutPooledGenericResponse returns a generic response to global pool
func PutPooledGenericResponse(resp *GenericResponse) {
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	GlobalResponsePoolManager.PutGenericResponse(resp)
}