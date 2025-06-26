# Dockerised GoLang api with VueJS Frontend. Comes with authentication already handled. 

# Base Laravel Go Project

## Getting Started

This project is a modern, Laravel-inspired Go API and Vue.js frontend for Base Laravel Go Project, featuring asynchronous job/event handling, Dockerized infrastructure, and a robust local development workflow.

---

## Prerequisites

- **Docker Desktop** (latest)
- **Go 1.23+**
- **Node.js 20+**
- **npm** (comes with Node.js)
- **Git**
- **VS Code** (recommended)

---

## Quick Start

### 1. Clone the Repository
```bash
git clone <repository-url>
cd base_lara_go_project
```

### 2. Generate SSL Certificates
```bash
./docker/ssl/gen_certs.sh
./docker/ssl/trust_certs_mac.sh  # (macOS only)
```

### 3. Environment Variables
Copy and edit the environment files:
```bash
# API environment
cp api/config/.env.local api/.env
nano api/.env

# Frontend environment
cp frontend/config/.env.local frontend/.env
nano frontend/.env
```

### 4. Start All Services
```bash
# First, install frontend dependencies in a standalone container to avoid ARM architecture issues
docker run --rm -v $(pwd)/frontend:/app -w /app node:20-alpine npm install

# Start all services
docker-compose up -d
```

### 5. Access the Application
The application uses nginx reverse proxy with automatic SSL and dnsmasq for local domain resolution. All services are accessible via their virtual host URLs:

| Service | URL | Description |
|---------|-----|-------------|
| Frontend | https://app.baselaragoproject.test | Vue.js application |
| API | https://api.baselaragoproject.test | Go API endpoints |
| MailHog | https://mail.baselaragoproject.test | Email testing interface |
| MinIO Console | https://s3.baselaragoproject.test | S3-compatible storage |
| Redis | https://redis.baselaragoproject.test | Cache service |
| ElasticMQ | https://sqs.baselaragoproject.test | Queue service |

---

## Infrastructure

### Nginx Reverse Proxy
- **Automatic SSL**: Self-signed certificates for all services
- **Service Discovery**: Automatically detects containers via Docker labels
- **Virtual Hosts**: Routes traffic based on `VIRTUAL_HOST` environment variables
- **Load Balancing**: Supports multiple backend services

### DNS Resolution
- **dnsmasq**: Local DNS server that resolves `.test` domains to localhost
- **Automatic**: No manual `/etc/hosts` configuration required
- **Container Discovery**: Services automatically register their hostnames

---

## Development

The development environment is fully containerized with hot reloading enabled:

### Hot Reloading
- **API**: Uses Air for Go hot reloading in development
- **Frontend**: Uses Vite for Vue.js hot module replacement
- **Worker**: Separate container for background job processing

### Local Development
All services run in Docker containers with live code reloading:
```bash
# Start development environment
docker-compose up -d

# View logs for specific services
docker-compose logs -f api      # API logs
docker-compose logs -f app      # Frontend logs
docker-compose logs -f worker   # Worker logs
```

### Frontend State Management
The frontend uses **Pinia** for state management:
- **Authentication Store**: Manages user login/logout state
- **Reactive State**: Automatic UI updates when state changes
- **DevTools Support**: Vue DevTools integration for debugging
- **TypeScript Support**: Full type safety for state management

---

## Testing

- Register a user via the frontend or API
- Check MailHog for welcome emails
- Use the API endpoints for authentication and registration

---

## Documentation

See `docs/ARCHITECTURE.md` for full technical and architectural details.