package exceptions_core

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

// Error implements the error interface
func (e *SimpleException) Error() string { return e.message }

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
func ConvertToException(err error) *SimpleException {
	// Try to convert to our exception type
	if ex, ok := err.(*SimpleException); ok {
		return ex
	}

	// Create a simple wrapper that implements the interface
	return NewSimpleException(err.Error(), 500)
}
