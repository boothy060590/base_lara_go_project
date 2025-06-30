package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Get retrieves an environment variable with a default value
func Get(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetInt retrieves an environment variable as an integer with a default value
func GetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetInt64 retrieves an environment variable as an int64 with a default value
func GetInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetFloat retrieves an environment variable as a float64 with a default value
func GetFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// GetBool retrieves an environment variable as a boolean with a default value
func GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetDuration retrieves an environment variable as a time.Duration with a default value
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Has checks if an environment variable exists
func Has(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// Missing checks if an environment variable is missing
func Missing(key string) bool {
	return !Has(key)
}

// Set sets an environment variable
func Set(key, value string) error {
	return os.Setenv(key, value)
}

// Unset removes an environment variable
func Unset(key string) error {
	return os.Unsetenv(key)
}

// All returns all environment variables as a map
func All() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}
	return env
}

// LoadFromFile loads environment variables from a .env file
func LoadFromFile(filename string) error {
	// TODO: Implement .env file loading
	// This would parse a .env file and set environment variables
	return nil
}

// LoadFromString loads environment variables from a string
func LoadFromString(content string) error {
	// TODO: Implement loading from string
	// This would parse environment variables from a string and set them
	return nil
}
