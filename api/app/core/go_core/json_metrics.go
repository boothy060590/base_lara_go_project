package go_core

import (
	"sync"
	"time"
)

// JSONMetricsCollector collects and aggregates JSON processing metrics
type JSONMetricsCollector struct {
	processor           *JSONProcessor
	responsePoolManager *ResponsePoolManager
	startTime          time.Time
	mu                 sync.RWMutex
}

// NewJSONMetricsCollector creates a new JSON metrics collector
func NewJSONMetricsCollector(processor *JSONProcessor, responsePoolManager *ResponsePoolManager) *JSONMetricsCollector {
	return &JSONMetricsCollector{
		processor:           processor,
		responsePoolManager: responsePoolManager,
		startTime:          time.Now(),
	}
}

// JSONPerformanceReport contains comprehensive JSON performance metrics
type JSONPerformanceReport struct {
	// Timing metrics
	StartTime       time.Time `json:"start_time"`
	ReportTime      time.Time `json:"report_time"`
	UptimeSeconds   int64     `json:"uptime_seconds"`
	
	// JSON processing metrics
	TotalEncodeOps      int64         `json:"total_encode_ops"`
	TotalDecodeOps      int64         `json:"total_decode_ops"`
	TotalBytesProcessed int64         `json:"total_bytes_processed"`
	AvgEncodeTime       time.Duration `json:"avg_encode_time"`
	AvgDecodeTime       time.Duration `json:"avg_decode_time"`
	
	// Buffer pool metrics
	BufferPoolHits      int64 `json:"buffer_pool_hits"`
	BufferPoolMisses    int64 `json:"buffer_pool_misses"`
	BufferPoolHitRate   float64 `json:"buffer_pool_hit_rate"`
	
	// Response pool metrics
	ResponsePoolStats   ResponsePoolStats `json:"response_pool_stats"`
	
	// Performance calculations
	EncodesPerSecond    float64 `json:"encodes_per_second"`
	DecodesPerSecond    float64 `json:"decodes_per_second"`
	BytesPerSecond      float64 `json:"bytes_per_second"`
	MemoryEfficiency    float64 `json:"memory_efficiency"`
}

// ResponsePoolStats contains response pool statistics
type ResponsePoolStats struct {
	LoginPoolHits      int64   `json:"login_pool_hits"`
	HealthPoolHits     int64   `json:"health_pool_hits"`
	ErrorPoolHits      int64   `json:"error_pool_hits"`
	ValidationPoolHits int64   `json:"validation_pool_hits"`
	UserPoolHits       int64   `json:"user_pool_hits"`
	GenericPoolHits    int64   `json:"generic_pool_hits"`
	TotalAllocations   int64   `json:"total_allocations"`
	TotalReuses        int64   `json:"total_reuses"`
	ReuseRate          float64 `json:"reuse_rate"`
}

// GenerateReport generates a comprehensive JSON performance report
func (jmc *JSONMetricsCollector) GenerateReport() JSONPerformanceReport {
	jmc.mu.RLock()
	defer jmc.mu.RUnlock()
	
	now := time.Now()
	uptime := now.Sub(jmc.startTime)
	
	// Get JSON processor metrics
	jsonMetrics := jmc.processor.GetMetrics()
	
	// Get response pool metrics
	poolMetrics := jmc.responsePoolManager.GetMetrics()
	
	// Calculate performance metrics
	var encodesPerSecond, decodesPerSecond, bytesPerSecond float64
	if uptime.Seconds() > 0 {
		encodesPerSecond = float64(jsonMetrics.EncodeOps) / uptime.Seconds()
		decodesPerSecond = float64(jsonMetrics.DecodeOps) / uptime.Seconds()
		bytesPerSecond = float64(jsonMetrics.BytesProcessed) / uptime.Seconds()
	}
	
	// Calculate buffer pool hit rate
	var bufferPoolHitRate float64
	totalBufferOps := jsonMetrics.BufferPoolHits + jsonMetrics.BufferPoolMisses
	if totalBufferOps > 0 {
		bufferPoolHitRate = float64(jsonMetrics.BufferPoolHits) / float64(totalBufferOps) * 100
	}
	
	// Calculate response pool reuse rate
	var reuseRate float64
	if poolMetrics.TotalAllocations > 0 {
		reuseRate = float64(poolMetrics.TotalReuses) / float64(poolMetrics.TotalAllocations) * 100
	}
	
	// Calculate memory efficiency (higher is better)
	var memoryEfficiency float64
	if poolMetrics.TotalAllocations > 0 {
		memoryEfficiency = (float64(poolMetrics.TotalReuses) / float64(poolMetrics.TotalAllocations)) * bufferPoolHitRate
	}
	
	return JSONPerformanceReport{
		StartTime:       jmc.startTime,
		ReportTime:      now,
		UptimeSeconds:   int64(uptime.Seconds()),
		
		TotalEncodeOps:      jsonMetrics.EncodeOps,
		TotalDecodeOps:      jsonMetrics.DecodeOps,
		TotalBytesProcessed: jsonMetrics.BytesProcessed,
		AvgEncodeTime:       jsonMetrics.AvgEncodeTime,
		AvgDecodeTime:       jsonMetrics.AvgDecodeTime,
		
		BufferPoolHits:      jsonMetrics.BufferPoolHits,
		BufferPoolMisses:    jsonMetrics.BufferPoolMisses,
		BufferPoolHitRate:   bufferPoolHitRate,
		
		ResponsePoolStats: ResponsePoolStats{
			LoginPoolHits:      poolMetrics.LoginPoolHits,
			HealthPoolHits:     poolMetrics.HealthPoolHits,
			ErrorPoolHits:      poolMetrics.ErrorPoolHits,
			ValidationPoolHits: poolMetrics.ValidationPoolHits,
			UserPoolHits:       poolMetrics.UserPoolHits,
			GenericPoolHits:    poolMetrics.GenericPoolHits,
			TotalAllocations:   poolMetrics.TotalAllocations,
			TotalReuses:        poolMetrics.TotalReuses,
			ReuseRate:          reuseRate,
		},
		
		EncodesPerSecond:    encodesPerSecond,
		DecodesPerSecond:    decodesPerSecond,
		BytesPerSecond:      bytesPerSecond,
		MemoryEfficiency:    memoryEfficiency,
	}
}

// GetCurrentMetrics returns current metrics snapshot
func (jmc *JSONMetricsCollector) GetCurrentMetrics() (JSONMetrics, ResponsePoolMetrics) {
	jmc.mu.RLock()
	defer jmc.mu.RUnlock()
	
	return jmc.processor.GetMetrics(), jmc.responsePoolManager.GetMetrics()
}

// ResetMetrics resets all metrics
func (jmc *JSONMetricsCollector) ResetMetrics() {
	jmc.mu.Lock()
	defer jmc.mu.Unlock()
	
	jmc.processor.ResetMetrics()
	jmc.responsePoolManager.ResetMetrics()
	jmc.startTime = time.Now()
}

// GetPerformanceSummary returns a human-readable performance summary
func (jmc *JSONMetricsCollector) GetPerformanceSummary() map[string]interface{} {
	report := jmc.GenerateReport()
	
	return map[string]interface{}{
		"uptime_hours":          float64(report.UptimeSeconds) / 3600,
		"total_operations":      report.TotalEncodeOps + report.TotalDecodeOps,
		"operations_per_second": report.EncodesPerSecond + report.DecodesPerSecond,
		"mb_processed":          float64(report.TotalBytesProcessed) / (1024 * 1024),
		"buffer_efficiency":     report.BufferPoolHitRate,
		"pool_efficiency":       report.ResponsePoolStats.ReuseRate,
		"memory_efficiency":     report.MemoryEfficiency,
		"avg_encode_time_ms":    float64(report.AvgEncodeTime.Nanoseconds()) / 1e6,
		"avg_decode_time_ms":    float64(report.AvgDecodeTime.Nanoseconds()) / 1e6,
	}
}

// JSONMetricsMonitor provides real-time monitoring of JSON performance
type JSONMetricsMonitor struct {
	collector *JSONMetricsCollector
	interval  time.Duration
	stopChan  chan bool
	running   bool
	mu        sync.RWMutex
}

// NewJSONMetricsMonitor creates a new JSON metrics monitor
func NewJSONMetricsMonitor(collector *JSONMetricsCollector, interval time.Duration) *JSONMetricsMonitor {
	return &JSONMetricsMonitor{
		collector: collector,
		interval:  interval,
		stopChan:  make(chan bool, 1),
	}
}

// Start starts the metrics monitor
func (jmm *JSONMetricsMonitor) Start() {
	jmm.mu.Lock()
	defer jmm.mu.Unlock()
	
	if jmm.running {
		return
	}
	
	jmm.running = true
	go jmm.monitorLoop()
}

// Stop stops the metrics monitor
func (jmm *JSONMetricsMonitor) Stop() {
	jmm.mu.Lock()
	defer jmm.mu.Unlock()
	
	if !jmm.running {
		return
	}
	
	jmm.running = false
	jmm.stopChan <- true
}

// monitorLoop runs the monitoring loop
func (jmm *JSONMetricsMonitor) monitorLoop() {
	ticker := time.NewTicker(jmm.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Log performance metrics periodically
			summary := jmm.collector.GetPerformanceSummary()
			// TODO: Integrate with your logging system
			_ = summary
			
		case <-jmm.stopChan:
			return
		}
	}
}

// IsRunning returns whether the monitor is running
func (jmm *JSONMetricsMonitor) IsRunning() bool {
	jmm.mu.RLock()
	defer jmm.mu.RUnlock()
	return jmm.running
}

// Global metrics collector and monitor
var (
	GlobalJSONMetricsCollector *JSONMetricsCollector
	GlobalJSONMetricsMonitor   *JSONMetricsMonitor
)

// InitializeGlobalJSONMetrics initializes global JSON metrics
func InitializeGlobalJSONMetrics() {
	if GlobalJSONProcessor == nil {
		InitializeGlobalJSONProcessor(nil)
	}
	if GlobalResponsePoolManager == nil {
		InitializeGlobalResponsePoolManager()
	}
	
	GlobalJSONMetricsCollector = NewJSONMetricsCollector(GlobalJSONProcessor, GlobalResponsePoolManager)
	GlobalJSONMetricsMonitor = NewJSONMetricsMonitor(GlobalJSONMetricsCollector, 5*time.Minute)
}

// GetGlobalJSONMetrics returns global JSON metrics
func GetGlobalJSONMetrics() JSONPerformanceReport {
	if GlobalJSONMetricsCollector == nil {
		InitializeGlobalJSONMetrics()
	}
	return GlobalJSONMetricsCollector.GenerateReport()
}

// GetGlobalJSONMetricsSummary returns global JSON metrics summary
func GetGlobalJSONMetricsSummary() map[string]interface{} {
	if GlobalJSONMetricsCollector == nil {
		InitializeGlobalJSONMetrics()
	}
	return GlobalJSONMetricsCollector.GetPerformanceSummary()
}

// StartGlobalJSONMetricsMonitor starts global JSON metrics monitoring
func StartGlobalJSONMetricsMonitor() {
	if GlobalJSONMetricsMonitor == nil {
		InitializeGlobalJSONMetrics()
	}
	GlobalJSONMetricsMonitor.Start()
}

// StopGlobalJSONMetricsMonitor stops global JSON metrics monitoring
func StopGlobalJSONMetricsMonitor() {
	if GlobalJSONMetricsMonitor != nil {
		GlobalJSONMetricsMonitor.Stop()
	}
}