# Multi-Worker Infrastructure & Multi-Channel Logging

This document explains the multi-worker infrastructure and multi-channel logging features implemented in the Base Laravel Go Project.

## Overview

The system supports both single-worker development mode and multi-worker production mode for horizontal scaling. Different queue types (jobs, mail, events) can be processed by separate workers, allowing for independent scaling based on demand.

## Architecture

### Single-Worker Mode (Development)
- One worker processes all queue types
- Suitable for development and low-traffic applications
- Simple configuration and deployment

### Multi-Worker Mode (Production)
- Separate workers for different queue types
- Independent scaling based on queue-specific demands
- Better resource utilization and fault isolation

## Queue Types

1. **Jobs Queue** - Long-running background jobs
   - High processing requirements
   - Longer timeouts
   - Moderate throughput

2. **Mail Queue** - Email sending operations
   - Moderate processing requirements
   - Shorter timeouts
   - Lower throughput

3. **Events Queue** - Event processing
   - High throughput requirements
   - Short processing times
   - High memory usage

## Configuration

### Environment Variables

```bash
# Multi-Worker Mode
QUEUE_MULTI_WORKER=false                    # Enable/disable multi-worker mode
QUEUE_DEFAULT_WORKER=default               # Default worker type

# Queue Names
SQS_QUEUE_JOBS=jobs                        # Jobs queue name
SQS_QUEUE_MAIL=mail                        # Mail queue name
SQS_QUEUE_EVENTS=events                    # Events queue name

# Default Worker Configuration
WORKER_MAX_JOBS=1000                       # Max jobs before restart
WORKER_MEMORY_LIMIT=128                    # Memory limit in MB
WORKER_TIMEOUT=60                          # Timeout in seconds
WORKER_SLEEP=3                             # Sleep between polls
WORKER_TRIES=3                             # Retry attempts

# Jobs Worker Configuration
JOBS_WORKER_MAX_JOBS=1000
JOBS_WORKER_MEMORY_LIMIT=128
JOBS_WORKER_TIMEOUT=60
JOBS_WORKER_SLEEP=3
JOBS_WORKER_TRIES=3

# Mail Worker Configuration
MAIL_WORKER_MAX_JOBS=500
MAIL_WORKER_MEMORY_LIMIT=64
MAIL_WORKER_TIMEOUT=30
MAIL_WORKER_SLEEP=3
MAIL_WORKER_TRIES=3

# Events Worker Configuration
EVENTS_WORKER_MAX_JOBS=2000
EVENTS_WORKER_MEMORY_LIMIT=256
EVENTS_WORKER_TIMEOUT=120
EVENTS_WORKER_SLEEP=1
EVENTS_WORKER_TRIES=1
```

## Usage

### Running Workers

#### Using the Management Script

```bash
# Development mode - single worker
./scripts/run-workers.sh single

# Production mode - separate workers
./scripts/run-workers.sh multi

# Run specific worker types
./scripts/run-workers.sh jobs
./scripts/run-workers.sh mail
./scripts/run-workers.sh events

# Check worker status
./scripts/run-workers.sh status

# Stop all workers
./scripts/run-workers.sh stop
```

#### Using Go Commands

```bash
# Single worker (default)
go run bootstrap/worker/main.go

# Specific worker types
go run bootstrap/worker/main.go -worker=jobs
go run bootstrap/worker/main.go -worker=mail
go run bootstrap/worker/main.go -worker=events
```

### Docker Deployment

For production deployment, you can run multiple worker containers:

```yaml
# docker-compose.yaml
services:
  worker-jobs:
    build: .
    command: go run bootstrap/worker/main.go -worker=jobs
    environment:
      - QUEUE_MULTI_WORKER=true
    scale: 2  # Scale jobs workers independently

  worker-mail:
    build: .
    command: go run bootstrap/worker/main.go -worker=mail
    environment:
      - QUEUE_MULTI_WORKER=true
    scale: 1  # Single mail worker

  worker-events:
    build: .
    command: go run bootstrap/worker/main.go -worker=events
    environment:
      - QUEUE_MULTI_WORKER=true
    scale: 3  # Scale events workers for high throughput
```

## Multi-Channel Logging

The system supports Laravel-style multi-channel logging with the ability to report exceptions to specific channels.

### Basic Logging

```go
import "base_lara_go_project/app/core/facades"

// Basic logging
facades.Info("User logged in", map[string]interface{}{
    "user_id": 123,
    "ip": "192.168.1.1",
})

facades.Error("Database connection failed", map[string]interface{}{
    "error": err.Error(),
})
```

### Multi-Channel Logging

```go
// Log to specific channel
logger, err := facades.Channel("slack")
if err == nil {
    logger.Error("Critical error occurred", context)
}

// Log to multiple channels
logger, err := facades.Stack("slack", "sentry")
if err == nil {
    logger.Error("Error occurred", context)
}
```

### Exception Reporting (Laravel-style)

```go
// Report exception to default channels
facades.Report(err)

// Report exception to specific channels
facades.Report(err, "slack", "sentry")

// Report with custom level
facades.ReportWithLevel(err, facades.Critical, "slack")

// Report less severe exceptions
facades.ReportWithLevel(err, facades.Warning, "sentry")
```

### Logging Configuration

```bash
# Logging channels
LOG_CHANNEL=stack                    # Default channel
LOG_LEVEL=debug                      # Log level
LOG_PATH=storage/logs/laravel.log    # Log file path

# Multi-channel configuration
LOG_STACK=single                     # Stack channel configuration
LOG_DEPRECATIONS_CHANNEL=null        # Deprecations channel

# Slack logging
LOG_SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
LOG_SLACK_USERNAME=Laravel Log
LOG_SLACK_EMOJI=:boom:

# Sentry logging
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project-id
```

## Worker Lifecycle Management

### Automatic Restart Conditions

Workers automatically restart when:
- Memory limit is exceeded
- Maximum jobs processed
- Timeout reached
- Manual restart signal

### Health Monitoring

```go
// Get worker statistics
stats := worker.GetStats()
fmt.Printf("Processed jobs: %d\n", stats["processed_jobs"])
fmt.Printf("Uptime: %s\n", stats["uptime"])
fmt.Printf("Memory usage: %d MB\n", stats["memory_usage"])
```

### Graceful Shutdown

```go
// Stop worker gracefully
worker.Stop()
```

## Best Practices

### Development
- Use single-worker mode for development
- Set lower memory limits and timeouts
- Use sync queue driver for testing

### Production
- Use multi-worker mode for production
- Scale workers based on queue-specific demands
- Monitor worker health and restart conditions
- Use appropriate queue drivers (SQS, Redis, etc.)

### Monitoring
- Monitor queue depths for each worker type
- Track worker restart frequency
- Monitor memory usage and processing times
- Set up alerts for worker failures

### Scaling Strategy
- **Jobs Workers**: Scale based on job complexity and processing time
- **Mail Workers**: Scale based on email volume and delivery requirements
- **Events Workers**: Scale based on event frequency and processing speed

## Troubleshooting

### Common Issues

1. **Worker not starting**
   - Check queue configuration
   - Verify environment variables
   - Check service provider registration

2. **High memory usage**
   - Adjust memory limits
   - Check for memory leaks in job processing
   - Monitor job complexity

3. **Queue processing delays**
   - Scale workers horizontally
   - Optimize job processing
   - Check queue driver performance

4. **Worker restarts**
   - Check restart conditions
   - Monitor system resources
   - Review job processing logic

### Debugging

```bash
# Enable debug logging
LOG_LEVEL=debug

# Check worker status
./scripts/run-workers.sh status

# View worker logs
tail -f storage/logs/laravel.log
```

## Migration from Single-Worker

1. Update environment configuration
2. Deploy new worker types
3. Monitor performance and adjust scaling
4. Gradually migrate queue assignments
5. Remove old single-worker instances

## Future Enhancements

- Worker pool management
- Dynamic scaling based on queue depth
- Worker load balancing
- Advanced monitoring and alerting
- Worker performance metrics
- Queue priority support
