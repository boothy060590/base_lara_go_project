package go_core

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// StatementCache manages prepared statements with caching and performance tracking
type StatementCache struct {
	enabled    bool
	maxSize    int
	ttl        time.Duration
	cache      map[string]*CachedStatement
	cacheMutex sync.RWMutex
	stats      *StatementCacheStats
	statsMutex sync.RWMutex
	config     map[string]any
}

// CachedStatement represents a cached prepared statement
type CachedStatement struct {
	Statement *sql.Stmt
	Created   time.Time
	LastUsed  time.Time
	UseCount  int64
}

// StatementCacheStats tracks statement cache statistics
type StatementCacheStats struct {
	CacheSize     int
	HitCount      int64
	MissCount     int64
	EvictionCount int64
	LastUpdated   time.Time
}

// NewStatementCache creates a new statement cache with configuration
func NewStatementCache(config map[string]any) *StatementCache {
	cache := &StatementCache{
		enabled: true,
		maxSize: 100, // Default max cache size
		ttl:     30 * time.Minute,
		cache:   make(map[string]*CachedStatement),
		stats:   &StatementCacheStats{},
		config:  config,
	}

	// Apply configuration
	if maxSize, ok := config["statement_cache_size"].(int); ok {
		cache.maxSize = maxSize
	}
	if ttl, ok := config["statement_cache_ttl"].(int); ok {
		cache.ttl = time.Duration(ttl) * time.Second
	}

	// Start cleanup routine
	go cache.cleanupRoutine()

	return cache
}

// GetStatement gets a prepared statement from cache or creates a new one
func (sc *StatementCache) GetStatement(ctx context.Context, query string) (*sql.Stmt, error) {
	if !sc.enabled {
		// Return nil to indicate no caching
		return nil, nil
	}

	// Check cache first
	sc.cacheMutex.RLock()
	if cached, exists := sc.cache[query]; exists {
		// Update usage statistics
		cached.LastUsed = time.Now()
		cached.UseCount++
		sc.cacheMutex.RUnlock()

		// Update stats
		sc.updateStats(true, false)
		return cached.Statement, nil
	}
	sc.cacheMutex.RUnlock()

	// Cache miss - create new statement
	sc.updateStats(false, true)
	return nil, nil // Return nil to indicate no cached statement
}

// CacheStatement caches a prepared statement
func (sc *StatementCache) CacheStatement(query string, stmt *sql.Stmt) error {
	if !sc.enabled {
		return nil
	}

	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()

	// Check if cache is full
	if len(sc.cache) >= sc.maxSize {
		sc.evictOldest()
	}

	// Add to cache
	sc.cache[query] = &CachedStatement{
		Statement: stmt,
		Created:   time.Now(),
		LastUsed:  time.Now(),
		UseCount:  1,
	}

	return nil
}

// evictOldest removes the oldest statement from cache
func (sc *StatementCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range sc.cache {
		if oldestKey == "" || cached.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.LastUsed
		}
	}

	if oldestKey != "" {
		// Close the statement
		if sc.cache[oldestKey].Statement != nil {
			sc.cache[oldestKey].Statement.Close()
		}
		delete(sc.cache, oldestKey)
		sc.updateEvictionStats()
	}
}

// cleanupRoutine periodically cleans up expired statements
func (sc *StatementCache) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sc.cleanup()
	}
}

// cleanup removes expired statements from cache
func (sc *StatementCache) cleanup() {
	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, cached := range sc.cache {
		if now.Sub(cached.LastUsed) > sc.ttl {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		// Close the statement
		if sc.cache[key].Statement != nil {
			sc.cache[key].Statement.Close()
		}
		delete(sc.cache, key)
		sc.updateEvictionStats()
	}
}

// updateStats updates cache statistics
func (sc *StatementCache) updateStats(hit, miss bool) {
	sc.statsMutex.Lock()
	defer sc.statsMutex.Unlock()

	if hit {
		sc.stats.HitCount++
	}
	if miss {
		sc.stats.MissCount++
	}
	sc.stats.CacheSize = len(sc.cache)
	sc.stats.LastUpdated = time.Now()
}

// updateEvictionStats updates eviction statistics
func (sc *StatementCache) updateEvictionStats() {
	sc.statsMutex.Lock()
	defer sc.statsMutex.Unlock()

	sc.stats.EvictionCount++
	sc.stats.CacheSize = len(sc.cache)
	sc.stats.LastUpdated = time.Now()
}

// GetStats returns statement cache statistics
func (sc *StatementCache) GetStats() map[string]any {
	sc.statsMutex.RLock()
	defer sc.statsMutex.RUnlock()

	hitRate := 0.0
	totalRequests := sc.stats.HitCount + sc.stats.MissCount
	if totalRequests > 0 {
		hitRate = float64(sc.stats.HitCount) / float64(totalRequests) * 100
	}

	return map[string]any{
		"enabled":        sc.enabled,
		"max_size":       sc.maxSize,
		"ttl":            sc.ttl.String(),
		"cache_size":     sc.stats.CacheSize,
		"hit_count":      sc.stats.HitCount,
		"miss_count":     sc.stats.MissCount,
		"hit_rate":       fmt.Sprintf("%.2f%%", hitRate),
		"eviction_count": sc.stats.EvictionCount,
		"last_updated":   sc.stats.LastUpdated,
	}
}

// IsEnabled returns whether statement caching is enabled
func (sc *StatementCache) IsEnabled() bool {
	return sc.enabled
}

// Clear clears all cached statements
func (sc *StatementCache) Clear() {
	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()

	for _, cached := range sc.cache {
		if cached.Statement != nil {
			cached.Statement.Close()
		}
	}

	sc.cache = make(map[string]*CachedStatement)
	sc.updateStats(false, false)
}

// Close closes all cached statements
func (sc *StatementCache) Close() error {
	sc.cacheMutex.Lock()
	defer sc.cacheMutex.Unlock()

	for _, cached := range sc.cache {
		if cached.Statement != nil {
			if err := cached.Statement.Close(); err != nil {
				return fmt.Errorf("failed to close statement: %w", err)
			}
		}
	}

	sc.cache = make(map[string]*CachedStatement)
	return nil
}
