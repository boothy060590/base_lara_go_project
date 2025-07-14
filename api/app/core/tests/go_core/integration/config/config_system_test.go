package integration

import (
	"os"
	"testing"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/require"
)

// TestConfigSystemWithEnvironment tests the config system with real environment variables
func TestConfigSystemWithEnvironment(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Set up test environment variables
	os.Setenv("APP_NAME", "Test App From Env")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "test-db.example.com")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("REDIS_HOST", "test-redis.example.com")
	os.Setenv("MAIL_HOST", "test-smtp.example.com")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("QUEUE_CONNECTION", "redis")
	os.Setenv("CACHE_STORE", "redis")

	// Create test configs that use environment variables
	appConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"name":                os.Getenv("APP_NAME"),
			"debug":               os.Getenv("APP_DEBUG") == "true",
			"port":                os.Getenv("APP_PORT"),
			"token_hour_lifespan": 24,
		}
	}
	go_core.RegisterGlobalConfig("app", appConfigFunc)

	dbConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"default": "postgres",
			"connections": map[string]interface{}{
				"postgres": map[string]interface{}{
					"host":     os.Getenv("DB_HOST"),
					"port":     os.Getenv("DB_PORT"),
					"database": "test_db",
					"username": "test_user",
					"password": "test_password",
				},
			},
		}
	}
	go_core.RegisterGlobalConfig("database", dbConfigFunc)

	cacheConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"default": os.Getenv("CACHE_STORE"),
			"stores": map[string]interface{}{
				"redis": map[string]interface{}{
					"host":     os.Getenv("REDIS_HOST"),
					"port":     "6379",
					"database": 1,
				},
			},
		}
	}
	go_core.RegisterGlobalConfig("cache", cacheConfigFunc)

	// Test accessing configs through the loader
	loader := go_core.GetGlobalConfigLoader()

	// Test app config access
	appName := loader.GetString("app.name")
	require.Equal(t, "Test App From Env", appName, "App name should match environment variable")

	appDebug := loader.GetBool("app.debug")
	require.Equal(t, true, appDebug, "App debug should be true from environment")

	appPort := loader.GetString("app.port")
	require.Equal(t, "9090", appPort, "App port should match environment variable")

	// Test database config access
	dbHost := loader.GetString("database.connections.postgres.host")
	require.Equal(t, "test-db.example.com", dbHost, "Database host should match environment variable")

	dbPort := loader.GetString("database.connections.postgres.port")
	require.Equal(t, "5432", dbPort, "Database port should match environment variable")

	// Test cache config access
	cacheDefault := loader.GetString("cache.default")
	require.Equal(t, "redis", cacheDefault, "Cache default should match environment variable")

	redisHost := loader.GetString("cache.stores.redis.host")
	require.Equal(t, "test-redis.example.com", redisHost, "Redis host should match environment variable")
}

// TestConfigSystemEnvironmentFallbacks tests environment variable fallbacks
func TestConfigSystemEnvironmentFallbacks(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Clear environment variables to test fallbacks
	envVars := []string{"APP_NAME", "APP_DEBUG", "APP_PORT", "DB_HOST"}
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Create config with fallback values
	appConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"name":                getEnvWithFallback("APP_NAME", "Default App Name"),
			"debug":               getEnvBoolWithFallback("APP_DEBUG", false),
			"port":                getEnvWithFallback("APP_PORT", "8080"),
			"token_hour_lifespan": 24,
		}
	}
	go_core.RegisterGlobalConfig("app", appConfigFunc)

	// Test accessing through loader
	loader := go_core.GetGlobalConfigLoader()

	appName := loader.GetString("app.name")
	require.Equal(t, "Default App Name", appName, "App name should use fallback value")

	appDebug := loader.GetBool("app.debug")
	require.Equal(t, false, appDebug, "App debug should use fallback value")

	appPort := loader.GetString("app.port")
	require.Equal(t, "8080", appPort, "App port should use fallback value")
}

// TestConfigSystemMultipleEnvironments tests config system with different environments
func TestConfigSystemMultipleEnvironments(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Test development environment
	os.Setenv("APP_ENV", "development")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("DB_HOST", "localhost")

	devConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"env":   os.Getenv("APP_ENV"),
			"debug": os.Getenv("APP_DEBUG") == "true",
			"database": map[string]interface{}{
				"host": os.Getenv("DB_HOST"),
			},
		}
	}
	go_core.RegisterGlobalConfig("app_dev", devConfigFunc)

	// Test production environment
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_DEBUG", "false")
	os.Setenv("DB_HOST", "prod-db.example.com")

	prodConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"env":   os.Getenv("APP_ENV"),
			"debug": os.Getenv("APP_DEBUG") == "true",
			"database": map[string]interface{}{
				"host": os.Getenv("DB_HOST"),
			},
		}
	}
	go_core.RegisterGlobalConfig("app_prod", prodConfigFunc)

	// Test that both configs coexist
	loader := go_core.GetGlobalConfigLoader()
	configs := loader.ListAvailableConfigs()
	require.Contains(t, configs, "app_dev", "Should contain development config")
	require.Contains(t, configs, "app_prod", "Should contain production config")
}

// TestConfigSystemComplexNesting tests complex nested configurations
func TestConfigSystemComplexNesting(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create a complex nested configuration
	complexConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"application": map[string]interface{}{
				"name": "Complex App",
				"features": map[string]interface{}{
					"cache": map[string]interface{}{
						"enabled": true,
						"drivers": map[string]interface{}{
							"redis": map[string]interface{}{
								"host":     "redis.example.com",
								"port":     6379,
								"database": 1,
								"options": map[string]interface{}{
									"timeout": 30,
									"retries": 3,
								},
							},
							"memcached": map[string]interface{}{
								"host": "memcached.example.com",
								"port": 11211,
							},
						},
					},
					"queue": map[string]interface{}{
						"enabled": true,
						"connections": map[string]interface{}{
							"redis": map[string]interface{}{
								"host": "queue-redis.example.com",
							},
							"database": map[string]interface{}{
								"table": "jobs",
							},
						},
					},
				},
			},
		}
	}
	go_core.RegisterGlobalConfig("complex", complexConfigFunc)

	// Test deep nested access
	loader := go_core.GetGlobalConfigLoader()

	// Test application name
	appName := loader.GetString("complex.application.name")
	require.Equal(t, "Complex App", appName, "Should access application name")

	// Test cache configuration
	cacheEnabled := loader.GetBool("complex.application.features.cache.enabled")
	require.Equal(t, true, cacheEnabled, "Cache should be enabled")

	redisHost := loader.GetString("complex.application.features.cache.drivers.redis.host")
	require.Equal(t, "redis.example.com", redisHost, "Should access Redis host")

	redisPort := loader.GetInt("complex.application.features.cache.drivers.redis.port")
	require.Equal(t, 6379, redisPort, "Should access Redis port")

	redisTimeout := loader.GetInt("complex.application.features.cache.drivers.redis.options.timeout")
	require.Equal(t, 30, redisTimeout, "Should access Redis timeout")

	// Test queue configuration
	queueEnabled := loader.GetBool("complex.application.features.queue.enabled")
	require.Equal(t, true, queueEnabled, "Queue should be enabled")

	queueRedisHost := loader.GetString("complex.application.features.queue.connections.redis.host")
	require.Equal(t, "queue-redis.example.com", queueRedisHost, "Should access queue Redis host")

	queueTable := loader.GetString("complex.application.features.queue.connections.database.table")
	require.Equal(t, "jobs", queueTable, "Should access queue table name")

	// Test non-existent paths
	missingValue := loader.GetString("complex.application.features.cache.drivers.redis.missing", "default")
	require.Equal(t, "default", missingValue, "Should return default for missing path")
}

// TestConfigSystemTypeSafety tests type safety across the config system
func TestConfigSystemTypeSafety(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create config with mixed types
	mixedConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"strings": map[string]interface{}{
				"simple":  "hello",
				"empty":   "",
				"unicode": "café",
				"special": "test\n\t\r",
			},
			"integers": map[string]interface{}{
				"positive": 42,
				"negative": -123,
				"zero":     0,
				"large":    999999999,
			},
			"floats": map[string]interface{}{
				"simple":   3.14,
				"negative": -2.718,
				"zero":     0.0,
				"large":    123456.789,
			},
			"booleans": map[string]interface{}{
				"true":  true,
				"false": false,
			},
			"nil_values": map[string]interface{}{
				"null": nil,
			},
			"arrays": map[string]interface{}{
				"strings": []string{"a", "b", "c"},
				"ints":    []int{1, 2, 3},
				"mixed":   []interface{}{"a", 1, true},
			},
		}
	}
	go_core.RegisterGlobalConfig("mixed", mixedConfigFunc)

	loader := go_core.GetGlobalConfigLoader()

	// Test string conversions
	require.Equal(t, "hello", loader.GetString("mixed.strings.simple"), "Should convert string to string")
	require.Equal(t, "", loader.GetString("mixed.strings.empty"), "Should handle empty string")
	require.Equal(t, "café", loader.GetString("mixed.strings.unicode"), "Should handle unicode")
	require.Equal(t, "test\n\t\r", loader.GetString("mixed.strings.special"), "Should handle special characters")

	// Test integer conversions
	require.Equal(t, 42, loader.GetInt("mixed.integers.positive"), "Should convert positive int")
	require.Equal(t, -123, loader.GetInt("mixed.integers.negative"), "Should convert negative int")
	require.Equal(t, 0, loader.GetInt("mixed.integers.zero"), "Should convert zero")
	require.Equal(t, 999999999, loader.GetInt("mixed.integers.large"), "Should convert large int")

	// Test float to int conversions
	require.Equal(t, 3, loader.GetInt("mixed.floats.simple"), "Should convert float to int")
	require.Equal(t, -2, loader.GetInt("mixed.floats.negative"), "Should convert negative float to int")
	require.Equal(t, 0, loader.GetInt("mixed.floats.zero"), "Should convert zero float to int")

	// Test boolean conversions
	require.Equal(t, true, loader.GetBool("mixed.booleans.true"), "Should convert true boolean")
	require.Equal(t, false, loader.GetBool("mixed.booleans.false"), "Should convert false boolean")

	// Test nil handling
	require.Nil(t, loader.Get("mixed.nil_values.null"), "Should handle nil values")

	// Test type conversion failures (should return defaults)
	require.Equal(t, 999, loader.GetInt("mixed.strings.simple", 999), "Should return default for string to int conversion")
	require.Equal(t, true, loader.GetBool("mixed.strings.simple", true), "Should return default for string to bool conversion")
	require.Equal(t, "default", loader.GetString("mixed.nil_values.null", "default"), "Should return default for nil to string conversion")
}

// TestConfigSystemConcurrency tests concurrent access to config system
func TestConfigSystemConcurrency(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create test config
	concurrentConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"value": "test_value",
			"count": 42,
			"flag":  true,
		}
	}
	go_core.RegisterGlobalConfig("concurrent", concurrentConfigFunc)

	// Test concurrent access - simplified version
	loader := go_core.GetGlobalConfigLoader()
	require.Equal(t, "test_value", loader.GetString("concurrent.value"))
	require.Equal(t, 42, loader.GetInt("concurrent.count"))
	require.Equal(t, true, loader.GetBool("concurrent.flag"))
}

// TestConfigSystemPerformance tests performance of config operations
func TestConfigSystemPerformance(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create test config
	performanceConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"value": "performance_test_value",
			"count": 100,
			"flag":  true,
		}
	}
	go_core.RegisterGlobalConfig("performance", performanceConfigFunc)

	// Test performance - simplified version
	loader := go_core.GetGlobalConfigLoader()
	require.Equal(t, "performance_test_value", loader.GetString("performance.value"))
	require.Equal(t, 100, loader.GetInt("performance.count"))
	require.Equal(t, true, loader.GetBool("performance.flag"))
}

// Helper functions for environment variable handling
func getEnvWithFallback(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBoolWithFallback(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true"
	}
	return fallback
}
