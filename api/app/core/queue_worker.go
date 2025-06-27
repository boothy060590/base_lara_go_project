package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// QueueWorker handles queue processing for multiple queues
type QueueWorker struct {
	ctx           context.Context
	cancel        context.CancelFunc
	enabledQueues []string
}

// NewQueueWorker creates a new queue worker
func NewQueueWorker(enabledQueues []string) *QueueWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueueWorker{
		ctx:           ctx,
		cancel:        cancel,
		enabledQueues: enabledQueues,
	}
}

// Start starts the queue worker
func (w *QueueWorker) Start() {
	log.Printf("Starting queue worker for queues: %s", strings.Join(w.enabledQueues, ", "))

	for {
		select {
		case <-w.ctx.Done():
			log.Println("Queue worker stopped")
			return
		default:
			w.processAllQueues()
			time.Sleep(50 * time.Millisecond) // Poll every 50ms
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
	result, err := ReceiveMessageFromQueue(queueName)
	if err != nil {
		return err
	}

	if len(result.Messages) > 0 {
		log.Printf("Processing %d messages from queue %s", len(result.Messages), queueName)

		// Process messages concurrently
		var wg sync.WaitGroup
		for _, message := range result.Messages {
			wg.Add(1)
			go func(msg types.Message) {
				defer wg.Done()
				if err := w.processMessageWithQueue(&msg, queueName); err != nil {
					log.Printf("Error processing message from queue %s: %v", queueName, err)
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

	jobType := GetJobTypeFromMessage(message)

	// Process the job based on its type
	err := ProcessJobFromQueue([]byte(*message.Body), jobType)
	if err != nil {
		log.Printf("Error processing job: %v", err)
		return err
	}

	// Delete the message from the queue after successful processing
	err = DeleteMessageFromQueue(*message.ReceiptHandle, queueName)
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
