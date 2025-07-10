package logging_core

import (
	"context"
	"fmt"
	"sync"
	"time"

	go_core "base_lara_go_project/app/core/go_core"
	config_core "base_lara_go_project/app/core/laravel_core/config"
	"sync/atomic"
)

// LoggingMonitor provides comprehensive monitoring for the logging system
type LoggingMonitor[T any] struct {
	config       *config_core.ConfigFacade
	metrics      *LoggingSystemMetrics
	healthChecks map[string]HealthCheck
	alerts       map[string]LoggingAlert
	mu           sync.RWMutex
	monitoring   int32
	ctx          context.Context
	cancel       context.CancelFunc
}

// LoggingSystemMetrics tracks comprehensive system metrics
type LoggingSystemMetrics struct {
	// Performance metrics
	TotalEntriesLogged  int64
	TotalEntriesDropped int64
	TotalHandlerErrors  int64
	AverageLatency      int64
	PeakLatency         int64
	Throughput          int64 // entries per second

	// System metrics
	ActiveHandlers int64
	QueueSize      int64
	CacheHitRate   float64
	MemoryUsage    int64
	CPUUsage       float64

	// Error metrics
	ErrorRate         float64
	LastErrorTime     time.Time
	ConsecutiveErrors int64

	// Health metrics
	IsHealthy       bool
	LastHealthCheck time.Time
	Uptime          time.Duration

	mu sync.RWMutex
}

// HealthCheck defines a health check function
type HealthCheck func() (bool, error)

// LoggingAlert represents a monitoring alert
type LoggingAlert struct {
	ID         string
	Type       string
	Message    string
	Severity   string
	Timestamp  time.Time
	Resolved   bool
	ResolvedAt *time.Time
}

// NewLoggingMonitor creates a new logging monitor
func NewLoggingMonitor[T any](config *config_core.ConfigFacade) *LoggingMonitor[T] {
	ctx, cancel := context.WithCancel(context.Background())

	return &LoggingMonitor[T]{
		config:       config,
		metrics:      &LoggingSystemMetrics{},
		healthChecks: make(map[string]HealthCheck),
		alerts:       make(map[string]LoggingAlert),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the monitoring system
func (m *LoggingMonitor[T]) Start() error {
	if atomic.CompareAndSwapInt32(&m.monitoring, 0, 1) {
		go m.monitoringLoop()
		return nil
	}
	return fmt.Errorf("monitor already running")
}

// Stop stops the monitoring system
func (m *LoggingMonitor[T]) Stop() error {
	atomic.StoreInt32(&m.monitoring, 0)
	m.cancel()
	return nil
}

// AddHealthCheck adds a health check
func (m *LoggingMonitor[T]) AddHealthCheck(name string, check HealthCheck) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthChecks[name] = check
}

// RemoveHealthCheck removes a health check
func (m *LoggingMonitor[T]) RemoveHealthCheck(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.healthChecks, name)
}

// AddAlert adds an alert
func (m *LoggingMonitor[T]) AddAlert(alert LoggingAlert) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts[alert.ID] = alert
}

// ResolveAlert resolves an alert
func (m *LoggingMonitor[T]) ResolveAlert(alertID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if alert, exists := m.alerts[alertID]; exists {
		now := time.Now()
		alert.Resolved = true
		alert.ResolvedAt = &now
		m.alerts[alertID] = alert
	}
}

// GetMetrics returns system metrics
func (m *LoggingMonitor[T]) GetMetrics() *LoggingSystemMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

// GetHealthStatus returns overall health status
func (m *LoggingMonitor[T]) GetHealthStatus() (bool, map[string]bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]bool)
	overallHealthy := true

	for name, check := range m.healthChecks {
		healthy, err := check()
		status[name] = healthy
		if !healthy {
			overallHealthy = false
			// Create alert for failed health check
			m.createHealthAlert(name, err)
		}
	}

	return overallHealthy, status
}

// GetAlerts returns all alerts
func (m *LoggingMonitor[T]) GetAlerts() map[string]LoggingAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make(map[string]LoggingAlert)
	for id, alert := range m.alerts {
		alerts[id] = alert
	}
	return alerts
}

// monitoringLoop continuously monitors the logging system
func (m *LoggingMonitor[T]) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for atomic.LoadInt32(&m.monitoring) == 1 {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthChecks()
			m.updateMetrics()
			m.checkAlerts()
		}
	}
}

// performHealthChecks runs all health checks
func (m *LoggingMonitor[T]) performHealthChecks() {
	m.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range m.healthChecks {
		checks[name] = check
	}
	m.mu.RUnlock()

	for name, check := range checks {
		healthy, err := check()
		if !healthy {
			m.createHealthAlert(name, err)
		}
	}
}

// updateMetrics updates system metrics
func (m *LoggingMonitor[T]) updateMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update uptime
	m.metrics.Uptime = time.Since(m.metrics.LastHealthCheck)
	m.metrics.LastHealthCheck = time.Now()

	// Update health status
	healthy, _ := m.GetHealthStatus()
	m.metrics.IsHealthy = healthy

	// Calculate error rate
	if m.metrics.TotalEntriesLogged > 0 {
		m.metrics.ErrorRate = float64(m.metrics.TotalHandlerErrors) / float64(m.metrics.TotalEntriesLogged)
	}

	// Update throughput (entries per second)
	if m.metrics.Uptime.Seconds() > 0 {
		m.metrics.Throughput = int64(float64(m.metrics.TotalEntriesLogged) / m.metrics.Uptime.Seconds())
	}
}

// checkAlerts checks for alert conditions
func (m *LoggingMonitor[T]) checkAlerts() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check error rate alert
	if m.metrics.ErrorRate > 0.1 { // 10% error rate threshold
		m.createAlert("high_error_rate", "High error rate detected", "warning")
	}

	// Check latency alert
	if m.metrics.AverageLatency > 1000 { // 1 second threshold
		m.createAlert("high_latency", "High latency detected", "warning")
	}

	// Check memory usage alert
	if m.metrics.MemoryUsage > 100*1024*1024 { // 100MB threshold
		m.createAlert("high_memory", "High memory usage detected", "warning")
	}
}

// createHealthAlert creates a health check alert
func (m *LoggingMonitor[T]) createHealthAlert(checkName string, err error) {
	message := fmt.Sprintf("Health check '%s' failed", checkName)
	if err != nil {
		message += fmt.Sprintf(": %v", err)
	}

	m.createAlert("health_check_"+checkName, message, "critical")
}

// createAlert creates a new alert
func (m *LoggingMonitor[T]) createAlert(alertType, message, severity string) {
	alertID := fmt.Sprintf("%s_%d", alertType, time.Now().Unix())

	alert := LoggingAlert{
		ID:        alertID,
		Type:      alertType,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	m.alerts[alertID] = alert
}

// LoggingMetricsCollector collects metrics from various logging components
type LoggingMetricsCollector[T any] struct {
	monitor  *LoggingMonitor[T]
	handlers map[string]go_core.LogHandler[T]
	mu       sync.RWMutex
}

// NewLoggingMetricsCollector creates a new metrics collector
func NewLoggingMetricsCollector[T any](monitor *LoggingMonitor[T]) *LoggingMetricsCollector[T] {
	return &LoggingMetricsCollector[T]{
		monitor:  monitor,
		handlers: make(map[string]go_core.LogHandler[T]),
	}
}

// AddHandler adds a handler to collect metrics from
func (c *LoggingMetricsCollector[T]) AddHandler(name string, handler go_core.LogHandler[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[name] = handler
}

// RemoveHandler removes a handler from metrics collection
func (c *LoggingMetricsCollector[T]) RemoveHandler(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.handlers, name)
}

// CollectMetrics collects metrics from all handlers
func (c *LoggingMetricsCollector[T]) CollectMetrics() {
	c.mu.RLock()
	handlers := make(map[string]go_core.LogHandler[T])
	for name, handler := range c.handlers {
		handlers[name] = handler
	}
	c.mu.RUnlock()

	var totalEntries int64
	var totalErrors int64
	var totalLatency int64
	var handlerCount int64

	for _, handler := range handlers {
		// Get handler metrics if available
		if metricsHandler, ok := handler.(interface {
			GetMetrics() *go_core.LoggingMetrics
		}); ok {
			metrics := metricsHandler.GetMetrics()
			atomic.AddInt64(&totalEntries, metrics.EntriesLogged)
			atomic.AddInt64(&totalErrors, metrics.HandlerErrors)
			atomic.AddInt64(&totalLatency, metrics.AverageLatency)
		}
		atomic.AddInt64(&handlerCount, 1)
	}

	// Update monitor metrics
	c.monitor.mu.Lock()
	c.monitor.metrics.TotalEntriesLogged = totalEntries
	c.monitor.metrics.TotalHandlerErrors = totalErrors
	c.monitor.metrics.AverageLatency = totalLatency
	c.monitor.metrics.ActiveHandlers = handlerCount
	c.monitor.mu.Unlock()
}

// LoggingPerformanceTracker tracks performance metrics
type LoggingPerformanceTracker[T any] struct {
	metrics   *LoggingSystemMetrics
	startTime time.Time
	mu        sync.RWMutex
}

// NewLoggingPerformanceTracker creates a new performance tracker
func NewLoggingPerformanceTracker[T any]() *LoggingPerformanceTracker[T] {
	return &LoggingPerformanceTracker[T]{
		metrics:   &LoggingSystemMetrics{},
		startTime: time.Now(),
	}
}

// TrackOperation tracks a single operation
func (t *LoggingPerformanceTracker[T]) TrackOperation(operation string, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	atomic.AddInt64(&t.metrics.TotalEntriesLogged, 1)

	// Update latency metrics
	latency := int64(duration)
	atomic.StoreInt64(&t.metrics.AverageLatency, latency)

	if latency > t.metrics.PeakLatency {
		t.metrics.PeakLatency = latency
	}
}

// TrackError tracks an error
func (t *LoggingPerformanceTracker[T]) TrackError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	atomic.AddInt64(&t.metrics.TotalHandlerErrors, 1)
	atomic.AddInt64(&t.metrics.ConsecutiveErrors, 1)
	t.metrics.LastErrorTime = time.Now()
}

// TrackSuccess tracks a successful operation
func (t *LoggingPerformanceTracker[T]) TrackSuccess() {
	t.mu.Lock()
	defer t.mu.Unlock()

	atomic.StoreInt64(&t.metrics.ConsecutiveErrors, 0)
}

// GetMetrics returns performance metrics
func (t *LoggingPerformanceTracker[T]) GetMetrics() *LoggingSystemMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.metrics
}

// LoggingHealthChecker provides health checking capabilities
type LoggingHealthChecker[T any] struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
}

// NewLoggingHealthChecker creates a new health checker
func NewLoggingHealthChecker[T any]() *LoggingHealthChecker[T] {
	return &LoggingHealthChecker[T]{
		checks: make(map[string]HealthCheck),
	}
}

// AddCheck adds a health check
func (h *LoggingHealthChecker[T]) AddCheck(name string, check HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// RemoveCheck removes a health check
func (h *LoggingHealthChecker[T]) RemoveCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checks, name)
}

// RunChecks runs all health checks
func (h *LoggingHealthChecker[T]) RunChecks() map[string]bool {
	h.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range h.checks {
		checks[name] = check
	}
	h.mu.RUnlock()

	results := make(map[string]bool)
	for name, check := range checks {
		healthy, _ := check()
		results[name] = healthy
	}

	return results
}

// IsHealthy returns overall health status
func (h *LoggingHealthChecker[T]) IsHealthy() bool {
	results := h.RunChecks()
	for _, healthy := range results {
		if !healthy {
			return false
		}
	}
	return true
}
