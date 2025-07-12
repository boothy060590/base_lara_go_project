//go:build mysql
// +build mysql

package go_core

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestModelMySQL is a MySQL-compatible version of TestModel
type TestModelMySQL struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"not null"`
	Email string `gorm:"uniqueIndex:idx_email,length:191"` // MySQL requires length for TEXT indexes
	Age   int
}

func setupMySQLBenchmarkDB() *gorm.DB {
	// Get MySQL connection details from environment or use defaults
	host := getEnv("MYSQL_HOST", "localhost")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "api_user")
	password := getEnv("MYSQL_PASSWORD", "b4s3L4r4G0212!")
	database := getEnv("MYSQL_DATABASE", "benchmark_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable logging for benchmarks
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to MySQL: %v", err))
	}

	// Auto migrate
	if err := db.AutoMigrate(&TestModelMySQL{}); err != nil {
		panic(fmt.Sprintf("failed to migrate: %v", err))
	}

	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func BenchmarkRepository_CRUD_MySQL(b *testing.B) {
	// Setup
	db := setupMySQLBenchmarkDB()
	repo := NewRepository[TestModelMySQL](db, nil, nil, nil)

	// Clean up
	defer func() {
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create
		model := &TestModelMySQL{
			Name:  "Benchmark Test",
			Email: fmt.Sprintf("benchmark%d@example.com", i),
			Age:   25,
		}
		if err := repo.Create(model); err != nil {
			b.Fatal(err)
		}

		// Read
		found, err := repo.Find(model.ID)
		if err != nil {
			b.Fatal(err)
		}
		if found == nil {
			b.Fatal("model not found")
		}

		// Update
		found.Name = "Updated Benchmark Test"
		if err := repo.Update(found); err != nil {
			b.Fatal(err)
		}

		// Delete
		if err := repo.Delete(model.ID); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_BatchOperations_MySQL(b *testing.B) {
	// Setup
	db := setupMySQLBenchmarkDB()
	repo := NewRepository[TestModelMySQL](db, nil, nil, nil)

	// Clean up
	defer func() {
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create 10 models
		models := make([]*TestModelMySQL, 10)
		for j := 0; j < 10; j++ {
			models[j] = &TestModelMySQL{
				Name:  "Batch Test",
				Email: fmt.Sprintf("batch%d_%d@example.com", i, j),
				Age:   25 + j,
			}
		}

		// Create all
		for _, model := range models {
			if err := repo.Create(model); err != nil {
				b.Fatal(err)
			}
		}

		// Find all
		results, err := repo.FindAll()
		if err != nil {
			b.Fatal(err)
		}
		if len(results) == 0 {
			b.Fatal("no results found")
		}

		// Query with conditions
		results, err = repo.Where(map[string]any{"age": 25}).Get()
		if err != nil {
			b.Fatal(err)
		}

		// Clean up
		db.Exec("DELETE FROM test_model_mysql")
	}
}

func BenchmarkRepository_QueryPerformance_MySQL(b *testing.B) {
	// Setup
	db := setupMySQLBenchmarkDB()
	repo := NewRepository[TestModelMySQL](db, nil, nil, nil)

	// Create test data
	for i := 0; i < 100; i++ {
		model := &TestModelMySQL{
			Name:  "Query Test",
			Email: fmt.Sprintf("query%d@example.com", i),
			Age:   20 + (i % 60),
		}
		if err := repo.Create(model); err != nil {
			b.Fatal(err)
		}
	}

	// Clean up
	defer func() {
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Test various query operations
		_, err := repo.FindAll()
		if err != nil {
			b.Fatal(err)
		}

		_, err = repo.Where(map[string]any{"age": 25}).Get()
		if err != nil {
			b.Fatal(err)
		}

		_, err = repo.Count()
		if err != nil {
			b.Fatal(err)
		}

		_, _, err = repo.Where(map[string]any{}).Paginate(1, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_Concurrency_MySQL(b *testing.B) {
	// Setup
	db := setupMySQLBenchmarkDB()
	repo := NewRepository[TestModelMySQL](db, nil, nil, nil)

	// Clean up
	defer func() {
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			counter++
			// Create
			model := &TestModelMySQL{
				Name:  "Concurrent Test",
				Email: fmt.Sprintf("concurrent%d@example.com", counter),
				Age:   25,
			}
			if err := repo.Create(model); err != nil {
				b.Fatal(err)
			}

			// Read
			_, err := repo.Find(model.ID)
			if err != nil {
				b.Fatal(err)
			}

			// Update
			model.Name = "Updated Concurrent Test"
			if err := repo.Update(model); err != nil {
				b.Fatal(err)
			}

			// Delete
			if err := repo.Delete(model.ID); err != nil {
				b.Fatal(err)
			}
		}
	})
}
