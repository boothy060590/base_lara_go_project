package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestModel represents a test model for integration testing
type TestModel struct {
	ID        uint      `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Age       int       `json:"age" db:"age"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// GetTableName returns the table name for TestModel
func (tm *TestModel) GetTableName() string {
	return "test_models"
}

// GetPrimaryKey returns the primary key field name
func (tm *TestModel) GetPrimaryKey() string {
	return "id"
}

// GetFillableFields returns fillable fields
func (tm *TestModel) GetFillableFields() []string {
	return []string{"name", "email", "age", "is_active"}
}

// BeforeSave is called before saving
func (tm *TestModel) BeforeSave() error {
	if tm.CreatedAt.IsZero() {
		tm.CreatedAt = time.Now()
	}
	tm.UpdatedAt = time.Now()
	return nil
}

// AfterSave is called after saving
func (tm *TestModel) AfterSave() error {
	return nil
}

// BeforeDelete is called before deleting
func (tm *TestModel) BeforeDelete() error {
	return nil
}

// AfterDelete is called after deleting
func (tm *TestModel) AfterDelete() error {
	return nil
}

// RepositoryIntegrationTestSuite defines the test suite
type RepositoryIntegrationTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo go_core.Repository[TestModel]
}

// SetupSuite runs once before all tests
func (suite *RepositoryIntegrationTestSuite) SetupSuite() {
	// Skip integration tests if no database is available
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		suite.T().Skip("Skipping integration tests: database not available")
		return
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		suite.T().Skip("Skipping integration tests: cannot connect to database")
		return
	}

	suite.db = db

	// Create repository with test configuration
	config := map[string]any{
		"statement_cache_enabled":    true,
		"statement_cache_size":       100,
		"statement_cache_ttl":        300,
		"field_validation_enabled":   true,
		"sql_validation_enabled":     true,
		"performance_tracking":       true,
		"connection_pool_enabled":    true,
		"connection_pool_max_open":   25,
		"connection_pool_max_idle":   10,
		"connection_pool_max_lifetime": 300,
	}

	// Create required dependencies with default configurations
	wsp := go_core.NewWorkStealingPool[any](go_core.DefaultWorkStealingConfig())
	ca := go_core.NewCustomAllocator[any](go_core.DefaultCustomAllocatorConfig())
	pgo := go_core.NewProfileGuidedOptimizer[any](go_core.DefaultProfileGuidedConfig())

	suite.repo = go_core.NewRepositoryWithConfig[TestModel](db, config, wsp, ca, pgo)

	// Create test table
	suite.createTestTable()
}

// TearDownSuite runs once after all tests
func (suite *RepositoryIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.dropTestTable()
		suite.db.Close()
	}
}

// SetupTest runs before each test
func (suite *RepositoryIntegrationTestSuite) SetupTest() {
	// Clean test table
	suite.cleanTestTable()
}

// createTestTable creates the test table
func (suite *RepositoryIntegrationTestSuite) createTestTable() {
	query := `
		CREATE TABLE IF NOT EXISTS test_models (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			age INT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL DEFAULT NULL,
			INDEX idx_email (email),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`
	_, err := suite.db.Exec(query)
	require.NoError(suite.T(), err, "Failed to create test table")
}

// dropTestTable drops the test table
func (suite *RepositoryIntegrationTestSuite) dropTestTable() {
	_, err := suite.db.Exec("DROP TABLE IF EXISTS test_models")
	require.NoError(suite.T(), err, "Failed to drop test table")
}

// cleanTestTable cleans the test table
func (suite *RepositoryIntegrationTestSuite) cleanTestTable() {
	_, err := suite.db.Exec("DELETE FROM test_models")
	require.NoError(suite.T(), err, "Failed to clean test table")
}

// TestBasicCRUDOperations tests basic CRUD operations
func (suite *RepositoryIntegrationTestSuite) TestBasicCRUDOperations() {
	t := suite.T()

	// Test Create
	model := &TestModel{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		IsActive: true,
	}

	err := suite.repo.Create(model)
	require.NoError(t, err, "Create should succeed")
	require.NotZero(t, model.ID, "ID should be set after create")
	require.False(t, model.CreatedAt.IsZero(), "CreatedAt should be set")
	require.False(t, model.UpdatedAt.IsZero(), "UpdatedAt should be set")

	// Test Find
	found, err := suite.repo.Find(model.ID)
	require.NoError(t, err, "Find should succeed")
	require.NotNil(t, found, "Found model should not be nil")
	assert.Equal(t, model.Name, found.Name)
	assert.Equal(t, model.Email, found.Email)
	assert.Equal(t, model.Age, found.Age)
	assert.Equal(t, model.IsActive, found.IsActive)

	// Test Update
	found.Name = "Jane Doe"
	found.Age = 25
	err = suite.repo.Update(found)
	require.NoError(t, err, "Update should succeed")

	// Verify update
	updated, err := suite.repo.Find(model.ID)
	require.NoError(t, err, "Find after update should succeed")
	assert.Equal(t, "Jane Doe", updated.Name)
	assert.Equal(t, 25, updated.Age)
	assert.True(t, updated.UpdatedAt.After(found.CreatedAt))

	// Test Delete (soft delete)
	err = suite.repo.Delete(model.ID)
	require.NoError(t, err, "Delete should succeed")

	// Verify soft delete
	deleted, err := suite.repo.Find(model.ID)
	assert.Error(t, err, "Should not find deleted record")
	assert.Nil(t, deleted)
}

// TestFindByOperations tests FindBy operations
func (suite *RepositoryIntegrationTestSuite) TestFindByOperations() {
	t := suite.T()

	// Create test data
	model := &TestModel{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		IsActive: true,
	}

	err := suite.repo.Create(model)
	require.NoError(t, err)

	// Test FindBy with email
	found, err := suite.repo.FindBy("email", "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, model.Email, found.Email)

	// Test FindBy with non-existent value
	notFound, err := suite.repo.FindBy("email", "nonexistent@example.com")
	assert.Error(t, err)
	assert.Nil(t, notFound)
}

// TestFindAllOperations tests FindAll operations
func (suite *RepositoryIntegrationTestSuite) TestFindAllOperations() {
	t := suite.T()

	// Create multiple test records
	models := []*TestModel{
		{Name: "User 1", Email: "user1@example.com", Age: 20, IsActive: true},
		{Name: "User 2", Email: "user2@example.com", Age: 25, IsActive: false},
		{Name: "User 3", Email: "user3@example.com", Age: 30, IsActive: true},
	}

	for _, model := range models {
		err := suite.repo.Create(model)
		require.NoError(t, err)
	}

	// Test FindAll
	all, err := suite.repo.FindAll()
	require.NoError(t, err)
	require.Len(t, all, 3)

	// Verify all records are present
	names := make([]string, len(all))
	for i, model := range all {
		names[i] = model.Name
	}
	assert.Contains(t, names, "User 1")
	assert.Contains(t, names, "User 2")
	assert.Contains(t, names, "User 3")
}

// TestQueryBuilderOperations tests query builder functionality
func (suite *RepositoryIntegrationTestSuite) TestQueryBuilderOperations() {
	t := suite.T()

	// Create test data
	models := []*TestModel{
		{Name: "Active User 1", Email: "active1@example.com", Age: 20, IsActive: true},
		{Name: "Active User 2", Email: "active2@example.com", Age: 25, IsActive: true},
		{Name: "Inactive User", Email: "inactive@example.com", Age: 30, IsActive: false},
	}

	for _, model := range models {
		err := suite.repo.Create(model)
		require.NoError(t, err)
	}

	// Test Where query
	activeUsers, err := suite.repo.Where(map[string]any{"is_active": true}).Get()
	require.NoError(t, err)
	require.Len(t, activeUsers, 2)

	// Test First query
	firstActive, err := suite.repo.Where(map[string]any{"is_active": true}).First()
	require.NoError(t, err)
	require.NotNil(t, firstActive)
	assert.True(t, firstActive.IsActive)

	// Test Paginate
	page1, total, err := suite.repo.Where(map[string]any{"is_active": true}).Paginate(1, 1)
	require.NoError(t, err)
	require.Len(t, page1, 1)
	assert.Equal(t, int64(2), total)
}

// TestBulkOperations tests bulk operations
func (suite *RepositoryIntegrationTestSuite) TestBulkOperations() {
	t := suite.T()

	// Test BulkCreate
	models := []*TestModel{
		{Name: "Bulk User 1", Email: "bulk1@example.com", Age: 20, IsActive: true},
		{Name: "Bulk User 2", Email: "bulk2@example.com", Age: 25, IsActive: true},
		{Name: "Bulk User 3", Email: "bulk3@example.com", Age: 30, IsActive: false},
	}

	err := suite.repo.BulkCreate(models)
	require.NoError(t, err)

	// Verify all models have IDs
	for _, model := range models {
		assert.NotZero(t, model.ID)
	}

	// Test BulkUpdate
	for _, model := range models {
		model.Age += 5
	}

	err = suite.repo.BulkUpdate(models)
	require.NoError(t, err)

	// Verify updates
	for _, model := range models {
		found, err := suite.repo.Find(model.ID)
		require.NoError(t, err)
		assert.Equal(t, model.Age, found.Age)
	}

	// Test BulkDelete
	ids := make([]uint, len(models))
	for i, model := range models {
		ids[i] = model.ID
	}

	err = suite.repo.BulkDelete(ids)
	require.NoError(t, err)

	// Verify deletion
	for _, id := range ids {
		found, err := suite.repo.Find(id)
		assert.Error(t, err)
		assert.Nil(t, found)
	}
}

// TestUtilityOperations tests utility operations
func (suite *RepositoryIntegrationTestSuite) TestUtilityOperations() {
	t := suite.T()

	// Create test data
	model := &TestModel{
		Name:     "Utility Test",
		Email:    "utility@example.com",
		Age:      25,
		IsActive: true,
	}

	err := suite.repo.Create(model)
	require.NoError(t, err)

	// Test Exists
	exists, err := suite.repo.Exists(model.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Exists with non-existent ID
	notExists, err := suite.repo.Exists(999)
	require.NoError(t, err)
	assert.False(t, notExists)

	// Test Count
	count, err := suite.repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Test CountWhere
	activeCount, err := suite.repo.CountWhere(map[string]any{"is_active": true})
	require.NoError(t, err)
	assert.Equal(t, int64(1), activeCount)

	inactiveCount, err := suite.repo.CountWhere(map[string]any{"is_active": false})
	require.NoError(t, err)
	assert.Equal(t, int64(0), inactiveCount)
}

// TestContextOperations tests context-aware operations
func (suite *RepositoryIntegrationTestSuite) TestContextOperations() {
	t := suite.T()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test context-aware operations
	model := &TestModel{
		Name:     "Context Test",
		Email:    "context@example.com",
		Age:      25,
		IsActive: true,
	}

	// Test CreateWithContext
	err := suite.repo.CreateWithContext(ctx, model)
	require.NoError(t, err)

	// Test FindWithContext
	found, err := suite.repo.FindWithContext(ctx, model.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, model.Email, found.Email)

	// Test UpdateWithContext
	found.Name = "Updated Context Test"
	err = suite.repo.UpdateWithContext(ctx, found)
	require.NoError(t, err)

	// Test DeleteWithContext
	err = suite.repo.DeleteWithContext(ctx, model.ID)
	require.NoError(t, err)
}

// TestTransactionOperations tests transaction operations
func (suite *RepositoryIntegrationTestSuite) TestTransactionOperations() {
	t := suite.T()

	// Test successful transaction
	err := suite.repo.Transaction(func(txRepo go_core.Repository[TestModel]) error {
		model1 := &TestModel{
			Name:     "TX User 1",
			Email:    "tx1@example.com",
			Age:      25,
			IsActive: true,
		}

		model2 := &TestModel{
			Name:     "TX User 2",
			Email:    "tx2@example.com",
			Age:      30,
			IsActive: true,
		}

		if err := txRepo.Create(model1); err != nil {
			return err
		}

		if err := txRepo.Create(model2); err != nil {
			return err
		}

		return nil
	})

	require.NoError(t, err)

	// Verify both records were created
	count, err := suite.repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Test transaction rollback
	err = suite.repo.Transaction(func(txRepo go_core.Repository[TestModel]) error {
		model3 := &TestModel{
			Name:     "TX User 3",
			Email:    "tx3@example.com",
			Age:      35,
			IsActive: true,
		}

		if err := txRepo.Create(model3); err != nil {
			return err
		}

		// Force rollback
		return fmt.Errorf("forced rollback")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "forced rollback")

	// Verify rollback - count should still be 2
	count, err = suite.repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// TestPerformanceMonitoring tests performance monitoring features
func (suite *RepositoryIntegrationTestSuite) TestPerformanceMonitoring() {
	t := suite.T()

	// Create some test data to generate performance stats
	model := &TestModel{
		Name:     "Performance Test",
		Email:    "perf@example.com",
		Age:      25,
		IsActive: true,
	}

	err := suite.repo.Create(model)
	require.NoError(t, err)

	_, err = suite.repo.Find(model.ID)
	require.NoError(t, err)

	// Test GetPerformanceStats
	perfStats := suite.repo.GetPerformanceStats()
	require.NotNil(t, perfStats)
	assert.Contains(t, perfStats, "queries_executed")
	assert.Contains(t, perfStats, "total_execution_time")

	// Test GetOptimizationStats
	optStats := suite.repo.GetOptimizationStats()
	require.NotNil(t, optStats)
	assert.Contains(t, optStats, "statement_cache_enabled")
	assert.Contains(t, optStats, "field_validation_enabled")
}

// TestRepositoryIntegration runs the integration test suite
func TestRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(RepositoryIntegrationTestSuite))
}