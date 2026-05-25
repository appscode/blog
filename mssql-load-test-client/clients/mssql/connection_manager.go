package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Neaj-Morshad-101/mssql-load-test/config"
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
	"time"
)

// ConnectionManager manages SQL Server connections with safety checks
type ConnectionManager struct {
	db     *sql.DB
	config *config.DBConfig
}

// ConnectionStats represents the current connection state
type ConnectionStats struct {
	MaxConnections       int32
	CurrentConnections   int32
	AvailableConnections int32
	CanConnect           bool
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(cfg *config.DBConfig) (*ConnectionManager, error) {
	cm := &ConnectionManager{
		config: cfg,
	}
	connStr := cfg.GetConnectionString()
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQL Server: %w", err)
	}
	cm.db = db
	// Check connection stats
	stats, err := cm.GetConnectionStats(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to get connection stats: %w", err)
	}
	if !stats.CanConnect {
		db.Close()
		return nil, fmt.Errorf(
			"insufficient available connections: max=%d, current=%d, available=%d, required_free=%d",
			stats.MaxConnections,
			stats.CurrentConnections,
			stats.AvailableConnections,
			cfg.MinFreeConns,
		)
	}
	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(15 * time.Minute)
	fmt.Printf("Connection Manager initialized successfully\n")
	fmt.Printf("  Max connections in DB: %d\n", stats.MaxConnections)
	fmt.Printf("  Current active connections: %d\n", stats.CurrentConnections)
	fmt.Printf("  Available connections: %d\n", stats.AvailableConnections)
	fmt.Printf("  Client pool size: %d (max open), %d (max idle)\n", cfg.MaxOpenConns, cfg.MaxIdleConns)
	return cm, nil
}

// GetConnectionStats retrieves current connection statistics from SQL Server
func (cm *ConnectionManager) GetConnectionStats(ctx context.Context) (*ConnectionStats, error) {
	stats := &ConnectionStats{}
	// @@MAX_CONNECTIONS works reliably from any database context (including AG databases).
	// When the server-level config is 0 (unlimited), SQL Server returns 32767 here.
	var maxConnsCfg int32
	err := cm.db.QueryRowContext(ctx, "SELECT @@MAX_CONNECTIONS").Scan(&maxConnsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get max connections: %w", err)
	}
	if maxConnsCfg == 0 {
		maxConnsCfg = 32767
	}
	stats.MaxConnections = maxConnsCfg
	// Count active user sessions across the whole server
	var currentConns int32
	err = cm.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sys.dm_exec_sessions WHERE is_user_process = 1",
	).Scan(&currentConns)
	if err != nil {
		return nil, fmt.Errorf("failed to get current connections: %w", err)
	}
	stats.CurrentConnections = currentConns
	stats.AvailableConnections = stats.MaxConnections - stats.CurrentConnections
	stats.CanConnect = stats.AvailableConnections >= int32(cm.config.MinFreeConns)
	return stats, nil
}

// GetDB returns the underlying database connection
func (cm *ConnectionManager) GetDB() *sql.DB {
	return cm.db
}

// Close closes the database connection
func (cm *ConnectionManager) Close() error {
	if cm.db != nil {
		return cm.db.Close()
	}
	return nil
}

// GetDBStats returns database/sql connection pool stats
func (cm *ConnectionManager) GetDBStats() sql.DBStats {
	return cm.db.Stats()
}

// MonitorConnections periodically monitors connection stats
func (cm *ConnectionManager) MonitorConnections(ctx context.Context, interval time.Duration, callback func(*ConnectionStats)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := cm.GetConnectionStats(ctx)
			if err != nil {
				fmt.Printf("Error getting connection stats: %v\n", err)
				continue
			}
			callback(stats)
		}
	}
}

// HealthCheck performs a health check on the connection
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	if err := cm.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	stats, err := cm.GetConnectionStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection stats: %w", err)
	}
	if !stats.CanConnect {
		return fmt.Errorf(
			"insufficient connections available: max=%d, current=%d, available=%d",
			stats.MaxConnections,
			stats.CurrentConnections,
			stats.AvailableConnections,
		)
	}
	return nil
}
