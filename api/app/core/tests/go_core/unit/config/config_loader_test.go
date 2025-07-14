package unit

import (
	"testing"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/require"
)

// TestConfigLoaderNew tests the NewConfigLoader function
func TestConfigLoaderNew(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Test with non-existent config directory to avoid global config interference
	loader := go_core.NewConfigLoader("/non/existent/path")
	require.NotNil(t, loader, "Loader should not be nil")

	// Test with custom config directory
	loader = go_core.NewConfigLoader("/custom/path")
	require.NotNil(t, loader, "Loader should not be nil")
}

// TestConfigLoaderRegisterConfig tests the RegisterConfig method
func TestConfigLoaderRegisterConfig(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test registering a simple config
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"key": "value",
		}
	}

	loader.RegisterConfig("test", configFunc)

	// Verify the config was registered
	configs := loader.ListAvailableConfigs()
	require.Contains(t, configs, "test", "Config should be in available configs list")
}

// TestConfigLoaderLoad tests the Load method
func TestConfigLoaderLoad(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	// Create a fresh loader with a non-existent directory to avoid global config interference
	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test loading non-existent config
	config, err := loader.Load("non_existent")
	require.Nil(t, err, "Load should not return error for non-existent config")
	require.Nil(t, config, "Load should return nil for non-existent config")

	// Test loading registered config
	expectedConfig := map[string]interface{}{
		"key": "value",
	}
	configFunc := func() map[string]interface{} {
		return expectedConfig
	}

	loader.RegisterConfig("test", configFunc)

	config, err = loader.Load("test")
	require.Nil(t, err, "Load should not return error for registered config")
	require.NotNil(t, config, "Load should return config for registered config")
	require.Equal(t, "value", config["key"], "Config should contain expected value")

	// Test loading the same config again (should use cache)
	config2, err := loader.Load("test")
	require.Nil(t, err, "Load should not return error for cached config")
	require.NotNil(t, config2, "Load should return cached config")
	require.Equal(t, "value", config2["key"], "Cached config should contain expected value")
}

// TestConfigLoaderGet tests the Get method
func TestConfigLoaderGet(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test getting non-existent key
	value := loader.Get("non_existent.key")
	require.Nil(t, value, "Get should return nil for non-existent key")

	// Test getting non-existent key with default
	value = loader.Get("non_existent.key", "default")
	require.Equal(t, "default", value, "Get should return default value for non-existent key")

	// Test getting registered config
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
		}
	}

	loader.RegisterConfig("test", configFunc)

	// Test getting top-level config
	value = loader.Get("test")
	require.NotNil(t, value, "Get should return config for top-level key")

	// Test getting nested key
	value = loader.Get("test.nested.key")
	require.Equal(t, "value", value, "Get should return nested value")

	// Test getting non-existent nested key
	value = loader.Get("test.nested.missing", "default")
	require.Equal(t, "default", value, "Get should return default for non-existent nested key")
}

// TestConfigLoaderGetString tests the GetString method
func TestConfigLoaderGetString(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test getting non-existent key
	value := loader.GetString("non_existent.key")
	require.Equal(t, "", value, "GetString should return empty string for non-existent key")

	// Test getting non-existent key with default
	value = loader.GetString("non_existent.key", "default")
	require.Equal(t, "default", value, "GetString should return default for non-existent key")

	// Test getting string value
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"string_key": "string_value",
		}
	}

	loader.RegisterConfig("test", configFunc)

	value = loader.GetString("test.string_key")
	require.Equal(t, "string_value", value, "GetString should return string value")

	// Test getting non-string value (should return default)
	value = loader.GetString("test.string_key", "default")
	require.Equal(t, "string_value", value, "GetString should return actual value when it exists")
}

// TestConfigLoaderGetInt tests the GetInt method
func TestConfigLoaderGetInt(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test getting non-existent key
	value := loader.GetInt("non_existent.key")
	require.Equal(t, 0, value, "GetInt should return 0 for non-existent key")

	// Test getting non-existent key with default
	value = loader.GetInt("non_existent.key", 42)
	require.Equal(t, 42, value, "GetInt should return default for non-existent key")

	// Test getting int value
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"int_key": 123,
		}
	}

	loader.RegisterConfig("test", configFunc)

	value = loader.GetInt("test.int_key")
	require.Equal(t, 123, value, "GetInt should return int value")

	// Test getting float value (should convert to int)
	configFunc = func() map[string]interface{} {
		return map[string]interface{}{
			"float_key": 123.45,
		}
	}

	loader.RegisterConfig("test_float", configFunc)

	value = loader.GetInt("test_float.float_key")
	require.Equal(t, 123, value, "GetInt should convert float to int")

	// Test getting non-numeric value (should return default)
	configFunc = func() map[string]interface{} {
		return map[string]interface{}{
			"string_key": "not_a_number",
		}
	}

	loader.RegisterConfig("test_string", configFunc)

	value = loader.GetInt("test_string.string_key", 999)
	require.Equal(t, 999, value, "GetInt should return default for non-numeric value")
}

// TestConfigLoaderGetBool tests the GetBool method
func TestConfigLoaderGetBool(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test getting non-existent key
	value := loader.GetBool("non_existent.key")
	require.Equal(t, false, value, "GetBool should return false for non-existent key")

	// Test getting non-existent key with default
	value = loader.GetBool("non_existent.key", true)
	require.Equal(t, true, value, "GetBool should return default for non-existent key")

	// Test getting bool value
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"bool_key": true,
		}
	}

	loader.RegisterConfig("test", configFunc)

	value = loader.GetBool("test.bool_key")
	require.Equal(t, true, value, "GetBool should return bool value")

	// Test getting non-bool value (should return default)
	configFunc = func() map[string]interface{} {
		return map[string]interface{}{
			"string_key": "not_a_bool",
		}
	}

	loader.RegisterConfig("test_string", configFunc)

	value = loader.GetBool("test_string.string_key", true)
	require.Equal(t, true, value, "GetBool should return default for non-bool value")
}

// TestConfigLoaderHas tests the Has method
func TestConfigLoaderHas(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test checking non-existent key
	exists := loader.Has("non_existent.key")
	require.False(t, exists, "Has should return false for non-existent key")

	// Test checking existing key
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
		}
	}

	loader.RegisterConfig("test", configFunc)

	exists = loader.Has("test.nested.key")
	require.True(t, exists, "Has should return true for existing key")

	exists = loader.Has("test.nested.missing")
	require.False(t, exists, "Has should return false for non-existent nested key")
}

// TestConfigLoaderClearCache tests the ClearCache method
func TestConfigLoaderClearCache(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Track config function calls
	callCount := 0
	configFunc := func() map[string]interface{} {
		callCount++
		return map[string]interface{}{
			"key": "value",
		}
	}

	loader.RegisterConfig("test", configFunc)

	// Load config multiple times
	loader.Load("test")
	loader.Load("test")
	loader.Load("test")

	// Should only call once due to caching
	require.Equal(t, 1, callCount, "Config function should be called once due to caching")

	// Clear cache
	loader.ClearCache()

	// Load again
	loader.Load("test")

	// Should call again after cache clear
	require.Equal(t, 2, callCount, "Config function should be called again after cache clear")
}

// TestConfigLoaderReload tests the Reload method
func TestConfigLoaderReload(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Track config function calls
	callCount := 0
	configFunc := func() map[string]interface{} {
		callCount++
		return map[string]interface{}{
			"key": "value",
		}
	}

	loader.RegisterConfig("test", configFunc)

	// Load config
	loader.Load("test")
	require.Equal(t, 1, callCount, "Config function should be called once")

	// Reload config
	err := loader.Reload("test")
	require.Nil(t, err, "Reload should not return error")
	require.Equal(t, 2, callCount, "Config function should be called again after reload")

	// Test reloading non-existent config
	err = loader.Reload("non_existent")
	require.Nil(t, err, "Reload should not return error for non-existent config")
}

// TestConfigLoaderListAvailableConfigs tests the ListAvailableConfigs method
func TestConfigLoaderListAvailableConfigs(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Test empty list
	configs := loader.ListAvailableConfigs()
	require.Empty(t, configs, "Should return empty list when no configs registered")

	// Register configs
	configFunc1 := func() map[string]interface{} {
		return map[string]interface{}{"key": "value1"}
	}
	configFunc2 := func() map[string]interface{} {
		return map[string]interface{}{"key": "value2"}
	}

	loader.RegisterConfig("config1", configFunc1)
	loader.RegisterConfig("config2", configFunc2)

	// Test populated list
	configs = loader.ListAvailableConfigs()
	require.Len(t, configs, 2, "Should return list with 2 configs")
	require.Contains(t, configs, "config1", "Should contain config1")
	require.Contains(t, configs, "config2", "Should contain config2")
}

// TestConfigLoaderDotNotation tests complex dot notation scenarios
func TestConfigLoaderDotNotation(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Create complex nested config
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"value": "deep_value",
						"array": []string{"item1", "item2", "item3"},
						"mixed": map[string]interface{}{
							"string": "string_value",
							"int":    42,
							"bool":   true,
						},
					},
				},
			},
		}
	}

	loader.RegisterConfig("complex", configFunc)

	// Test deep nested access
	value := loader.GetString("complex.level1.level2.level3.value")
	require.Equal(t, "deep_value", value, "Should access deep nested string value")

	// Test array access (should return default since we can't access array elements)
	value = loader.GetString("complex.level1.level2.level3.array", "default")
	require.Equal(t, "default", value, "Should return default for array access")

	// Test mixed type access
	strValue := loader.GetString("complex.level1.level2.level3.mixed.string")
	require.Equal(t, "string_value", strValue, "Should access nested string")

	intValue := loader.GetInt("complex.level1.level2.level3.mixed.int")
	require.Equal(t, 42, intValue, "Should access nested int")

	boolValue := loader.GetBool("complex.level1.level2.level3.mixed.bool")
	require.Equal(t, true, boolValue, "Should access nested bool")

	// Test non-existent deep path
	value = loader.GetString("complex.level1.level2.level3.missing", "default")
	require.Equal(t, "default", value, "Should return default for non-existent deep path")
}

// TestConfigLoaderTypeConversion tests type conversion edge cases
func TestConfigLoaderTypeConversion(t *testing.T) {
	// Clear any existing configs
	go_core.ClearGlobalConfigs()

	loader := go_core.NewConfigLoader("/non/existent/path")

	// Create config with various types
	configFunc := func() map[string]interface{} {
		return map[string]interface{}{
			"string_value": "hello",
			"int_value":    42,
			"float_value":  3.14,
			"bool_value":   true,
			"nil_value":    nil,
			"map_value":    map[string]interface{}{"key": "value"},
			"slice_value":  []string{"item1", "item2"},
		}
	}

	loader.RegisterConfig("types", configFunc)

	// Test string conversion
	strValue := loader.GetString("types.string_value")
	require.Equal(t, "hello", strValue, "Should convert string to string")

	// Test int conversion
	intValue := loader.GetInt("types.int_value")
	require.Equal(t, 42, intValue, "Should convert int to int")

	// Test float to int conversion
	floatToInt := loader.GetInt("types.float_value")
	require.Equal(t, 3, floatToInt, "Should convert float to int")

	// Test bool conversion
	boolValue := loader.GetBool("types.bool_value")
	require.Equal(t, true, boolValue, "Should convert bool to bool")

	// Test nil handling
	nilValue := loader.Get("types.nil_value")
	require.Nil(t, nilValue, "Should handle nil values")

	// Test non-convertible types
	nonConvertibleInt := loader.GetInt("types.string_value", 999)
	require.Equal(t, 999, nonConvertibleInt, "Should return default for non-convertible string to int")

	nonConvertibleBool := loader.GetBool("types.string_value", true)
	require.Equal(t, true, nonConvertibleBool, "Should return default for non-convertible string to bool")
}
