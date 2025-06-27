# Dockerised GoLang api with VueJS Frontend. Comes with authentication already handled. 

# Base Laravel Go Project

A Go-based web application with Laravel-style architecture, featuring async event processing, mail queue management, and modern web development practices.

## 🚀 Features

- **Laravel-Style Architecture**: Familiar patterns and structure for Laravel developers
- **Async Event Processing**: Event-driven architecture with queue-based processing
- **Mail Queue Management**: Asynchronous email sending via dedicated mail queue
- **Multi-Queue System**: Separate queues for jobs, mail, and events
- **Real-time Queue Processing**: Ultra-fast concurrent queue processing
- **JWT Authentication**: Secure token-based authentication
- **Database Integration**: GORM v2 with MySQL support
- **Docker Development**: Complete containerized development environment
- **Vue.js Frontend**: Modern reactive frontend with form validation

## 🏗️ Architecture

### Core Components

- **Event System**: Async event dispatching and processing
- **Queue System**: Multi-queue processing with ElasticMQ
- **Mail System**: Template-based email sending with queue support
- **Job System**: Background job processing
- **Authentication**: JWT-based user authentication and authorization

### Queue Structure

- **Events Queue**: Handles application events (user registration, etc.)
- **Mail Queue**: Processes email sending tasks
- **Jobs Queue**: General background job processing

## 🛠️ Technology Stack

### Backend
- **Go 1.24+**: Core application language
- **Gin**: HTTP web framework
- **GORM v2**: Database ORM
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

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd base_lara_go_project
   ```

2. **Start the development environment**
   ```bash
   docker-compose up -d
   ```

3. **Access the application**
   - **Frontend**: https://app.baselaragoproject.test
   - **API**: https://api.baselaragoproject.test
   - **Mail Testing**: http://mail.baselaragoproject.test:8025

### Development

The application uses a hot-reload system for both frontend and backend:

- **Frontend**: Automatic reloading with Vite
- **Backend**: Air for Go hot-reloading
- **Worker**: Automatic restart on code changes

## 📁 Project Structure

```
base_lara_go_project/
├── api/                    # Go backend application
│   ├── app/
│   │   ├── core/          # Core business logic
│   │   ├── events/        # Event definitions
│   │   ├── jobs/          # Background jobs
│   │   ├── listeners/     # Event listeners
│   │   ├── models/        # Data models
│   │   └── providers/     # Service providers
│   ├── bootstrap/         # Application bootstrap
│   ├── config/           # Configuration files
│   └── routes/           # API routes
├── frontend/              # Vue.js frontend
├── docker/               # Docker configuration
└── docs/                 # Documentation
```

## 🔧 Configuration

### Environment Variables

Key environment variables for the API:

```env
# Application
APP_NAME=Base Laravel Go Project
APP_ENV=development
APP_DEBUG=false
APP_URL=http://localhost

# Database
DB_CONNECTION=mysql
DB_HOST=db
DB_PORT=3306
DB_NAME=dev_base_lara_go
DB_USER=api_user
DB_PASSWORD=b4s3L4r4G0212!

# Queue
QUEUE_CONNECTION=sqs
SQS_ENDPOINT=http://sqs.baselaragoproject.test:9324
SQS_QUEUE_JOBS=jobs
SQS_QUEUE_MAIL=mail
SQS_QUEUE_EVENTS=events

# Mail
MAIL_MAILER=smtp
MAIL_HOST=mail.baselaragoproject.test
MAIL_PORT=1025
MAIL_FROM_ADDRESS=no-reply@baselaragoproject.test
```

## 📚 Usage Examples

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

## 🔄 Queue Processing

The application uses a multi-queue system with ultra-fast processing:

- **Concurrent Queue Processing**: All queues processed simultaneously
- **Zero Wait Time**: Instant message polling
- **Concurrent Message Processing**: Multiple messages processed concurrently
- **50ms Polling Cycle**: Ultra-responsive queue monitoring

### Queue Flow

1. **User Registration** → API creates user and dispatches `UserCreated` event
2. **Event Queue** → Event sent to `events` queue with `job_type: event`
3. **Event Processing** → Worker processes event from `events` queue
4. **Email Queueing** → Event listener queues email to `mail` queue
5. **Email Processing** → Worker processes email from `mail` queue
6. **Email Sending** → Email sent via SMTP

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

## 📖 Documentation

- [Architecture Documentation](docs/ARCHITECTURE.md)
- [API Documentation](docs/API.md)
- [Development Guide](docs/DEVELOPMENT.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.