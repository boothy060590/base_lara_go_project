# Base Laravel Go Project - Architecture Documentation

## Table of Contents

1. [Software Architecture](#software-architecture)
2. [API Implementation](#api-implementation)
3. [Frontend Implementation](#frontend-implementation)
4. [Docker Infrastructure](#docker-infrastructure)
5. [Development Workflow](#development-workflow)

---

## Software Architecture

### Overview

This application follows a **Laravel-inspired architecture** implemented in Go, featuring:

- **Event-driven architecture** with asynchronous job processing
- **Service-oriented design** with dependency injection
- **Facade pattern** for simplified service access
- **Repository pattern** for data access
- **Middleware-based HTTP handling**
- **Queue-based background processing**

### Core Architecture Principles

#### 1. Separation of Concerns
- **API Layer**: HTTP controllers and middleware
- **Service Layer**: Business logic and orchestration
- **Data Layer**: Models, DTOs, and database operations
- **Infrastructure Layer**: External services, queues, and mail

#### 2. Dependency Injection
- Service providers register dependencies
- Interfaces define contracts
- Facades provide simplified access
- Core package contains fundamental interfaces

#### 3. Event-Driven Design
- Events decouple business logic
- Listeners handle side effects
- Asynchronous processing via queues
- Laravel-style event dispatching

### Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Vue.js)                    │
├─────────────────────────────────────────────────────────────┤
│                        API Gateway (Nginx)                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   API       │  │   Worker    │  │   MailHog   │         │
│  │  (Gin)      │  │  (Queue)    │  │  (SMTP)     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   MySQL     │  │ ElasticMQ   │  │   Redis     │         │
│  │  (Database) │  │   (Queue)   │  │  (Cache)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

---

## API Implementation

### Technology Stack

- **Language**: Go 1.23.0
- **Framework**: Gin (HTTP router)
- **ORM**: GORM with MySQL driver
- **Authentication**: JWT tokens
- **Validation**: go-playground/validator
- **Queue**: AWS SQS (ElasticMQ for development)
- **Mail**: SMTP (MailHog for development)

### Project Structure

```
api/
├── app/
│   ├── core/                    # Core interfaces and base types
│   │   ├── base_dto.go         # Base DTO interface
│   │   ├── base_model.go       # Base model with common fields
│   │   ├── database.go         # Database connection and configuration
│   │   ├── event_dispatcher.go # Event dispatching system
│   │   ├── event_registry.go   # Event factory registry
│   │   ├── interfaces.go       # Core service interfaces
│   │   ├── queue_worker.go     # Queue worker implementation
│   │   ├── register.go         # Service registration
│   │   └── registry.go         # Service registry
│   ├── data_objects/           # Data Transfer Objects
│   │   └── auth/
│   │       └── user_dto.go     # User DTO implementation
│   ├── events/                 # Event definitions
│   │   └── auth/
│   │       └── user_created.go # User creation event
│   ├── facades/                # Service facades
│   │   ├── database.go         # Database facade
│   │   ├── event.go            # Event facade
│   │   ├── job.go              # Job facade
│   │   └── mail.go             # Mail facade
│   ├── http/                   # HTTP layer
│   │   ├── controllers/        # HTTP controllers
│   │   ├── middlewares/        # HTTP middlewares
│   │   └── requests/           # Request validation
│   ├── jobs/                   # Background jobs
│   │   └── auth/
│   │       ├── create_user.go
│   │       ├── get_logged_in_user.go
│   │       └── login_user.go
│   ├── listeners/              # Event listeners
│   │   ├── base_listener.go
│   │   └── send_email_confirmation.go
│   ├── models/                 # Database models
│   │   ├── interfaces/         # Model interfaces
│   │   ├── category.go
│   │   ├── permission.go
│   │   ├── role.go
│   │   ├── service.go
│   │   └── user.go
│   ├── providers/              # Service providers
│   │   ├── database_service_provider.go
│   │   ├── event_dispatcher_provider.go
│   │   ├── form_field_validators_provider.go
│   │   ├── job_dispatcher_provider.go
│   │   ├── listener_service_provider.go
│   │   ├── mail_service_provider.go
│   │   ├── message_processor_provider.go
│   │   ├── queue_service_provider.go
│   │   └── router_service_provider.go
│   ├── transformers/           # Data transformers
│   │   └── user_transformer.go
│   ├── utils/                  # Utility functions
│   │   └── token/
│   │       └── token.go
│   └── validators/             # Custom validators
│       └── name_field_validator.go
├── bootstrap/                  # Application bootstrap
│   ├── api/
│   │   └── main.go            # API entry point
│   └── worker/
│       └── main.go            # Worker entry point
├── config/                     # Configuration files
├── database/                   # Database migrations
│   └── migrations/
└── routes/                     # Route definitions
    └── api/
        └── v1/
            └── auth/
                └── auth.go
```

### Key Design Patterns

#### 1. Service Provider Pattern
Service providers register dependencies and configure services:

```go
// Example: Database Service Provider
func (p *DatabaseServiceProvider) Register() {
    // Register database connection
    // Configure GORM
    // Set up migrations
}
```

#### 2. Facade Pattern
Facades provide simplified access to complex services:

```go
// Example: Database Facade
func GetDB() *gorm.DB {
    return database.GetDB()
}
```

#### 3. Event-Driven Architecture
Events decouple business logic from side effects:

```go
// Dispatch event
providers.DispatchEvent("UserCreated", userData)

// Listen for event
type SendEmailConfirmationListener struct{}

func (l *SendEmailConfirmationListener) Handle(event interface{}) error {
    // Send welcome email
    return nil
}
```

#### 4. Repository Pattern
Models implement interfaces for data access:

```go
type UserInterface interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    FindByEmail(email string) (*User, error)
}
```

### Authentication & Authorization

#### JWT Implementation
- **Token Generation**: Custom token utility
- **Middleware**: JWT validation middleware
- **Role-Based Access**: Role middleware for route protection

#### User Roles & Permissions
- **Roles**: Admin, Customer, Engineer
- **Permissions**: Granular permission system
- **Pivot Tables**: Many-to-many relationships

### Database Design

#### Core Tables
- **users**: User accounts and profiles
- **roles**: User roles (admin, customer, engineer)
- **permissions**: System permissions
- **role_user**: Role assignments
- **permission_role**: Permission assignments

#### Migration System
- **GORMigrate**: Database migration tool
- **Version Control**: Migration files with timestamps
- **Rollback Support**: Migration rollback capabilities

### Queue System

#### Architecture
- **Producer**: API dispatches jobs/events
- **Consumer**: Worker processes messages
- **Queue**: ElasticMQ (SQS-compatible)

#### Job Processing
```go
// Job structure
type Job struct {
    Type    string      `json:"type"`
    Data    interface{} `json:"data"`
    Attempt int         `json:"attempt"`
}

// Event structure
type Event struct {
    Type    string      `json:"type"`
    Data    interface{} `json:"data"`
    Time    time.Time   `json:"time"`
}
```

#### Worker Implementation
- **Polling**: Continuous message polling
- **Processing**: Job/event type routing
- **Error Handling**: Retry logic with backoff
- **Graceful Shutdown**: Signal handling

### Mail System

#### SMTP Configuration
- **Development**: MailHog for testing
- **Production**: Configurable SMTP server
- **Templates**: HTML email templates

#### Mail Facade
```go
// Synchronous sending
err := providers.SendMail(recipients, subject, body)

// Asynchronous sending
err := providers.SendMailAsync(recipients, subject, body)
```

---

## Frontend Implementation

### Technology Stack

- **Framework**: Vue.js 3.2.13
- **Build Tool**: Vite 7.0.0
- **State Management**: Pinia 2.3.1 + Vuex 4.1.0
- **Routing**: Vue Router 4.3.2
- **UI Framework**: Bootstrap 5.3.3
- **Validation**: Vee-validate 4.12.8 + Yup 1.4.0
- **HTTP Client**: Axios 1.7.2
- **Styling**: SCSS with custom components

### Project Structure

```
frontend/
├── public/                     # Static assets
├── src/
│   ├── assets/                 # Static resources
│   │   ├── logo.png
│   │   └── scss/               # SCSS stylesheets
│   │       ├── button.scss
│   │       ├── card.scss
│   │       ├── form.scss
│   │       └── utilities.scss
│   ├── components/             # Reusable components
│   │   └── form/               # Form components
│   │       ├── EmailFormField.vue
│   │       ├── PasswordFormField.vue
│   │       ├── TelephoneFormField.vue
│   │       └── TextFormField.vue
│   ├── config.js               # Application configuration
│   ├── form_validators/        # Form validation
│   │   ├── index.js
│   │   ├── login_validator.js
│   │   └── register_validator.js
│   ├── helpers/                # Utility functions
│   │   └── api/
│   │       ├── api.js          # Base API configuration
│   │       └── auth/
│   │           └── authApi.js  # Auth-specific API calls
│   ├── Pages/                  # Page components
│   │   ├── auth/
│   │   │   ├── login/
│   │   │   │   ├── login.scss
│   │   │   │   └── Login.vue
│   │   │   └── register/
│   │   │       ├── Register.scss
│   │   │       └── Register.vue
│   │   └── home/
│   │       ├── admin/
│   │       │   └── Admin.vue
│   │       ├── customer/
│   │       │   └── Customer.vue
│   │       ├── engineer/
│   │       │   └── Engineer.vue
│   │       └── Home.vue
│   ├── plugins/                # Vue plugins
│   │   └── font-awesome.js
│   ├── router.js               # Vue Router configuration
│   ├── store/                  # State management
│   │   └── auth.js
│   ├── App.vue                 # Root component
│   └── main.js                 # Application entry point
├── index.html                  # HTML template
├── package.json                # Dependencies
├── vite.config.js              # Vite configuration
└── jsconfig.json               # JavaScript configuration
```

### Component Architecture

#### 1. Form Components
Reusable form fields with validation:

```vue
<!-- EmailFormField.vue -->
<template>
  <div class="form-group">
    <label :for="id">{{ label }}</label>
    <input
      :id="id"
      type="email"
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
      class="form-control"
      :class="{ 'is-invalid': error }"
    />
    <div v-if="error" class="invalid-feedback">{{ error }}</div>
  </div>
</template>
```

#### 2. Page Components
Role-based page components:

- **Admin.vue**: Administrator dashboard
- **Customer.vue**: Customer portal
- **Engineer.vue**: Engineer interface
- **Login.vue**: Authentication form
- **Register.vue**: User registration

#### 3. State Management
Pinia store for authentication:

```javascript
// store/auth.js
export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    token: null,
    isAuthenticated: false
  }),
  actions: {
    async login(credentials) {
      // Login logic
    },
    async register(userData) {
      // Registration logic
    },
    logout() {
      // Logout logic
    }
  }
})
```

### API Integration

#### HTTP Client Configuration
```javascript
// helpers/api/api.js
import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 10000
})

// Request interceptor for authentication
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
```

#### Auth API Module
```javascript
// helpers/api/auth/authApi.js
export const authApi = {
  login: (credentials) => api.post('/v1/auth/login', credentials),
  register: (userData) => api.post('/v1/auth/register', userData),
  getProfile: () => api.get('/v1/auth/profile')
}
```

### Form Validation

#### Validation Schema
```javascript
// form_validators/register_validator.js
import * as yup from 'yup'

export const registerSchema = yup.object({
  firstName: yup.string().required('First name is required'),
  lastName: yup.string().required('Last name is required'),
  email: yup.string().email('Invalid email').required('Email is required'),
  mobileNumber: yup.string().required('Mobile number is required'),
  password: yup.string().min(8, 'Password must be at least 8 characters').required('Password is required'),
  passwordConfirmation: yup.string().oneOf([yup.ref('password')], 'Passwords must match')
})
```

#### Component Integration
```vue
<template>
  <Form @submit="onSubmit" :validation-schema="registerSchema">
    <EmailFormField name="email" label="Email" />
    <PasswordFormField name="password" label="Password" />
    <!-- Other fields -->
  </Form>
</template>
```

### Styling Architecture

#### SCSS Structure
- **Component-specific styles**: Scoped to components
- **Global utilities**: Reusable utility classes
- **Bootstrap integration**: Custom Bootstrap overrides
- **Responsive design**: Mobile-first approach

#### Custom Components
- **Button styles**: Custom button variants
- **Card components**: Consistent card layouts
- **Form styling**: Enhanced form appearance
- **Utility classes**: Helper classes for common patterns

---

## Docker Infrastructure

### Overview

The application uses a **microservices architecture** with Docker containers:

- **Reverse Proxy**: Nginx with automatic SSL
- **API Service**: Go API with hot reloading
- **Worker Service**: Background job processing
- **Frontend**: Vue.js development server
- **Database**: MySQL with persistent storage
- **Queue**: ElasticMQ (SQS-compatible)
- **Mail**: MailHog for development
- **Cache**: Redis for session storage
- **DNS**: dnsmasq for local domain resolution

### Container Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Nginx Reverse Proxy                      │
│              (SSL Termination + Load Balancing)             │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Frontend  │  │     API     │  │   Worker    │         │
│  │   (Vue.js)  │  │   (Gin)     │  │  (Queue)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │    MySQL    │  │  ElasticMQ  │  │   Redis     │         │
│  │  (Database) │  │   (Queue)   │  │  (Cache)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   MailHog   │  │   MinIO     │  │  dnsmasq    │         │
│  │   (SMTP)    │  │  (S3)       │  │   (DNS)     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### Service Configuration

#### 1. Nginx Reverse Proxy
```yaml
nginx:
  image: nginxproxy/nginx-proxy:alpine
  ports:
    - "80:80"
    - "443:443"
  volumes:
    - /var/run/docker.sock:/tmp/docker.sock:ro
    - ./docker/ssl/certs:/etc/nginx/certs:ro
```

**Features:**
- **Automatic SSL**: Self-signed certificates
- **Service Discovery**: Automatic container detection
- **Load Balancing**: Multiple backend support
- **Virtual Hosts**: Domain-based routing

#### 2. API Service
```yaml
api:
  build:
    context: ./api
    dockerfile: ../docker/api/Dockerfile
  volumes:
    - ./api:/usr/src/app
  environment:
    - VIRTUAL_HOST=api.baselaragoproject.test
    - VIRTUAL_PORT=8080
    - HTTPS_METHOD=static
```

**Features:**
- **Hot Reloading**: Air for development
- **Multi-stage Build**: Optimized production images
- **Volume Mounting**: Live code changes
- **Environment Variables**: Configurable settings

#### 3. Worker Service
```yaml
worker:
  build:
    context: ./api
    dockerfile: ../docker/worker/Dockerfile
  environment:
    - VIRTUAL_HOST=worker.baselaragoproject.test
    - VIRTUAL_PORT=8081
  restart: unless-stopped
```

**Features:**
- **Background Processing**: Queue message handling
- **Auto-restart**: Automatic recovery
- **Health Monitoring**: Process monitoring
- **Graceful Shutdown**: Signal handling

#### 4. Frontend Service
```yaml
app:
  build:
    context: ./frontend
    dockerfile: ../docker/frontend/Dockerfile
  environment:
    - VIRTUAL_HOST=app.baselaragoproject.test
    - VIRTUAL_PORT=5173
    - VIRTUAL_PROTO=https
```

**Features:**
- **Vite Dev Server**: Fast development
- **HTTPS Support**: SSL termination
- **Hot Module Replacement**: Live updates
- **Build Optimization**: Production builds

### SSL Configuration

#### Certificate Generation
```bash
# Generate self-signed certificates
./docker/ssl/gen_certs.sh

# Trust certificates on macOS
./docker/ssl/trust_certs_mac.sh
```

#### SSL Structure
```
docker/ssl/
├── certs/                      # Certificate files
│   ├── baselaragoproject.test.crt
│   ├── baselaragoproject.test.key
│   └── dhparam.pem
├── dhparam/                    # Diffie-Hellman parameters
├── gen_certs.sh               # Certificate generation script
├── trust_certs_mac.sh         # macOS trust script
└── vhost.d/                   # Virtual host configurations
```

### DNS Configuration

#### dnsmasq Setup
```yaml
dnsmasq:
  image: 4km3/dnsmasq:latest
  ports:
    - "54:53/udp"
  volumes:
    - ./docker/dnsmasq/dnsmasq.conf:/etc/dnsmasq.conf
```

#### DNS Configuration
```conf
# docker/dnsmasq/dnsmasq.conf
address=/baselaragoproject.test/127.0.0.1
listen-address=0.0.0.0
bind-interfaces
```

**Features:**
- **Local Domains**: `.test` domain resolution
- **Container Discovery**: Automatic service discovery
- **Port Forwarding**: Local development access
- **Cross-platform**: Works on macOS, Linux, Windows

### Queue Configuration

#### ElasticMQ Setup
```yaml
elasticmq:
  image: softwaremill/elasticmq
  ports:
    - "9324:9324"
  volumes:
    - ./docker/elasticmq/elasticmq.conf:/opt/elasticmq.conf
```

#### Queue Configuration
```hocon
# docker/elasticmq/elasticmq.conf
include classpath("application.conf")

node-address {
    protocol = http
    host = "0.0.0.0"
    port = 9324
    context-path = ""
}

rest-sqs {
    enabled = true
    bind-port = 9324
    bind-hostname = "0.0.0.0"
    sqs-limits = strict
}

queues {
    default {
        defaultVisibilityTimeout = 10 seconds
        delay = 5 seconds
        receiveMessageWait = 0 seconds
        deadLettersQueue {
            name = "default-dead-letters"
            maxReceiveCount = 3
        }
    }
}
```

### Database Configuration

#### MySQL Setup
```yaml
db:
  image: mysql
  ports:
    - "3309:3306"
  volumes:
    - ./my-db:/var/lib/mysql
  environment:
    - MYSQL_ROOT_PASSWORD=D3vB4s3L4r4G0!
    - MYSQL_DATABASE=dev_base_lara_go
    - MYSQL_USER=api_user
    - MYSQL_PASSWORD=b4s3L4r4G0212!
```

**Features:**
- **Persistent Storage**: Data persistence across restarts
- **Port Mapping**: Local development access
- **Environment Variables**: Configurable credentials
- **Volume Mounting**: Data directory persistence

### Mail Configuration

#### MailHog Setup
```yaml
mailhog:
  image: jcalonso/mailhog
  environment:
    - VIRTUAL_HOST=mail.baselaragoproject.test
    - VIRTUAL_PORT=8025
```

**Features:**
- **SMTP Server**: Development mail server
- **Web Interface**: Email viewing interface
- **Message Capture**: All outgoing emails
- **No External Dependencies**: Self-contained

### Development Workflow

#### 1. Local Development
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Access services
open https://app.baselaragoproject.test
open https://api.baselaragoproject.test
open https://mail.baselaragoproject.test
```

#### 2. Code Changes
- **API**: Automatic reloading with Air
- **Frontend**: Hot module replacement
- **Database**: Migration-based schema changes
- **Configuration**: Environment variable updates

#### 3. Production Deployment
```bash
# Build production images
docker-compose -f docker-compose.prod.yml build

# Deploy to production
docker-compose -f docker-compose.prod.yml up -d
```

### Security Considerations

#### 1. SSL/TLS
- **Self-signed Certificates**: Development only
- **Certificate Authority**: Production certificates
- **HTTPS Enforcement**: All traffic encrypted
- **Certificate Renewal**: Automated renewal process

#### 2. Network Security
- **Container Isolation**: Network segmentation
- **Port Exposure**: Minimal port exposure
- **Service Discovery**: Internal communication
- **Firewall Rules**: Network-level protection

#### 3. Data Security
- **Database Encryption**: At-rest encryption
- **Password Hashing**: Secure password storage
- **JWT Tokens**: Secure authentication
- **Environment Variables**: Sensitive data protection

---

## Development Workflow

### Prerequisites

#### Required Software
- **Docker Desktop**: Container runtime
- **Go 1.23+**: Backend development
- **Node.js 18+**: Frontend development
- **Git**: Version control
- **Code Editor**: VS Code recommended

#### System Requirements
- **macOS**: 10.15+ (Catalina)
- **Linux**: Ubuntu 20.04+
- **Windows**: Windows 10+ with WSL2
- **Memory**: 8GB RAM minimum
- **Storage**: 10GB free space

### Environment Setup

#### 1. Clone Repository
```bash
git clone <repository-url>
cd base_lara_go_project
```

#### 2. Generate SSL Certificates
```bash
# Generate certificates
./docker/ssl/gen_certs.sh

# Trust certificates (macOS)
./docker/ssl/trust_certs_mac.sh
```

#### 3. Configure DNS
```bash
# Add to /etc/hosts (Linux/macOS)
echo "127.0.0.1 baselaragoproject.test" | sudo tee -a /etc/hosts
echo "127.0.0.1 api.baselaragoproject.test" | sudo tee -a /etc/hosts
echo "127.0.0.1 app.baselaragoproject.test" | sudo tee -a /etc/hosts
echo "127.0.0.1 mail.baselaragoproject.test" | sudo tee -a /etc/hosts
```

#### 4. Environment Configuration
```bash
# Copy environment template
cp api/.env.example api/.env

# Edit environment variables
nano api/.env
```

### Development Commands

#### Docker Operations
```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Rebuild services
docker-compose build

# View logs
docker-compose logs -f [service-name]

# Execute commands in containers
docker-compose exec api go run main.go
docker-compose exec app npm run dev
```

#### API Development
```bash
# Run API locally
cd api
go run bootstrap/api/main.go

# Run worker locally
cd api
go run bootstrap/worker/main.go

# Run tests
go test ./...

# Run migrations
go run main.go migrate
```

#### Frontend Development
```bash
# Install dependencies
cd frontend
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm test
```

### Testing

#### API Testing
```bash
# Test endpoints
curl -X POST https://api.baselaragoproject.test/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"firstName":"John","lastName":"Doe","email":"john@example.com","mobileNumber":"1234567890","password":"password123"}'

# Test mail functionality
curl -X POST https://api.baselaragoproject.test/v1/auth/test-mail
```

#### Frontend Testing
```bash
# Access frontend
open https://app.baselaragoproject.test

# Access mail interface
open https://mail.baselaragoproject.test
```

### Deployment

#### Production Build
```bash
# Build production images
docker-compose -f docker-compose.prod.yml build

# Deploy to production
docker-compose -f docker-compose.prod.yml up -d
```

#### Environment Configuration
```bash
# Production environment
cp api/.env.example api/.env.prod
# Edit production variables
nano api/.env.prod
```

### Troubleshooting

#### Common Issues

1. **SSL Certificate Errors**
   ```bash
   # Regenerate certificates
   ./docker/ssl/gen_certs.sh
   ./docker/ssl/trust_certs_mac.sh
   ```

2. **DNS Resolution Issues**
   ```bash
   # Check dnsmasq
   docker-compose logs dnsmasq
   
   # Test DNS resolution
   nslookup api.baselaragoproject.test
   ```

3. **Database Connection Issues**
   ```bash
   # Check database logs
   docker-compose logs db
   
   # Test connection
   docker-compose exec api go run main.go migrate
   ```

4. **Queue Processing Issues**
   ```bash
   # Check worker logs
   docker-compose logs worker
   
   # Check queue status
   curl http://sqs.baselaragoproject.test:9324
   ```

#### Performance Optimization

1. **Docker Resource Limits**
   ```yaml
   # docker-compose.yaml
   services:
     api:
       deploy:
         resources:
           limits:
             memory: 1G
             cpus: '0.5'
   ```

---

## Conclusion

This architecture provides a robust, scalable foundation for the Base Laravel Go Project application. The Laravel-inspired patterns in Go, combined with modern frontend technologies and containerized infrastructure, create a development experience that is both powerful and maintainable.

Key benefits of this architecture:

- **Separation of Concerns**: Clear boundaries between layers
- **Event-Driven Design**: Loose coupling and extensibility
- **Containerized Deployment**: Consistent environments
- **Development Experience**: Hot reloading and fast feedback
- **Scalability**: Microservices architecture
- **Security**: SSL/TLS and proper authentication
- **Maintainability**: Well-structured codebase

The documentation provides comprehensive guidance for development, deployment, and maintenance of the application.