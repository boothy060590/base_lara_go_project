package core

import (
	"context"
	"log"
	"time"
)

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	ProcessMessages() error
}

// QueueWorker handles generic queue processing
type QueueWorker struct {
	ctx       context.Context
	cancel    context.CancelFunc
	processor MessageProcessor
}

// NewQueueWorker creates a new queue worker
func NewQueueWorker(processor MessageProcessor) *QueueWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueueWorker{
		ctx:       ctx,
		cancel:    cancel,
		processor: processor,
	}
}

// Start starts the queue worker
func (w *QueueWorker) Start() {
	log.Println("Starting queue worker...")

	for {
		select {
		case <-w.ctx.Done():
			log.Println("Queue worker stopped")
			return
		default:
			if err := w.processor.ProcessMessages(); err != nil {
				log.Printf("Error processing messages: %v", err)
			}
			time.Sleep(1 * time.Second) // Poll every second
		}
	}
}

// Stop stops the queue worker
func (w *QueueWorker) Stop() {
	w.cancel()
}
