package core

import (
	"gorm.io/gorm"
)

// DatabaseInterface defines the core database operations
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

	// Get underlying GORM DB instance
	GetDB() *gorm.DB
}

// DatabaseProvider implements the core DatabaseInterface
type DatabaseProvider struct {
	db *gorm.DB
}

// NewDatabaseProvider creates a new database provider
func NewDatabaseProvider(db *gorm.DB) *DatabaseProvider {
	return &DatabaseProvider{db: db}
}

// Basic operations that are used by the facade
func (d *DatabaseProvider) Create(value interface{}) error {
	return d.db.Create(value).Error
}

func (d *DatabaseProvider) First(dest interface{}, conds ...interface{}) error {
	return d.db.First(dest, conds...).Error
}

func (d *DatabaseProvider) Find(dest interface{}, conds ...interface{}) error {
	return d.db.Find(dest, conds...).Error
}

func (d *DatabaseProvider) Save(value interface{}) error {
	return d.db.Save(value).Error
}

func (d *DatabaseProvider) Delete(value interface{}, conds ...interface{}) error {
	return d.db.Delete(value, conds...).Error
}

// Query builder methods that are used by the facade
func (d *DatabaseProvider) Table(tableName string) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Table(tableName)}
}

func (d *DatabaseProvider) Where(query interface{}, args ...interface{}) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Where(query, args...)}
}

func (d *DatabaseProvider) Preload(query string, args ...interface{}) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Preload(query, args...)}
}

func (d *DatabaseProvider) Model(value interface{}) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Model(value)}
}

// Additional methods that might be needed by the facade
func (d *DatabaseProvider) Order(value interface{}) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Order(value)}
}

func (d *DatabaseProvider) Limit(limit int) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Limit(limit)}
}

func (d *DatabaseProvider) Offset(offset int) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Offset(offset)}
}

// Additional methods required by the interface
func (d *DatabaseProvider) Or(query interface{}, args ...interface{}) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Or(query, args...)}
}

func (d *DatabaseProvider) Joins(query string, args ...interface{}) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Joins(query, args...)}
}

func (d *DatabaseProvider) Transaction(fc func(tx DatabaseInterface) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		txProvider := &DatabaseProvider{db: tx}
		return fc(txProvider)
	})
}

func (d *DatabaseProvider) Raw(sql string, values ...interface{}) DatabaseInterface {
	return &DatabaseProvider{db: d.db.Raw(sql, values...)}
}

func (d *DatabaseProvider) Exec(sql string, values ...interface{}) error {
	return d.db.Exec(sql, values...).Error
}

func (d *DatabaseProvider) Migrate() error {
	// This would be implemented to run migrations
	// For now, we'll return nil as migrations are handled separately
	return nil
}

// GetDB returns the underlying GORM DB instance
func (d *DatabaseProvider) GetDB() *gorm.DB {
	return d.db
}

// DatabaseProvider interface for database configuration
type DatabaseProviderInterface interface {
	Connect() error
	GetConnection() DatabaseInterface
	Close() error
}

// Global database instance
var DatabaseInstance DatabaseInterface

// Helper functions for models to use (avoiding import cycles)

// DB returns the global database instance
func DB() DatabaseInterface {
	return DatabaseInstance
}

// Create creates a new record
func Create(value interface{}) error {
	return DatabaseInstance.Create(value)
}

// First retrieves the first record
func First(dest interface{}, conds ...interface{}) error {
	return DatabaseInstance.First(dest, conds...)
}

// Find retrieves all records
func Find(dest interface{}, conds ...interface{}) error {
	return DatabaseInstance.Find(dest, conds...)
}

// Save saves a record
func Save(value interface{}) error {
	return DatabaseInstance.Save(value)
}

// Delete deletes a record
func Delete(value interface{}, conds ...interface{}) error {
	return DatabaseInstance.Delete(value, conds...)
}

// Model starts a new query for a model
func Model(value interface{}) DatabaseInterface {
	return DatabaseInstance.Model(value)
}

// Where adds a where clause to the query
func Where(query interface{}, args ...interface{}) DatabaseInterface {
	return DatabaseInstance.Where(query, args...)
}

// Preload preloads associations
func Preload(query string, args ...interface{}) DatabaseInterface {
	return DatabaseInstance.Preload(query, args...)
}
