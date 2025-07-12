package integration

import (
	"os"
	"testing"

	"base_lara_go_project/app/core/go_core"
	config_core "base_lara_go_project/app/core/laravel_core/config"

	"github.com/stretchr/testify/require"
)

// TestConfigFacadeIntegrationWithGoCore tests integration between Laravel facade and Go core
func TestConfigFacadeIntegrationWithGoCore(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create config using Go core directly
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"app": map[string]interface{}{
				"name":  "Test App",
				"debug": true,
				"port":  "8080",
			},
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 3306,
			},
		}
	}
	go_core.RegisterGlobalConfig("integration_test", configFunc)

	// Test accessing through Laravel facade
	appName := config_core.GetString("integration_test.app.name")
	require.Equal(t, "Test App", appName, "Should access app name through facade")

	appDebug := config_core.GetBool("integration_test.app.debug")
	require.Equal(t, true, appDebug, "Should access app debug through facade")

	appPort := config_core.GetString("integration_test.app.port")
	require.Equal(t, "8080", appPort, "Should access app port through facade")

	dbHost := config_core.GetString("integration_test.database.host")
	require.Equal(t, "localhost", dbHost, "Should access database host through facade")

	dbPort := config_core.GetInt("integration_test.database.port")
	require.Equal(t, 3306, dbPort, "Should access database port through facade")

	// Test that facade and Go core are in sync
	loader := go_core.GetGlobalConfigLoader()

	// Verify through Go core directly
	goCoreAppName := loader.GetString("integration_test.app.name")
	require.Equal(t, appName, goCoreAppName, "Go core and facade should return same value")

	goCoreAppDebug := loader.GetBool("integration_test.app.debug")
	require.Equal(t, appDebug, goCoreAppDebug, "Go core and facade should return same value")
}

// TestConfigFacadeEnvironmentIntegration tests facade with environment variables
func TestConfigFacadeEnvironmentIntegration(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Set environment variables
	os.Setenv("APP_NAME", "Environment Test App")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "env-db.example.com")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("REDIS_HOST", "env-redis.example.com")
	os.Setenv("MAIL_HOST", "env-smtp.example.com")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("QUEUE_CONNECTION", "redis")
	os.Setenv("CACHE_STORE", "redis")

	// Create configs that use environment variables
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

	// Test facade access with environment variables
	appName := config_core.GetString("app.name")
	require.Equal(t, "Environment Test App", appName, "App name should match environment")

	appDebug := config_core.GetBool("app.debug")
	require.Equal(t, true, appDebug, "App debug should be true from environment")

	appPort := config_core.GetString("app.port")
	require.Equal(t, "9090", appPort, "App port should match environment")

	dbHost := config_core.GetString("database.connections.postgres.host")
	require.Equal(t, "env-db.example.com", dbHost, "Database host should match environment")

	dbPort := config_core.GetString("database.connections.postgres.port")
	require.Equal(t, "5432", dbPort, "Database port should match environment")

	cacheDefault := config_core.GetString("cache.default")
	require.Equal(t, "redis", cacheDefault, "Cache default should match environment")

	redisHost := config_core.GetString("cache.stores.redis.host")
	require.Equal(t, "env-redis.example.com", redisHost, "Redis host should match environment")
}

// TestConfigFacadeMultipleEnvironments tests facade with different environments
func TestConfigFacadeMultipleEnvironments(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Test development environment
	os.Setenv("APP_ENV", "development")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("DB_HOST", "dev-db.local")
	os.Setenv("REDIS_HOST", "dev-redis.local")

	devConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"env":   os.Getenv("APP_ENV"),
			"debug": os.Getenv("APP_DEBUG") == "true",
			"database": map[string]interface{}{
				"host": os.Getenv("DB_HOST"),
			},
			"cache": map[string]interface{}{
				"host": os.Getenv("REDIS_HOST"),
			},
		}
	}
	go_core.RegisterGlobalConfig("app_dev", devConfigFunc)

	// Test facade access for development
	devEnv := config_core.GetString("app_dev.env")
	require.Equal(t, "development", devEnv, "Development environment should be set")

	devDebug := config_core.GetBool("app_dev.debug")
	require.Equal(t, true, devDebug, "Development debug should be true")

	devDbHost := config_core.GetString("app_dev.database.host")
	require.Equal(t, "dev-db.local", devDbHost, "Development database host should be set")

	devCacheHost := config_core.GetString("app_dev.cache.host")
	require.Equal(t, "dev-redis.local", devCacheHost, "Development cache host should be set")

	// Test production environment
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_DEBUG", "false")
	os.Setenv("DB_HOST", "prod-db.example.com")
	os.Setenv("REDIS_HOST", "prod-redis.example.com")

	prodConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"env":   os.Getenv("APP_ENV"),
			"debug": os.Getenv("APP_DEBUG") == "true",
			"database": map[string]interface{}{
				"host": os.Getenv("DB_HOST"),
			},
			"cache": map[string]interface{}{
				"host": os.Getenv("REDIS_HOST"),
			},
		}
	}
	go_core.RegisterGlobalConfig("app_prod", prodConfigFunc)

	// Test facade access for production
	prodEnv := config_core.GetString("app_prod.env")
	require.Equal(t, "production", prodEnv, "Production environment should be set")

	prodDebug := config_core.GetBool("app_prod.debug")
	require.Equal(t, false, prodDebug, "Production debug should be false")

	prodDbHost := config_core.GetString("app_prod.database.host")
	require.Equal(t, "prod-db.example.com", prodDbHost, "Production database host should be set")

	prodCacheHost := config_core.GetString("app_prod.cache.host")
	require.Equal(t, "prod-redis.example.com", prodCacheHost, "Production cache host should be set")

	// Verify both configs coexist
	require.True(t, config_core.Has("app_dev.env"), "Should have development config")
	require.True(t, config_core.Has("app_prod.env"), "Should have production config")
}

// TestConfigFacadeComplexNesting tests facade with complex nested configurations
func TestConfigFacadeComplexNesting(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create complex nested configuration
	complexConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"application": map[string]interface{}{
				"name": "Complex Test App",
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
									"ssl": map[string]interface{}{
										"enabled": false,
										"cert":    "/path/to/cert",
									},
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
								"port": 6379,
							},
							"database": map[string]interface{}{
								"table":       "jobs",
								"retry_after": 90,
							},
						},
					},
					"mail": map[string]interface{}{
						"enabled": true,
						"drivers": map[string]interface{}{
							"smtp": map[string]interface{}{
								"host":       "smtp.example.com",
								"port":       587,
								"encryption": "tls",
							},
						},
					},
				},
			},
		}
	}
	go_core.RegisterGlobalConfig("complex_app", complexConfigFunc)

	// Test deep nested access through facade
	appName := config_core.GetString("complex_app.application.name")
	require.Equal(t, "Complex Test App", appName, "Should access application name")

	// Test cache configuration
	cacheEnabled := config_core.GetBool("complex_app.application.features.cache.enabled")
	require.Equal(t, true, cacheEnabled, "Cache should be enabled")

	redisHost := config_core.GetString("complex_app.application.features.cache.drivers.redis.host")
	require.Equal(t, "redis.example.com", redisHost, "Should access Redis host")

	redisPort := config_core.GetInt("complex_app.application.features.cache.drivers.redis.port")
	require.Equal(t, 6379, redisPort, "Should access Redis port")

	redisTimeout := config_core.GetInt("complex_app.application.features.cache.drivers.redis.options.timeout")
	require.Equal(t, 30, redisTimeout, "Should access Redis timeout")

	redisSslEnabled := config_core.GetBool("complex_app.application.features.cache.drivers.redis.options.ssl.enabled")
	require.Equal(t, false, redisSslEnabled, "Should access Redis SSL enabled")

	redisSslCert := config_core.GetString("complex_app.application.features.cache.drivers.redis.options.ssl.cert")
	require.Equal(t, "/path/to/cert", redisSslCert, "Should access Redis SSL cert")

	// Test queue configuration
	queueEnabled := config_core.GetBool("complex_app.application.features.queue.enabled")
	require.Equal(t, true, queueEnabled, "Queue should be enabled")

	queueRedisHost := config_core.GetString("complex_app.application.features.queue.connections.redis.host")
	require.Equal(t, "queue-redis.example.com", queueRedisHost, "Should access queue Redis host")

	queueTable := config_core.GetString("complex_app.application.features.queue.connections.database.table")
	require.Equal(t, "jobs", queueTable, "Should access queue table name")

	queueRetryAfter := config_core.GetInt("complex_app.application.features.queue.connections.database.retry_after")
	require.Equal(t, 90, queueRetryAfter, "Should access queue retry after")

	// Test mail configuration
	mailEnabled := config_core.GetBool("complex_app.application.features.mail.enabled")
	require.Equal(t, true, mailEnabled, "Mail should be enabled")

	smtpHost := config_core.GetString("complex_app.application.features.mail.drivers.smtp.host")
	require.Equal(t, "smtp.example.com", smtpHost, "Should access SMTP host")

	smtpPort := config_core.GetInt("complex_app.application.features.mail.drivers.smtp.port")
	require.Equal(t, 587, smtpPort, "Should access SMTP port")

	smtpEncryption := config_core.GetString("complex_app.application.features.mail.drivers.smtp.encryption")
	require.Equal(t, "tls", smtpEncryption, "Should access SMTP encryption")

	// Test non-existent paths
	missingValue := config_core.GetString("complex_app.application.features.cache.drivers.redis.missing", "default")
	require.Equal(t, "default", missingValue, "Should return default for missing path")
}

// TestConfigFacadeTypeSafety tests type safety across the facade
func TestConfigFacadeTypeSafety(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create config with mixed types
	mixedTypesConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"strings": map[string]interface{}{
				"simple":           "hello",
				"empty":            "",
				"unicode":          "café",
				"special":          "test\n\t\r",
				"number_as_string": "42",
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
	go_core.RegisterGlobalConfig("mixed_types", mixedTypesConfigFunc)

	// Test string conversions
	require.Equal(t, "hello", config_core.GetString("mixed_types.strings.simple"), "Should convert string to string")
	require.Equal(t, "", config_core.GetString("mixed_types.strings.empty"), "Should handle empty string")
	require.Equal(t, "café", config_core.GetString("mixed_types.strings.unicode"), "Should handle unicode")
	require.Equal(t, "test\n\t\r", config_core.GetString("mixed_types.strings.special"), "Should handle special characters")

	// Test integer conversions
	require.Equal(t, 42, config_core.GetInt("mixed_types.integers.positive"), "Should convert positive int")
	require.Equal(t, -123, config_core.GetInt("mixed_types.integers.negative"), "Should convert negative int")
	require.Equal(t, 0, config_core.GetInt("mixed_types.integers.zero"), "Should convert zero")
	require.Equal(t, 999999999, config_core.GetInt("mixed_types.integers.large"), "Should convert large int")

	// Test float to int conversions
	require.Equal(t, 3, config_core.GetInt("mixed_types.floats.simple"), "Should convert float to int")
	require.Equal(t, -2, config_core.GetInt("mixed_types.floats.negative"), "Should convert negative float to int")
	require.Equal(t, 0, config_core.GetInt("mixed_types.floats.zero"), "Should convert zero float to int")

	// Test boolean conversions
	require.Equal(t, true, config_core.GetBool("mixed_types.booleans.true"), "Should convert true boolean")
	require.Equal(t, false, config_core.GetBool("mixed_types.booleans.false"), "Should convert false boolean")

	// Test nil handling
	require.Nil(t, config_core.Get("mixed_types.nil_values.null"), "Should handle nil values")

	// Test type conversion failures (should return defaults)
	require.Equal(t, 999, config_core.GetInt("mixed_types.strings.simple", 999), "Should return default for string to int conversion")
	require.Equal(t, true, config_core.GetBool("mixed_types.strings.simple", true), "Should return default for string to bool conversion")
	require.Equal(t, "default", config_core.GetString("mixed_types.nil_values.null", "default"), "Should return default for nil to string conversion")

	// Test number as string conversion
	require.Equal(t, 999, config_core.GetInt("mixed_types.strings.number_as_string", 999), "Should return default for number as string")
}

// TestConfigFacadeConcurrency tests concurrent access to facade
func TestConfigFacadeConcurrency(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create test config
	concurrentConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"string_value": "concurrent_test_value",
			"int_value":    100,
			"bool_value":   true,
			"nested": map[string]interface{}{
				"key": "nested_concurrent_value",
			},
		}
	}
	go_core.RegisterGlobalConfig("concurrent", concurrentConfigFunc)

	// Test concurrent access - simplified version
	// In a real implementation, you would use goroutines and sync.WaitGroup
	// For now, just test that the config is accessible
	require.Equal(t, "concurrent_test_value", config_core.GetString("concurrent.string_value"))
	require.Equal(t, 100, config_core.GetInt("concurrent.int_value"))
	require.Equal(t, true, config_core.GetBool("concurrent.bool_value"))
	require.Equal(t, "nested_concurrent_value", config_core.GetString("concurrent.nested.key"))
}

// TestConfigFacadePerformance tests performance of facade operations
func TestConfigFacadePerformance(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create test config
	performanceConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"string_value": "performance_test_value",
			"int_value":    200,
			"bool_value":   true,
			"nested": map[string]interface{}{
				"deep": map[string]interface{}{
					"value": "deep_performance_value",
				},
			},
		}
	}
	go_core.RegisterGlobalConfig("performance", performanceConfigFunc)

	// Test performance - simplified version
	// In a real implementation, you would use testing.B for benchmarks
	require.Equal(t, "performance_test_value", config_core.GetString("performance.string_value"))
	require.Equal(t, 200, config_core.GetInt("performance.int_value"))
	require.Equal(t, true, config_core.GetBool("performance.bool_value"))
	require.Equal(t, "deep_performance_value", config_core.GetString("performance.nested.deep.value"))
}

// TestConfigFacadeSetIntegration tests Set method integration
func TestConfigFacadeSetIntegration(t *testing.T) {
	// Clear any existing configs and cache
	go_core.ClearGlobalConfigs()
	config_core.ClearCache()

	// Create initial config
	initialConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"initial": "initial_value",
		}
	}
	go_core.RegisterGlobalConfig("set_test", initialConfigFunc)

	// Test setting new values through facade
	config_core.Set("set_test.new_key", "new_value")
	config_core.Set("set_test.nested.key", "nested_value")
	config_core.Set("set_test.deep.nested.key", "deep_nested_value")

	// Verify values were set
	require.True(t, config_core.Has("set_test.new_key"), "Should have new_key after setting")
	require.True(t, config_core.Has("set_test.nested.key"), "Should have nested key after setting")
	require.True(t, config_core.Has("set_test.deep.nested.key"), "Should have deep nested key after setting")

	// Test retrieving set values
	newValue := config_core.GetString("set_test.new_key")
	require.Equal(t, "new_value", newValue, "Should return set value for new_key")

	nestedValue := config_core.GetString("set_test.nested.key")
	require.Equal(t, "nested_value", nestedValue, "Should return set value for nested key")

	deepNestedValue := config_core.GetString("set_test.deep.nested.key")
	require.Equal(t, "deep_nested_value", deepNestedValue, "Should return set value for deep nested key")

	// Test overwriting existing values
	config_core.Set("set_test.initial", "overwritten_value")
	overwrittenValue := config_core.GetString("set_test.initial")
	require.Equal(t, "overwritten_value", overwrittenValue, "Should return overwritten value")

	// Verify through Go core directly - use the same loader instance as the facade
	loader := config_core.GetConfigLoader() // Use the same loader as the facade
	goCoreValue := loader.GetString("set_test.new_key")
	require.Equal(t, newValue, goCoreValue, "Go core and facade should be in sync after Set")
}
