//go:build mysql
// +build mysql

package go_core

import (
	"testing"
)

func TestMySQLConnection(t *testing.T) {
	// Test that we can connect to MySQL
	db := setupMySQLBenchmarkDB()
	if db == nil {
		t.Fatal("Failed to connect to MySQL")
	}

	// Test that we can create a table
	if err := db.AutoMigrate(&TestModelMySQL{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Test that we can create a record
	model := &TestModelMySQL{
		Name:  "Test",
		Email: "test@example.com",
		Age:   25,
	}

	if err := db.Create(model).Error; err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Clean up
	db.Exec("DELETE FROM test_model_mysql")

	t.Log("MySQL connection test passed")
}
