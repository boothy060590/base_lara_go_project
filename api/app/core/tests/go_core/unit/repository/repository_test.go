package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel represents a test model for repository testing
type TestModel struct {
	ID        uint      `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
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
	return []string{"name", "email"}
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

// setupMockDB creates a mock database connection
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")
	return db, mock
}

// createTestRepository creates a repository for testing
func createTestRepository(db *sql.DB) go_core.Repository[TestModel] {
	config := map[string]any{
		"statement_cache_enabled": true,
		"statement_cache_size":    100,
		"field_validation_enabled": true,
		"sql_validation_enabled":   true,
	}
	
	// Create required dependencies with default configurations
	wsp := go_core.NewWorkStealingPool[any](go_core.DefaultWorkStealingConfig())
	ca := go_core.NewCustomAllocator[any](go_core.DefaultCustomAllocatorConfig())
	pgo := go_core.NewProfileGuidedOptimizer[any](go_core.DefaultProfileGuidedConfig())
	
	return go_core.NewRepositoryWithConfig[TestModel](db, config, wsp, ca, pgo)
}

// TestRepositoryFind tests the Find method
func TestRepositoryFind(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	// Test successful find
	t.Run("SuccessfulFind", func(t *testing.T) {
		expectedModel := &TestModel{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock the query
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(expectedModel.ID, expectedModel.Name, expectedModel.Email, 
				   expectedModel.CreatedAt, expectedModel.UpdatedAt, nil)

		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs(1).
			WillReturnRows(rows)

		// Execute find
		result, err := repo.Find(1)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedModel.ID, result.ID)
		assert.Equal(t, expectedModel.Name, result.Name)
		assert.Equal(t, expectedModel.Email, result.Email)
	})

	// Test record not found
	t.Run("RecordNotFound", func(t *testing.T) {
		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs(999).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}))

		result, err := repo.Find(999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no records found")
	})

	// Verify all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryFindBy tests the FindBy method
func TestRepositoryFindBy(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("SuccessfulFindBy", func(t *testing.T) {
		expectedModel := &TestModel{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(expectedModel.ID, expectedModel.Name, expectedModel.Email,
				   expectedModel.CreatedAt, expectedModel.UpdatedAt, nil)

		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE email = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs("test@example.com").
			WillReturnRows(rows)

		result, err := repo.FindBy("email", "test@example.com")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedModel.Email, result.Email)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryFindAll tests the FindAll method
func TestRepositoryFindAll(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("SuccessfulFindAll", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "User 1", "user1@example.com", time.Now(), time.Now(), nil).
			AddRow(2, "User 2", "user2@example.com", time.Now(), time.Now(), nil)

		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE deleted_at IS NULL").
			ExpectQuery().
			WillReturnRows(rows)

		results, err := repo.FindAll()

		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, "User 1", results[0].Name)
		assert.Equal(t, "User 2", results[1].Name)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryCreate tests the Create method
func TestRepositoryCreate(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("SuccessfulCreate", func(t *testing.T) {
		model := &TestModel{
			Name:  "New User",
			Email: "new@example.com",
		}

		mock.ExpectPrepare("INSERT INTO test_models").
			ExpectExec().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(model)

		require.NoError(t, err)
		assert.Equal(t, uint(1), model.ID)
		assert.False(t, model.CreatedAt.IsZero())
		assert.False(t, model.UpdatedAt.IsZero())
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryUpdate tests the Update method
func TestRepositoryUpdate(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("SuccessfulUpdate", func(t *testing.T) {
		model := &TestModel{
			ID:        1,
			Name:      "Updated User",
			Email:     "updated@example.com",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
		}

		mock.ExpectPrepare("UPDATE test_models SET").
			ExpectExec().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		originalUpdatedAt := model.UpdatedAt
		err := repo.Update(model)

		require.NoError(t, err)
		assert.True(t, model.UpdatedAt.After(originalUpdatedAt))
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryDelete tests the Delete method
func TestRepositoryDelete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("SuccessfulDelete", func(t *testing.T) {
		mock.ExpectPrepare("UPDATE test_models SET deleted_at = \\? WHERE id = \\?").
			ExpectExec().
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(1)

		require.NoError(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryWithContext tests context-aware methods
func TestRepositoryWithContext(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("FindWithContext", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "Test User", "test@example.com", time.Now(), time.Now(), nil)

		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs(1).
			WillReturnRows(rows)

		result, err := repo.FindWithContext(ctx, 1)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Test User", result.Name)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryQuery tests query builder functionality
func TestRepositoryQuery(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("WhereQuery", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "Test User", "test@example.com", time.Now(), time.Now(), nil)

		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE name = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs("Test User").
			WillReturnRows(rows)

		results, err := repo.Where(map[string]any{"name": "Test User"}).Get()

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Test User", results[0].Name)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryBulkOperations tests bulk operations
func TestRepositoryBulkOperations(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("BulkCreate", func(t *testing.T) {
		models := []*TestModel{
			{Name: "User 1", Email: "user1@example.com"},
			{Name: "User 2", Email: "user2@example.com"},
		}

		// Mock bulk insert
		for i := range models {
			mock.ExpectPrepare("INSERT INTO test_models").
				ExpectExec().
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
		}

		err := repo.BulkCreate(models)

		require.NoError(t, err)
		for i, model := range models {
			assert.Equal(t, uint(i+1), model.ID)
		}
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryUtilityMethods tests utility methods
func TestRepositoryUtilityMethods(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("Exists", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

		mock.ExpectPrepare("SELECT COUNT\\(\\*\\) FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs(1).
			WillReturnRows(rows)

		exists, err := repo.Exists(1)

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Count", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(5)

		mock.ExpectPrepare("SELECT COUNT\\(\\*\\) FROM test_models WHERE deleted_at IS NULL").
			ExpectQuery().
			WillReturnRows(rows)

		count, err := repo.Count()

		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryPerformanceStats tests performance monitoring
func TestRepositoryPerformanceStats(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("GetPerformanceStats", func(t *testing.T) {
		stats := repo.GetPerformanceStats()

		require.NotNil(t, stats)
		// Basic stats should exist
		assert.Contains(t, stats, "queries_executed")
		assert.Contains(t, stats, "total_execution_time")
	})

	t.Run("GetOptimizationStats", func(t *testing.T) {
		stats := repo.GetOptimizationStats()

		require.NotNil(t, stats)
		// Optimization stats should exist
		assert.Contains(t, stats, "statement_cache_enabled")
		assert.Contains(t, stats, "field_validation_enabled")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryErrorHandling tests error conditions
func TestRepositoryErrorHandling(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)

		result, err := repo.Find(1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "sql: connection is already closed")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryScanRowToStruct tests the scanRowToStruct implementation
func TestRepositoryScanRowToStruct(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("ComplexTypeScanning", func(t *testing.T) {
		// Test with various SQL types including NULL values
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "Test User", "test@example.com", time.Now(), time.Now(), nil)

		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			ExpectQuery().
			WithArgs(1).
			WillReturnRows(rows)

		result, err := repo.Find(1)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, "Test User", result.Name)
		assert.Equal(t, "test@example.com", result.Email)
		assert.Nil(t, result.DeletedAt) // Should handle NULL properly
	})

	require.NoError(t, mock.ExpectationsWereMet())
}