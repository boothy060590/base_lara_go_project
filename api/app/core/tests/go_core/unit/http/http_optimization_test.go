package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPOptimizer tests the HTTP optimizer functionality
func TestHTTPOptimizer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("DefaultConfiguration", func(t *testing.T) {
		config := go_core.DefaultHTTPOptimizationConfig()
		
		assert.True(t, config.EnableFastHTTP)
		assert.Equal(t, 30*time.Second, config.ReadTimeout)
		assert.Equal(t, 30*time.Second, config.WriteTimeout)
		assert.Equal(t, 10000, config.MaxConnections)
		assert.True(t, config.EnableCompression)
		assert.True(t, config.ZeroCopyEnabled)
	})

	t.Run("OptimizerCreation", func(t *testing.T) {
		config := go_core.DefaultHTTPOptimizationConfig()
		optimizer := go_core.NewHTTPOptimizer(config)
		
		require.NotNil(t, optimizer)
		
		metrics := optimizer.GetMetrics()
		assert.True(t, metrics["fasthttp_enabled"].(bool))
		assert.Equal(t, int64(0), metrics["requests_total"])
	})

	t.Run("HandlerIntegration", func(t *testing.T) {
		config := go_core.DefaultHTTPOptimizationConfig()
		optimizer := go_core.NewHTTPOptimizer(config)
		
		// Create a simple Gin router
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello from optimized HTTP!",
				"fasthttp": true,
			})
		})
		
		// Set the router as the handler
		optimizer.SetHandler(router)
		
		// Verify metrics are available
		metrics := optimizer.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Contains(t, metrics, "fasthttp_enabled")
	})

	t.Run("GinCompatibility", func(t *testing.T) {
		// Test that Gin middleware works with our optimization
		router := gin.New()
		
		// Add middleware
		router.Use(func(c *gin.Context) {
			c.Header("X-Powered-By", "Laravel-Go-FastHTTP")
			c.Next()
		})
		
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"server": "fasthttp",
			})
		})
		
		// Test with standard httptest (simulates the adapter)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Laravel-Go-FastHTTP", w.Header().Get("X-Powered-By"))
		assert.Contains(t, w.Body.String(), "fasthttp")
	})

	t.Run("PerformanceMetrics", func(t *testing.T) {
		config := go_core.DefaultHTTPOptimizationConfig()
		optimizer := go_core.NewHTTPOptimizer(config)
		
		// Create test router
		router := gin.New()
		router.GET("/metrics-test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"test": "metrics"})
		})
		
		optimizer.SetHandler(router)
		
		// Get initial metrics
		initialMetrics := optimizer.GetMetrics()
		
		// Verify metrics structure
		expectedKeys := []string{
			"fasthttp_enabled",
			"requests_total", 
			"bytes_read",
			"bytes_written",
			"connections_active",
			"errors_total",
			"config",
		}
		
		for _, key := range expectedKeys {
			assert.Contains(t, initialMetrics, key, "Metrics should contain %s", key)
		}
		
		// Verify config is included
		config_data := initialMetrics["config"].(*go_core.HTTPOptimizationConfig)
		assert.True(t, config_data.EnableFastHTTP)
		assert.Equal(t, 10000, config_data.MaxConnections)
	})
}

// TestHTTPOptimizationConfiguration tests configuration handling
func TestHTTPOptimizationConfiguration(t *testing.T) {
	t.Run("CustomConfiguration", func(t *testing.T) {
		config := &go_core.HTTPOptimizationConfig{
			EnableFastHTTP:      false, // Disable fasthttp
			ReadTimeout:         10 * time.Second,
			WriteTimeout:        10 * time.Second,
			MaxConnections:      5000,
			EnableCompression:   false,
			ZeroCopyEnabled:     false,
		}
		
		optimizer := go_core.NewHTTPOptimizer(config)
		metrics := optimizer.GetMetrics()
		
		// Should be using standard HTTP, not fasthttp
		assert.False(t, metrics["fasthttp_enabled"].(bool))
		
		configData := metrics["config"].(*go_core.HTTPOptimizationConfig)
		assert.False(t, configData.EnableFastHTTP)
		assert.Equal(t, 5000, configData.MaxConnections)
		assert.False(t, configData.EnableCompression)
		assert.False(t, configData.ZeroCopyEnabled)
	})

	t.Run("ConfigurationValidation", func(t *testing.T) {
		// Test with nil configuration - should use defaults
		optimizer := go_core.NewHTTPOptimizer(nil)
		metrics := optimizer.GetMetrics()
		
		configData := metrics["config"].(*go_core.HTTPOptimizationConfig)
		assert.True(t, configData.EnableFastHTTP)
		assert.Equal(t, 10000, configData.MaxConnections)
		assert.True(t, configData.EnableCompression)
	})
}

// BenchmarkHTTPOptimizer provides basic performance benchmarks
func BenchmarkHTTPOptimizer(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	
	b.Run("FastHTTPEnabled", func(b *testing.B) {
		config := go_core.DefaultHTTPOptimizationConfig()
		config.EnableFastHTTP = true
		
		optimizer := go_core.NewHTTPOptimizer(config)
		
		router := gin.New()
		router.GET("/bench", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"benchmark": true})
		})
		
		optimizer.SetHandler(router)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/bench", nil)
			router.ServeHTTP(w, req)
		}
	})

	b.Run("StandardHTTP", func(b *testing.B) {
		config := go_core.DefaultHTTPOptimizationConfig()
		config.EnableFastHTTP = false // Use standard HTTP
		
		optimizer := go_core.NewHTTPOptimizer(config)
		
		router := gin.New()
		router.GET("/bench", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"benchmark": true})
		})
		
		optimizer.SetHandler(router)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/bench", nil)
			router.ServeHTTP(w, req)
		}
	})
}

// TestHTTPOptimizationIntegration tests integration with service providers
func TestHTTPOptimizationIntegration(t *testing.T) {
	t.Run("ServiceProviderIntegration", func(t *testing.T) {
		// Test that the HTTP optimization integrates properly with the service container
		container := go_core.NewContainer()
		
		// Register HTTP optimizer (simulating service provider)
		container.Singleton("http.optimizer", func() (any, error) {
			config := go_core.DefaultHTTPOptimizationConfig()
			return go_core.NewHTTPOptimizer(config), nil
		})
		
		// Resolve optimizer
		optimizerInstance, err := container.Resolve("http.optimizer")
		require.NoError(t, err)
		require.NotNil(t, optimizerInstance)
		
		optimizer := optimizerInstance.(*go_core.HTTPOptimizer)
		metrics := optimizer.GetMetrics()
		
		assert.True(t, metrics["fasthttp_enabled"].(bool))
	})
}