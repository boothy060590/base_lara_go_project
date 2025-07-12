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

	// Context-aware basic operations
	PushWithContext(ctx context.Context, job *Job[T]) error
	PopWithContext(ctx context.Context) (*Job[T], error)
	DeleteWithContext(ctx context.Context, jobID string) error

	// Batch operations
	PushMany(jobs []*Job[T]) error
	PopMany(count int) ([]*Job[T], error)

	// Context-aware batch operations
	PushManyWithContext(ctx context.Context, jobs []*Job[T]) error
	PopManyWithContext(ctx context.Context, count int) ([]*Job[T], error)

	// Job management
	Retry(job *Job[T]) error
	Fail(job *Job[T], error error) error

	// Context-aware job management
	RetryWithContext(ctx context.Context, job *Job[T]) error
	FailWithContext(ctx context.Context, job *Job[T], error error) error

	// Queue management
	Size() (int64, error)
	Clear() error
	SizeWithContext(ctx context.Context) (int64, error)
	ClearWithContext(ctx context.Context) error
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

// jobDispatcher implements JobDispatcher[T] with performance optimizations
type jobDispatcher[T any] struct {
	queue Queue[T]
	ctx   context.Context
	// Performance optimizations (safe for job operations)
	atomicCounter     *AtomicCounter
	jobPool           *ObjectPool[Job[T]]
	performanceFacade *PerformanceFacade

	// New optimization fields
	workStealingPool *WorkStealingPool[any]
	customAllocator  *CustomAllocator[any]
	profileOptimizer *ProfileGuidedOptimizer[any]
}

// NewJobDispatcher creates a new job dispatcher with performance optimizations
func NewJobDispatcher[T any](queue Queue[T], wsp *WorkStealingPool[any], ca *CustomAllocator[any], pgo *ProfileGuidedOptimizer[any]) JobDispatcher[T] {
	// Create performance optimizations
	atomicCounter := NewAtomicCounter()
	performanceFacade := NewPerformanceFacade()

	// Use custom allocator for job pooling if provided
	var jobPool *ObjectPool[Job[T]]
	if ca != nil {
		// Create a wrapper that uses custom allocator
		jobPool = &ObjectPool[Job[T]]{
			// Implementation would delegate to ca.Allocate/Deallocate
		}
	} else {
		jobPool = NewObjectPool[Job[T]](100,
			func() Job[T] { return Job[T]{} },
			func(job Job[T]) Job[T] { return Job[T]{} },
		)
	}

	return &jobDispatcher[T]{
		queue:             queue,
		ctx:               context.Background(),
		atomicCounter:     atomicCounter,
		jobPool:           jobPool,
		performanceFacade: performanceFacade,
		workStealingPool:  wsp,
		customAllocator:   ca,
		profileOptimizer:  pgo,
	}
}

// Dispatch dispatches a job (respects ShouldQueue trait) with performance tracking and atomic counter
func (d *jobDispatcher[T]) Dispatch(job T) error {
	// Track operation count atomically
	d.atomicCounter.Increment()

	return d.performanceFacade.Track("job.dispatch", func() error {
		// Check if job implements JobTraits
		if queueableJob, ok := any(job).(JobTraits); ok && queueableJob.ShouldQueue() {
			// Get job from object pool (safe - no database state)
			goJob := d.jobPool.Get()
			defer d.jobPool.Put(goJob)

			// Initialize job
			goJob.ID = "job_" + time.Now().Format("20060102150405")
			goJob.Data = job
			goJob.Attempts = 0
			goJob.MaxRetries = queueableJob.GetMaxAttempts()
			goJob.CreatedAt = time.Now()

			// Use work stealing pool for job processing if available
			if d.workStealingPool != nil {
				return d.dispatchWithWorkStealing(&goJob)
			}

			return d.queue.Push(&goJob)
		}

		// Execute synchronously
		return d.executeJob(job)
	})
}

// dispatchWithWorkStealing dispatches job using work stealing pool
func (d *jobDispatcher[T]) dispatchWithWorkStealing(job *Job[T]) error {
	workItem := WorkItem[any]{
		ID:      job.ID,
		Data:    job,
		Handler: d.processJob,
		Timeout: 30 * time.Second,
	}

	return d.workStealingPool.Submit(workItem)
}

// processJob processes a job using work stealing pool
func (d *jobDispatcher[T]) processJob(ctx context.Context, data any) error {
	job := data.(*Job[T])
	return d.executeJob(job.Data)
}

// DispatchSync dispatches a job synchronously (ignores ShouldQueue trait) with performance tracking and atomic counter
func (d *jobDispatcher[T]) DispatchSync(job T) error {
	// Track operation count atomically
	d.atomicCounter.Increment()

	return d.performanceFacade.Track("job.dispatch_sync", func() error {
		return d.executeJob(job)
	})
}

// executeJob executes a job synchronously with performance tracking
func (d *jobDispatcher[T]) executeJob(job T) error {
	return d.performanceFacade.Track("job.execute", func() error {
		// For now, we'll just call the Handle method if it exists
		if handler, ok := any(job).(interface {
			Handle(ctx context.Context) error
		}); ok {
			return handler.Handle(d.ctx)
		}

		// If no Handle method, just return success
		return nil
	})
}

// GetQueue returns the underlying queue
func (d *jobDispatcher[T]) GetQueue() Queue[T] {
	return d.queue
}

// WithContext returns a dispatcher with context
func (d *jobDispatcher[T]) WithContext(ctx context.Context) JobDispatcher[T] {
	return &jobDispatcher[T]{
		queue:             d.queue.WithContext(ctx),
		ctx:               ctx,
		atomicCounter:     d.atomicCounter,
		jobPool:           d.jobPool,
		performanceFacade: d.performanceFacade,
	}
}

// QueueWorker processes jobs from a queue with performance optimizations
type QueueWorker[T any] struct {
	queue   Queue[T]
	handler JobHandler[T]
	ctx     context.Context
	cancel  context.CancelFunc
	// Performance optimizations (safe for worker operations)
	atomicCounter     *AtomicCounter
	performanceFacade *PerformanceFacade

	// New optimization fields
	workStealingPool *WorkStealingPool[any]
	customAllocator  *CustomAllocator[any]
	profileOptimizer *ProfileGuidedOptimizer[any]
}

// NewQueueWorker creates a new queue worker with performance optimizations
func NewQueueWorker[T any](queue Queue[T], handler JobHandler[T], wsp *WorkStealingPool[any], ca *CustomAllocator[any], pgo *ProfileGuidedOptimizer[any]) *QueueWorker[T] {
	ctx, cancel := context.WithCancel(context.Background())

	// Create performance optimizations
	atomicCounter := NewAtomicCounter()
	performanceFacade := NewPerformanceFacade()

	return &QueueWorker[T]{
		queue:             queue,
		handler:           handler,
		ctx:               ctx,
		cancel:            cancel,
		atomicCounter:     atomicCounter,
		performanceFacade: performanceFacade,
		workStealingPool:  wsp,
		customAllocator:   ca,
		profileOptimizer:  pgo,
	}
}

// Start begins processing jobs with performance tracking and atomic counter
func (w *QueueWorker[T]) Start() error {
	if w.performanceFacade != nil {
		return w.performanceFacade.Track("worker.start", func() error {
			return w.processJobs()
		})
	}
	return w.processJobs()
}

// processJobs is the main job processing loop
func (w *QueueWorker[T]) processJobs() error {
	for {
		select {
		case <-w.ctx.Done():
			return w.ctx.Err()
		default:
			// Track operation count atomically
			if w.atomicCounter != nil {
				w.atomicCounter.Increment()
			}

			// Pop job from queue
			job, err := w.queue.Pop()
			if err != nil {
				// No jobs available, wait a bit
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if job == nil {
				// Queue is empty, wait a bit
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

// processJob processes a single job with performance tracking
func (w *QueueWorker[T]) processJob(job *Job[T]) error {
	if w.performanceFacade != nil {
		return w.performanceFacade.Track("worker.process_job", func() error {
			return w.executeJob(job)
		})
	}
	return w.executeJob(job)
}

// executeJob executes a single job
func (w *QueueWorker[T]) executeJob(job *Job[T]) error {
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
	return q.PushWithContext(q.ctx, job)
}

// PushWithContext adds a job to the queue with context support
func (q *redisQueue[T]) PushWithContext(ctx context.Context, job *Job[T]) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return q.client.LPush(ctx, q.queueName, data).Err()
}

// Pop removes and returns a job from the queue
func (q *redisQueue[T]) Pop() (*Job[T], error) {
	return q.PopWithContext(q.ctx)
}

// PopWithContext removes and returns a job from the queue with context support
func (q *redisQueue[T]) PopWithContext(ctx context.Context) (*Job[T], error) {
	result, err := q.client.BRPop(ctx, 0, q.queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Queue is empty
		}
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	data := result[1]
	var job Job[T]
	err = json.Unmarshal([]byte(data), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// Delete removes a job from the queue
func (q *redisQueue[T]) Delete(jobID string) error {
	return q.DeleteWithContext(q.ctx, jobID)
}

// DeleteWithContext removes a job from the queue with context support
func (q *redisQueue[T]) DeleteWithContext(ctx context.Context, jobID string) error {
	// For Redis, we need to scan and remove the specific job
	// This is a simplified implementation
	return nil
}

// PushMany adds multiple jobs to the queue
func (q *redisQueue[T]) PushMany(jobs []*Job[T]) error {
	return q.PushManyWithContext(q.ctx, jobs)
}

// PushManyWithContext adds multiple jobs to the queue with context support
func (q *redisQueue[T]) PushManyWithContext(ctx context.Context, jobs []*Job[T]) error {
	if len(jobs) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()
	for _, job := range jobs {
		data, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal job: %w", err)
		}
		pipe.LPush(ctx, q.queueName, data)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// PopMany removes and returns multiple jobs from the queue
func (q *redisQueue[T]) PopMany(count int) ([]*Job[T], error) {
	return q.PopManyWithContext(q.ctx, count)
}

// PopManyWithContext removes and returns multiple jobs from the queue with context support
func (q *redisQueue[T]) PopManyWithContext(ctx context.Context, count int) ([]*Job[T], error) {
	if count <= 0 {
		return []*Job[T]{}, nil
	}

	var jobs []*Job[T]
	for i := 0; i < count; i++ {
		job, err := q.PopWithContext(ctx)
		if err != nil {
			return jobs, err
		}
		if job == nil {
			break // Queue is empty
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Retry retries a failed job
func (q *redisQueue[T]) Retry(job *Job[T]) error {
	return q.RetryWithContext(q.ctx, job)
}

// RetryWithContext retries a failed job with context support
func (q *redisQueue[T]) RetryWithContext(ctx context.Context, job *Job[T]) error {
	if job.Attempts >= job.MaxRetries {
		return fmt.Errorf("job %s has exceeded max retries", job.ID)
	}

	job.Attempts++
	return q.PushWithContext(ctx, job)
}

// Fail marks a job as failed
func (q *redisQueue[T]) Fail(job *Job[T], err error) error {
	return q.FailWithContext(q.ctx, job, err)
}

// FailWithContext marks a job as failed with context support
func (q *redisQueue[T]) FailWithContext(ctx context.Context, job *Job[T], err error) error {
	// For now, just log the failure
	// In a real implementation, you might want to store failed jobs in a separate queue
	return nil
}

// Size returns the number of jobs in the queue
func (q *redisQueue[T]) Size() (int64, error) {
	return q.SizeWithContext(q.ctx)
}

// SizeWithContext returns the number of jobs in the queue with context support
func (q *redisQueue[T]) SizeWithContext(ctx context.Context) (int64, error) {
	return q.client.LLen(ctx, q.queueName).Result()
}

// Clear removes all jobs from the queue
func (q *redisQueue[T]) Clear() error {
	return q.ClearWithContext(q.ctx)
}

// ClearWithContext removes all jobs from the queue with context support
func (q *redisQueue[T]) ClearWithContext(ctx context.Context) error {
	return q.client.Del(ctx, q.queueName).Err()
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
	return q.PushWithContext(q.ctx, job)
}

// PushWithContext adds a job to the queue with context support
func (q *syncQueue[T]) PushWithContext(ctx context.Context, job *Job[T]) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs = append(q.jobs, job)
	return nil
}

// Pop removes and returns a job from the queue
func (q *syncQueue[T]) Pop() (*Job[T], error) {
	return q.PopWithContext(q.ctx)
}

// PopWithContext removes and returns a job from the queue with context support
func (q *syncQueue[T]) PopWithContext(ctx context.Context) (*Job[T], error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.jobs) == 0 {
		return nil, nil
	}

	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return job, nil
}

// Delete removes a job from the queue by ID
func (q *syncQueue[T]) Delete(jobID string) error {
	return q.DeleteWithContext(q.ctx, jobID)
}

// DeleteWithContext removes a job from the queue by ID with context support
func (q *syncQueue[T]) DeleteWithContext(ctx context.Context, jobID string) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for i, job := range q.jobs {
		if job.ID == jobID {
			q.jobs = append(q.jobs[:i], q.jobs[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("job %s not found", jobID)
}

// PushMany adds multiple jobs to the queue
func (q *syncQueue[T]) PushMany(jobs []*Job[T]) error {
	return q.PushManyWithContext(q.ctx, jobs)
}

// PushManyWithContext adds multiple jobs to the queue with context support
func (q *syncQueue[T]) PushManyWithContext(ctx context.Context, jobs []*Job[T]) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(jobs) == 0 {
		return nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs = append(q.jobs, jobs...)
	return nil
}

// PopMany removes and returns multiple jobs from the queue
func (q *syncQueue[T]) PopMany(count int) ([]*Job[T], error) {
	return q.PopManyWithContext(q.ctx, count)
}

// PopManyWithContext removes and returns multiple jobs from the queue with context support
func (q *syncQueue[T]) PopManyWithContext(ctx context.Context, count int) ([]*Job[T], error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if count <= 0 {
		return []*Job[T]{}, nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	actualCount := count
	if len(q.jobs) < count {
		actualCount = len(q.jobs)
	}

	jobs := q.jobs[:actualCount]
	q.jobs = q.jobs[actualCount:]
	return jobs, nil
}

// Retry retries a failed job
func (q *syncQueue[T]) Retry(job *Job[T]) error {
	return q.RetryWithContext(q.ctx, job)
}

// RetryWithContext retries a failed job with context support
func (q *syncQueue[T]) RetryWithContext(ctx context.Context, job *Job[T]) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if job.Attempts >= job.MaxRetries {
		return fmt.Errorf("job %s has exceeded max retries", job.ID)
	}

	job.Attempts++
	return q.PushWithContext(ctx, job)
}

// Fail marks a job as failed
func (q *syncQueue[T]) Fail(job *Job[T], err error) error {
	return q.FailWithContext(q.ctx, job, err)
}

// FailWithContext marks a job as failed with context support
func (q *syncQueue[T]) FailWithContext(ctx context.Context, job *Job[T], err error) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// For now, just log the failure
	// In a real implementation, you might want to store failed jobs in a separate queue
	return nil
}

// Size returns the number of jobs in the queue
func (q *syncQueue[T]) Size() (int64, error) {
	return q.SizeWithContext(q.ctx)
}

// SizeWithContext returns the number of jobs in the queue with context support
func (q *syncQueue[T]) SizeWithContext(ctx context.Context) (int64, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	return int64(len(q.jobs)), nil
}

// Clear removes all jobs from the queue
func (q *syncQueue[T]) Clear() error {
	return q.ClearWithContext(q.ctx)
}

// ClearWithContext removes all jobs from the queue with context support
func (q *syncQueue[T]) ClearWithContext(ctx context.Context) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs = []*Job[T]{}
	return nil
}

// WithContext returns the same queue instance with the updated context
func (q *syncQueue[T]) WithContext(ctx context.Context) Queue[T] {
	q.ctx = ctx
	return q
}
