package go_core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// HTTPOptimizationConfig defines configuration for HTTP optimizations
type HTTPOptimizationConfig struct {
	// Server configuration
	EnableFastHTTP           bool          `json:"enable_fasthttp"`
	ReadTimeout             time.Duration `json:"read_timeout"`
	WriteTimeout            time.Duration `json:"write_timeout"`
	IdleTimeout             time.Duration `json:"idle_timeout"`
	MaxRequestBodySize      int           `json:"max_request_body_size"`
	
	// Connection pooling
	MaxConnections          int           `json:"max_connections"`
	MaxIdleConnections      int           `json:"max_idle_connections"`
	ConnectionTimeout       time.Duration `json:"connection_timeout"`
	
	// Performance optimizations
	EnableCompression       bool          `json:"enable_compression"`
	EnableKeepAlive        bool          `json:"enable_keep_alive"`
	PreAllocatedBuffers    int           `json:"pre_allocated_buffers"`
	ZeroCopyEnabled        bool          `json:"zero_copy_enabled"`
	
	// CORS configuration
	CORSAllowedOrigins     []string      `json:"cors_allowed_origins"`
	CORSAllowedMethods     []string      `json:"cors_allowed_methods"`
	CORSAllowedHeaders     []string      `json:"cors_allowed_headers"`
	CORSExposedHeaders     []string      `json:"cors_exposed_headers"`
	CORSAllowCredentials   bool          `json:"cors_allow_credentials"`
	CORSMaxAge            int           `json:"cors_max_age"`
	
	// Monitoring
	EnableMetrics          bool          `json:"enable_metrics"`
	MetricsPath           string        `json:"metrics_path"`
}

// DefaultHTTPOptimizationConfig returns sensible defaults for HTTP optimization
func DefaultHTTPOptimizationConfig() *HTTPOptimizationConfig {
	return &HTTPOptimizationConfig{
		EnableFastHTTP:         true,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		MaxRequestBodySize:    10 * 1024 * 1024, // 10MB
		MaxConnections:        10000,
		MaxIdleConnections:    1000,
		ConnectionTimeout:     5 * time.Second,
		EnableCompression:     true,
		EnableKeepAlive:      true,
		PreAllocatedBuffers:  100,
		ZeroCopyEnabled:      true,
		// CORS defaults
		CORSAllowedOrigins:   []string{"*"},
		CORSAllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		CORSAllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		CORSExposedHeaders:   []string{"Content-Length"},
		CORSAllowCredentials: true,
		CORSMaxAge:          86400, // 24 hours
		EnableMetrics:        true,
		MetricsPath:          "/metrics",
	}
}

// HTTPOptimizer manages HTTP performance optimizations
type HTTPOptimizer struct {
	config           *HTTPOptimizationConfig
	fasthttpServer   *fasthttp.Server
	standardServer   *http.Server
	bufferPool       *sync.Pool
	metrics          *HTTPMetrics
	adapter          *FastHTTPAdapter
	
	// State
	isRunning        bool
	stopChan         chan struct{}
	mu               sync.RWMutex
}

// HTTPMetrics tracks HTTP performance metrics
type HTTPMetrics struct {
	RequestsTotal        int64         `json:"requests_total"`
	RequestDuration      time.Duration `json:"request_duration"`
	BytesRead           int64         `json:"bytes_read"`
	BytesWritten        int64         `json:"bytes_written"`
	ConnectionsActive   int64         `json:"connections_active"`
	ConnectionsTotal    int64         `json:"connections_total"`
	ErrorsTotal         int64         `json:"errors_total"`
	LastUpdated         time.Time     `json:"last_updated"`
	
	mu                  sync.RWMutex
}

// FastHTTPAdapter adapts fasthttp requests to standard http.Handler interface
type FastHTTPAdapter struct {
	handler     http.Handler
	bufferPool  *sync.Pool
	metrics     *HTTPMetrics
	config      *HTTPOptimizationConfig
}

// NewHTTPOptimizer creates a new HTTP optimizer with configuration
func NewHTTPOptimizer(config *HTTPOptimizationConfig) *HTTPOptimizer {
	if config == nil {
		config = DefaultHTTPOptimizationConfig()
	}

	optimizer := &HTTPOptimizer{
		config:   config,
		metrics:  &HTTPMetrics{},
		stopChan: make(chan struct{}),
	}

	// Initialize buffer pool for zero-copy operations
	optimizer.bufferPool = &sync.Pool{
		New: func() any {
			return make([]byte, 4096) // 4KB buffers
		},
	}

	// Create fasthttp server if enabled
	if config.EnableFastHTTP {
		optimizer.fasthttpServer = &fasthttp.Server{
			ReadTimeout:         config.ReadTimeout,
			WriteTimeout:        config.WriteTimeout,
			IdleTimeout:         config.IdleTimeout,
			MaxRequestBodySize:  config.MaxRequestBodySize,
			Concurrency:         config.MaxConnections,
			DisableKeepalive:    !config.EnableKeepAlive,
			GetOnly:            false,
			DisablePreParseMultipartForm: true,
		}
	}

	return optimizer
}

// SetHandler sets the HTTP handler for the optimizer
func (o *HTTPOptimizer) SetHandler(handler http.Handler) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.config.EnableFastHTTP {
		// Create adapter for fasthttp
		o.adapter = &FastHTTPAdapter{
			handler:    handler,
			bufferPool: o.bufferPool,
			metrics:    o.metrics,
			config:     o.config,
		}
		o.fasthttpServer.Handler = o.adapter.FastHTTPHandler
	} else {
		// Use standard HTTP server
		o.standardServer = &http.Server{
			Handler:      handler,
			ReadTimeout:  o.config.ReadTimeout,
			WriteTimeout: o.config.WriteTimeout,
			IdleTimeout:  o.config.IdleTimeout,
		}
	}
}

// ListenAndServe starts the optimized HTTP server
func (o *HTTPOptimizer) ListenAndServe(addr string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.isRunning {
		return fmt.Errorf("server is already running")
	}

	o.isRunning = true

	if o.config.EnableFastHTTP && o.fasthttpServer != nil {
		fmt.Printf("🚀 Starting FASTHTTP server on %s (high-performance mode)\n", addr)
		fmt.Printf("   - Max connections: %d\n", o.config.MaxConnections)
		fmt.Printf("   - Read timeout: %v\n", o.config.ReadTimeout)
		fmt.Printf("   - Write timeout: %v\n", o.config.WriteTimeout)
		fmt.Printf("   - Zero-copy enabled: %v\n", o.config.ZeroCopyEnabled)
		return o.fasthttpServer.ListenAndServe(addr)
	} else if o.standardServer != nil {
		fmt.Printf("📡 Starting standard HTTP server on %s (fallback mode)\n", addr)
		return o.standardServer.ListenAndServe()
	}

	return fmt.Errorf("no server configured")
}

// UpdateConfig updates the HTTP optimizer configuration
// Note: This should only be called before the server starts
func (o *HTTPOptimizer) UpdateConfig(newConfig *HTTPOptimizationConfig) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.isRunning {
		return fmt.Errorf("cannot update configuration while server is running")
	}

	// Update the configuration
	o.config = newConfig

	// Recreate fasthttp server with new configuration if needed
	if newConfig.EnableFastHTTP {
		o.fasthttpServer = &fasthttp.Server{
			ReadTimeout:         newConfig.ReadTimeout,
			WriteTimeout:        newConfig.WriteTimeout,
			IdleTimeout:         newConfig.IdleTimeout,
			MaxRequestBodySize:  newConfig.MaxRequestBodySize,
			Concurrency:         newConfig.MaxConnections,
			DisableKeepalive:    !newConfig.EnableKeepAlive,
			GetOnly:            false,
			DisablePreParseMultipartForm: true,
		}

		// If we already have an adapter, update its configuration
		if o.adapter != nil {
			o.fasthttpServer.Handler = o.adapter.FastHTTPHandler
		}
	} else {
		// Clear fasthttp server if disabled
		o.fasthttpServer = nil
	}

	return nil
}

// GetConfig returns the current configuration
func (o *HTTPOptimizer) GetConfig() *HTTPOptimizationConfig {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.config
}

// Shutdown gracefully shuts down the HTTP server
func (o *HTTPOptimizer) Shutdown() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.isRunning {
		return nil
	}

	close(o.stopChan)
	o.isRunning = false

	if o.config.EnableFastHTTP && o.fasthttpServer != nil {
		return o.fasthttpServer.Shutdown()
	} else if o.standardServer != nil {
		return o.standardServer.Close()
	}

	return nil
}

// GetMetrics returns current HTTP performance metrics
func (o *HTTPOptimizer) GetMetrics() map[string]any {
	o.metrics.mu.RLock()
	defer o.metrics.mu.RUnlock()

	return map[string]any{
		"fasthttp_enabled":     o.config.EnableFastHTTP,
		"requests_total":       o.metrics.RequestsTotal,
		"request_duration_ms":  o.metrics.RequestDuration.Milliseconds(),
		"bytes_read":          o.metrics.BytesRead,
		"bytes_written":       o.metrics.BytesWritten,
		"connections_active":  o.metrics.ConnectionsActive,
		"connections_total":   o.metrics.ConnectionsTotal,
		"errors_total":        o.metrics.ErrorsTotal,
		"last_updated":        o.metrics.LastUpdated,
		"config":              o.config,
	}
}

// FastHTTPHandler adapts fasthttp requests to standard http.Handler
func (a *FastHTTPAdapter) FastHTTPHandler(ctx *fasthttp.RequestCtx) {
	start := time.Now()

	// Add CORS headers
	a.handleCORS(ctx)

	// Handle preflight requests
	if string(ctx.Method()) == "OPTIONS" {
		ctx.SetStatusCode(fasthttp.StatusOK)
		return
	}

	// Update metrics
	a.updateMetrics(func(m *HTTPMetrics) {
		m.RequestsTotal++
		m.ConnectionsActive++
		m.LastUpdated = time.Now()
	})

	defer func() {
		duration := time.Since(start)
		a.updateMetrics(func(m *HTTPMetrics) {
			m.ConnectionsActive--
			m.RequestDuration = duration
			m.BytesRead += int64(len(ctx.Request.Body()))
			m.BytesWritten += int64(len(ctx.Response.Body()))
		})
	}()

	// Convert fasthttp request to standard http.Request
	req, err := a.convertRequest(ctx)
	if err != nil {
		a.updateMetrics(func(m *HTTPMetrics) { m.ErrorsTotal++ })
		ctx.Error("Failed to convert request", fasthttp.StatusInternalServerError)
		return
	}

	// Create response adapter
	respAdapter := &responseAdapter{
		ctx:        ctx,
		bufferPool: a.bufferPool,
	}

	// Call the handler
	a.handler.ServeHTTP(respAdapter, req)
}

// convertRequest converts fasthttp.RequestCtx to http.Request
func (a *FastHTTPAdapter) convertRequest(ctx *fasthttp.RequestCtx) (*http.Request, error) {
	// Create URL
	uri := ctx.URI()
	url := &url.URL{
		Scheme:   string(uri.Scheme()),
		Host:     string(uri.Host()),
		Path:     string(uri.Path()),
		RawQuery: string(uri.QueryString()),
	}

	// Create request
	req := &http.Request{
		Method:     string(ctx.Method()),
		URL:        url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       string(ctx.Host()),
		RemoteAddr: ctx.RemoteAddr().String(),
	}

	// Convert headers - use VisitAll as All() doesn't take parameters in this version
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	// Handle body for POST/PUT/PATCH requests
	if len(ctx.PostBody()) > 0 {
		req.Body = io.NopCloser(strings.NewReader(string(ctx.PostBody())))
		req.ContentLength = int64(len(ctx.PostBody()))
	}

	return req, nil
}

// responseAdapter adapts http.ResponseWriter to fasthttp response
type responseAdapter struct {
	ctx        *fasthttp.RequestCtx
	bufferPool *sync.Pool
	written    bool
}

// Header returns the response headers
func (r *responseAdapter) Header() http.Header {
	headers := make(http.Header)
	r.ctx.Response.Header.VisitAll(func(key, value []byte) {
		headers.Add(string(key), string(value))
	})
	return headers
}

// Write writes data to the response
func (r *responseAdapter) Write(data []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	
	r.ctx.Response.AppendBody(data)
	return len(data), nil
}

// WriteHeader sets the response status code
func (r *responseAdapter) WriteHeader(statusCode int) {
	if r.written {
		return
	}
	
	r.written = true
	r.ctx.Response.SetStatusCode(statusCode)
	
	// Copy headers from http.Header to fasthttp
	for key, values := range r.Header() {
		for _, value := range values {
			r.ctx.Response.Header.Add(key, value)
		}
	}
}

// handleCORS adds CORS headers to the response
func (a *FastHTTPAdapter) handleCORS(ctx *fasthttp.RequestCtx) {
	// Set CORS headers from configuration
	if len(a.config.CORSAllowedOrigins) > 0 {
		if len(a.config.CORSAllowedOrigins) == 1 && a.config.CORSAllowedOrigins[0] == "*" {
			ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		} else {
			// Check if the request origin is in the allowed list
			origin := string(ctx.Request.Header.Peek("Origin"))
			for _, allowedOrigin := range a.config.CORSAllowedOrigins {
				if origin == allowedOrigin {
					ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}
	}
	
	if len(a.config.CORSAllowedMethods) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(a.config.CORSAllowedMethods, ", "))
	}
	
	if len(a.config.CORSAllowedHeaders) > 0 {
		ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(a.config.CORSAllowedHeaders, ", "))
	}
	
	if len(a.config.CORSExposedHeaders) > 0 {
		ctx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(a.config.CORSExposedHeaders, ", "))
	}
	
	if a.config.CORSAllowCredentials {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}
	
	if a.config.CORSMaxAge > 0 {
		ctx.Response.Header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", a.config.CORSMaxAge))
	}
}

// updateMetrics safely updates HTTP metrics
func (a *FastHTTPAdapter) updateMetrics(fn func(*HTTPMetrics)) {
	a.metrics.mu.Lock()
	defer a.metrics.mu.Unlock()
	fn(a.metrics)
}