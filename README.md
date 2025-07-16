Base Laravel Go Project
A Go-based web application with Laravel-style architecture, featuring async event processing, mail queue management, and modern web development practices.

🚀 Features
Laravel-Style Architecture: Familiar patterns and structure for Laravel developers
Dynamic Configuration System: Zero-config, automatic loading with Laravel-style dot notation
Service Layer Architecture: Clean separation between business logic and data access
Service Facades: Laravel-style static access to services
Service Decorators: Cross-cutting concerns (logging, caching, auditing)
Async Event Processing: Event-driven architecture with queue-based processing
Mail Queue Management: Asynchronous email sending via dedicated mail queue
Multi-Queue System: Separate queues for jobs, mail, and events
Real-time Queue Processing: Ultra-fast concurrent queue processing
JWT Authentication: Secure token-based authentication
Database Integration: GORM v2 with MySQL support
Docker Development: Complete containerized development environment
Vue.js Frontend: Modern reactive frontend with form validation

🏗️ Architecture
Core Components
Dynamic Config System: Zero-configuration, automatic loading with Laravel-style access
Service Layer: Business logic with proper separation from data access
Repository Layer: Data persistence and retrieval with caching
Service Facades: Laravel-style static access to services
Service Decorators: Cross-cutting concerns (logging, caching, auditing)
Event System: Async event dispatching and processing
Queue System: Multi-queue processing with ElasticMQ
Mail System: Template-based email sending with queue support
Job System: Background job processing
Authentication: JWT-based user authentication and authorization

Controllers → Services → Repositories → Models
↓ ↓ ↓ ↓
Facades Business Logic CRUD Cache/DB
↓ ↓
Decorators Cross-Cutting

### Configuration System

The framework features a **dynamic configuration system** that provides:

- **Zero Configuration**: Drop Go config files in `api/config/` and they're automatically available
- **Laravel-Style Access**: Use dot notation like `config.Get("app.name")` or `config.GetString("database.connections.mysql.host")`
- **Environment Integration**: Built-in environment variable fallbacks with sensible defaults
- **Type-Safe Access**: Helper methods for strings, integers, and booleans
- **Automatic Caching**: Config values are cached for performance
- **Thread-Safe**: All operations are concurrency-safe

```go
// Access config values with dot notation
appName := config.GetString("app.name")
debugMode := config.GetBool("app.debug")
port := config.GetInt("app.port")

// With default values
timeout := config.GetInt("my.timeout", 30)
enabled := config.GetBool("my.enabled", true)

// Check if config exists
if config.Has("my.feature") {
    // Use the config
}
```

### Queue Structure

- **Events Queue**: Handles application events (user registration, etc.)
- **Mail Queue**: Processes email sending tasks
- **Jobs Queue**: General background job processing

---

## 🛠️ Technology Stack

### Backend
- **Go 1.24+**: Core application language
- **FastHTTP**: High-performance HTTP server with Gin routing adapter
- **Custom Repository Layer**: Optimized SQL with connection pooling and statement caching
- **ElasticMQ**: Message queue (SQS-compatible)
- **JWT**: Authentication tokens

### Frontend
- **Vue.js 3**: Reactive frontend framework
- **Vite**: Build tool and dev server
- **SCSS**: Styling with modern CSS features

### Infrastructure
- **Docker**: Containerization
- **MySQL**: Primary database
- **Nginx**: Reverse proxy
- **MailHog**: Email testing
- **Redis**: Cache store
- **Sentry**: Error logging (optional, choose at install)

---

## 🚀 Quick Start for Developers

### One-Shot Setup (Recommended)

**If you have make:**

```sh
make clean
make install_dev
```

- You'll be prompted for:
  - Queue mode: SQS/ElasticMQ (multi-worker) or sync (single worker)
  - Logging: Sentry or local
- All config, env, and Docker Compose files are generated from templates.
- All containers and services are started automatically.

**If you don't have make:**

```sh
bash setup/clean.sh
bash setup/install.sh
```

---

## 🛠️ Configuration & Templates

- **Dynamic Configuration**: All configuration is Go-native with automatic loading from `api/config/` directory
- **Environment-Driven**: Configuration values fall back to environment variables with sensible defaults
- **Laravel-Style Access**: Use dot notation to access nested configuration values
- **Zero Configuration**: Just drop Go config files and they're automatically available
- The main env file is generated as `api/env/.env.worker`.
- For multi-worker setups, additional envs are generated as needed.
- All environment and config files are generated from `.template` files (e.g. `.env.template`, `docker-compose.template.yaml`).
- **Only template files are committed to git; generated files are ignored.**
- To change domains or environments, use `make switch_domain` and rerun the install.

See [`/docs/config/CONFIGURATION.md`](docs/config/CONFIGURATION.md) and [`/api/config/README.md`](api/config/README.md) for details.

---

## 🛠️ Multi-Worker & Logging Options

- At install, choose between:
  - **SQS/ElasticMQ (multi-worker):** Generates multiple worker envs and Docker Compose services.
  - **Sync (single worker):** Simpler, local-only queue processing.
  - **Sentry or Local Logging:** Choose Sentry for error reporting, or local for file-based logs.

See [`/docs/queues/MULTI_WORKER_INFRASTRUCTURE.md`](docs/queues/MULTI_WORKER_INFRASTRUCTURE.md) for advanced queue/worker setup.

---

## 📚 Documentation Structure

- [`/docs/architecture/`](docs/architecture/) — Architecture, service vs repository, etc.
- [`/docs/config/`](docs/config/) — Configuration system and environment variables
- [`/api/config/README.md`](api/config/README.md) — Dynamic configuration system documentation
- [`/docs/setup/`](docs/setup/) — Setup scripts, Sentry, and install flow
- [`/docs/queues/`](docs/queues/) — Multi-worker and queue infrastructure
- [`/docs/performance/`](docs/performance/) — Performance analysis and optimization

---

## 🧩 Project Structure


base_lara_go_project/
├── api/ # Go backend application
│ ├── app/
│ ├── bootstrap/
│ ├── config/ # Dynamic configuration files
│ ├── env/
│ ├── ...
├── frontend/ # Vue.js frontend
├── docker/ # Docker configuration
├── setup/ # Modular install and setup scripts
├── docs/ # Documentation (see above)

---

## 🔒 SSL & Health Check

- Local SSL certs are generated and trusted automatically.
- See [`/docs/setup/SETUP_SCRIPTS.md`](docs/setup/SETUP_SCRIPTS.md) for details.

---

## 🐳 Docker & Troubleshooting

- If you hit Docker Hub rate limits, run `docker login` and try again.
- For any issues, rerun `make clean` and `make install_dev`.
- If you see SSL warnings in your browser, proceed past them for local development.

---

## 📚 Usage Examples

### Configuration Usage

```go
import "base_lara_go_project/config"

// Laravel-style config access
appName := config.GetString("app.name")
debugMode := config.GetBool("app.debug")
port := config.GetInt("app.port")

// With default values
timeout := config.GetInt("my.timeout", 30)
enabled := config.GetBool("my.enabled", true)

// Check if config exists
if config.Has("my.feature") {
    // Use the config
}

// Using facades
import facades_core "base_lara_go_project/app/core/laravel_core/facades"
appName := facades_core.GetString("app.name")
```

### Service Layer Usage

```go
// Laravel-style facade usage
user, err := facades.CreateUser(userData, roles)
user, err := facades.AuthenticateUser(email, password)

// Service with decorators for cross-cutting concerns
loggingDecorator := core.NewLoggingDecorator[interfaces.UserInterface](userService, logger)
cachingDecorator := core.NewCachingDecorator[interfaces.UserInterface](userService, cache, 30*time.Minute)

// Use decorated service
user, err := loggingDecorator.CreateUser(userData, roles) // Automatically logged
user, err := cachingDecorator.AuthenticateUser(email, password) // Automatically cached
```

### Event Processing

```go
// Dispatch an event asynchronously
event := &authEvents.UserCreated{User: user}
core.DispatchEventAsync(event)
```

### Mail Sending

```go
// Send email asynchronously
facades.MailAsync([]string{"user@example.com"}, "Subject", "Body")

// Send templated email
data := core.EmailTemplateData{
    Subject: "Welcome!",
    User:    user,
}
facades.MailTemplateAsync([]string{user.Email}, "auth/welcome", data)
```

### Job Processing

```go
// Dispatch a background job
job := &jobs.CreateUser{UserData: userData}
facades.Dispatch(job)
```

---

## 🔄 Queue Processing

The application uses a multi-queue system with ultra-fast processing:

- **Concurrent Queue Processing**: All queues processed simultaneously
- **Zero Wait Time**: Instant message polling
- **Concurrent Message Processing**: Multiple messages processed concurrently
- **50ms Polling Cycle**: Ultra-responsive queue monitoring

See [`/docs/queues/MULTI_WORKER_INFRASTRUCTURE.md`](docs/queues/MULTI_WORKER_INFRASTRUCTURE.md) for more.

---

## ⚡ Performance

Our Laravel-inspired Go architecture provides exceptional performance while maintaining developer productivity.

See [`/docs/performance/PERFORMANCE.md`](docs/performance/PERFORMANCE.md) for benchmarks and details.

---

## 🧪 Testing

### API Testing

```bash
# Test user registration
curl -X POST https://api.baselaragoproject.test/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "User",
    "email": "test@example.com",
    "mobile_number": "+1234567890",
    "password": "password123",
    "password_confirmation": "password123"
  }' \
  -k
```

### Email Testing

- Check MailHog at http://mail.baselaragoproject.test:8025
- All emails are captured for testing

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## ⚡ For More

- See [`/docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md) for a deep dive into the system design.
- See [`/docs/setup/SETUP_SCRIPTS.md`](docs/setup/SETUP_SCRIPTS.md) for all setup and install scripts.
- See [`/docs/queues/MULTI_WORKER_INFRASTRUCTURE.md`](docs/queues/MULTI_WORKER_INFRASTRUCTURE.md) for advanced queue/worker setup.
- See [`/docs/config/CONFIGURATION.md`](docs/config/CONFIGURATION.md) for environment and config details.

---

Happy hacking!