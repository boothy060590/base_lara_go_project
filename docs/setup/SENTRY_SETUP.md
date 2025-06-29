> **Docker Compose Version Requirement:**  
> This setup uses the `include` feature in `docker-compose.yaml`.  
> You must use Docker Compose v2.20.0 or newer.

# Sentry Setup for Local Development

This project includes a local Sentry setup for error tracking during development.

## SSL Certificates for Sentry

Generate and trust a self-signed SSL certificate for Sentry:

```bash
./docker/ssl/gen_certs.sh sentry.baselaragoproject.test
./docker/ssl/trust_certs_mac.sh sentry.baselaragoproject.test
```

This ensures your browser and Docker containers trust the local HTTPS endpoint.

## Architecture

The Sentry setup is split into two Docker Compose files:

- `docker/sentry/docker-compose.yml` - Infrastructure services (PostgreSQL, Redis, ClickHouse, Worker, Cron)
- `docker-compose.yaml` - Main services including Sentry web interface
- `docker/sentry/envs/sentry.env` - Shared environment variables for Sentry services

## Setup Instructions

### 1. Generate Secret Key

Run the script to generate a secure secret key:

```bash
./scripts/generate-sentry-key.sh
```

### 2. Update Configuration

Update the following files with the generated secret key:

1. **docker/sentry/envs/sentry.env** - Replace the SENTRY_SECRET_KEY value
2. **api/.env** - Add `SENTRY_SECRET_KEY=your-generated-key`
3. **api/env/.env.local** - Add `SENTRY_SECRET_KEY=your-generated-key`
4. **api/env/.env.staging** - Add `SENTRY_SECRET_KEY=your-generated-key`
5. **api/env/.env.production** - Add `SENTRY_SECRET_KEY=your-generated-key`

### 3. Start Services

```bash
# Start all services including Sentry
docker-compose up -d

# Or start just Sentry infrastructure first
docker-compose -f docker/sentry/docker-compose.yml up -d
docker-compose up -d sentry
```

### 4. Access Sentry

- **Web Interface**: https://sentry.baselaragoproject.test
- **Default Admin**: Create on first visit

## Services

### Infrastructure (docker/sentry/docker-compose.yml)
- **sentry-postgres**: PostgreSQL database
- **sentry-redis**: Redis cache
- **sentry-clickhouse**: ClickHouse for analytics
- **sentry-worker**: Background job processing
- **sentry-cron**: Scheduled tasks

### Web Interface (docker-compose.yaml)
- **sentry**: Web interface accessible at `sentry.baselaragoproject.test`

## Environment Variables

### Shared Infrastructure Variables (docker/sentry/envs/sentry.env)
```env
# Sentry Secret Key (shared across all services)
SENTRY_SECRET_KEY=your-secret-key-here

# PostgreSQL Configuration
SENTRY_POSTGRES_HOST=sentry-postgres
SENTRY_POSTGRES_PORT=5432
SENTRY_POSTGRES_DBNAME=sentry
SENTRY_POSTGRES_USERNAME=sentry
SENTRY_POSTGRES_PASSWORD=sentry_password

# Redis Configuration
SENTRY_REDIS_HOST=sentry-redis
SENTRY_REDIS_PORT=6379

# ClickHouse Configuration
SENTRY_CLICKHOUSE_HOST=sentry-clickhouse
SENTRY_CLICKHOUSE_PORT=9000
SENTRY_CLICKHOUSE_DBNAME=sentry
SENTRY_CLICKHOUSE_USERNAME=sentry
SENTRY_CLICKHOUSE_PASSWORD=sentry_password
```

### Application Environment Variables
```env
# Sentry Configuration
SENTRY_DSN=http://sentry.baselaragoproject.test/1
SENTRY_ENVIRONMENT=local
SENTRY_SECRET_KEY=your-secret-key-here
SENTRY_DEBUG=false
SENTRY_TRACES_SAMPLE_RATE=1.0
SENTRY_PROFILES_SAMPLE_RATE=1.0
SENTRY_REPLAYS_SESSION_SAMPLE_RATE=0.1
SENTRY_REPLAYS_ON_ERROR_SAMPLE_RATE=1.0
```

## Integration with Go Application

Once Sentry is running, you can integrate it with your Go application using the official Sentry Go SDK.

### Example Usage
```go
import (
    "github.com/getsentry/sentry-go"
)

func init() {
    err := sentry.Init(sentry.ClientOptions{
        Dsn: "http://your-dsn@sentry.baselaragoproject.test/1",
    })
    if err != nil {
        log.Fatalf("sentry.Init: %s", err)
    }
}

func reportError(err error, message string) {
    sentry.CaptureException(err)
    sentry.Flush(2 * time.Second)
}
```

## Troubleshooting

### Common Issues

1. **Sentry not starting**: Check that all infrastructure services are running
2. **Database connection errors**: Ensure PostgreSQL is fully started before Sentry
3. **Secret key issues**: Make sure the same secret key is used across all services

### Logs
```bash
# Check Sentry web service logs
docker-compose logs sentry

# Check infrastructure logs
docker-compose -f docker/sentry/docker-compose.yml logs

# Check specific service logs
docker-compose -f docker/sentry/docker-compose.yml logs sentry-postgres
```

## Cleanup

To remove Sentry completely:

```bash
# Stop and remove all Sentry services
docker-compose down sentry
docker-compose -f docker/sentry/docker-compose.yml down

# Remove volumes (WARNING: This will delete all data)
docker volume rm base_lara_go_project_sentry-postgres-data
docker volume rm base_lara_go_project_sentry-redis-data
docker volume rm base_lara_go_project_sentry-clickhouse-data
``` 