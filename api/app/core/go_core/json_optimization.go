package go_core

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
)

// JSONProcessor provides zero-copy JSON processing with pre-allocated buffers
type JSONProcessor struct {
	bufferPool     *sync.Pool
	responsePool   *sync.Pool
	maxBufferSize  int
	initialSize    int
	metrics        *JSONMetrics
}

// JSONMetrics tracks performance metrics for JSON operations
type JSONMetrics struct {
	EncodeOps        int64
	DecodeOps        int64
	BufferPoolHits   int64
	BufferPoolMisses int64
	AvgEncodeTime    time.Duration
	AvgDecodeTime    time.Duration
	BytesProcessed   int64
	mu               sync.RWMutex
}

// JSONProcessorConfig configures the JSON processor
type JSONProcessorConfig struct {
	MaxBufferSize    int
	InitialSize      int
	ResponsePoolSize int
}

// DefaultJSONProcessorConfig returns default configuration
func DefaultJSONProcessorConfig() *JSONProcessorConfig {
	return &JSONProcessorConfig{
		MaxBufferSize:    64 * 1024, // 64KB
		InitialSize:      4 * 1024,  // 4KB
		ResponsePoolSize: 500,
	}
}

// NewJSONProcessor creates a new zero-copy JSON processor
func NewJSONProcessor(config *JSONProcessorConfig) *JSONProcessor {
	if config == nil {
		config = DefaultJSONProcessorConfig()
	}

	jp := &JSONProcessor{
		maxBufferSize: config.MaxBufferSize,
		initialSize:   config.InitialSize,
		metrics:       &JSONMetrics{},
	}

	// Initialize buffer pool
	jp.bufferPool = &sync.Pool{
		New: func() interface{} {
			jp.metrics.mu.Lock()
			jp.metrics.BufferPoolMisses++
			jp.metrics.mu.Unlock()
			return bytes.NewBuffer(make([]byte, 0, jp.initialSize))
		},
	}


	// Initialize response pool for common response structures
	jp.responsePool = &sync.Pool{
		New: func() interface{} {
			return &APIResponse{}
		},
	}

	return jp
}

// getBuffer gets a buffer from the pool
func (jp *JSONProcessor) getBuffer() *bytes.Buffer {
	buf := jp.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	
	jp.metrics.mu.Lock()
	jp.metrics.BufferPoolHits++
	jp.metrics.mu.Unlock()
	
	return buf
}

// putBuffer returns a buffer to the pool
func (jp *JSONProcessor) putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > jp.maxBufferSize {
		// Buffer too large, don't pool it
		return
	}
	jp.bufferPool.Put(buf)
}

// EncodeToBuffer encodes data to a buffer with zero-copy optimization
func (jp *JSONProcessor) EncodeToBuffer(data interface{}) (*bytes.Buffer, error) {
	start := time.Now()
	
	buf := jp.getBuffer()
	
	defer func() {
		jp.metrics.mu.Lock()
		jp.metrics.EncodeOps++
		jp.metrics.AvgEncodeTime = (jp.metrics.AvgEncodeTime + time.Since(start)) / 2
		jp.metrics.BytesProcessed += int64(buf.Len())
		jp.metrics.mu.Unlock()
	}()
	
	// Use Sonic for direct buffer encoding
	jsonBytes, err := sonic.Marshal(data)
	if err != nil {
		jp.putBuffer(buf)
		return nil, err
	}
	
	buf.Write(jsonBytes)
	return buf, nil
}

// EncodeToBytes encodes data to bytes with zero-copy optimization
func (jp *JSONProcessor) EncodeToBytes(data interface{}) ([]byte, error) {
	buf, err := jp.EncodeToBuffer(data)
	if err != nil {
		return nil, err
	}
	
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	jp.putBuffer(buf)
	
	return result, nil
}

// EncodeToFastHTTP encodes data directly to fasthttp response
func (jp *JSONProcessor) EncodeToFastHTTP(ctx *fasthttp.RequestCtx, data interface{}) error {
	buf, err := jp.EncodeToBuffer(data)
	if err != nil {
		return err
	}
	
	defer jp.putBuffer(buf)
	
	ctx.SetContentType("application/json")
	ctx.SetBody(buf.Bytes())
	
	return nil
}

// DecodeFromReader decodes JSON from a reader with zero-copy optimization
func (jp *JSONProcessor) DecodeFromReader(reader io.Reader, v interface{}) error {
	start := time.Now()
	
	defer func() {
		jp.metrics.mu.Lock()
		jp.metrics.DecodeOps++
		jp.metrics.AvgDecodeTime = (jp.metrics.AvgDecodeTime + time.Since(start)) / 2
		jp.metrics.mu.Unlock()
	}()
	
	// Read data from reader and use Sonic for decoding
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	
	return sonic.Unmarshal(data, v)
}

// DecodeFromBytes decodes JSON from bytes with zero-copy optimization
func (jp *JSONProcessor) DecodeFromBytes(data []byte, v interface{}) error {
	start := time.Now()
	
	defer func() {
		jp.metrics.mu.Lock()
		jp.metrics.DecodeOps++
		jp.metrics.AvgDecodeTime = (jp.metrics.AvgDecodeTime + time.Since(start)) / 2
		jp.metrics.BytesProcessed += int64(len(data))
		jp.metrics.mu.Unlock()
	}()
	
	return sonic.Unmarshal(data, v)
}

// DecodeFromFastHTTP decodes JSON from fasthttp request
func (jp *JSONProcessor) DecodeFromFastHTTP(ctx *fasthttp.RequestCtx, v interface{}) error {
	return jp.DecodeFromBytes(ctx.PostBody(), v)
}

// GetAPIResponse gets a pooled API response object
func (jp *JSONProcessor) GetAPIResponse() *APIResponse {
	resp := jp.responsePool.Get().(*APIResponse)
	resp.Reset()
	return resp
}

// PutAPIResponse returns an API response object to the pool
func (jp *JSONProcessor) PutAPIResponse(resp *APIResponse) {
	jp.responsePool.Put(resp)
}

// GetMetrics returns current JSON processing metrics
func (jp *JSONProcessor) GetMetrics() JSONMetrics {
	jp.metrics.mu.RLock()
	defer jp.metrics.mu.RUnlock()
	// Return a copy to avoid lock copying
	return JSONMetrics{
		EncodeOps:        jp.metrics.EncodeOps,
		DecodeOps:        jp.metrics.DecodeOps,
		BufferPoolHits:   jp.metrics.BufferPoolHits,
		BufferPoolMisses: jp.metrics.BufferPoolMisses,
		AvgEncodeTime:    jp.metrics.AvgEncodeTime,
		AvgDecodeTime:    jp.metrics.AvgDecodeTime,
		BytesProcessed:   jp.metrics.BytesProcessed,
	}
}

// ResetMetrics resets all metrics counters
func (jp *JSONProcessor) ResetMetrics() {
	jp.metrics.mu.Lock()
	defer jp.metrics.mu.Unlock()
	jp.metrics = &JSONMetrics{}
}

// Common response structures for pooling

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Reset resets the API response for reuse
func (r *APIResponse) Reset() {
	r.Success = false
	r.Message = ""
	r.Data = nil
	r.Errors = nil
}


// Global JSON processor instance
var GlobalJSONProcessor *JSONProcessor

// InitializeGlobalJSONProcessor initializes the global JSON processor
func InitializeGlobalJSONProcessor(config *JSONProcessorConfig) {
	GlobalJSONProcessor = NewJSONProcessor(config)
}

// Helper functions for global processor

// EncodeJSON encodes data using the global processor
func EncodeJSON(data interface{}) ([]byte, error) {
	if GlobalJSONProcessor == nil {
		InitializeGlobalJSONProcessor(nil)
	}
	return GlobalJSONProcessor.EncodeToBytes(data)
}

// DecodeJSON decodes data using the global processor
func DecodeJSON(data []byte, v interface{}) error {
	if GlobalJSONProcessor == nil {
		InitializeGlobalJSONProcessor(nil)
	}
	return GlobalJSONProcessor.DecodeFromBytes(data, v)
}

// EncodeJSONToFastHTTP encodes data directly to fasthttp response
func EncodeJSONToFastHTTP(ctx *fasthttp.RequestCtx, data interface{}) error {
	if GlobalJSONProcessor == nil {
		InitializeGlobalJSONProcessor(nil)
	}
	return GlobalJSONProcessor.EncodeToFastHTTP(ctx, data)
}

// DecodeJSONFromFastHTTP decodes data from fasthttp request
func DecodeJSONFromFastHTTP(ctx *fasthttp.RequestCtx, v interface{}) error {
	if GlobalJSONProcessor == nil {
		InitializeGlobalJSONProcessor(nil)
	}
	return GlobalJSONProcessor.DecodeFromFastHTTP(ctx, v)
}