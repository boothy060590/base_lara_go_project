package unit

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	go_core "base_lara_go_project/app/core/go_core"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel represents a test model for repository testing
type TestModel struct {
	ID        uint       `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
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
	return go_core.NewRepository[TestModel](db)
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

		// Mock the query - FastPathExecutor uses direct queries
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(expectedModel.ID, expectedModel.Name, expectedModel.Email,
				expectedModel.CreatedAt, expectedModel.UpdatedAt, nil)

		// FastPathExecutor uses direct queries, not prepared statements
		mock.ExpectQuery("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
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
		mock.ExpectQuery("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(999).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}))

		result, err := repo.Find(999)

		assert.Error(t, err)
		assert.Nil(t, result)
		// Accept either sql.ErrNoRows or the error string
		assert.True(t, errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows in result set"))
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

		// New system uses dynamic queries for FindBy
		mock.ExpectQuery("SELECT \\* FROM test_models WHERE email = \\? AND deleted_at IS NULL LIMIT 1").
			WithArgs("test@example.com").
			WillReturnRows(rows)

		result, err := repo.FindBy("email", "test@example.com")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedModel.Email, result.Email)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryFindAll tests the FindAll method using Where
func TestRepositoryFindAll(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("SuccessfulFindAll", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "User 1", "user1@example.com", time.Now(), time.Now(), nil).
			AddRow(2, "User 2", "user2@example.com", time.Now(), time.Now(), nil)

		// New system uses balanced path for Where queries
		mock.ExpectPrepare("SELECT \\* FROM test_models WHERE deleted_at IS NULL").
			ExpectQuery().
			WillReturnRows(rows)

		results, err := repo.Where(map[string]any{}).Get()

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

		// New system uses dynamic queries for Create
		mock.ExpectExec("INSERT INTO test_models").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(model)
		require.NoError(t, err)
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
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// New system uses dynamic queries for Update
		mock.ExpectExec("UPDATE test_models SET").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(model)
		require.NoError(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryDelete tests the Delete method
func TestRepositoryDelete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("SuccessfulDelete", func(t *testing.T) {
		// New system uses dynamic queries for Delete (soft delete)
		mock.ExpectExec("UPDATE test_models SET deleted_at = NOW\\(\\) WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(1)
		require.NoError(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryWithContext tests context integration
func TestRepositoryWithContext(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("FindWithContext", func(t *testing.T) {
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

		// Context-aware queries use direct queries like FastPathExecutor
		mock.ExpectQuery("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(1).
			WillReturnRows(rows)

		ctx := context.Background()
		ctxRepo := repo.WithContext(ctx)
		result, err := ctxRepo.Find(1)

		require.NoError(t, err)
		assert.Equal(t, expectedModel.ID, result.ID)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryQuery tests query building
func TestRepositoryQuery(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("WhereQuery", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "Test User", "test@example.com", time.Now(), time.Now(), nil)

		// Where queries use balanced path with dynamic SQL
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

	t.Run("BulkCreate", func(t *testing.T) {
		// Bulk operations are not implemented in the current system
		// This test should be updated when bulk operations are added
		t.Skip("Bulk operations not implemented yet")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryUtilityMethods tests utility methods
func TestRepositoryUtilityMethods(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("Exists", func(t *testing.T) {
		// Exists uses dynamic queries
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM test_models WHERE id = \\? AND deleted_at IS NULL\\)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.Exists(1)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Count", func(t *testing.T) {
		// Count uses dynamic queries
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM test_models WHERE deleted_at IS NULL").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		count, err := repo.Count()
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryPerformanceStats tests performance statistics
func TestRepositoryPerformanceStats(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	t.Run("GetPerformanceStats", func(t *testing.T) {
		// Performance stats are not implemented in the current system
		// This test should be updated when performance stats are added
		t.Skip("Performance stats not implemented yet")
	})

	t.Run("GetOptimizationStats", func(t *testing.T) {
		// Optimization stats are not implemented in the current system
		// This test should be updated when optimization stats are added
		t.Skip("Optimization stats not implemented yet")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryErrorHandling tests error handling
func TestRepositoryErrorHandling(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("DatabaseError", func(t *testing.T) {
		// Mock a database error
		mock.ExpectQuery("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)

		result, err := repo.Find(1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to scan row")
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryScanRowToStruct tests row scanning
func TestRepositoryScanRowToStruct(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := createTestRepository(db)

	t.Run("ComplexTypeScanning", func(t *testing.T) {
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

		mock.ExpectQuery("SELECT \\* FROM test_models WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(1).
			WillReturnRows(rows)

		result, err := repo.Find(1)
		require.NoError(t, err)
		assert.Equal(t, expectedModel.ID, result.ID)
		assert.Equal(t, expectedModel.Name, result.Name)
		assert.Equal(t, expectedModel.Email, result.Email)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
