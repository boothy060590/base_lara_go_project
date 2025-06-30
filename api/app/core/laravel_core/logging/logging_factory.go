package logging_core

import (
	"base_lara_go_project/config"
	"fmt"
)

// LoggingFactory creates and manages logging instances
type LoggingFactory struct {
	config map[string]interface{}
}

// NewLoggingFactory creates a new logging factory
func NewLoggingFactory() *LoggingFactory {
	return &LoggingFactory{
		config: config.LoggingConfig(),
	}
}

// CreateLogger creates a logger based on configuration
func (f *LoggingFactory) CreateLogger() (LoggerInterface, error) {
	defaultChannel := f.config["default"].(string)
	channels := f.config["channels"].(map[string]interface{})

	// Get default channel configuration
	channelConfig, exists := channels[defaultChannel]
	if !exists {
		return nil, fmt.Errorf("default channel '%s' not found in configuration", defaultChannel)
	}

	configMap := channelConfig.(map[string]interface{})

	// Create logger configuration
	loggerConfig := &LoggerConfig{
		Driver:   configMap["driver"].(string),
		Level:    ParseLogLevel(configMap["level"].(string)),
		Channels: []string{defaultChannel},
		Options:  make(map[string]string),
	}

	// Set path for file-based drivers
	if path, exists := configMap["path"]; exists {
		loggerConfig.Path = path.(string)
	}

	// Set max files for daily driver
	if maxFiles, exists := configMap["days"]; exists {
		if maxFilesInt, ok := maxFiles.(int); ok {
			loggerConfig.MaxFiles = maxFilesInt
		}
	}

	// Set max size for file drivers
	if maxSize, exists := configMap["max_size"]; exists {
		if maxSizeInt, ok := maxSize.(int64); ok {
			loggerConfig.MaxSize = maxSizeInt
		}
	}

	// Handle stack driver
	if loggerConfig.Driver == "stack" {
		stackChannels := configMap["channels"].([]interface{})
		channelsList := make([]string, len(stackChannels))
		for i, ch := range stackChannels {
			channelsList[i] = ch.(string)
		}
		loggerConfig.Channels = channelsList
	}

	// Create logger
	logger := NewLogger(loggerConfig)

	return logger, nil
}

// CreateLoggerForChannel creates a logger for a specific channel
func (f *LoggingFactory) CreateLoggerForChannel(channelName string) (LoggerInterface, error) {
	channels := f.config["channels"].(map[string]interface{})

	channelConfig, exists := channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel '%s' not found in configuration", channelName)
	}

	configMap := channelConfig.(map[string]interface{})

	// Create logger configuration
	loggerConfig := &LoggerConfig{
		Driver:   configMap["driver"].(string),
		Level:    ParseLogLevel(configMap["level"].(string)),
		Channels: []string{channelName},
		Options:  make(map[string]string),
	}

	// Set path for file-based drivers
	if path, exists := configMap["path"]; exists {
		loggerConfig.Path = path.(string)
	}

	// Create logger
	logger := NewLogger(loggerConfig)

	return logger, nil
}

// GetAvailableChannels returns all available logging channels
func (f *LoggingFactory) GetAvailableChannels() []string {
	channels := f.config["channels"].(map[string]interface{})
	channelNames := make([]string, 0, len(channels))

	for channelName := range channels {
		channelNames = append(channelNames, channelName)
	}

	return channelNames
}

// IsChannelAvailable checks if a channel is available
func (f *LoggingFactory) IsChannelAvailable(channelName string) bool {
	channels := f.config["channels"].(map[string]interface{})
	_, exists := channels[channelName]
	return exists
}

// GetChannelConfig returns the configuration for a specific channel
func (f *LoggingFactory) GetChannelConfig(channelName string) (map[string]interface{}, error) {
	channels := f.config["channels"].(map[string]interface{})

	channelConfig, exists := channels[channelName]
	if !exists {
		return nil, fmt.Errorf("channel '%s' not found in configuration", channelName)
	}

	return channelConfig.(map[string]interface{}), nil
}

// ValidateConfiguration validates the logging configuration
func (f *LoggingFactory) ValidateConfiguration() error {
	// Check if default channel exists
	defaultChannel := f.config["default"].(string)
	channels := f.config["channels"].(map[string]interface{})

	if _, exists := channels[defaultChannel]; !exists {
		return fmt.Errorf("default channel '%s' not found in channels configuration", defaultChannel)
	}

	// Validate each channel configuration
	for channelName, channelConfig := range channels {
		configMap := channelConfig.(map[string]interface{})

		// Check required fields
		if _, exists := configMap["driver"]; !exists {
			return fmt.Errorf("channel '%s' missing required field 'driver'", channelName)
		}

		if _, exists := configMap["level"]; !exists {
			return fmt.Errorf("channel '%s' missing required field 'level'", channelName)
		}

		// Validate driver
		driver := configMap["driver"].(string)
		validDrivers := []string{"single", "daily", "sentry", "stack", "null"}
		isValidDriver := false

		for _, validDriver := range validDrivers {
			if driver == validDriver {
				isValidDriver = true
				break
			}
		}

		if !isValidDriver {
			return fmt.Errorf("channel '%s' has invalid driver '%s'", channelName, driver)
		}

		// Validate level
		level := configMap["level"].(string)
		validLevels := []string{"debug", "info", "warning", "error", "fatal"}
		isValidLevel := false

		for _, validLevel := range validLevels {
			if level == validLevel {
				isValidLevel = true
				break
			}
		}

		if !isValidLevel {
			return fmt.Errorf("channel '%s' has invalid level '%s'", channelName, level)
		}

		// Validate stack driver configuration
		if driver == "stack" {
			if stackChannels, exists := configMap["channels"]; exists {
				channelsList := stackChannels.([]interface{})
				for _, ch := range channelsList {
					channelName := ch.(string)
					if _, exists := channels[channelName]; !exists {
						return fmt.Errorf("stack channel '%s' references non-existent channel '%s'", channelName, channelName)
					}
				}
			} else {
				return fmt.Errorf("stack channel '%s' missing required field 'channels'", channelName)
			}
		}

		// Validate file-based drivers
		if driver == "single" || driver == "daily" {
			if _, exists := configMap["path"]; !exists {
				return fmt.Errorf("file-based channel '%s' missing required field 'path'", channelName)
			}
		}
	}

	return nil
}

// Global logging factory instance
var LoggingFactoryInstance *LoggingFactory

// InitializeLogging initializes the global logging system
func InitializeLogging() error {
	factory := NewLoggingFactory()

	// Validate configuration
	if err := factory.ValidateConfiguration(); err != nil {
		return fmt.Errorf("invalid logging configuration: %v", err)
	}

	// Create logger
	logger, err := factory.CreateLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %v", err)
	}

	// Set global instances
	LoggingFactoryInstance = factory
	SetLogger(logger)

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() LoggerInterface {
	return LoggerInstance
}

// GetLoggerForChannel returns a logger for a specific channel
func GetLoggerForChannel(channelName string) (LoggerInterface, error) {
	if LoggingFactoryInstance == nil {
		return nil, fmt.Errorf("logging factory not initialized")
	}

	return LoggingFactoryInstance.CreateLoggerForChannel(channelName)
}

// Add this helper at the top or near the LogLevel definition
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warning":
		return LogLevelWarning
	case "error":
		return LogLevelError
	case "fatal":
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}
