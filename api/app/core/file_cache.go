package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// fileCacheItem represents a cached item stored in a file
type fileCacheItem struct {
	Value      interface{} `json:"value"`
	Expiration time.Time   `json:"expiration"`
}

// FileCacheDriver implements file-based caching
type FileCacheDriver struct {
	*BaseCacheProvider
	path string
}

// NewFileCacheDriver creates a new file cache driver
func NewFileCacheDriver(path, prefix string, ttl time.Duration) *FileCacheDriver {
	return &FileCacheDriver{
		BaseCacheProvider: NewBaseCacheProvider(prefix, ttl),
		path:              path,
	}
}

// Get retrieves a value from file cache
func (d *FileCacheDriver) Get(key string) (interface{}, bool) {
	fullKey := d.GetFullKey(key)
	filePath := d.getFilePath(fullKey)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, false
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false
	}

	// Unmarshal cache item
	var item fileCacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, false
	}

	// Check expiration
	if time.Now().After(item.Expiration) {
		// Clean up expired file
		os.Remove(filePath)
		return nil, false
	}

	return item.Value, true
}

// Set stores a value in file cache
func (d *FileCacheDriver) Set(key string, value interface{}, ttl ...time.Duration) error {
	fullKey := d.GetFullKey(key)
	filePath := d.getFilePath(fullKey)
	duration := d.GetEffectiveTTL(ttl...)

	// Create cache item
	item := fileCacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration),
	}

	// Marshal to JSON
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filePath, data, 0644)
}

// Delete removes a value from file cache
func (d *FileCacheDriver) Delete(key string) error {
	fullKey := d.GetFullKey(key)
	filePath := d.getFilePath(fullKey)

	return os.Remove(filePath)
}

// Has checks if a key exists in file cache
func (d *FileCacheDriver) Has(key string) bool {
	_, exists := d.Get(key)
	return exists
}

// Flush clears all file cache
func (d *FileCacheDriver) Flush() error {
	return os.RemoveAll(d.path)
}

// getFilePath returns the full file path for a cache key
func (d *FileCacheDriver) getFilePath(key string) string {
	// Create a hash or use the key directly for the filename
	// For simplicity, we'll use the key as filename (sanitized)
	safeKey := filepath.Base(key) // Remove any path separators
	return filepath.Join(d.path, safeKey+".cache")
}

// GetStats returns cache statistics
func (d *FileCacheDriver) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"path": d.path,
	}

	// Count files in cache directory
	if entries, err := os.ReadDir(d.path); err == nil {
		stats["total_files"] = len(entries)
	} else {
		stats["total_files"] = 0
	}

	return stats
}
