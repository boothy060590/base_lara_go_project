package go_core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Job represents a generic job that can be queued
type Job[T any] struct {
	ID          string     `json:"id"`
	Data        T          `json:"data"`
	Attempts    int        `json:"attempts"`
	MaxRetries  int        `json:"max_retries"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// JobTraits defines traits that jobs can implement
type JobTraits interface {
	ShouldQueue() bool
	GetQueueName() string
	GetMaxAttempts() int
	GetRetryDelay() int
}

// QueueableJob represents a job that can be queued with traits
type QueueableJob[T any] struct {
	Job[T]
	Traits JobTraits
}

// Queue defines a generic queue interface for any job type
type Queue[T any] interface {
	// Basic operations
	Push(job *Job[T]) error
	Pop() (*Job[T], error)
	Delete(jobID string) error

	// Batch operations
	PushMany(jobs []*Job[T]) error
	PopMany(count int) ([]*Job[T], error)

	// Job management
	Retry(job *Job[T]) error
	Fail(job *Job[T], error error) error

	// Queue management
	Size() (int64, error)
	Clear() error
	WithContext(ctx context.Context) Queue[T]
}

// JobHandler defines a function that processes a job
type JobHandler[T any] func(ctx context.Context, job *Job[T]) error

// JobDispatcher defines a generic job dispatcher interface
type JobDispatcher[T any] interface {
	// Dispatch methods
	Dispatch(job T) error
	DispatchSync(job T) error

	// Queue management
	GetQueue() Queue[T]
	WithContext(ctx context.Context) JobDispatcher[T]
}

// jobDispatcher implements JobDispatcher[T]
type jobDispatcher[T any] struct {
	queue Queue[T]
	ctx   context.Context
}

// NewJobDispatcher creates a new job dispatcher
func NewJobDispatcher[T any](queue Queue[T]) JobDispatcher[T] {
	return &jobDispatcher[T]{
		queue: queue,
		ctx:   context.Background(),
	}
}

// Dispatch dispatches a job (respects ShouldQueue trait)
func (d *jobDispatcher[T]) Dispatch(job T) error {
	// Check if job implements JobTraits
	if queueableJob, ok := any(job).(JobTraits); ok && queueableJob.ShouldQueue() {
		// Queue the job
		goJob := &Job[T]{
			ID:         "job_" + time.Now().Format("20060102150405"),
			Data:       job,
			Attempts:   0,
			MaxRetries: queueableJob.GetMaxAttempts(),
			CreatedAt:  time.Now(),
		}
		return d.queue.Push(goJob)
	}

	// Execute synchronously
	return d.executeJob(job)
}

// DispatchSync dispatches a job synchronously (ignores ShouldQueue trait)
func (d *jobDispatcher[T]) DispatchSync(job T) error {
	return d.executeJob(job)
}

// executeJob executes a job synchronously
func (d *jobDispatcher[T]) executeJob(job T) error {
	// For now, we'll just call the Handle method if it exists
	if handler, ok := any(job).(interface {
		Handle(ctx context.Context) error
	}); ok {
		return handler.Handle(d.ctx)
	}

	// If no Handle method, just return success
	return nil
}

// GetQueue returns the underlying queue
func (d *jobDispatcher[T]) GetQueue() Queue[T] {
	return d.queue
}

// WithContext returns a dispatcher with context
func (d *jobDispatcher[T]) WithContext(ctx context.Context) JobDispatcher[T] {
	return &jobDispatcher[T]{
		queue: d.queue.WithContext(ctx),
		ctx:   ctx,
	}
}

// QueueWorker processes jobs from a queue
type QueueWorker[T any] struct {
	queue   Queue[T]
	handler JobHandler[T]
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewQueueWorker creates a new queue worker
func NewQueueWorker[T any](queue Queue[T], handler JobHandler[T]) *QueueWorker[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueueWorker[T]{
		queue:   queue,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start begins processing jobs
func (w *QueueWorker[T]) Start() error {
	for {
		select {
		case <-w.ctx.Done():
			return w.ctx.Err()
		default:
			// Pop job from queue
			job, err := w.queue.Pop()
			if err != nil {
				// No jobs available, wait a bit
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Process job
			err = w.processJob(job)
			if err != nil {
				// Handle job failure
				w.handleJobFailure(job, err)
			}
		}
	}
}

// Stop stops the worker
func (w *QueueWorker[T]) Stop() {
	w.cancel()
}

// processJob processes a single job
func (w *QueueWorker[T]) processJob(job *Job[T]) error {
	// Update processed time
	now := time.Now()
	job.ProcessedAt = &now

	// Call handler
	return w.handler(w.ctx, job)
}

// handleJobFailure handles job processing failures
func (w *QueueWorker[T]) handleJobFailure(job *Job[T], err error) {
	job.Attempts++

	if job.Attempts <= job.MaxRetries {
		// Retry job
		w.queue.Retry(job)
	} else {
		// Job failed permanently
		w.queue.Fail(job, err)
	}
}

// redisQueue implements Queue[T] with Redis
type redisQueue[T any] struct {
	client    *redis.Client
	queueName string
	ctx       context.Context
}

// NewRedisQueue creates a new Redis queue instance
func NewRedisQueue[T any](client *redis.Client, queueName string) Queue[T] {
	return &redisQueue[T]{
		client:    client,
		queueName: queueName,
		ctx:       context.Background(),
	}
}

// Push adds a job to the queue
func (q *redisQueue[T]) Push(job *Job[T]) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return q.client.LPush(q.ctx, q.queueName, data).Err()
}

// Pop retrieves a job from the queue
func (q *redisQueue[T]) Pop() (*Job[T], error) {
	result, err := q.client.BRPop(q.ctx, 0, q.queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	var job Job[T]
	err = json.Unmarshal([]byte(result[1]), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// Delete removes a job from the queue
func (q *redisQueue[T]) Delete(jobID string) error {
	// For Redis, we can't delete specific jobs easily
	// This would require scanning the queue, which is inefficient
	// In practice, jobs are removed when popped
	return nil
}

// PushMany adds multiple jobs to the queue
func (q *redisQueue[T]) PushMany(jobs []*Job[T]) error {
	if len(jobs) == 0 {
		return nil
	}

	// Prepare pipeline
	pipe := q.client.Pipeline()

	for _, job := range jobs {
		data, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal job: %w", err)
		}

		pipe.LPush(q.ctx, q.queueName, data)
	}

	// Execute pipeline
	_, err := pipe.Exec(q.ctx)
	return err
}

// PopMany retrieves multiple jobs from the queue
func (q *redisQueue[T]) PopMany(count int) ([]*Job[T], error) {
	jobs := make([]*Job[T], 0, count)

	for i := 0; i < count; i++ {
		job, err := q.Pop()
		if err != nil {
			return jobs, err
		}

		if job == nil {
			break // No more jobs
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Retry retries a failed job
func (q *redisQueue[T]) Retry(job *Job[T]) error {
	// For Redis, we'll just push it back to the queue
	return q.Push(job)
}

// Fail marks a job as permanently failed
func (q *redisQueue[T]) Fail(job *Job[T], err error) error {
	// For Redis, we could move it to a failed jobs queue
	// For now, we'll just log the failure
	return nil
}

// Size returns the number of jobs in the queue
func (q *redisQueue[T]) Size() (int64, error) {
	return q.client.LLen(q.ctx, q.queueName).Result()
}

// Clear removes all jobs from the queue
func (q *redisQueue[T]) Clear() error {
	return q.client.Del(q.ctx, q.queueName).Err()
}

// WithContext returns a queue with context
func (q *redisQueue[T]) WithContext(ctx context.Context) Queue[T] {
	return &redisQueue[T]{
		client:    q.client,
		queueName: q.queueName,
		ctx:       ctx,
	}
}

// syncQueue implements Queue[T] with in-memory storage
type syncQueue[T any] struct {
	jobs []*Job[T]
	mu   sync.RWMutex
	ctx  context.Context
}

// NewSyncQueue creates a new in-memory queue
func NewSyncQueue[T any]() Queue[T] {
	return &syncQueue[T]{
		jobs: make([]*Job[T], 0),
		ctx:  context.Background(),
	}
}

// Push adds a job to the queue
func (q *syncQueue[T]) Push(job *Job[T]) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs = append(q.jobs, job)
	return nil
}

// Pop retrieves a job from the queue
func (q *syncQueue[T]) Pop() (*Job[T], error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.jobs) == 0 {
		return nil, nil
	}

	job := q.jobs[0]
	q.jobs = q.jobs[1:]

	return job, nil
}

// Delete removes a job from the queue
func (q *syncQueue[T]) Delete(jobID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, job := range q.jobs {
		if job.ID == jobID {
			q.jobs = append(q.jobs[:i], q.jobs[i+1:]...)
			break
		}
	}

	return nil
}

// PushMany adds multiple jobs to the queue
func (q *syncQueue[T]) PushMany(jobs []*Job[T]) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs = append(q.jobs, jobs...)
	return nil
}

// PopMany retrieves multiple jobs from the queue
func (q *syncQueue[T]) PopMany(count int) ([]*Job[T], error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.jobs) == 0 {
		return []*Job[T]{}, nil
	}

	if count > len(q.jobs) {
		count = len(q.jobs)
	}

	jobs := q.jobs[:count]
	q.jobs = q.jobs[count:]

	return jobs, nil
}

// Retry retries a failed job
func (q *syncQueue[T]) Retry(job *Job[T]) error {
	return q.Push(job)
}

// Fail marks a job as permanently failed
func (q *syncQueue[T]) Fail(job *Job[T], err error) error {
	// For sync queue, we'll just log the failure
	return nil
}

// Size returns the number of jobs in the queue
func (q *syncQueue[T]) Size() (int64, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return int64(len(q.jobs)), nil
}

// Clear removes all jobs from the queue
func (q *syncQueue[T]) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs = make([]*Job[T], 0)
	return nil
}

// WithContext returns a queue with context
func (q *syncQueue[T]) WithContext(ctx context.Context) Queue[T] {
	return &syncQueue[T]{
		jobs: q.jobs,
		ctx:  ctx,
	}
}
