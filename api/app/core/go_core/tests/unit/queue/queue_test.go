package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"base_lara_go_project/app/core/go_core"
	"sync/atomic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data structures
type TestJob struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

func (j TestJob) Handle(ctx context.Context) error {
	// Simulate some work
	time.Sleep(1 * time.Millisecond)
	return nil
}

type TestQueueableJob struct {
	TestJob
	shouldQueue bool
	maxAttempts int
	retryDelay  int
}

func (j TestQueueableJob) ShouldQueue() bool {
	return j.shouldQueue
}

func (j TestQueueableJob) GetQueueName() string {
	return "test-queue"
}

func (j TestQueueableJob) GetMaxAttempts() int {
	return j.maxAttempts
}

func (j TestQueueableJob) GetRetryDelay() int {
	return j.retryDelay
}

// Helper function to create a job
func createJob(id, data string) *go_core.Job[TestJob] {
	return &go_core.Job[TestJob]{
		ID:         id,
		Data:       TestJob{ID: id, Data: data},
		Attempts:   0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}

// TestQueueInterface tests the basic queue interface
func TestQueueInterface(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	// Test Push and Pop
	job := createJob("test-1", "test data")
	err := queue.Push(job)
	require.NoError(t, err)

	poppedJob, err := queue.Pop()
	require.NoError(t, err)
	assert.Equal(t, job.ID, poppedJob.ID)
	assert.Equal(t, job.Data, poppedJob.Data)

	// Test empty queue
	emptyJob, err := queue.Pop()
	require.NoError(t, err)
	assert.Nil(t, emptyJob)
}

// TestQueueContextCancellation tests context cancellation for all operations
func TestQueueContextCancellation(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	// Test PushWithContext cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	job := createJob("test-1", "test data")
	err := queue.PushWithContext(ctx, job)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test PopWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = queue.PopWithContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test DeleteWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = queue.DeleteWithContext(ctx, "test-1")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test PushManyWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	jobs := []*go_core.Job[TestJob]{
		createJob("test-1", "data1"),
		createJob("test-2", "data2"),
	}
	err = queue.PushManyWithContext(ctx, jobs)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test PopManyWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = queue.PopManyWithContext(ctx, 5)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test RetryWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	job = createJob("test-1", "test data")
	err = queue.RetryWithContext(ctx, job)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test FailWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = queue.FailWithContext(ctx, job, fmt.Errorf("test error"))
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test SizeWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = queue.SizeWithContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Test ClearWithContext cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = queue.ClearWithContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestQueueContextTimeout tests context timeout for all operations
func TestQueueContextTimeout(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	// Test PushWithContext timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout

	job := createJob("test-1", "test data")
	err := queue.PushWithContext(ctx, job)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)

	// Test PopWithContext timeout
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout

	_, err = queue.PopWithContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

// TestQueueRaceConditions tests race conditions with context-aware operations
func TestQueueRaceConditions(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent PushWithContext operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			job := createJob(fmt.Sprintf("key-%d", id), fmt.Sprintf("value-%d", id))
			_ = queue.PushWithContext(ctx, job)
		}(i)
	}
	wg.Wait()

	// Verify all jobs were pushed
	size, err := queue.Size()
	require.NoError(t, err)
	assert.Equal(t, int64(numGoroutines), size)

	// Test concurrent PopWithContext operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			_, _ = queue.PopWithContext(ctx)
		}()
	}
	wg.Wait()

	// Verify all jobs were popped
	size, err = queue.Size()
	require.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

// TestQueueMixedOperations tests mixing context-aware and non-context operations
func TestQueueMixedOperations(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	// Push using non-context method
	job := createJob("test", "value")
	err := queue.Push(job)
	assert.NoError(t, err)

	// Pop using context method
	ctx := context.Background()
	poppedJob, err := queue.PopWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, job.ID, poppedJob.ID)

	// Push using context method
	err = queue.PushWithContext(ctx, job)
	assert.NoError(t, err)

	// Pop using non-context method
	poppedJob, err = queue.Pop()
	assert.NoError(t, err)
	assert.Equal(t, job.ID, poppedJob.ID)
}

// TestQueueBatchOperations tests context-aware batch operations
func TestQueueBatchOperations(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	ctx := context.Background()
	jobs := []*go_core.Job[TestJob]{
		createJob("batch1", "value1"),
		createJob("batch2", "value2"),
		createJob("batch3", "value3"),
	}

	// Push many jobs
	err := queue.PushManyWithContext(ctx, jobs)
	assert.NoError(t, err)

	// Verify size
	size, err := queue.SizeWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), size)

	// Pop many jobs
	poppedJobs, err := queue.PopManyWithContext(ctx, 2)
	assert.NoError(t, err)
	assert.Len(t, poppedJobs, 2)

	// Verify remaining size
	size, err = queue.SizeWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), size)

	// Pop remaining job
	poppedJobs, err = queue.PopManyWithContext(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, poppedJobs, 1)

	// Verify empty queue
	size, err = queue.SizeWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

// TestQueueRetryLogic tests retry logic with context support
func TestQueueRetryLogic(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	ctx := context.Background()
	job := createJob("retry-test", "data")
	job.MaxRetries = 2

	// Push job
	err := queue.PushWithContext(ctx, job)
	assert.NoError(t, err)

	// Pop and retry
	poppedJob, err := queue.PopWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, poppedJob.Attempts)

	// Retry job
	err = queue.RetryWithContext(ctx, poppedJob)
	assert.NoError(t, err)

	// Pop again and verify attempt count
	poppedJob, err = queue.PopWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, poppedJob.Attempts)

	// Retry again
	err = queue.RetryWithContext(ctx, poppedJob)
	assert.NoError(t, err)

	// Pop and verify attempt count
	poppedJob, err = queue.PopWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, poppedJob.Attempts)

	// Try to retry beyond max attempts
	err = queue.RetryWithContext(ctx, poppedJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeded max retries")
}

// TestQueueDeleteOperations tests delete operations with context support
func TestQueueDeleteOperations(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	ctx := context.Background()
	jobs := []*go_core.Job[TestJob]{
		createJob("delete1", "value1"),
		createJob("delete2", "value2"),
		createJob("delete3", "value3"),
	}

	// Push jobs
	err := queue.PushManyWithContext(ctx, jobs)
	assert.NoError(t, err)

	// Delete specific job
	err = queue.DeleteWithContext(ctx, "delete2")
	assert.NoError(t, err)

	// Verify size
	size, err := queue.SizeWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), size)

	// Try to delete non-existent job
	err = queue.DeleteWithContext(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestQueueClearOperations tests clear operations with context support
func TestQueueClearOperations(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	ctx := context.Background()
	jobs := []*go_core.Job[TestJob]{
		createJob("clear1", "value1"),
		createJob("clear2", "value2"),
		createJob("clear3", "value3"),
	}

	// Push jobs
	err := queue.PushManyWithContext(ctx, jobs)
	assert.NoError(t, err)

	// Verify size
	size, err := queue.SizeWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), size)

	// Clear queue
	err = queue.ClearWithContext(ctx)
	assert.NoError(t, err)

	// Verify empty
	size, err = queue.SizeWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

// TestQueueWithContext tests the WithContext method
func TestQueueWithContext(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	// Create context-aware queue
	ctx := context.Background()
	contextQueue := queue.WithContext(ctx)

	// Test operations on context-aware queue
	job := createJob("context-test", "data")
	err := contextQueue.Push(job)
	assert.NoError(t, err)

	poppedJob, err := contextQueue.Pop()
	assert.NoError(t, err)
	assert.Equal(t, job.ID, poppedJob.ID)
}

// TestQueueEdgeCases tests edge cases and error conditions
func TestQueueEdgeCases(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()

	ctx := context.Background()

	// Test PopMany with count 0
	jobs, err := queue.PopManyWithContext(ctx, 0)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)

	// Test PopMany with negative count
	jobs, err = queue.PopManyWithContext(ctx, -1)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)

	// Test PushMany with empty slice
	err = queue.PushManyWithContext(ctx, []*go_core.Job[TestJob]{})
	assert.NoError(t, err)

	// Test PopMany on empty queue
	jobs, err = queue.PopManyWithContext(ctx, 5)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)
}

// TestQueueConcurrencyStress tests stress testing with high concurrency
func TestQueueConcurrencyStress(t *testing.T) {
	queue := go_core.NewSyncQueue[TestJob]()
	var wg sync.WaitGroup
	numGoroutines := 1000
	numOperations := 10

	// Track total jobs pushed and popped
	var totalPushed int64
	var totalPopped int64

	// Stress test with mixed operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()

			for j := 0; j < numOperations; j++ {
				job := createJob(fmt.Sprintf("stress-%d-%d", id, j), fmt.Sprintf("data-%d-%d", id, j))
				err := queue.PushWithContext(ctx, job)
				if err != nil {
					t.Errorf("Push failed: %v", err)
					return
				}
				atomic.AddInt64(&totalPushed, 1)

				poppedJob, err := queue.PopWithContext(ctx)
				if err != nil {
					t.Errorf("Pop failed: %v", err)
					return
				}

				if poppedJob == nil {
					t.Errorf("Expected job, got nil")
					return
				}

				// Don't check specific job ID - just verify we got a valid job
				if poppedJob.ID == "" {
					t.Errorf("Got job with empty ID")
					return
				}

				atomic.AddInt64(&totalPopped, 1)
			}
		}(i)
	}
	wg.Wait()

	// Verify queue is empty
	size, err := queue.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), size)

	// Verify we pushed and popped the same number of jobs
	assert.Equal(t, totalPushed, totalPopped)
	assert.Equal(t, int64(numGoroutines*numOperations), totalPushed)
}
