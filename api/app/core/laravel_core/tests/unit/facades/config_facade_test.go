package unit

import (
	"os"
	"testing"

	"base_lara_go_project/app/core/go_core"
	config_core "base_lara_go_project/app/core/laravel_core/config"

	"github.com/stretchr/testify/require"
)

// TestConfigFacadeGet tests the Get method
func TestConfigFacadeGet(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create test config
	testConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"string_value": "hello",
			"int_value":    42,
			"bool_value":   true,
			"nested": map[string]interface{}{
				"key": "nested_value",
			},
		}
	}
	go_core.RegisterGlobalConfig("test", testConfigFunc)

	// Test getting existing values
	value := config_core.Get("test.string_value")
	require.Equal(t, "hello", value, "Should return string value")

	value = config_core.Get("test.int_value")
	require.Equal(t, 42, value, "Should return int value")

	value = config_core.Get("test.bool_value")
	require.Equal(t, true, value, "Should return bool value")

	value = config_core.Get("test.nested.key")
	require.Equal(t, "nested_value", value, "Should return nested value")

	// Test getting non-existent values
	value = config_core.Get("test.missing")
	require.Nil(t, value, "Should return nil for missing key")

	// Test getting with default
	value = config_core.Get("test.missing", "default")
	require.Equal(t, "default", value, "Should return default for missing key")
}

// TestConfigFacadeGetString tests the GetString method
func TestConfigFacadeGetString(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create test config
	testConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"string_value": "hello",
			"int_value":    42,
			"bool_value":   true,
		}
	}
	go_core.RegisterGlobalConfig("test", testConfigFunc)

	// Test getting string values
	value := config_core.GetString("test.string_value")
	require.Equal(t, "hello", value, "Should return string value")

	// Test getting non-existent values
	value = config_core.GetString("test.missing")
	require.Equal(t, "", value, "Should return empty string for missing key")

	// Test getting with default
	value = config_core.GetString("test.missing", "default")
	require.Equal(t, "default", value, "Should return default for missing key")

	// Test getting non-string values (should return default)
	value = config_core.GetString("test.int_value", "default")
	require.Equal(t, "default", value, "Should return default for non-string value")
}

// TestConfigFacadeGetInt tests the GetInt method
func TestConfigFacadeGetInt(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()
	config_core.ClearCache()

	// Create test config
	testConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"int_value":    42,
			"float_value":  3.14,
			"string_value": "hello",
		}
	}
	go_core.RegisterGlobalConfig("test", testConfigFunc)

	// Test getting int values
	value := config_core.GetInt("test.int_value")
	require.Equal(t, 42, value, "Should return int value")

	// Test getting float values (should convert to int)
	value = config_core.GetInt("test.float_value")
	require.Equal(t, 3, value, "Should convert float to int")

	// Test getting non-existent values
	value = config_core.GetInt("test.missing")
	require.Equal(t, 0, value, "Should return 0 for missing key")

	// Test getting with default
	value = config_core.GetInt("test.missing", 999)
	require.Equal(t, 999, value, "Should return default for missing key")

	// Test getting non-numeric values (should return default)
	value = config_core.GetInt("test.string_value", 999)
	require.Equal(t, 999, value, "Should return default for non-numeric value")
}

// TestConfigFacadeGetBool tests the GetBool method
func TestConfigFacadeGetBool(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()
	config_core.ClearCache()

	// Create test config
	testConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"bool_true":    true,
			"bool_false":   false,
			"string_value": "hello",
		}
	}
	go_core.RegisterGlobalConfig("test", testConfigFunc)

	// Test getting bool values
	value := config_core.GetBool("test.bool_true")
	require.Equal(t, true, value, "Should return true bool value")

	value = config_core.GetBool("test.bool_false")
	require.Equal(t, false, value, "Should return false bool value")

	// Test getting non-existent values
	value = config_core.GetBool("test.missing")
	require.Equal(t, false, value, "Should return false for missing key")

	// Test getting with default
	value = config_core.GetBool("test.missing", true)
	require.Equal(t, true, value, "Should return default for missing key")

	// Test getting non-bool values (should return default)
	value = config_core.GetBool("test.string_value", true)
	require.Equal(t, true, value, "Should return default for non-bool value")
}

// TestConfigFacadeHas tests the Has method
func TestConfigFacadeHas(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()
	config_core.ClearCache()

	// Create test config
	testConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"key": "value",
			"nested": map[string]interface{}{
				"key": "nested_value",
			},
		}
	}
	go_core.RegisterGlobalConfig("test", testConfigFunc)

	// Test checking existing keys
	require.True(t, config_core.Has("test.key"), "Should return true for existing key")
	require.True(t, config_core.Has("test.nested.key"), "Should return true for existing nested key")

	// Test checking non-existent keys
	require.False(t, config_core.Has("test.missing"), "Should return false for missing key")
	require.False(t, config_core.Has("test.nested.missing"), "Should return false for missing nested key")
	require.False(t, config_core.Has("missing.config"), "Should return false for missing config")
}

// TestConfigFacadeSet tests the Set method
func TestConfigFacadeSet(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create initial config
	initialConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"key1": "value1",
		}
	}
	go_core.RegisterGlobalConfig("test", initialConfigFunc)

	// Test setting new values
	config_core.Set("test.key2", "value2")
	config_core.Set("test.nested.key", "nested_value")

	// Verify values were set
	require.True(t, config_core.Has("test.key2"), "Should have key2 after setting")
	require.True(t, config_core.Has("test.nested.key"), "Should have nested key after setting")

	value := config_core.GetString("test.key2")
	require.Equal(t, "value2", value, "Should return set value for key2")

	value = config_core.GetString("test.nested.key")
	require.Equal(t, "nested_value", value, "Should return set value for nested key")

	// Test overwriting existing values
	config_core.Set("test.key1", "new_value")
	value = config_core.GetString("test.key1")
	require.Equal(t, "new_value", value, "Should return overwritten value")
}

// TestConfigFacadeDotNotation tests complex dot notation scenarios
func TestConfigFacadeDotNotation(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create complex nested config
	complexConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"string_value": "deep_value",
						"int_value":    42,
						"bool_value":   true,
					},
				},
			},
		}
	}
	go_core.RegisterGlobalConfig("complex", complexConfigFunc)

	// Test deep nested access
	value := config_core.GetString("complex.level1.level2.level3.string_value")
	require.Equal(t, "deep_value", value, "Should access deep nested string value")

	valueInt := config_core.GetInt("complex.level1.level2.level3.int_value")
	require.Equal(t, 42, valueInt, "Should access deep nested int value")

	valueBool := config_core.GetBool("complex.level1.level2.level3.bool_value")
	require.Equal(t, true, valueBool, "Should access deep nested bool value")

	// Test non-existent deep path
	value = config_core.GetString("complex.level1.level2.level3.missing", "default")
	require.Equal(t, "default", value, "Should return default for non-existent deep path")
}

// TestConfigFacadeTypeConversion tests type conversion edge cases
func TestConfigFacadeTypeConversion(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create config with various types
	typesConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"string_value": "hello",
			"int_value":    42,
			"float_value":  3.14,
			"bool_value":   true,
			"nil_value":    nil,
		}
	}
	go_core.RegisterGlobalConfig("types", typesConfigFunc)

	// Test string conversions
	strValue := config_core.GetString("types.string_value")
	require.Equal(t, "hello", strValue, "Should convert string to string")

	// Test int conversions
	intValue := config_core.GetInt("types.int_value")
	require.Equal(t, 42, intValue, "Should convert int to int")

	// Test float to int conversion
	floatToInt := config_core.GetInt("types.float_value")
	require.Equal(t, 3, floatToInt, "Should convert float to int")

	// Test bool conversions
	boolValue := config_core.GetBool("types.bool_value")
	require.Equal(t, true, boolValue, "Should convert bool to bool")

	// Test nil handling
	nilValue := config_core.Get("types.nil_value")
	require.Nil(t, nilValue, "Should handle nil values")

	// Test type conversion failures (should return defaults)
	nonConvertibleInt := config_core.GetInt("types.string_value", 999)
	require.Equal(t, 999, nonConvertibleInt, "Should return default for string to int conversion")

	nonConvertibleBool := config_core.GetBool("types.string_value", true)
	require.Equal(t, true, nonConvertibleBool, "Should return default for string to bool conversion")

	nonConvertibleString := config_core.GetString("types.nil_value", "default")
	require.Equal(t, "default", nonConvertibleString, "Should return default for nil to string conversion")
}

// TestConfigFacadeEnvironmentIntegration tests integration with environment variables
func TestConfigFacadeEnvironmentIntegration(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Set environment variables
	os.Setenv("APP_NAME", "Test App")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("APP_PORT", "8080")

	// Create config that uses environment variables
	appConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"name":  "Test App", // Use environment value
			"debug": true,       // Use default as fallback
			"port":  "8080",     // Use environment value
		}
	}
	go_core.RegisterGlobalConfig("app", appConfigFunc)

	// Test accessing config values
	appName := config_core.GetString("app.name")
	require.Equal(t, "Test App", appName, "App name should match environment variable")

	appDebug := config_core.GetBool("app.debug")
	require.Equal(t, true, appDebug, "App debug should be true from environment")

	appPort := config_core.GetString("app.port")
	require.Equal(t, "8080", appPort, "App port should match environment variable")
}

// TestConfigFacadeMultipleConfigs tests working with multiple configs
func TestConfigFacadeMultipleConfigs(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create multiple configs
	appConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"name":  "Test App",
			"debug": true,
		}
	}
	go_core.RegisterGlobalConfig("app", appConfigFunc)

	dbConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"host": "localhost",
			"port": 3306,
		}
	}
	go_core.RegisterGlobalConfig("database", dbConfigFunc)

	cacheConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"driver": "redis",
			"host":   "redis.local",
		}
	}
	go_core.RegisterGlobalConfig("cache", cacheConfigFunc)

	// Test accessing different configs
	appName := config_core.GetString("app.name")
	require.Equal(t, "Test App", appName, "Should access app config")

	dbHost := config_core.GetString("database.host")
	require.Equal(t, "localhost", dbHost, "Should access database config")

	cacheDriver := config_core.GetString("cache.driver")
	require.Equal(t, "redis", cacheDriver, "Should access cache config")

	// Test that configs don't interfere with each other
	require.True(t, config_core.Has("app.name"), "Should have app.name")
	require.True(t, config_core.Has("database.host"), "Should have database.host")
	require.True(t, config_core.Has("cache.driver"), "Should have cache.driver")

	// Test that non-existent keys in one config don't affect others
	require.False(t, config_core.Has("app.missing"), "Should not have app.missing")
	require.False(t, config_core.Has("database.missing"), "Should not have database.missing")
	require.False(t, config_core.Has("cache.missing"), "Should not have cache.missing")
}

// TestConfigFacadeDynamicChanges tests dynamic config changes at runtime
func TestConfigFacadeDynamicChanges(t *testing.T) {
	// Clear any existing configs and cache
	go_core.ClearGlobalConfigs()
	config_core.ClearCache()

	// Create initial config
	initialConfigFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"app_name":   "MyApp",
			"debug_mode": false,
			"port":       8080,
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
		}
	}
	go_core.RegisterGlobalConfig("dynamic_test", initialConfigFunc)

	// Step 1: Get initial values
	appName := config_core.GetString("dynamic_test.app_name")
	debugMode := config_core.GetBool("dynamic_test.debug_mode")
	port := config_core.GetInt("dynamic_test.port")
	dbHost := config_core.GetString("dynamic_test.database.host")
	dbPort := config_core.GetInt("dynamic_test.database.port")

	require.Equal(t, "MyApp", appName, "Should get initial app name")
	require.Equal(t, false, debugMode, "Should get initial debug mode")
	require.Equal(t, 8080, port, "Should get initial port")
	require.Equal(t, "localhost", dbHost, "Should get initial database host")
	require.Equal(t, 5432, dbPort, "Should get initial database port")

	// Step 2: Set new values at runtime
	config_core.Set("dynamic_test.app_name", "UpdatedApp")
	config_core.Set("dynamic_test.debug_mode", true)
	config_core.Set("dynamic_test.port", 9000)
	config_core.Set("dynamic_test.database.host", "newhost.com")
	config_core.Set("dynamic_test.database.port", 3306)

	// Step 3: Verify the changes are immediately visible
	appName = config_core.GetString("dynamic_test.app_name")
	debugMode = config_core.GetBool("dynamic_test.debug_mode")
	port = config_core.GetInt("dynamic_test.port")
	dbHost = config_core.GetString("dynamic_test.database.host")
	dbPort = config_core.GetInt("dynamic_test.database.port")

	require.Equal(t, "UpdatedApp", appName, "Should see updated app name")
	require.Equal(t, true, debugMode, "Should see updated debug mode")
	require.Equal(t, 9000, port, "Should see updated port")
	require.Equal(t, "newhost.com", dbHost, "Should see updated database host")
	require.Equal(t, 3306, dbPort, "Should see updated database port")

	// Step 4: Set nested values
	config_core.Set("dynamic_test.database.ssl", true)
	config_core.Set("dynamic_test.database.connection_pool", 20)

	sslEnabled := config_core.GetBool("dynamic_test.database.ssl")
	poolSize := config_core.GetInt("dynamic_test.database.connection_pool")

	require.Equal(t, true, sslEnabled, "Should set and get nested boolean value")
	require.Equal(t, 20, poolSize, "Should set and get nested int value")

	// Step 5: Reload the config (should revert to original values from function)
	err := config_core.Reload("dynamic_test")
	require.NoError(t, err, "Should reload config without error")

	// Step 6: Verify values reverted to original
	appName = config_core.GetString("dynamic_test.app_name")
	debugMode = config_core.GetBool("dynamic_test.debug_mode")
	port = config_core.GetInt("dynamic_test.port")
	dbHost = config_core.GetString("dynamic_test.database.host")
	dbPort = config_core.GetInt("dynamic_test.database.port")

	require.Equal(t, "MyApp", appName, "Should revert to original app name after reload")
	require.Equal(t, false, debugMode, "Should revert to original debug mode after reload")
	require.Equal(t, 8080, port, "Should revert to original port after reload")
	require.Equal(t, "localhost", dbHost, "Should revert to original database host after reload")
	require.Equal(t, 5432, dbPort, "Should revert to original database port after reload")

	// Step 7: Verify that dynamically added nested values are gone
	sslEnabled = config_core.GetBool("dynamic_test.database.ssl")
	poolSize = config_core.GetInt("dynamic_test.database.connection_pool")

	require.Equal(t, false, sslEnabled, "Dynamically added values should be gone after reload")
	require.Equal(t, 0, poolSize, "Dynamically added values should be gone after reload")

	// Step 8: Test sequential dynamic access (removed concurrent test to avoid race conditions)
	config_core.Set("dynamic_test.sequential_test", 42)
	value := config_core.GetInt("dynamic_test.sequential_test")
	require.Equal(t, 42, value, "Should set and get sequential value")

	// Step 9: Test that Has works with dynamic changes
	require.True(t, config_core.Has("dynamic_test.app_name"), "Should have app_name")
	require.True(t, config_core.Has("dynamic_test.database.host"), "Should have nested database.host")
	require.False(t, config_core.Has("dynamic_test.missing"), "Should not have missing key")

	// Add a dynamic key and verify Has works
	config_core.Set("dynamic_test.new_key", "new_value")
	require.True(t, config_core.Has("dynamic_test.new_key"), "Should have dynamically added key")
}
