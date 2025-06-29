package queue_core

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	app_core "base_lara_go_project/app/core/app"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// QueueWorker handles queue processing for multiple queues
type QueueWorker struct {
	ctx              context.Context
	cancel           context.CancelFunc
	enabledQueues    []string
	queueService     app_core.QueueService
	jobDispatcher    app_core.JobDispatcherService
	messageProcessor app_core.MessageProcessorService
	config           map[string]interface{}
	processedJobs    int64
	startTime        time.Time
}

// NewQueueWorker creates a new queue worker
func NewQueueWorker(enabledQueues []string, queueService app_core.QueueService, jobDispatcher app_core.JobDispatcherService, messageProcessor app_core.MessageProcessorService, config map[string]interface{}) *QueueWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueueWorker{
		ctx:              ctx,
		cancel:           cancel,
		enabledQueues:    enabledQueues,
		queueService:     queueService,
		jobDispatcher:    jobDispatcher,
		messageProcessor: messageProcessor,
		config:           config,
		processedJobs:    0,
		startTime:        time.Now(),
	}
}

// Start starts the queue worker
func (w *QueueWorker) Start() {
	log.Printf("Starting queue worker for queues: %s", strings.Join(w.enabledQueues, ", "))

	// Get configuration values
	sleepTime := w.getConfigInt("sleep", 3)
	maxJobs := w.getConfigInt("max_jobs", 1000)
	memoryLimit := w.getConfigInt("memory_limit", 128)
	timeout := w.getConfigInt("timeout", 60)
	tries := w.getConfigInt("tries", 3)

	log.Printf("Worker configuration: sleep=%ds, max_jobs=%d, memory_limit=%dMB, timeout=%ds, tries=%d",
		sleepTime, maxJobs, memoryLimit, timeout, tries)

	for {
		select {
		case <-w.ctx.Done():
			log.Println("Queue worker stopped")
			return
		default:
			// Check memory usage
			if w.shouldRestart(memoryLimit) {
				log.Printf("Memory limit reached (%dMB), restarting worker", memoryLimit)
				return
			}

			// Check if max jobs reached
			if w.processedJobs >= int64(maxJobs) {
				log.Printf("Max jobs reached (%d), stopping worker", maxJobs)
				return
			}

			// Check if timeout reached
			if time.Since(w.startTime) > time.Duration(timeout)*time.Second {
				log.Printf("Timeout reached (%ds), stopping worker", timeout)
				return
			}

			w.processAllQueues()
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}
}

// processAllQueues processes messages from all enabled queues
func (w *QueueWorker) processAllQueues() {
	// Process all queues concurrently
	var wg sync.WaitGroup
	for _, queueName := range w.enabledQueues {
		wg.Add(1)
		go func(queue string) {
			defer wg.Done()
			if err := w.processQueue(queue); err != nil {
				log.Printf("Error processing queue %s: %v", queue, err)
			}
		}(queueName)
	}
	wg.Wait()
}

// processQueue processes messages from a specific queue
func (w *QueueWorker) processQueue(queueName string) error {
	// Receive messages from the queue
	result, err := w.queueService.ReceiveMessageFromQueue(queueName)
	if err != nil {
		return err
	}

	// Type assertion for SQS result
	sqsResult, ok := result.(*sqs.ReceiveMessageOutput)
	if !ok {
		return fmt.Errorf("unexpected result type from queue service")
	}

	if len(sqsResult.Messages) > 0 {
		log.Printf("Processing %d messages from queue %s", len(sqsResult.Messages), queueName)

		// Process messages concurrently
		var wg sync.WaitGroup
		for _, message := range sqsResult.Messages {
			wg.Add(1)
			go func(msg types.Message) {
				defer wg.Done()
				if err := w.processMessageWithQueue(&msg, queueName); err != nil {
					log.Printf("Error processing message from queue %s: %v", queueName, err)
				} else {
					// Increment processed jobs counter
					w.processedJobs++
				}
			}(message)
		}
		wg.Wait()
	}

	return nil
}

// processMessageWithQueue processes a message with queue context
func (w *QueueWorker) processMessageWithQueue(message *types.Message, queueName string) error {
	if message.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	jobType := w.messageProcessor.GetJobTypeFromMessage(message)
	tries := w.getConfigInt("tries", 3)

	// Process the job with retry logic
	var err error
	for attempt := 1; attempt <= tries; attempt++ {
		err = w.jobDispatcher.ProcessJobFromQueue([]byte(*message.Body), jobType)
		if err == nil {
			break
		}

		if attempt < tries {
			log.Printf("Job processing failed (attempt %d/%d): %v, retrying...", attempt, tries, err)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}
	}

	if err != nil {
		log.Printf("Job processing failed after %d attempts: %v", tries, err)
		return err
	}

	// Delete the message from the queue after successful processing
	err = w.queueService.DeleteMessageFromQueue(*message.ReceiptHandle, queueName)
	if err != nil {
		log.Printf("Error deleting message from queue: %v", err)
		return err
	}

	return nil
}

// Stop stops the queue worker
func (w *QueueWorker) Stop() {
	w.cancel()
}

// GetStats returns worker statistics
func (w *QueueWorker) GetStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"processed_jobs": w.processedJobs,
		"uptime":         time.Since(w.startTime).String(),
		"memory_usage":   m.Alloc / 1024 / 1024, // MB
		"queues":         w.enabledQueues,
	}
}

// shouldRestart checks if the worker should restart due to memory limits
func (w *QueueWorker) shouldRestart(memoryLimit int) bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryUsageMB := m.Alloc / 1024 / 1024
	return memoryUsageMB > uint64(memoryLimit)
}

// getConfigInt gets an integer configuration value with fallback
func (w *QueueWorker) getConfigInt(key string, fallback int) int {
	if w.config == nil {
		return fallback
	}

	if value, exists := w.config[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}

	return fallback
}
