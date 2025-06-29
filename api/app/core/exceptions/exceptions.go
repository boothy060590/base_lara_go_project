package exceptions_core

import (
	"fmt"
	"runtime"
	"time"

	app_core "base_lara_go_project/app/core/app"
)

// Exception represents a framework exception with Laravel-style features
type Exception struct {
	Message    string                 `json:"message"`
	Code       int                    `json:"code"`
	File       string                 `json:"file"`
	Line       int                    `json:"line"`
	Trace      []string               `json:"trace"`
	Context    map[string]interface{} `json:"context"`
	Previous   error                  `json:"-"`
	Reported   bool                   `json:"reported"`
	ReportedAt *time.Time             `json:"reported_at"`
}

// Error implements the error interface
func (e *Exception) Error() string {
	return e.Message
}

// GetCode returns the exception code
func (e *Exception) GetCode() int {
	return e.Code
}

// GetFile returns the file where the exception occurred
func (e *Exception) GetFile() string {
	return e.File
}

// GetLine returns the line where the exception occurred
func (e *Exception) GetLine() int {
	return e.Line
}

// GetTrace returns the stack trace
func (e *Exception) GetTrace() []string {
	return e.Trace
}

// GetContext returns the exception context
func (e *Exception) GetContext() map[string]interface{} {
	return e.Context
}

// GetPrevious returns the previous exception
func (e *Exception) GetPrevious() error {
	return e.Previous
}

// IsReported returns whether the exception has been reported
func (e *Exception) IsReported() bool {
	return e.Reported
}

// MarkAsReported marks the exception as reported
func (e *Exception) MarkAsReported() {
	e.Reported = true
	now := time.Now()
	e.ReportedAt = &now
}

// AddContext adds context to the exception
func (e *Exception) AddContext(key string, value interface{}) {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
}

// NewException creates a new exception
func NewException(message string, code int) *Exception {
	_, file, line, _ := runtime.Caller(1)
	return &Exception{
		Message: message,
		Code:    code,
		File:    file,
		Line:    line,
		Context: make(map[string]interface{}),
	}
}

// NewExceptionWithContext creates a new exception with context
func NewExceptionWithContext(message string, code int, context map[string]interface{}) *Exception {
	_, file, line, _ := runtime.Caller(1)
	return &Exception{
		Message: message,
		Code:    code,
		File:    file,
		Line:    line,
		Context: context,
	}
}

// WrapException wraps an existing error as an exception
func WrapException(err error, message string, code int) *Exception {
	_, file, line, _ := runtime.Caller(1)
	exception := &Exception{
		Message:  message,
		Code:     code,
		File:     file,
		Line:     line,
		Previous: err,
		Context:  make(map[string]interface{}),
	}

	// If the wrapped error is already an exception, preserve its context
	if existingException, ok := err.(*Exception); ok {
		for k, v := range existingException.Context {
			exception.Context[k] = v
		}
	}

	return exception
}

// Common exception types (Laravel-style)

// ModelNotFoundException represents a model not found error
type ModelNotFoundException struct {
	*Exception
	Model string      `json:"model"`
	ID    interface{} `json:"id"`
}

// NewModelNotFoundException creates a new model not found exception
func NewModelNotFoundException(model string, id interface{}) *ModelNotFoundException {
	exception := NewException(fmt.Sprintf("No query results for model [%s] %v", model, id), 404)
	return &ModelNotFoundException{
		Exception: exception,
		Model:     model,
		ID:        id,
	}
}

// ValidationException represents a validation error
type ValidationException struct {
	*Exception
	Errors map[string][]string `json:"errors"`
}

// NewValidationException creates a new validation exception
func NewValidationException(errors map[string][]string) *ValidationException {
	exception := NewException("The given data was invalid.", 422)
	return &ValidationException{
		Exception: exception,
		Errors:    errors,
	}
}

// AuthenticationException represents an authentication error
type AuthenticationException struct {
	*Exception
}

// NewAuthenticationException creates a new authentication exception
func NewAuthenticationException(message string) *AuthenticationException {
	if message == "" {
		message = "Unauthenticated."
	}
	exception := NewException(message, 401)
	return &AuthenticationException{
		Exception: exception,
	}
}

// AuthorizationException represents an authorization error
type AuthorizationException struct {
	*Exception
}

// NewAuthorizationException creates a new authorization exception
func NewAuthorizationException(message string) *AuthorizationException {
	if message == "" {
		message = "This action is unauthorized."
	}
	exception := NewException(message, 403)
	return &AuthorizationException{
		Exception: exception,
	}
}

// QueryException represents a database query error
type QueryException struct {
	*Exception
	SQL  string        `json:"sql"`
	Bind []interface{} `json:"bind"`
}

// NewQueryException creates a new query exception
func NewQueryException(message string, sql string, bind []interface{}) *QueryException {
	exception := NewException(message, 500)
	return &QueryException{
		Exception: exception,
		SQL:       sql,
		Bind:      bind,
	}
}

// FileNotFoundException represents a file not found error
type FileNotFoundException struct {
	*Exception
	Path string `json:"path"`
}

// NewFileNotFoundException creates a new file not found exception
func NewFileNotFoundException(path string) *FileNotFoundException {
	exception := NewException(fmt.Sprintf("File not found: %s", path), 404)
	return &FileNotFoundException{
		Exception: exception,
		Path:      path,
	}
}

// ConfigurationException represents a configuration error
type ConfigurationException struct {
	*Exception
	Key string `json:"key"`
}

// NewConfigurationException creates a new configuration exception
func NewConfigurationException(key string, message string) *ConfigurationException {
	if message == "" {
		message = fmt.Sprintf("Configuration key '%s' not found or invalid", key)
	}
	exception := NewException(message, 500)
	return &ConfigurationException{
		Exception: exception,
		Key:       key,
	}
}

// ExceptionHandler defines the interface for exception handling
type ExceptionHandler interface {
	Report(exception error) error
	Render(exception error) interface{}
	ShouldReport(exception error) bool
	ShouldRender(exception error) bool
}

// DefaultExceptionHandler provides default exception handling
type DefaultExceptionHandler struct {
	logger app_core.LoggerInterface
}

// NewDefaultExceptionHandler creates a new default exception handler
func NewDefaultExceptionHandler(logger app_core.LoggerInterface) *DefaultExceptionHandler {
	return &DefaultExceptionHandler{
		logger: logger,
	}
}

// Report reports an exception to the logging system
func (h *DefaultExceptionHandler) Report(exception error) error {
	if !h.ShouldReport(exception) {
		return nil
	}

	// Mark as reported if it's our exception type
	if ex, ok := exception.(*Exception); ok {
		ex.MarkAsReported()
	}

	// Log the exception
	return h.logger.Error(exception.Error(), map[string]interface{}{
		"exception": exception,
		"file":      getExceptionFile(exception),
		"line":      getExceptionLine(exception),
		"trace":     getExceptionTrace(exception),
	})
}

// Render renders an exception for display
func (h *DefaultExceptionHandler) Render(exception error) interface{} {
	if !h.ShouldRender(exception) {
		return nil
	}

	// Convert to our exception type if needed
	ex := convertToException(exception)

	return map[string]interface{}{
		"message": ex.Message,
		"code":    ex.Code,
		"file":    ex.File,
		"line":    ex.Line,
		"trace":   ex.Trace,
		"context": ex.Context,
	}
}

// ShouldReport determines if an exception should be reported
func (h *DefaultExceptionHandler) ShouldReport(exception error) bool {
	// Don't report if already reported
	if ex, ok := exception.(*Exception); ok && ex.IsReported() {
		return false
	}

	// Don't report certain exception types in production
	if _, ok := exception.(*ValidationException); ok {
		return false // Validation errors are usually not reported
	}

	return true
}

// ShouldRender determines if an exception should be rendered
func (h *DefaultExceptionHandler) ShouldRender(exception error) bool {
	// Always render in development
	// In production, only render certain types
	return true
}

// Helper functions

func getExceptionFile(exception error) string {
	if ex, ok := exception.(*Exception); ok {
		return ex.File
	}
	return ""
}

func getExceptionLine(exception error) int {
	if ex, ok := exception.(*Exception); ok {
		return ex.Line
	}
	return 0
}

func getExceptionTrace(exception error) []string {
	if ex, ok := exception.(*Exception); ok {
		return ex.Trace
	}
	return []string{}
}

func convertToException(err error) *Exception {
	if ex, ok := err.(*Exception); ok {
		return ex
	}

	// Convert standard errors to exceptions
	return WrapException(err, err.Error(), 500)
}

// Global exception handler instance
var ExceptionHandlerInstance ExceptionHandler

// SetExceptionHandler sets the global exception handler
func SetExceptionHandler(handler ExceptionHandler) {
	ExceptionHandlerInstance = handler
}

// Report reports an exception (Laravel-style report() function)
func Report(exception error) error {
	if ExceptionHandlerInstance == nil {
		return fmt.Errorf("exception handler not set")
	}
	return ExceptionHandlerInstance.Report(exception)
}

// ReportError reports an error with additional context
func ReportError(err error, context map[string]interface{}) error {
	if ex, ok := err.(*Exception); ok {
		for k, v := range context {
			ex.AddContext(k, v)
		}
	} else {
		// Wrap standard errors
		ex := WrapException(err, err.Error(), 500)
		for k, v := range context {
			ex.AddContext(k, v)
		}
		err = ex
	}

	return Report(err)
}

// Render renders an exception for display
func Render(exception error) interface{} {
	if ExceptionHandlerInstance == nil {
		return nil
	}
	return ExceptionHandlerInstance.Render(exception)
}
