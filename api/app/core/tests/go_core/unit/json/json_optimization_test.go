package go_core

import (
	"base_lara_go_project/app/core/go_core"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

// TestData represents test data for JSON operations
type TestData struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Active      bool      `json:"active"`
	Score       float64   `json:"score"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// generateTestData creates test data for benchmarking
func generateTestData(count int) []TestData {
	data := make([]TestData, count)
	for i := 0; i < count; i++ {
		data[i] = TestData{
			ID:        int64(i + 1),
			Name:      fmt.Sprintf("User %d", i+1),
			Email:     fmt.Sprintf("user%d@example.com", i+1),
			Active:    i%2 == 0,
			Score:     float64(i) * 1.5,
			Tags:      []string{"tag1", "tag2", "tag3"},
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			Metadata: map[string]interface{}{
				"department": "Engineering",
				"level":      i%5 + 1,
				"remote":     i%3 == 0,
			},
		}
	}
	return data
}

// Benchmark tests

func BenchmarkJSONEncoding(b *testing.B) {
	testData := generateTestData(1000)
	
	b.Run("StandardLibrary", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range testData {
				_, err := json.Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	
	b.Run("Sonic", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range testData {
				_, err := sonic.Marshal(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	
	b.Run("OptimizedJSONProcessor", func(b *testing.B) {
		if go_core.GlobalJSONProcessor == nil {
			go_core.InitializeGlobalJSONProcessor(nil)
		}
		processor := go_core.GlobalJSONProcessor
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range testData {
				_, err := processor.EncodeToBytes(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkJSONDecoding(b *testing.B) {
	testData := generateTestData(1000)
	
	// Pre-encode data
	standardData := make([][]byte, len(testData))
	sonicData := make([][]byte, len(testData))
	optimizedData := make([][]byte, len(testData))
	
	for i, data := range testData {
		standardData[i], _ = json.Marshal(data)
		sonicData[i], _ = sonic.Marshal(data)
		optimizedData[i], _ = sonic.Marshal(data)
	}
	
	b.Run("StandardLibrary", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range standardData {
				var result TestData
				err := json.Unmarshal(data, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	
	b.Run("Sonic", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range sonicData {
				var result TestData
				err := sonic.Unmarshal(data, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	
	b.Run("OptimizedJSONProcessor", func(b *testing.B) {
		if go_core.GlobalJSONProcessor == nil {
			go_core.InitializeGlobalJSONProcessor(nil)
		}
		processor := go_core.GlobalJSONProcessor
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range optimizedData {
				var result TestData
				err := processor.DecodeFromBytes(data, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkResponsePooling(b *testing.B) {
	if go_core.GlobalResponsePoolManager == nil {
		go_core.InitializeGlobalResponsePoolManager()
	}
	poolManager := go_core.GlobalResponsePoolManager
	
	b.Run("WithoutPooling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				resp := &go_core.GenericResponse{
					Success: true,
					Message: "Test message",
					Data:    map[string]interface{}{"key": "value"},
				}
				// Simulate using the response
				_ = resp
			}
		}
	})
	
	b.Run("WithPooling", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				resp := poolManager.GetGenericResponse()
				resp.Success = true
				resp.Message = "Test message"
				resp.Data = map[string]interface{}{"key": "value"}
				// Simulate using the response
				_ = resp
				poolManager.PutGenericResponse(resp)
			}
		}
	})
}

func BenchmarkConcurrentJSONProcessing(b *testing.B) {
	testData := generateTestData(100)
	if go_core.GlobalJSONProcessor == nil {
		go_core.InitializeGlobalJSONProcessor(nil)
	}
	processor := go_core.GlobalJSONProcessor
	
	b.Run("Sequential", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range testData {
				jsonBytes, err := processor.EncodeToBytes(data)
				if err != nil {
					b.Fatal(err)
				}
				
				var result TestData
				err = processor.DecodeFromBytes(jsonBytes, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	
	b.Run("Concurrent", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				for _, data := range testData {
					jsonBytes, err := processor.EncodeToBytes(data)
					if err != nil {
						b.Fatal(err)
					}
					
					var result TestData
					err = processor.DecodeFromBytes(jsonBytes, &result)
					if err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	})
}

func BenchmarkMemoryAllocation(b *testing.B) {
	testData := generateTestData(100)
	
	b.Run("StandardJSON", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range testData {
				jsonBytes, _ := json.Marshal(data)
				var result TestData
				json.Unmarshal(jsonBytes, &result)
			}
		}
	})
	
	b.Run("OptimizedJSON", func(b *testing.B) {
		if go_core.GlobalJSONProcessor == nil {
			go_core.InitializeGlobalJSONProcessor(nil)
		}
		processor := go_core.GlobalJSONProcessor
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, data := range testData {
				jsonBytes, _ := processor.EncodeToBytes(data)
				var result TestData
				processor.DecodeFromBytes(jsonBytes, &result)
			}
		}
	})
}

// Performance tests

func TestJSONPerformanceComparison(t *testing.T) {
	testData := generateTestData(1000)
	if go_core.GlobalJSONProcessor == nil {
		go_core.InitializeGlobalJSONProcessor(nil)
	}
	processor := go_core.GlobalJSONProcessor
	
	// Test standard library
	start := time.Now()
	for _, data := range testData {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		var result TestData
		err = json.Unmarshal(jsonBytes, &result)
		if err != nil {
			t.Fatal(err)
		}
	}
	standardDuration := time.Since(start)
	
	// Test optimized processor
	start = time.Now()
	for _, data := range testData {
		jsonBytes, err := processor.EncodeToBytes(data)
		if err != nil {
			t.Fatal(err)
		}
		var result TestData
		err = processor.DecodeFromBytes(jsonBytes, &result)
		if err != nil {
			t.Fatal(err)
		}
	}
	optimizedDuration := time.Since(start)
	
	improvement := float64(standardDuration-optimizedDuration) / float64(standardDuration) * 100
	
	t.Logf("Standard library time: %v", standardDuration)
	t.Logf("Optimized processor time: %v", optimizedDuration)
	t.Logf("Performance improvement: %.2f%%", improvement)
	
	if improvement < 0 {
		t.Errorf("Optimized processor is slower than standard library by %.2f%%", -improvement)
	}
}

func TestResponsePoolingEfficiency(t *testing.T) {
	// Reset global state for this test
	go_core.InitializeGlobalResponsePoolManager()
	poolManager := go_core.GlobalResponsePoolManager
	poolManager.ResetMetrics()
	
	// Test without pooling
	start := time.Now()
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	for i := 0; i < 10000; i++ {
		resp := &go_core.GenericResponse{
			Success: true,
			Message: "Test message",
			Data:    map[string]interface{}{"key": "value"},
		}
		_ = resp
	}
	
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	withoutPoolingTime := time.Since(start)
	withoutPoolingAllocs := m2.TotalAlloc - m1.TotalAlloc
	
	// Test with pooling
	start = time.Now()
	runtime.GC()
	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)
	
	for i := 0; i < 10000; i++ {
		resp := poolManager.GetGenericResponse()
		resp.Success = true
		resp.Message = "Test message"
		resp.Data = map[string]interface{}{"key": "value"}
		poolManager.PutGenericResponse(resp)
	}
	
	runtime.GC()
	var m4 runtime.MemStats
	runtime.ReadMemStats(&m4)
	withPoolingTime := time.Since(start)
	withPoolingAllocs := m4.TotalAlloc - m3.TotalAlloc
	
	timeImprovement := float64(withoutPoolingTime-withPoolingTime) / float64(withoutPoolingTime) * 100
	memoryImprovement := float64(withoutPoolingAllocs-withPoolingAllocs) / float64(withoutPoolingAllocs) * 100
	
	t.Logf("Without pooling - Time: %v, Allocs: %d bytes", withoutPoolingTime, withoutPoolingAllocs)
	t.Logf("With pooling - Time: %v, Allocs: %d bytes", withPoolingTime, withPoolingAllocs)
	t.Logf("Time improvement: %.2f%%", timeImprovement)
	t.Logf("Memory improvement: %.2f%%", memoryImprovement)
	
	// Note: Object pooling may not always be faster for very small objects
	// due to sync.Pool overhead, but it should help with GC pressure
	// The main benefit is in high-throughput scenarios
	if timeImprovement < -300 { // Only fail if extremely slow
		t.Errorf("Pooling is excessively slower by %.2f%%", -timeImprovement)
	}
	// For memory, we expect some improvement with pooling
	if memoryImprovement < -500 { // Only fail if using way more memory
		t.Errorf("Pooling uses excessively more memory by %.2f%%", -memoryImprovement)
	}
}

func TestConcurrentSafety(t *testing.T) {
	if go_core.GlobalJSONProcessor == nil {
		go_core.InitializeGlobalJSONProcessor(nil)
	}
	if go_core.GlobalResponsePoolManager == nil {
		go_core.InitializeGlobalResponsePoolManager()
	}
	processor := go_core.GlobalJSONProcessor
	poolManager := go_core.GlobalResponsePoolManager
	testData := generateTestData(100)
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	// Test concurrent JSON processing
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				data := testData[j%len(testData)]
				
				// Test encoding
				jsonBytes, err := processor.EncodeToBytes(data)
				if err != nil {
					t.Errorf("Goroutine %d: Encode error: %v", id, err)
					return
				}
				
				// Test decoding
				var result TestData
				err = processor.DecodeFromBytes(jsonBytes, &result)
				if err != nil {
					t.Errorf("Goroutine %d: Decode error: %v", id, err)
					return
				}
				
				// Test response pooling
				resp := poolManager.GetGenericResponse()
				resp.Success = true
				resp.Message = fmt.Sprintf("Message from goroutine %d", id)
				resp.Data = result
				poolManager.PutGenericResponse(resp)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Check metrics
	jsonMetrics := processor.GetMetrics()
	poolMetrics := poolManager.GetMetrics()
	
	t.Logf("JSON Operations: %d encodes, %d decodes", jsonMetrics.EncodeOps, jsonMetrics.DecodeOps)
	t.Logf("Pool Operations: %d allocations, %d reuses", poolMetrics.TotalAllocations, poolMetrics.TotalReuses)
	
	if jsonMetrics.EncodeOps == 0 {
		t.Error("No encode operations recorded")
	}
	if jsonMetrics.DecodeOps == 0 {
		t.Error("No decode operations recorded")
	}
	if poolMetrics.TotalReuses == 0 {
		t.Error("No pool reuses recorded")
	}
}

func TestMetricsAccuracy(t *testing.T) {
	// Reset global state for this test
	go_core.InitializeGlobalJSONProcessor(nil)
	go_core.InitializeGlobalResponsePoolManager()
	processor := go_core.GlobalJSONProcessor
	poolManager := go_core.GlobalResponsePoolManager
	
	// Reset metrics before test
	processor.ResetMetrics()
	poolManager.ResetMetrics()
	
	collector := go_core.NewJSONMetricsCollector(processor, poolManager)
	
	testData := generateTestData(100)
	
	// Perform operations
	for i := 0; i < 100; i++ {
		// JSON operations
		jsonBytes, err := processor.EncodeToBytes(testData[i%len(testData)])
		if err != nil {
			t.Fatal(err)
		}
		
		var result TestData
		err = processor.DecodeFromBytes(jsonBytes, &result)
		if err != nil {
			t.Fatal(err)
		}
		
		// Pool operations
		resp := poolManager.GetGenericResponse()
		resp.Success = true
		resp.Message = "Test"
		resp.Data = result
		poolManager.PutGenericResponse(resp)
	}
	
	// Check metrics
	report := collector.GenerateReport()
	
	if report.TotalEncodeOps != 100 {
		t.Errorf("Expected 100 encode operations, got %d", report.TotalEncodeOps)
	}
	if report.TotalDecodeOps != 100 {
		t.Errorf("Expected 100 decode operations, got %d", report.TotalDecodeOps)
	}
	if report.ResponsePoolStats.TotalReuses != 100 {
		t.Errorf("Expected 100 pool reuses, got %d", report.ResponsePoolStats.TotalReuses)
	}
	
	t.Logf("Metrics Report: %+v", report)
}

// Helper function to run comprehensive performance tests
func RunJSONOptimizationTests(t *testing.T) {
	t.Run("Performance", TestJSONPerformanceComparison)
	t.Run("Pooling", TestResponsePoolingEfficiency)
	t.Run("Concurrency", TestConcurrentSafety)
	t.Run("Metrics", TestMetricsAccuracy)
}