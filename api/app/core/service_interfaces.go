package core

import (
	"context"
)

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
