package go_core

import (
	"io"
	"unsafe"

	"github.com/valyala/fasthttp"
)

// StreamingJSONEncoder provides zero-copy streaming JSON encoding
type StreamingJSONEncoder struct {
	processor *JSONProcessor
}

// NewStreamingJSONEncoder creates a new streaming JSON encoder
func NewStreamingJSONEncoder(processor *JSONProcessor) *StreamingJSONEncoder {
	return &StreamingJSONEncoder{
		processor: processor,
	}
}

// StreamResponse streams a response directly to the writer with zero-copy optimization
func (sje *StreamingJSONEncoder) StreamResponse(writer io.Writer, resp interface{}) error {
	// Get buffer from pool
	buf := sje.processor.getBuffer()
	defer sje.processor.putBuffer(buf)
	
	// Encode to buffer using JSON processor
	if encodeBuffer, err := sje.processor.EncodeToBuffer(resp); err != nil {
		return err
	} else {
		defer sje.processor.putBuffer(encodeBuffer)
		// Write directly to output with zero-copy
		_, err := writer.Write(encodeBuffer.Bytes())
		return err
	}
}

// StreamToFastHTTP streams response directly to FastHTTP context
func (sje *StreamingJSONEncoder) StreamToFastHTTP(ctx *fasthttp.RequestCtx, resp interface{}) error {
	// Use JSON processor to encode
	if jsonBytes, err := sje.processor.EncodeToBytes(resp); err != nil {
		return err
	} else {
		// Set content type and body directly
		ctx.SetContentType("application/json")
		ctx.SetBody(jsonBytes)
		return nil
	}
}

// StreamArray streams an array of objects with zero-copy optimization
func (sje *StreamingJSONEncoder) StreamArray(writer io.Writer, items []interface{}) error {
	// Write opening bracket
	if _, err := writer.Write([]byte("[")); err != nil {
		return err
	}
	
	// Stream each item
	for i, item := range items {
		if i > 0 {
			if _, err := writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		
		if err := sje.StreamResponse(writer, item); err != nil {
			return err
		}
	}
	
	// Write closing bracket
	_, err := writer.Write([]byte("]"))
	return err
}

// StreamObjectField streams a single object field with zero-copy
func (sje *StreamingJSONEncoder) StreamObjectField(writer io.Writer, key string, value interface{}) error {
	// Write key
	if _, err := writer.Write([]byte(`"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(key)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(`":`)); err != nil {
		return err
	}
	
	// Stream value
	return sje.StreamResponse(writer, value)
}

// SchemaBasedJSONBuilder provides schema-based JSON building
type SchemaBasedJSONBuilder struct {
	processor *JSONProcessor
}

// NewSchemaBasedJSONBuilder creates a new schema-based JSON builder
func NewSchemaBasedJSONBuilder(processor *JSONProcessor) *SchemaBasedJSONBuilder {
	return &SchemaBasedJSONBuilder{
		processor: processor,
	}
}

// ZeroCopyJSONBuilder provides zero-copy JSON building using schema manager
type ZeroCopyJSONBuilder struct {
	processor *JSONProcessor
}

// NewZeroCopyJSONBuilder creates a new zero-copy JSON builder
func NewZeroCopyJSONBuilder(processor *JSONProcessor) *ZeroCopyJSONBuilder {
	return &ZeroCopyJSONBuilder{
		processor: processor,
	}
}

// BuildLoginResponse builds a login response using schema-based rendering
func (zcjb *ZeroCopyJSONBuilder) BuildLoginResponse(writer io.Writer, message, token string, user interface{}) error {
	// Use schema manager if available
	if schemaManager := GetGlobalSchemaManager(); schemaManager != nil {
		data := map[string]interface{}{
			"message": message,
			"token":   token,
			"user":    user,
		}
		return schemaManager.RenderSchemaToWriter("login_success", data, writer)
	}
	
	// Fallback to manual building
	return zcjb.buildLoginResponseFallback(writer, message, token, user)
}

// buildLoginResponseFallback is a fallback when schema manager is not available
func (zcjb *ZeroCopyJSONBuilder) buildLoginResponseFallback(writer io.Writer, message, token string, user interface{}) error {
	if _, err := writer.Write([]byte(`{"message":"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(`","user":`)); err != nil {
		return err
	}
	if err := zcjb.streamValue(writer, user); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(`,"token":"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(token)); err != nil {
		return err
	}
	_, err := writer.Write([]byte(`"}`))
	return err
}

// BuildErrorResponse builds an error response using schema-based rendering
func (zcjb *ZeroCopyJSONBuilder) BuildErrorResponse(writer io.Writer, message string, errors interface{}) error {
	// Use schema manager if available
	if schemaManager := GetGlobalSchemaManager(); schemaManager != nil {
		data := map[string]interface{}{
			"message": message,
			"errors":  errors,
		}
		return schemaManager.RenderSchemaToWriter("api_error", data, writer)
	}
	
	// Fallback to manual building
	return zcjb.buildErrorResponseFallback(writer, message, errors)
}

// buildErrorResponseFallback is a fallback when schema manager is not available
func (zcjb *ZeroCopyJSONBuilder) buildErrorResponseFallback(writer io.Writer, message string, errors interface{}) error {
	if _, err := writer.Write([]byte(`{"success":false,"message":"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(`","errors":`)); err != nil {
		return err
	}
	if err := zcjb.streamValue(writer, errors); err != nil {
		return err
	}
	_, err := writer.Write([]byte(`}`))
	return err
}

// BuildHealthResponse builds a health response using schema-based rendering
func (zcjb *ZeroCopyJSONBuilder) BuildHealthResponse(writer io.Writer, status string, timestamp int64, service, server string) error {
	// Use schema manager if available
	if schemaManager := GetGlobalSchemaManager(); schemaManager != nil {
		data := map[string]interface{}{
			"status":    status,
			"timestamp": timestamp,
			"service":   service,
			"server":    server,
		}
		return schemaManager.RenderSchemaToWriter("health_check", data, writer)
	}
	
	// Fallback to manual building
	return zcjb.buildHealthResponseFallback(writer, status, timestamp, service, server)
}

// buildHealthResponseFallback is a fallback when schema manager is not available
func (zcjb *ZeroCopyJSONBuilder) buildHealthResponseFallback(writer io.Writer, status string, timestamp int64, service, server string) error {
	if _, err := writer.Write([]byte(`{"status":"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(status)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(`","timestamp":`)); err != nil {
		return err
	}
	timestampStr := int64ToString(timestamp)
	if _, err := writer.Write(timestampStr); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(`,"service":"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(service)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(`","server":"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(server)); err != nil {
		return err
	}
	_, err := writer.Write([]byte(`"}`))
	return err
}

// BuildGenericResponse builds a generic response using schema-based rendering
func (zcjb *ZeroCopyJSONBuilder) BuildGenericResponse(writer io.Writer, success bool, message string, data interface{}) error {
	// Use schema manager if available
	if schemaManager := GetGlobalSchemaManager(); schemaManager != nil {
		schemaName := "api_success"
		if !success {
			schemaName = "api_error"
		}
		
		schemaData := map[string]interface{}{
			"message": message,
		}
		
		if success {
			schemaData["data"] = data
		} else {
			schemaData["errors"] = data
		}
		
		return schemaManager.RenderSchemaToWriter(schemaName, schemaData, writer)
	}
	
	// Fallback to manual building
	return zcjb.buildGenericResponseFallback(writer, success, message, data)
}

// buildGenericResponseFallback is a fallback when schema manager is not available
func (zcjb *ZeroCopyJSONBuilder) buildGenericResponseFallback(writer io.Writer, success bool, message string, data interface{}) error {
	if _, err := writer.Write([]byte(`{"success":`)); err != nil {
		return err
	}
	if success {
		if _, err := writer.Write([]byte("true")); err != nil {
			return err
		}
	} else {
		if _, err := writer.Write([]byte("false")); err != nil {
			return err
		}
	}
	if _, err := writer.Write([]byte(`,"message":"`)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if data != nil {
		if _, err := writer.Write([]byte(`","data":`)); err != nil {
			return err
		}
		if err := zcjb.streamValue(writer, data); err != nil {
			return err
		}
		_, err := writer.Write([]byte(`}`))
		return err
	}
	_, err := writer.Write([]byte(`"`))
	return err
}

// streamValue streams a value using the JSON processor
func (zcjb *ZeroCopyJSONBuilder) streamValue(writer io.Writer, value interface{}) error {
	// Use JSON processor to encode
	if jsonBytes, err := zcjb.processor.EncodeToBytes(value); err != nil {
		return err
	} else {
		// Write content (remove trailing newline if present)
		content := jsonBytes
		if len(content) > 0 && content[len(content)-1] == '\n' {
			content = content[:len(content)-1]
		}
		
		_, err := writer.Write(content)
		return err
	}
}

// int64ToString converts int64 to string with zero-copy optimization
func int64ToString(n int64) []byte {
	// Simple implementation for positive numbers
	if n == 0 {
		return []byte("0")
	}
	
	// For negative numbers, handle sign
	negative := n < 0
	if negative {
		n = -n
	}
	
	// Convert to string
	var buf [20]byte // Max digits for int64
	i := len(buf) - 1
	
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	
	if negative {
		buf[i] = '-'
		i--
	}
	
	// Return slice of buffer
	return buf[i+1:]
}

// UnsafeStringToBytes converts string to bytes with zero-copy (use with caution)
func UnsafeStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// UnsafeBytesToString converts bytes to string with zero-copy (use with caution)
func UnsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Global instances
var (
	GlobalStreamingEncoder *StreamingJSONEncoder
	GlobalZeroCopyBuilder  *ZeroCopyJSONBuilder
	GlobalSchemaBuilder    *SchemaBasedJSONBuilder
)

// InitializeGlobalStreamingComponents initializes global streaming components
func InitializeGlobalStreamingComponents() {
	if GlobalJSONProcessor == nil {
		InitializeGlobalJSONProcessor(nil)
	}
	
	GlobalStreamingEncoder = NewStreamingJSONEncoder(GlobalJSONProcessor)
	GlobalZeroCopyBuilder = NewZeroCopyJSONBuilder(GlobalJSONProcessor)
	GlobalSchemaBuilder = NewSchemaBasedJSONBuilder(GlobalJSONProcessor)
}

// Helper functions for global streaming components

// StreamJSONResponse streams a response using global encoder
func StreamJSONResponse(writer io.Writer, resp interface{}) error {
	if GlobalStreamingEncoder == nil {
		InitializeGlobalStreamingComponents()
	}
	return GlobalStreamingEncoder.StreamResponse(writer, resp)
}

// StreamJSONToFastHTTP streams response to FastHTTP context
func StreamJSONToFastHTTP(ctx *fasthttp.RequestCtx, resp interface{}) error {
	if GlobalStreamingEncoder == nil {
		InitializeGlobalStreamingComponents()
	}
	return GlobalStreamingEncoder.StreamToFastHTTP(ctx, resp)
}

// BuildZeroCopyLoginResponse builds login response with zero-copy
func BuildZeroCopyLoginResponse(writer io.Writer, message, token string, user interface{}) error {
	if GlobalZeroCopyBuilder == nil {
		InitializeGlobalStreamingComponents()
	}
	return GlobalZeroCopyBuilder.BuildLoginResponse(writer, message, token, user)
}

// BuildZeroCopyErrorResponse builds error response with zero-copy
func BuildZeroCopyErrorResponse(writer io.Writer, message string, errors interface{}) error {
	if GlobalZeroCopyBuilder == nil {
		InitializeGlobalStreamingComponents()
	}
	return GlobalZeroCopyBuilder.BuildErrorResponse(writer, message, errors)
}

// BuildZeroCopyHealthResponse builds health response with zero-copy
func BuildZeroCopyHealthResponse(writer io.Writer, status string, timestamp int64, service, server string) error {
	if GlobalZeroCopyBuilder == nil {
		InitializeGlobalStreamingComponents()
	}
	return GlobalZeroCopyBuilder.BuildHealthResponse(writer, status, timestamp, service, server)
}

// BuildZeroCopyGenericResponse builds generic response with zero-copy
func BuildZeroCopyGenericResponse(writer io.Writer, success bool, message string, data interface{}) error {
	if GlobalZeroCopyBuilder == nil {
		InitializeGlobalStreamingComponents()
	}
	return GlobalZeroCopyBuilder.BuildGenericResponse(writer, success, message, data)
}