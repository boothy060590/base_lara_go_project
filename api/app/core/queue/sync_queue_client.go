package queue_core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	app_core "base_lara_go_project/app/core/app"
)

// SyncQueueClient provides synchronous queue functionality
type SyncQueueClient struct {
	*BaseQueueClient
	jobProcessor app_core.JobProcessor
	storagePath  string
}

// NewSyncQueueClient creates a new synchronous queue client
func NewSyncQueueClient(config *app_core.ClientConfig) *SyncQueueClient {
	storagePath := "storage/events"
	if configPath, ok := config.Options["storage_path"].(string); ok {
		storagePath = configPath
	}

	return &SyncQueueClient{
		BaseQueueClient: NewBaseQueueClient(config, "sync"),
		storagePath:     storagePath,
	}
}

// Connect establishes the queue connection (no-op for sync queue)
func (c *SyncQueueClient) Connect() error {
	// Ensure storage directories exist
	if err := c.ensureStorageDirectories(); err != nil {
		return fmt.Errorf("failed to create storage directories: %v", err)
	}
	return c.BaseClient.Connect()
}

// ensureStorageDirectories creates the necessary storage directories
func (c *SyncQueueClient) ensureStorageDirectories() error {
	directories := []string{
		filepath.Join(c.storagePath, "jobs", "events_started"),
		filepath.Join(c.storagePath, "jobs", "events_completed"),
		filepath.Join(c.storagePath, "jobs", "events_failed"),
		filepath.Join(c.storagePath, "events", "events_started"),
		filepath.Join(c.storagePath, "events", "events_completed"),
		filepath.Join(c.storagePath, "events", "events_failed"),
		filepath.Join(c.storagePath, "mail", "events_started"),
		filepath.Join(c.storagePath, "mail", "events_completed"),
		filepath.Join(c.storagePath, "mail", "events_failed"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// writeEventState writes an event state to the appropriate directory
func (c *SyncQueueClient) writeEventState(queue string, eventType string, job interface{}, metadata map[string]interface{}) error {
	// Determine the base directory based on queue type
	var baseDir string
	switch queue {
	case "jobs":
		baseDir = filepath.Join(c.storagePath, "jobs")
	case "events":
		baseDir = filepath.Join(c.storagePath, "events")
	case "mail":
		baseDir = filepath.Join(c.storagePath, "mail")
	default:
		baseDir = filepath.Join(c.storagePath, "jobs") // default to jobs
	}

	// Create the event state file
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.json", timestamp, eventType)
	filepath := filepath.Join(baseDir, eventType, filename)

	// Prepare the event data
	eventData := map[string]interface{}{
		"timestamp":  time.Now().Unix(),
		"queue":      queue,
		"event_type": eventType,
		"job":        job,
		"metadata":   metadata,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(eventData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write event file: %v", err)
	}

	return nil
}

// logToFile logs a message to storage/logs
func (c *SyncQueueClient) logToFile(level, message string, context map[string]interface{}) error {
	logDir := "storage/logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, "laravel.log")
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logEntry := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
	if len(context) > 0 {
		logEntry += fmt.Sprintf(" | Context: %+v", context)
	}
	logEntry += "\n"

	// Append to log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(logEntry)
	return err
}

// Disconnect closes the queue connection (no-op for sync queue)
func (c *SyncQueueClient) Disconnect() error {
	return c.BaseClient.Disconnect()
}

// Push adds a job to the queue (processes immediately)
func (c *SyncQueueClient) Push(queue string, job interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("queue client not connected")
	}

	// Log job started
	startContext := map[string]interface{}{
		"queue": queue,
		"job":   job,
	}
	c.logToFile("INFO", "Processing job synchronously", startContext)

	// Write job started event
	if err := c.writeEventState(queue, "events_started", job, startContext); err != nil {
		c.logToFile("ERROR", "Failed to write job started event", map[string]interface{}{"error": err.Error()})
	}

	// Process the job immediately
	if c.jobProcessor != nil {
		// Convert job to bytes for processing
		jobData := []byte(fmt.Sprintf("%v", job))

		// Process the job
		err := c.jobProcessor.Process(jobData)

		if err != nil {
			// Job failed
			failContext := map[string]interface{}{
				"queue": queue,
				"job":   job,
				"error": err.Error(),
			}
			c.logToFile("ERROR", "Job processing failed", failContext)

			// Write job failed event
			if writeErr := c.writeEventState(queue, "events_failed", job, failContext); writeErr != nil {
				c.logToFile("ERROR", "Failed to write job failed event", map[string]interface{}{"error": writeErr.Error()})
			}

			return err
		}

		// Job completed successfully
		completeContext := map[string]interface{}{
			"queue": queue,
			"job":   job,
		}
		c.logToFile("INFO", "Job processing completed successfully", completeContext)

		// Write job completed event
		if writeErr := c.writeEventState(queue, "events_completed", job, completeContext); writeErr != nil {
			c.logToFile("ERROR", "Failed to write job completed event", map[string]interface{}{"error": writeErr.Error()})
		}

		return nil
	}

	// No job processor registered
	noProcessorError := fmt.Errorf("no job processor registered")
	c.logToFile("ERROR", "No job processor registered", map[string]interface{}{
		"queue": queue,
		"job":   job,
	})

	// Write job failed event
	failContext := map[string]interface{}{
		"queue": queue,
		"job":   job,
		"error": noProcessorError.Error(),
	}
	if writeErr := c.writeEventState(queue, "events_failed", job, failContext); writeErr != nil {
		c.logToFile("ERROR", "Failed to write job failed event", map[string]interface{}{"error": writeErr.Error()})
	}

	return noProcessorError
}

// Pop retrieves a job from the queue (not applicable for sync)
func (c *SyncQueueClient) Pop(queue string) (interface{}, error) {
	return nil, fmt.Errorf("pop not supported for synchronous queue")
}

// Delete removes a job from the queue (not applicable for sync)
func (c *SyncQueueClient) Delete(queue string, job interface{}) error {
	return fmt.Errorf("delete not supported for synchronous queue")
}

// Size returns the number of jobs in the queue (always 0 for sync)
func (c *SyncQueueClient) Size(queue string) (int, error) {
	return 0, nil
}

// Clear clears all jobs from the queue (no-op for sync)
func (c *SyncQueueClient) Clear(queue string) error {
	return nil
}

// GetStats returns queue statistics
func (c *SyncQueueClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"status": "connected",
		"driver": "sync",
		"mode":   "synchronous",
	}
}

// SetJobProcessor sets the job processor for this queue
func (c *SyncQueueClient) SetJobProcessor(processor app_core.JobProcessor) {
	c.jobProcessor = processor
}
