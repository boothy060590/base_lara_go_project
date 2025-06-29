package exceptions_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// SimpleException is a basic implementation of the Exception interface
type SimpleException struct {
	code    int
	message string
	file    string
	line    int
	trace   []string
	context map[string]interface{}
}

// NewSimpleException creates a new simple exception
func NewSimpleException(message string, code int) *SimpleException {
	return &SimpleException{
		code:    code,
		message: message,
		file:    "",
		line:    0,
		trace:   []string{},
		context: map[string]interface{}{},
	}
}

// GetCode implements app_core.Exception interface
func (e *SimpleException) GetCode() int { return e.code }

// GetMessage implements app_core.Exception interface
func (e *SimpleException) GetMessage() string { return e.message }

// GetFile implements app_core.Exception interface
func (e *SimpleException) GetFile() string { return e.file }

// GetLine implements app_core.Exception interface
func (e *SimpleException) GetLine() int { return e.line }

// GetTrace implements app_core.Exception interface
func (e *SimpleException) GetTrace() []string { return e.trace }

// GetContext implements app_core.Exception interface
func (e *SimpleException) GetContext() map[string]interface{} { return e.context }

// ConvertToException converts a generic error to our framework exception
func ConvertToException(err error) app_core.Exception {
	// Try to convert to our exception type
	if ex, ok := err.(app_core.Exception); ok {
		return ex
	}

	// Create a simple wrapper that implements the interface
	return NewSimpleException(err.Error(), 500)
}
