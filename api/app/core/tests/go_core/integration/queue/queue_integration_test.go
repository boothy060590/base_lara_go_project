package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"base_lara_go_project/app/core/go_core"

	"github.com/stretchr/testify/assert"
)

// Test data structures for integration tests
type IntegrationTestJob struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

func (j IntegrationTestJob) Handle(ctx context.Context) error {
	// Simulate some work
	time.Sleep(1 * time.Millisecond)
	return nil
}

// Implement JobTraits interface
func (j IntegrationTestJob) ShouldQueue() bool {
	return true // Default to queuing
}

func (j IntegrationTestJob) GetQueueName() string {
	return "integration-test-queue"
}

func (j IntegrationTestJob) GetMaxAttempts() int {
	return 3
}

func (j IntegrationTestJob) GetRetryDelay() int {
	return 5
}

type IntegrationQueueableJob struct {
	IntegrationTestJob
	shouldQueue bool
	maxAttempts int
	retryDelay  int
}

func (j IntegrationQueueableJob) ShouldQueue() bool {
	return j.shouldQueue
}

func (j IntegrationQueueableJob) GetQueueName() string {
	return "integration-test-queue"
}

func (j IntegrationQueueableJob) GetMaxAttempts() int {
	return j.maxAttempts
}

func (j IntegrationQueueableJob) GetRetryDelay() int {
	return j.retryDelay
}

// Helper function to create integration test jobs
func createIntegrationJob(id, data string) *go_core.Job[IntegrationTestJob] {
	return &go_core.Job[IntegrationTestJob]{
		ID:         id,
		Data:       IntegrationTestJob{ID: id, Data: data},
		Attempts:   0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}

// TestJobDispatcherIntegration tests the integration between JobDispatcher and Queue
func TestJobDispatcherIntegration(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Create job dispatcher with performance optimizations
	dispatcher := go_core.NewJobDispatcher[IntegrationTestJob](queue, nil, nil, nil)

	// Test dispatching a job that should be queued (default behavior)
	queueableJob := IntegrationTestJob{ID: "test-1", Data: "test data"}

	err := dispatcher.Dispatch(queueableJob)
	assert.NoError(t, err)

	// Verify job was queued
	size, err := queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), size)

	// Test DispatchSync (should always execute synchronously)
	err = dispatcher.DispatchSync(queueableJob)
	assert.NoError(t, err)

	// Verify no additional job was queued
	size, err = queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), size)
}

// TestQueueWorkerIntegration tests the integration between QueueWorker and Queue
func TestQueueWorkerIntegration(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Track processed jobs
	var processedJobs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create job handler
	handler := func(ctx context.Context, job *go_core.Job[IntegrationTestJob]) error {
		mu.Lock()
		defer mu.Unlock()
		processedJobs = append(processedJobs, job.ID)
		wg.Done() // Signal that a job was processed
		return nil
	}

	// Create queue worker
	worker := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)

	// Add some jobs to the queue
	jobs := []*go_core.Job[IntegrationTestJob]{
		createIntegrationJob("worker-1", "data1"),
		createIntegrationJob("worker-2", "data2"),
		createIntegrationJob("worker-3", "data3"),
	}

	// Set up wait group for job processing
	wg.Add(len(jobs))

	for _, job := range jobs {
		err := queue.Push(job)
		assert.NoError(t, err)
	}

	// Start worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker failed: %v", err)
		}
	}()

	// Wait for all jobs to be processed
	wg.Wait()

	// Stop worker
	worker.Stop()

	// Verify all jobs were processed
	assert.Len(t, processedJobs, 3)
	assert.Contains(t, processedJobs, "worker-1")
	assert.Contains(t, processedJobs, "worker-2")
	assert.Contains(t, processedJobs, "worker-3")

	// Verify queue is empty
	size, err := queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

// TestQueueWorkerWithContextIntegration tests queue worker with context support
func TestQueueWorkerWithContextIntegration(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Track processed jobs
	var processedJobs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create job handler
	handler := func(ctx context.Context, job *go_core.Job[IntegrationTestJob]) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		mu.Lock()
		defer mu.Unlock()
		processedJobs = append(processedJobs, job.ID)
		wg.Done() // Signal that a job was processed
		return nil
	}

	// Create queue worker
	worker := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)

	// Add some jobs to the queue
	jobs := []*go_core.Job[IntegrationTestJob]{
		createIntegrationJob("ctx-1", "data1"),
		createIntegrationJob("ctx-2", "data2"),
	}

	// Set up wait group for job processing
	wg.Add(len(jobs))

	for _, job := range jobs {
		err := queue.PushWithContext(context.Background(), job)
		assert.NoError(t, err)
	}

	// Start worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker failed: %v", err)
		}
	}()

	// Wait for all jobs to be processed
	wg.Wait()

	// Stop worker
	worker.Stop()

	// Verify jobs were processed
	assert.Len(t, processedJobs, 2)
	assert.Contains(t, processedJobs, "ctx-1")
	assert.Contains(t, processedJobs, "ctx-2")
}

// TestQueueWorkerContextCancellation tests queue worker with context cancellation
func TestQueueWorkerContextCancellation(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Track processed jobs
	var processedJobs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create job handler that respects context
	handler := func(ctx context.Context, job *go_core.Job[IntegrationTestJob]) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		mu.Lock()
		defer mu.Unlock()
		processedJobs = append(processedJobs, job.ID)
		wg.Done() // Signal that a job was processed
		return nil
	}

	// Create queue worker
	worker := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)

	// Add a job to the queue
	job := createIntegrationJob("cancel-test", "data")
	err := queue.Push(job)
	assert.NoError(t, err)

	// Set up wait group for job processing
	wg.Add(1)

	// Start worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker failed: %v", err)
		}
	}()

	// Wait for job to be processed
	wg.Wait()

	// Stop worker (this cancels the context)
	worker.Stop()

	// Verify job was processed before cancellation
	assert.Len(t, processedJobs, 1)
	assert.Contains(t, processedJobs, "cancel-test")
}

// TestQueueRetryIntegration tests retry logic integration
func TestQueueRetryIntegration(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Track processing attempts
	var attempts map[string]int
	var mu sync.Mutex
	var wg sync.WaitGroup
	attempts = make(map[string]int)

	// Create job handler that fails on first attempt
	handler := func(ctx context.Context, job *go_core.Job[IntegrationTestJob]) error {
		mu.Lock()
		defer mu.Unlock()

		attempts[job.ID]++

		// Fail on first attempt, succeed on second
		if attempts[job.ID] == 1 {
			return fmt.Errorf("simulated failure")
		}

		// Signal completion on successful attempt
		wg.Done()
		return nil
	}

	// Create queue worker
	worker := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)

	// Add a job to the queue
	job := createIntegrationJob("retry-test", "data")
	job.MaxRetries = 2
	err := queue.Push(job)
	assert.NoError(t, err)

	// Set up wait group for successful completion
	wg.Add(1)

	// Start worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker failed: %v", err)
		}
	}()

	// Wait for successful completion
	wg.Wait()

	// Stop worker
	worker.Stop()

	// Verify job was retried and eventually succeeded
	assert.Equal(t, 2, attempts["retry-test"])

	// Verify queue is empty
	size, err := queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

// TestQueueBatchOperationsIntegration tests batch operations integration
func TestQueueBatchOperationsIntegration(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Track processed jobs
	var processedJobs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create job handler
	handler := func(ctx context.Context, job *go_core.Job[IntegrationTestJob]) error {
		mu.Lock()
		defer mu.Unlock()
		processedJobs = append(processedJobs, job.ID)
		wg.Done() // Signal that a job was processed
		return nil
	}

	// Create queue worker
	worker := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)

	// Add jobs in batches
	batch1 := []*go_core.Job[IntegrationTestJob]{
		createIntegrationJob("batch1-1", "data1"),
		createIntegrationJob("batch1-2", "data2"),
		createIntegrationJob("batch1-3", "data3"),
	}

	batch2 := []*go_core.Job[IntegrationTestJob]{
		createIntegrationJob("batch2-1", "data4"),
		createIntegrationJob("batch2-2", "data5"),
	}

	// Set up wait group for job processing
	wg.Add(len(batch1) + len(batch2))

	// Push batches
	err := queue.PushManyWithContext(context.Background(), batch1)
	assert.NoError(t, err)

	err = queue.PushManyWithContext(context.Background(), batch2)
	assert.NoError(t, err)

	// Verify total size
	size, err := queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), size)

	// Start worker in a goroutine
	go func() {
		err := worker.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker failed: %v", err)
		}
	}()

	// Wait for all jobs to be processed
	wg.Wait()

	// Stop worker
	worker.Stop()

	// Verify all jobs were processed
	assert.Len(t, processedJobs, 5)
	assert.Contains(t, processedJobs, "batch1-1")
	assert.Contains(t, processedJobs, "batch1-2")
	assert.Contains(t, processedJobs, "batch1-3")
	assert.Contains(t, processedJobs, "batch2-1")
	assert.Contains(t, processedJobs, "batch2-2")

	// Verify queue is empty
	size, err = queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

// TestQueueConcurrentWorkersIntegration tests multiple workers processing the same queue
func TestQueueConcurrentWorkersIntegration(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Track processed jobs
	var processedJobs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create job handler
	handler := func(ctx context.Context, job *go_core.Job[IntegrationTestJob]) error {
		mu.Lock()
		defer mu.Unlock()
		processedJobs = append(processedJobs, job.ID)
		wg.Done() // Signal that a job was processed
		return nil
	}

	// Create multiple workers
	worker1 := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)
	worker2 := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)
	worker3 := go_core.NewQueueWorker[IntegrationTestJob](queue, handler, nil, nil, nil)

	// Add jobs to the queue
	numJobs := 30
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		job := createIntegrationJob(fmt.Sprintf("concurrent-%d", i), fmt.Sprintf("data-%d", i))
		err := queue.Push(job)
		assert.NoError(t, err)
	}

	// Start all workers
	go func() {
		err := worker1.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker1 failed: %v", err)
		}
	}()

	go func() {
		err := worker2.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker2 failed: %v", err)
		}
	}()

	go func() {
		err := worker3.Start()
		if err != nil && err != context.Canceled {
			t.Errorf("Worker3 failed: %v", err)
		}
	}()

	// Wait for all jobs to be processed
	wg.Wait()

	// Stop all workers
	worker1.Stop()
	worker2.Stop()
	worker3.Stop()

	// Verify all jobs were processed
	assert.Len(t, processedJobs, numJobs)

	// Verify queue is empty
	size, err := queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

// TestQueueWithContextMethodIntegration tests the WithContext method integration
func TestQueueWithContextMethodIntegration(t *testing.T) {
	queue := go_core.NewSyncQueue[IntegrationTestJob]()

	// Create context-aware queue
	ctx := context.Background()
	contextQueue := queue.WithContext(ctx)

	// Test operations on context-aware queue
	job := createIntegrationJob("with-context", "data")
	err := contextQueue.Push(job)
	assert.NoError(t, err)

	// Verify job was added to original queue
	size, err := queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), size)

	// Pop from context queue
	poppedJob, err := contextQueue.Pop()
	assert.NoError(t, err)
	assert.Equal(t, job.ID, poppedJob.ID)

	// Verify original queue is empty
	size, err = queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)
}
