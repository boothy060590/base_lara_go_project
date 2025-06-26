package core

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
}

// DatabaseProvider interface for database configuration
type DatabaseProvider interface {
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
