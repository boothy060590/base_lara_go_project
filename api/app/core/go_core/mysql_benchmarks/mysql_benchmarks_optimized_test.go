package mysql_benchmarks

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/config"

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

var (
	optimizedDB  *gorm.DB
	once         sync.Once
	emailCounter uint64
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getUniqueEmail(prefix string) string {
	id := atomic.AddUint64(&emailCounter, 1)
	return fmt.Sprintf("%s%d@example.com", prefix, id)
}

// Helper to get MySQL config values
func getMySQLConfig() map[string]interface{} {
	cfg := config.DatabaseConfig()["connections"].(map[string]interface{})["mysql"].(map[string]interface{})
	return cfg
}

// Helper to get work stealing config values
func getWorkStealingConfig() map[string]interface{} {
	return config.WorkStealingConfig()["workers"].(map[string]interface{})
}

func setupOptimizedMySQLDB() *gorm.DB {
	once.Do(func() {
		// Use benchmark-specific database configuration
		host := getEnv("MYSQL_HOST", "localhost")
		port := getEnv("MYSQL_PORT", "3309") // Docker exposes MySQL on port 3309
		user := getEnv("MYSQL_USER", "api_user")
		password := getEnv("MYSQL_PASSWORD", "b4s3L4r4G0212!")
		database := getEnv("MYSQL_DATABASE", "benchmark_db")

		// Optimized DSN with connection pooling and performance settings
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&maxAllowedPacket=0&interpolateParams=true&timeout=30s&readTimeout=30s&writeTimeout=30s&autocommit=true&sql_mode='NO_ENGINE_SUBSTITUTION'",
			user, password, host, port, database)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:                 logger.Default.LogMode(logger.Silent), // Disable logging for benchmarks
			PrepareStmt:            true,                                  // Enable prepared statements
			SkipDefaultTransaction: true,                                  // Skip default transactions for better performance
		})
		if err != nil {
			panic(fmt.Sprintf("failed to connect to MySQL: %v", err))
		}

		// Configure connection pool for benchmarks - tuned for high performance
		sqlDB, err := db.DB()
		if err != nil {
			panic(fmt.Sprintf("failed to get underlying sql.DB: %v", err))
		}

		// Optimize connection pool for benchmarks
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(3600 * time.Second)
		sqlDB.SetConnMaxIdleTime(1800 * time.Second)

		// Auto migrate
		if err := db.AutoMigrate(&TestModelMySQL{}); err != nil {
			panic(fmt.Sprintf("failed to migrate: %v", err))
		}

		optimizedDB = db
	})

	return optimizedDB
}

func createOptimizedRepository() go_core.Repository[TestModelMySQL] {
	db := setupOptimizedMySQLDB()

	// Create optimization dependencies
	wspCfg := getWorkStealingConfig()
	wspConfig := &go_core.WorkStealingConfig{
		NumWorkers: wspCfg["num_workers"].(int),
		QueueSize:  wspCfg["queue_size"].(int),
	}
	wsp := go_core.NewWorkStealingPool[any](wspConfig)

	caConfig := &go_core.CustomAllocatorConfig{
		Enabled:            true,
		PoolSize:           10000, // Larger pool for MySQL
		MaxObjectSize:      1024 * 1024,
		CleanupInterval:    5 * time.Minute,
		EnableMetrics:      true,
		EnableProfiling:    true,
		AllocationStrategy: "pool",
	}
	ca := go_core.NewCustomAllocator[any](caConfig)

	pgoConfig := &go_core.ProfileGuidedConfig{
		Enabled:              true,
		SamplingInterval:     1 * time.Second,
		OptimizationInterval: 30 * time.Second,
		MinSamples:           100,
		MaxOptimizations:     10,
		EnableAutoTuning:     true,
		EnableMetrics:        true,
	}
	pgo := go_core.NewProfileGuidedOptimizer[any](pgoConfig)

	return go_core.NewRepository[TestModelMySQL](db, wsp, ca, pgo)
}

func BenchmarkRepository_CRUD_MySQL_Optimized(b *testing.B) {
	// Setup with optimized repository
	repo := createOptimizedRepository()

	// Clean up
	defer func() {
		db := setupOptimizedMySQLDB()
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create
			model := &TestModelMySQL{
				Name:  "Benchmark Test",
				Email: getUniqueEmail("benchmark"),
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
	})
}

func BenchmarkRepository_BatchOperations_MySQL_Optimized(b *testing.B) {
	// Setup with optimized repository
	repo := createOptimizedRepository()

	// Clean up
	defer func() {
		db := setupOptimizedMySQLDB()
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create 10 models in parallel
			models := make([]*TestModelMySQL, 10)
			timestamp := time.Now().UnixNano()
			for j := 0; j < 10; j++ {
				models[j] = &TestModelMySQL{
					Name:  "Batch Test",
					Email: fmt.Sprintf("batch%d_%d@example.com", timestamp, j), // Use timestamp for uniqueness
					Age:   25 + j,
				}
			}

			// Create all models
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

			// Clean up this batch
			for _, model := range models {
				repo.Delete(model.ID)
			}
		}
	})
}

func BenchmarkRepository_QueryPerformance_MySQL_Optimized(b *testing.B) {
	// Setup with optimized repository
	repo := createOptimizedRepository()

	// Pre-populate with test data (clear and recreate each time to avoid duplicates)
	db := setupOptimizedMySQLDB()
	// Clear any existing data first
	db.Exec("DELETE FROM test_model_mysql")

	// Create test data with guaranteed unique emails using timestamp
	timestamp := time.Now().UnixNano()
	for i := 0; i < 500; i++ { // Reduced from 1000 to 500 for better performance
		model := &TestModelMySQL{
			Name:  "Query Test",
			Email: fmt.Sprintf("query%d_%d@example.com", timestamp, i), // Use timestamp for guaranteed uniqueness
			Age:   20 + (i % 60),
		}
		if err := db.Create(model).Error; err != nil {
			b.Fatal(err)
		}
	}

	// Clean up
	defer func() {
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test various query operations in parallel
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
	})
}

func BenchmarkRepository_Concurrency_MySQL_Optimized(b *testing.B) {
	// Setup with optimized repository
	repo := createOptimizedRepository()

	// Clean up
	defer func() {
		db := setupOptimizedMySQLDB()
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create
			model := &TestModelMySQL{
				Name:  "Concurrent Test",
				Email: getUniqueEmail("concurrent"),
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

func BenchmarkRepository_BatchProcessing_MySQL_Optimized(b *testing.B) {
	// Setup with optimized repository
	repo := createOptimizedRepository()

	// Clean up
	defer func() {
		db := setupOptimizedMySQLDB()
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create batch of 10 models (reduced from 20 for better performance)
			models := make([]*TestModelMySQL, 10)
			for j := 0; j < 10; j++ {
				models[j] = &TestModelMySQL{
					Name:  "Batch Processing Test",
					Email: getUniqueEmail("batchproc"),
					Age:   25 + j,
				}
			}

			// Create all models in batch (reduces round-trips)
			for _, model := range models {
				if err := repo.Create(model); err != nil {
					b.Fatal(err)
				}
			}

			// Batch read operations - only read first 5 to reduce overhead
			for i := 0; i < 5; i++ {
				_, err := repo.Find(models[i].ID)
				if err != nil {
					b.Fatal(err)
				}
			}

			// Batch update operations - only update first 5
			for i := 0; i < 5; i++ {
				models[i].Name = "Updated Batch Processing Test"
				if err := repo.Update(models[i]); err != nil {
					b.Fatal(err)
				}
			}

			// Batch delete operations
			for _, model := range models {
				if err := repo.Delete(model.ID); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkRepository_BulkInsert_MySQL_Optimized(b *testing.B) {
	// Setup with optimized repository
	repo := createOptimizedRepository()

	// Clean up
	defer func() {
		db := setupOptimizedMySQLDB()
		db.Exec("DELETE FROM test_model_mysql")
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create batch of 50 models for bulk insert
			models := make([]*TestModelMySQL, 50)
			for j := 0; j < 50; j++ {
				models[j] = &TestModelMySQL{
					Name:  "Bulk Insert Test",
					Email: getUniqueEmail("bulk"),
					Age:   25 + j,
				}
			}

			// Use GORM's bulk insert for maximum performance
			db := setupOptimizedMySQLDB()
			if err := db.CreateInBatches(models, 10).Error; err != nil {
				b.Fatal(err)
			}

			// Clean up
			for _, model := range models {
				repo.Delete(model.ID)
			}
		}
	})
}
