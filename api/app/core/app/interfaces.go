package app_core

import (
	"context"
	"time"
)

// EventInterface defines the interface for all events
type EventInterface interface {
	GetEventName() string
}

// ListenerInterface defines the interface for all listeners
type ListenerInterface interface {
	Handle(mailService interface{}) error
}

// JobInterface defines the interface for all jobs
type JobInterface interface {
	Handle() (any, error)
}

// CacheInterface defines the interface for cache operations
type CacheInterface interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl ...time.Duration) error
	Delete(key string) error
	Has(key string) bool
	Flush() error
}

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	// Basic operations
	Create(value interface{}) error
	First(dest interface{}, conds ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	Save(value interface{}) error
	Delete(value interface{}, conds ...interface{}) error

	// Query builder
	Table(tableName string) DatabaseInterface
	Where(query interface{}, args ...interface{}) DatabaseInterface
	Or(query interface{}, args ...interface{}) DatabaseInterface
	Order(value interface{}) DatabaseInterface
	Limit(limit int) DatabaseInterface
	Offset(offset int) DatabaseInterface
	Preload(query string, args ...interface{}) DatabaseInterface
	Joins(query string, args ...interface{}) DatabaseInterface

	// Model operations
	Model(value interface{}) DatabaseInterface

	// Transaction support
	Transaction(fc func(tx DatabaseInterface) error) error

	// Raw query support
	Raw(sql string, values ...interface{}) DatabaseInterface
	Exec(sql string, values ...interface{}) error

	// Migration support
	Migrate() error

	// Get underlying DB instance
	GetDB() interface{}
}

// ModelInterface defines the interface for all models
type ModelInterface interface {
	GetID() uint
	GetTableName() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
}

// BaseModelInterface extends ModelInterface with base functionality
type BaseModelInterface interface {
	ModelInterface
	Set(key string, value interface{})
	Get(key string) interface{}
	Has(key string) bool
	GetData() map[string]interface{}
}

// RepositoryInterface defines the interface for repository operations
type RepositoryInterface interface {
	Find(id uint) (ModelInterface, error)
	FindAll() ([]ModelInterface, error)
	Create(model ModelInterface) error
	Update(model ModelInterface) error
	Delete(model ModelInterface) error
	Where(query interface{}, args ...interface{}) RepositoryInterface
	First(dest interface{}, conds ...interface{}) error
	Save(value interface{}) error
	Transaction(fc func(tx RepositoryInterface) error) error
	Raw(sql string, values ...interface{}) RepositoryInterface
	Exec(sql string, values ...interface{}) error
	Preload(query string, args ...interface{}) RepositoryInterface
	Order(value interface{}) RepositoryInterface
	Limit(limit int) RepositoryInterface
	Offset(offset int) RepositoryInterface
	GetDB() interface{}
}

// ServiceContainerInterface interface for dependency injection
type ServiceContainerInterface interface {
	// Basic operations
	Get(key string) interface{}
	Set(key string, value interface{})
	Has(key string) bool

	// Laravel-style operations
	Bind(abstract string, concrete interface{})
	BindWithResolver(abstract string, resolver func() interface{})
	Singleton(abstract string, concrete interface{})
	SingletonWithResolver(abstract string, resolver func() interface{})
	Resolve(abstract string) (interface{}, error)
	ResolveOrFail(abstract string) interface{}
	Forget(abstract string)
	Flush()
}

// ClientConfig defines the configuration for clients
type ClientConfig struct {
	Driver   string                 `json:"driver"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Username string                 `json:"username"`
	Password string                 `json:"password"`
	Database string                 `json:"database"`
	SSLMode  string                 `json:"ssl_mode"`
	Options  map[string]interface{} `json:"options"`
	Timeout  time.Duration          `json:"timeout"`
	Retries  int                    `json:"retries"`
}

// BaseClient defines the base client interface
type BaseClient interface {
	GetConfig() *ClientConfig
	Connect() error
	Disconnect() error
	IsConnected() bool
}

// ===== SERVICE INTERFACES =====

// BaseServiceInterface defines common CRUD operations for all services
type BaseServiceInterface[T any] interface {
	// Create operations
	Create(data map[string]interface{}) (T, error)
	CreateWithContext(ctx context.Context, data map[string]interface{}) (T, error)

	// Read operations
	FindByID(id uint) (T, error)
	FindByIDWithContext(ctx context.Context, id uint) (T, error)
	FindByField(field string, value interface{}) (T, error)
	FindByFieldWithContext(ctx context.Context, field string, value interface{}) (T, error)
	All() ([]T, error)
	AllWithContext(ctx context.Context) ([]T, error)
	Paginate(page, perPage int) ([]T, int64, error)
	PaginateWithContext(ctx context.Context, page, perPage int) ([]T, int64, error)

	// Update operations
	Update(id uint, data map[string]interface{}) (T, error)
	UpdateWithContext(ctx context.Context, id uint, data map[string]interface{}) (T, error)
	UpdateOrCreate(conditions map[string]interface{}, data map[string]interface{}) (T, error)
	UpdateOrCreateWithContext(ctx context.Context, conditions map[string]interface{}, data map[string]interface{}) (T, error)

	// Delete operations
	Delete(id uint) error
	DeleteWithContext(ctx context.Context, id uint) error
	DeleteWhere(conditions map[string]interface{}) error
	DeleteWhereWithContext(ctx context.Context, conditions map[string]interface{}) error

	// Utility operations
	Exists(id uint) (bool, error)
	ExistsWithContext(ctx context.Context, id uint) (bool, error)
	Count() (int64, error)
	CountWithContext(ctx context.Context) (int64, error)
	CountWhere(conditions map[string]interface{}) (int64, error)
	CountWhereWithContext(ctx context.Context, conditions map[string]interface{}) (int64, error)
}

// CacheableServiceInterface extends BaseServiceInterface with caching capabilities
type CacheableServiceInterface[T any] interface {
	BaseServiceInterface[T]

	// Cache operations
	FindByIDCached(id uint) (T, error)
	FindByIDCachedWithContext(ctx context.Context, id uint) (T, error)
	FindByFieldCached(field string, value interface{}) (T, error)
	FindByFieldCachedWithContext(ctx context.Context, field string, value interface{}) (T, error)
	AllCached() ([]T, error)
	AllCachedWithContext(ctx context.Context) ([]T, error)

	// Cache invalidation
	InvalidateCache(id uint) error
	InvalidateCacheWithContext(ctx context.Context, id uint) error
	InvalidateAllCache() error
	InvalidateAllCacheWithContext(ctx context.Context) error
}

// SearchableServiceInterface extends BaseServiceInterface with search capabilities
type SearchableServiceInterface[T any] interface {
	BaseServiceInterface[T]

	// Search operations
	Search(query string, fields []string) ([]T, error)
	SearchWithContext(ctx context.Context, query string, fields []string) ([]T, error)
	SearchPaginated(query string, fields []string, page, perPage int) ([]T, int64, error)
	SearchPaginatedWithContext(ctx context.Context, query string, fields []string, page, perPage int) ([]T, int64, error)
}

// AuditableServiceInterface extends BaseServiceInterface with audit capabilities
type AuditableServiceInterface[T any] interface {
	BaseServiceInterface[T]

	// Audit operations
	GetAuditLog(id uint) ([]AuditLog, error)
	GetAuditLogWithContext(ctx context.Context, id uint) ([]AuditLog, error)
	GetAuditLogByField(field string, value interface{}) ([]AuditLog, error)
	GetAuditLogByFieldWithContext(ctx context.Context, field string, value interface{}) ([]AuditLog, error)
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        uint        `json:"id"`
	UserID    *uint       `json:"user_id"`
	Action    string      `json:"action"`
	Table     string      `json:"table"`
	RecordID  uint        `json:"record_id"`
	OldValues interface{} `json:"old_values"`
	NewValues interface{} `json:"new_values"`
	CreatedAt string      `json:"created_at"`
}

// ServiceOptions provides configuration options for services
type ServiceOptions struct {
	EnableCache  bool
	EnableAudit  bool
	EnableSearch bool
	CacheTTL     int64
	SearchFields []string
	AuditFields  []string
}

// ServiceFactory creates services with specific options
type ServiceFactory[T any] interface {
	Create(options *ServiceOptions) (BaseServiceInterface[T], error)
	CreateWithContext(ctx context.Context, options *ServiceOptions) (BaseServiceInterface[T], error)
}

// ===== LOGGING INTERFACES =====

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

// LoggerInterface defines the interface for logging operations
type LoggerInterface interface {
	Debug(message string, context map[string]interface{}) error
	Info(message string, context map[string]interface{}) error
	Warning(message string, context map[string]interface{}) error
	Error(message string, context map[string]interface{}) error
	Fatal(message string, context map[string]interface{}) error

	// Laravel-style methods
	Log(level LogLevel, message string, context map[string]interface{}) error
	Emergency(message string, context map[string]interface{}) error
	Alert(message string, context map[string]interface{}) error
	Critical(message string, context map[string]interface{}) error
	Notice(message string, context map[string]interface{}) error

	// Context methods
	WithContext(ctx context.Context) LoggerInterface
	WithFields(fields map[string]interface{}) LoggerInterface
}

// LogDriver defines the interface for log drivers
type LogDriver interface {
	Log(level LogLevel, message string, context map[string]interface{}) error
	Close() error
}

// ===== CLIENT INTERFACES =====

// ClientInterface defines the base interface for all clients
type ClientInterface interface {
	// Connect establishes a connection to the service
	Connect() error

	// Disconnect closes the connection
	Disconnect() error

	// IsConnected checks if the client is connected
	IsConnected() bool

	// GetConfig returns the client configuration
	GetConfig() *ClientConfig

	// GetName returns the client name/type
	GetName() string
}

// HTTPClientInterface defines the interface for HTTP clients
type HTTPClientInterface interface {
	ClientInterface

	// Request makes an HTTP request
	Request(method, url string, headers map[string]string, body interface{}) ([]byte, error)

	// Get makes a GET request
	Get(url string, headers map[string]string) ([]byte, error)

	// Post makes a POST request
	Post(url string, headers map[string]string, body interface{}) ([]byte, error)

	// Put makes a PUT request
	Put(url string, headers map[string]string, body interface{}) ([]byte, error)

	// Delete makes a DELETE request
	Delete(url string, headers map[string]string) ([]byte, error)

	// SetTimeout sets the request timeout
	SetTimeout(timeout time.Duration)

	// SetRetries sets the number of retries
	SetRetries(retries int)
}

// LoggingClientInterface defines the interface for logging clients
type LoggingClientInterface interface {
	ClientInterface

	// Log logs a message with the specified level
	Log(level string, message string, context map[string]interface{}) error

	// LogException logs an exception
	LogException(exception error) error

	// Flush flushes any buffered logs
	Flush() error

	// SetLevel sets the logging level
	SetLevel(level string) error

	// GetLevel returns the current logging level
	GetLevel() string
}

// DatabaseClientInterface defines the interface for database clients
type DatabaseClientInterface interface {
	ClientInterface

	// Query executes a query and returns results
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)

	// Execute executes a query without returning results
	Execute(query string, args ...interface{}) (int64, error)

	// BeginTransaction begins a new transaction
	BeginTransaction() (TransactionInterface, error)

	// Ping checks if the database is reachable
	Ping() error

	// GetStats returns database statistics
	GetStats() map[string]interface{}
}

// TransactionInterface defines the interface for database transactions
type TransactionInterface interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// Query executes a query within the transaction
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)

	// Execute executes a query within the transaction
	Execute(query string, args ...interface{}) (int64, error)
}

// CacheClientInterface defines the interface for cache clients
type CacheClientInterface interface {
	ClientInterface

	// Get retrieves a value from cache
	Get(key string) (interface{}, bool, error)

	// Set stores a value in cache
	Set(key string, value interface{}, ttl ...int) error

	// Delete removes a value from cache
	Delete(key string) error

	// Has checks if a key exists in cache
	Has(key string) (bool, error)

	// Clear clears all cache entries
	Clear() error

	// Flush clears all cache entries
	Flush() error

	// Increment increments a numeric value
	Increment(key string, value int) (int, error)

	// Decrement decrements a numeric value
	Decrement(key string, value int) (int, error)

	// GetStats returns cache statistics
	GetStats() map[string]interface{}
}

// QueueClientInterface defines the interface for queue clients
type QueueClientInterface interface {
	ClientInterface

	// Push adds a job to the queue
	Push(queue string, job interface{}) error

	// Pop retrieves a job from the queue
	Pop(queue string) (interface{}, error)

	// Delete removes a job from the queue
	Delete(queue string, job interface{}) error

	// Size returns the number of jobs in the queue
	Size(queue string) (int, error)

	// Clear clears all jobs from the queue
	Clear(queue string) error

	// GetStats returns queue statistics
	GetStats() map[string]interface{}
}

// FileSystemClientInterface defines the interface for filesystem clients
type FileSystemClientInterface interface {
	ClientInterface

	// Read reads a file
	Read(path string) ([]byte, error)

	// Write writes data to a file
	Write(path string, data []byte) error

	// Delete deletes a file
	Delete(path string) error

	// Exists checks if a file exists
	Exists(path string) bool

	// List lists files in a directory
	List(path string) ([]string, error)

	// Mkdir creates a directory
	Mkdir(path string, mode int) error

	// Rmdir removes a directory
	Rmdir(path string) error

	// GetSize returns the size of a file
	GetSize(path string) (int64, error)

	// GetModifiedTime returns the last modified time of a file
	GetModifiedTime(path string) (time.Time, error)
}

// MailClientInterface defines the interface for mail clients
type MailClientInterface interface {
	ClientInterface

	// Send sends an email
	Send(to []string, subject string, body string, options map[string]interface{}) error

	// SendWithAttachments sends an email with attachments
	SendWithAttachments(to []string, subject string, body string, attachments []string, options map[string]interface{}) error

	// GetStats returns mail statistics
	GetStats() map[string]interface{}
}

// ===== CACHE MODEL INTERFACES =====

// CacheModelInterface defines the interface for cacheable models
type CacheModelInterface interface {
	BaseModelInterface
	GetBaseKey() string
	GetCacheKey() string
	GetCacheTTL() time.Duration
	GetCacheData() interface{}
	GetCacheTags() []string
	FromCacheData(data map[string]interface{}) error
}

// ===== DATABASE MODEL INTERFACES =====

// DatabaseModelInterface defines the interface for database models
type DatabaseModelInterface interface {
	BaseModelInterface
	GetTableName() string
	GetPrimaryKey() string
	GetConnection() string
	GetFillable() []string
	GetHidden() []string
	GetDates() []string
	GetCasts() map[string]string
	GetRelations() map[string]interface{}
}

// ===== DATABASE PROVIDER INTERFACES =====

// DatabaseProviderInterface defines the interface for database providers
type DatabaseProviderInterface interface {
	Connect() error
	GetConnection() DatabaseInterface
	Close() error
}

// DatabaseProviderServiceInterface defines the interface for database provider services
type DatabaseProviderServiceInterface interface {
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)
	Execute(query string, args ...interface{}) (int64, error)
	BeginTransaction() (TransactionInterface, error)
	Ping() error
	GetStats() map[string]interface{}
}

// ===== JOB PROCESSING INTERFACES =====

// JobProcessor defines the interface for processing specific job types
type JobProcessor interface {
	CanProcess(jobType string) bool
	Process(jobData []byte) error
}

// JobDispatcherService defines the interface for job dispatching operations
type JobDispatcherService interface {
	Dispatch(job JobInterface) error
	DispatchSync(job JobInterface) (any, error)
	DispatchJob(job interface{}, queueName string) error
	DispatchJobWithAttributes(job interface{}, attributes map[string]string, queueName string) error
	ProcessJobFromQueue(jobData []byte, jobType string) error
	RegisterJobProcessor(processor JobProcessor)
}

// ===== QUEUE INTERFACES =====

// QueueService defines the interface for queue operations
type QueueService interface {
	SendMessage(messageBody string) error
	SendMessageToQueue(messageBody string, queueName string) error
	SendMessageWithAttributes(messageBody string, attributes map[string]string) error
	SendMessageToQueueWithAttributes(messageBody string, attributes map[string]string, queueName string) error
	ReceiveMessage() (interface{}, error)
	ReceiveMessageFromQueue(queueName string) (interface{}, error)
	DeleteMessage(receiptHandle string) error
	DeleteMessageFromQueue(receiptHandle string, queueName string) error
}

// QueueWorker defines the interface for queue workers
type QueueWorker interface {
	Start()
	Stop()
}

// ===== OBSERVER INTERFACES =====

// ModelObserver interface for observing model events
type ModelObserver interface {
	// Model events
	Created(tx interface{}) error
	Updated(tx interface{}) error
	Deleted(tx interface{}) error
	Saved(tx interface{}) error
}

// CacheableModel defines the interface for cacheable models
type CacheableModel interface {
	GetCacheKey() string
	GetCacheTags() []string
}

// ===== LOGGING HANDLER INTERFACES =====

// LogHandler defines the interface for log handlers
type LogHandler interface {
	Handle(level string, message string, context map[string]interface{}) error
	ShouldHandle(level string) bool
	GetLevel() string
	Flush() error
	Close() error
}

// ExceptionHandler defines the interface for exception handlers
type ExceptionHandler interface {
	Report(exception error) error
	Render(exception error) interface{}
	ShouldReport(exception error) bool
	ShouldRender(exception error) bool
	Flush() error
}

// ===== MESSAGE PROCESSING INTERFACES =====

// MessageProcessorService defines the interface for message processing operations
type MessageProcessorService interface {
	ProcessMessage(message interface{}) error
	ProcessMessages(messages []interface{}) error
	GetJobTypeFromMessage(message interface{}) string
	GetQueueNameFromMessage(message interface{}) string
}

// ===== MAIL INTERFACES =====

// MailService defines the interface for mail operations
type MailService interface {
	SendMail(to []string, subject, body string) error
	SendMailAsync(to []string, subject, body string, queueName string) error
	ProcessMailJobFromQueue(jobData []byte) error
}

// EmailTemplateEngine defines the interface for email template rendering
type EmailTemplateEngine interface {
	Render(templateName string, data interface{}) (string, error)
	PreloadTemplates() error
}

// ===== EXCEPTION INTERFACES =====

// Exception defines the interface for framework exceptions
type Exception interface {
	GetCode() int
	GetMessage() string
	GetFile() string
	GetLine() int
	GetTrace() []string
	GetContext() map[string]interface{}
}

// CacheProviderServiceInterface defines the interface for cache provider services
type CacheProviderServiceInterface interface {
	Connect() error
	Disconnect() error
	Get(key string) (interface{}, bool, error)
	Set(key string, value interface{}, ttl ...int) error
	Delete(key string) error
	Clear() error
	Has(key string) (bool, error)
	Increment(key string, value int) (int, error)
	Decrement(key string, value int) (int, error)
	GetStats() map[string]interface{}
	GetClient() CacheClientInterface
}

// MailProviderServiceInterface defines the interface for mail provider services
type MailProviderServiceInterface interface {
	Connect() error
	Disconnect() error
	Send(to []string, subject string, body string, options map[string]interface{}) error
	SendWithAttachments(to []string, subject string, body string, attachments []string, options map[string]interface{}) error
	GetStats() map[string]interface{}
	GetClient() MailClientInterface
}

// QueueProviderServiceInterface defines the interface for queue provider services
type QueueProviderServiceInterface interface {
	Connect() error
	Disconnect() error
	Push(queue string, job interface{}) error
	Pop(queue string) (interface{}, error)
	Delete(queue string, job interface{}) error
	Size(queue string) (int, error)
	Clear(queue string) error
	GetStats() map[string]interface{}
	GetClient() QueueClientInterface
}
