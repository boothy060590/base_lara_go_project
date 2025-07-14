package go_core

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// ConnectionPool manages database connections with pooling and performance tracking
type ConnectionPool struct {
	enabled     bool
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	config      map[string]interface{}
	stats       *ConnectionPoolStats
	statsMutex  sync.RWMutex
}

// ConnectionPoolStats tracks connection pool statistics
type ConnectionPoolStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
	LastUpdated        time.Time
}

// NewConnectionPool creates a new connection pool with configuration
func NewConnectionPool(config map[string]interface{}) *ConnectionPool {
	pool := &ConnectionPool{
		enabled:     true,
		maxOpen:     25, // Default max open connections
		maxIdle:     5,  // Default max idle connections
		maxLifetime: 5 * time.Minute,
		config:      config,
		stats:       &ConnectionPoolStats{},
	}

	// Apply configuration
	if maxOpen, ok := config["max_open_connections"].(int); ok {
		pool.maxOpen = maxOpen
	}
	if maxIdle, ok := config["max_idle_connections"].(int); ok {
		pool.maxIdle = maxIdle
	}
	if maxLifetime, ok := config["max_lifetime"].(int); ok {
		pool.maxLifetime = time.Duration(maxLifetime) * time.Second
	}

	return pool
}

// ConfigurePool configures the database connection pool
func (cp *ConnectionPool) ConfigurePool(db *sql.DB) error {
	if !cp.enabled {
		return nil
	}

	db.SetMaxOpenConns(cp.maxOpen)
	db.SetMaxIdleConns(cp.maxIdle)
	db.SetConnMaxLifetime(cp.maxLifetime)

	// Start stats collection
	go cp.collectStats(db)

	return nil
}

// collectStats periodically collects connection pool statistics
func (cp *ConnectionPool) collectStats(db *sql.DB) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cp.updateStats(db)
	}
}

// updateStats updates connection pool statistics
func (cp *ConnectionPool) updateStats(db *sql.DB) {
	cp.statsMutex.Lock()
	defer cp.statsMutex.Unlock()

	cp.stats.MaxOpenConnections = db.Stats().MaxOpenConnections
	cp.stats.OpenConnections = db.Stats().OpenConnections
	cp.stats.InUse = db.Stats().InUse
	cp.stats.Idle = db.Stats().Idle
	cp.stats.WaitCount = db.Stats().WaitCount
	cp.stats.WaitDuration = db.Stats().WaitDuration
	cp.stats.MaxIdleClosed = db.Stats().MaxIdleClosed
	cp.stats.MaxLifetimeClosed = db.Stats().MaxLifetimeClosed
	cp.stats.LastUpdated = time.Now()
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	cp.statsMutex.RLock()
	defer cp.statsMutex.RUnlock()

	return map[string]interface{}{
		"enabled":              cp.enabled,
		"max_open":             cp.maxOpen,
		"max_idle":             cp.maxIdle,
		"max_lifetime":         cp.maxLifetime.String(),
		"max_open_connections": cp.stats.MaxOpenConnections,
		"open_connections":     cp.stats.OpenConnections,
		"in_use":               cp.stats.InUse,
		"idle":                 cp.stats.Idle,
		"wait_count":           cp.stats.WaitCount,
		"wait_duration":        cp.stats.WaitDuration.String(),
		"max_idle_closed":      cp.stats.MaxIdleClosed,
		"max_lifetime_closed":  cp.stats.MaxLifetimeClosed,
		"last_updated":         cp.stats.LastUpdated,
	}
}

// IsEnabled returns whether connection pooling is enabled
func (cp *ConnectionPool) IsEnabled() bool {
	return cp.enabled
}

// GetConnection gets a connection from the pool with context
func (cp *ConnectionPool) GetConnection(ctx context.Context, db *sql.DB) (*sql.Conn, error) {
	if !cp.enabled {
		return db.Conn(ctx)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return conn, nil
}

// ReleaseConnection releases a connection back to the pool
func (cp *ConnectionPool) ReleaseConnection(conn *sql.Conn) error {
	if !cp.enabled {
		return nil
	}

	return conn.Close()
}
