package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the load testing client
type Config struct {
	DB       DBConfig
	Load     LoadConfig
	Workload WorkloadConfig
}

// DBConfig contains database connection information
type DBConfig struct {
	Host                   string
	Port                   int
	User                   string
	Password               string
	DBName                 string
	Encrypt                string // disable, false, true
	TrustServerCertificate bool
	Certificate            string
	MaxOpenConns           int
	MaxIdleConns           int
	MinFreeConns           int
}

// LoadConfig contains load testing parameters
type LoadConfig struct {
	ConcurrentWriters int
	Duration          time.Duration
	BatchSize         int
	ReportInterval    time.Duration
}

// WorkloadConfig defines the workload distribution
type WorkloadConfig struct {
	ReadPercent   int
	InsertPercent int
	UpdatePercent int
	TableName     string
	SeedDataRows  int
	ReadBatchSize int
	SkipCleanup   bool // if true, do not drop the table after the test
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnvAsInt("DB_PORT", 1433)
	cfg.DB.User = getEnv("DB_USER", "sa")
	cfg.DB.Password = getEnv("DB_PASSWORD", "")
	cfg.DB.DBName = getEnv("DB_NAME", "testdb")
	cfg.DB.Encrypt = getEnv("DB_ENCRYPT", "disable")
	cfg.DB.TrustServerCertificate = getEnv("DB_TRUST_SERVER_CERT", "true") == "true"
	cfg.DB.Certificate = getEnv("DB_CERTIFICATE", "")
	cfg.DB.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 50)
	cfg.DB.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	cfg.DB.MinFreeConns = getEnvAsInt("DB_MIN_FREE_CONNS", 5)
	cfg.Load.ConcurrentWriters = getEnvAsInt("CONCURRENT_WRITERS", 10)
	durationSecs := getEnvAsInt("TEST_RUN_DURATION", 300)
	cfg.Load.Duration = time.Duration(durationSecs) * time.Second
	cfg.Load.BatchSize = getEnvAsInt("BATCH_SIZE", 100)
	reportIntervalSecs := getEnvAsInt("REPORT_INTERVAL", 10)
	cfg.Load.ReportInterval = time.Duration(reportIntervalSecs) * time.Second
	cfg.Workload.ReadPercent = getEnvAsInt("READ_PERCENT", 0)
	cfg.Workload.InsertPercent = getEnvAsInt("INSERT_PERCENT", 70)
	cfg.Workload.UpdatePercent = getEnvAsInt("UPDATE_PERCENT", 30)
	cfg.Workload.TableName = getEnv("TABLE_NAME", "load_test_data")
	cfg.Workload.ReadBatchSize = getEnvAsInt("READ_BATCH_SIZE", 10)
	cfg.Workload.SeedDataRows = getEnvAsInt("SEED_DATA_ROWS", 50000)
	cfg.Workload.SkipCleanup = getEnv("SKIP_CLEANUP", "false") == "true"
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DB.Host == "" {
		return fmt.Errorf("DB_HOST cannot be empty")
	}
	if c.DB.User == "" {
		return fmt.Errorf("DB_USER cannot be empty")
	}
	if c.DB.DBName == "" {
		return fmt.Errorf("DB_NAME cannot be empty")
	}
	if c.DB.MinFreeConns < 1 {
		return fmt.Errorf("DB_MIN_FREE_CONNS must be at least 1")
	}
	if c.Load.ConcurrentWriters < 1 {
		return fmt.Errorf("CONCURRENT_WRITERS must be at least 1")
	}
	if c.Load.Duration < time.Second {
		return fmt.Errorf("TEST_RUN_DURATION must be at least 1 second")
	}
	if c.Load.BatchSize < 1 {
		return fmt.Errorf("BATCH_SIZE must be at least 1")
	}
	// SQL Server limit: 2100 parameters per query; 9 cols/row => max 233 rows
	if c.Load.BatchSize > 200 {
		return fmt.Errorf("BATCH_SIZE must be <= 200 for SQL Server parameter limit compliance (2100 params max)")
	}
	totalPercent := c.Workload.ReadPercent + c.Workload.InsertPercent + c.Workload.UpdatePercent
	if totalPercent != 100 {
		return fmt.Errorf("READ_PERCENT + INSERT_PERCENT + UPDATE_PERCENT must equal 100, got %d", totalPercent)
	}
	if c.Workload.ReadBatchSize < 1 {
		return fmt.Errorf("READ_BATCH_SIZE must be at least 1")
	}
	return nil
}

// GetConnectionString returns the SQL Server DSN connection string
func (c *DBConfig) GetConnectionString() string {
	query := url.Values{}
	query.Set("database", c.DBName)
	query.Set("connection timeout", "30")
	query.Set("dial timeout", "30")
	query.Set("encrypt", c.Encrypt)
	if c.TrustServerCertificate {
		query.Set("TrustServerCertificate", "true")
	} else {
		query.Set("TrustServerCertificate", "false")
	}
	if c.Certificate != "" {
		query.Set("certificate", c.Certificate)
	}
	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		RawQuery: query.Encode(),
	}
	return u.String()
}
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
