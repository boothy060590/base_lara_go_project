package logging_core

import (
	"time"

	app_core "base_lara_go_project/app/core/app"
	exceptions_core "base_lara_go_project/app/core/exceptions"

	"github.com/getsentry/sentry-go"
)

// SentryExceptionHandler provides Sentry integration for exception handling
type SentryExceptionHandler struct {
	logger app_core.LoggerInterface
}

// NewSentryExceptionHandler creates a new Sentry exception handler
func NewSentryExceptionHandler(logger app_core.LoggerInterface) *SentryExceptionHandler {
	return &SentryExceptionHandler{
		logger: logger,
	}
}

// Report reports an exception to Sentry and logging system
func (h *SentryExceptionHandler) Report(exception error) error {
	if !h.ShouldReport(exception) {
		return nil
	}

	// Mark as reported if it's our exception type
	if ex, ok := exception.(*exceptions_core.Exception); ok {
		ex.MarkAsReported()
	}

	// Log the exception first
	ex := exceptions_core.ConvertToException(exception)
	if err := h.logger.Error(exception.Error(), map[string]interface{}{
		"exception": exception,
		"file":      ex.GetFile(),
		"line":      ex.GetLine(),
		"trace":     ex.GetTrace(),
	}); err != nil {
		return err
	}

	// Report to Sentry
	return h.reportToSentry(exception)
}

// reportToSentry sends the exception to Sentry
func (h *SentryExceptionHandler) reportToSentry(exception error) error {
	// Create Sentry event
	event := sentry.NewEvent()

	// Set the error message
	event.Message = exception.Error()
	event.Level = sentry.LevelError
	event.Timestamp = time.Now()

	// Add exception context
	if ex, ok := exception.(*exceptions_core.Exception); ok {
		// Add exception context as extra data
		if len(ex.Context) > 0 {
			event.Extra = ex.Context
		}

		// Add exception metadata
		event.Extra["exception_code"] = ex.GetCode()
		event.Extra["exception_file"] = ex.GetFile()
		event.Extra["exception_line"] = ex.GetLine()
		event.Extra["exception_trace"] = ex.GetTrace()

		// Set appropriate Sentry level based on exception code
		switch ex.GetCode() {
		case 404:
			event.Level = sentry.LevelWarning
		case 401, 403:
			event.Level = sentry.LevelInfo
		case 422:
			event.Level = sentry.LevelWarning
		case 500:
			event.Level = sentry.LevelError
		default:
			event.Level = sentry.LevelError
		}

		// Add tags for better categorization
		event.Tags = map[string]string{
			"exception_type": "framework_exception",
			"exception_code": string(rune(ex.GetCode())),
		}

		// Add fingerprint for grouping similar exceptions
		event.Fingerprint = []string{
			ex.GetFile(),
			string(rune(ex.GetCode())),
		}
	} else {
		// Handle standard errors
		event.Extra = map[string]interface{}{
			"error_type": "standard_error",
		}
	}

	// Capture the event
	sentry.CaptureEvent(event)

	return nil
}

// Render renders an exception for display
func (h *SentryExceptionHandler) Render(exception error) interface{} {
	if !h.ShouldRender(exception) {
		return nil
	}

	// Convert to our exception type if needed
	ex := exceptions_core.ConvertToException(exception)

	return map[string]interface{}{
		"message": ex.GetMessage(),
		"context": ex.GetContext(),
	}
}

// ShouldReport determines if an exception should be reported
func (h *SentryExceptionHandler) ShouldReport(exception error) bool {
	// Don't report if already reported
	if ex, ok := exception.(*exceptions_core.Exception); ok && ex.IsReported() {
		return false
	}

	// Don't report certain exception types
	if _, ok := exception.(*exceptions_core.ValidationException); ok {
		return false // Validation errors are usually not reported to Sentry
	}

	// Don't report authentication exceptions to Sentry (security)
	if _, ok := exception.(*exceptions_core.AuthenticationException); ok {
		return false
	}

	return true
}

// ShouldRender determines if an exception should be rendered
func (h *SentryExceptionHandler) ShouldRender(exception error) bool {
	// Always render in development
	// In production, only render certain types
	return true
}

// Flush flushes Sentry events
func (h *SentryExceptionHandler) Flush() error {
	// Flush Sentry events
	sentry.Flush(2 * time.Second)
	return nil
}
