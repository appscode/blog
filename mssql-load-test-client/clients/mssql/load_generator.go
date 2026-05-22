package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Neaj-Morshad-101/mssql-load-test/config"
	"github.com/Neaj-Morshad-101/mssql-load-test/metrics"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// LoadGenerator generates write load on SQL Server
type LoadGenerator struct {
	cm        *ConnectionManager
	config    *config.Config
	metrics   *metrics.Metrics
	wg        sync.WaitGroup
	stopChan  chan struct{}
	stopOnce  sync.Once
	tableName string
}

// TestRecord represents a sample record for load testing
type TestRecord struct {
	ID          int64
	Name        string
	Email       string
	Age         int
	Address     string
	PhoneNumber string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Data        string
	Status      string
	Score       int
}

// TotalBytes returns an approximate byte count for the record
func (t TestRecord) TotalBytes() int64 {
	total := 0
	total += 8 // ID
	total += 8 // Age
	total += 8 // Score
	total += 8 // CreatedAt
	total += 8 // UpdatedAt
	total += len(t.Name)
	total += len(t.Email)
	total += len(t.Address)
	total += len(t.PhoneNumber)
	total += len(t.Data)
	total += len(t.Status)
	return int64(total)
}

// NewLoadGenerator creates a new load generator
func NewLoadGenerator(cm *ConnectionManager, cfg *config.Config, m *metrics.Metrics) *LoadGenerator {
	return &LoadGenerator{
		cm:        cm,
		config:    cfg,
		metrics:   m,
		stopChan:  make(chan struct{}),
		tableName: cfg.Workload.TableName,
	}
}

// Initialize sets up the test table and prepares the database
func (lg *LoadGenerator) Initialize(ctx context.Context) error {
	fmt.Println("Initializing SQL Server load generator...")
	// Create table if it doesn't exist
	createTableSQL := fmt.Sprintf(`
IF OBJECT_ID(N'%s', N'U') IS NULL
CREATE TABLE %s (
id          BIGINT IDENTITY(1,1) PRIMARY KEY,
name        NVARCHAR(255) NOT NULL,
email       NVARCHAR(255) NOT NULL,
age         INT NOT NULL,
address     NVARCHAR(MAX),
phone_number NVARCHAR(20),
created_at  DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
updated_at  DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
data        NVARCHAR(MAX),
status      NVARCHAR(50) DEFAULT 'active',
score       INT DEFAULT 0
)
`, lg.tableName, lg.tableName)
	if _, err := lg.cm.GetDB().ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	// Create indexes if they don't exist
	indexes := []struct{ name, def string }{
		{"idx_" + lg.tableName + "_email", fmt.Sprintf("CREATE INDEX idx_%s_email ON %s(email)", lg.tableName, lg.tableName)},
		{"idx_" + lg.tableName + "_created_at", fmt.Sprintf("CREATE INDEX idx_%s_created_at ON %s(created_at)", lg.tableName, lg.tableName)},
	}
	for _, idx := range indexes {
		sql := fmt.Sprintf(`
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = '%s' AND object_id = OBJECT_ID('%s'))
%s`, idx.name, lg.tableName, idx.def)
		if _, err := lg.cm.GetDB().ExecContext(ctx, sql); err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}
	// Seed initial data if table is empty
	var count int64
	if err := lg.cm.GetDB().QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", lg.tableName)).Scan(&count); err != nil {
		return fmt.Errorf("failed to count records: %w", err)
	}
	if count == 0 {
		fmt.Println("Seeding table with initial data for update operations...")
		if err := lg.seedInitialData(ctx, 10000); err != nil {
			return fmt.Errorf("failed to seed initial data: %w", err)
		}
		fmt.Printf("Seeded %d initial records\n", 10000)
	} else {
		fmt.Printf("Table already contains %d records\n", count)
	}
	fmt.Println("Load generator initialized successfully")
	return nil
}
func (lg *LoadGenerator) seedInitialData(ctx context.Context, count int) error {
	batchSize := 100
	for i := 0; i < count; i += batchSize {
		remaining := count - i
		if remaining > batchSize {
			remaining = batchSize
		}
		records := make([]TestRecord, remaining)
		for j := 0; j < remaining; j++ {
			records[j] = lg.generateRecord()
		}
		if err := lg.batchInsertWithoutTracking(ctx, records); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the load generation with multiple workers
func (lg *LoadGenerator) Start(ctx context.Context) {
	fmt.Printf("Starting %d concurrent workers...\n", lg.config.Load.ConcurrentWriters)
	for i := 0; i < lg.config.Load.ConcurrentWriters; i++ {
		lg.wg.Add(1)
		go lg.worker(ctx, i)
	}
	fmt.Println("All workers started successfully")
}
func (lg *LoadGenerator) worker(ctx context.Context, workerID int) {
	defer lg.wg.Done()
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
	for {
		select {
		case <-ctx.Done():
			return
		case <-lg.stopChan:
			return
		default:
			roll := rng.Intn(100)
			if roll < lg.config.Workload.InsertPercent {
				lg.performInsert(ctx, rng)
			} else {
				lg.performUpdate(ctx, rng)
			}
		}
	}
}
func (lg *LoadGenerator) performInsert(ctx context.Context, rng *rand.Rand) {
	start := time.Now()
	records := make([]TestRecord, lg.config.Load.BatchSize)
	bytesWritten := int64(0)
	for i := 0; i < lg.config.Load.BatchSize; i++ {
		records[i] = lg.generateRecord()
		bytesWritten += records[i].TotalBytes()
	}
	err := lg.batchInsert(ctx, records)
	latency := time.Since(start)
	if err != nil {
		lg.metrics.RecordError()
		return
	}
	lg.metrics.RecordInsert(latency, bytesWritten)
}

// batchInsert inserts records and captures generated IDs via OUTPUT clause
func (lg *LoadGenerator) batchInsert(ctx context.Context, records []TestRecord) error {
	if len(records) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(records))
	valueArgs := make([]interface{}, 0, len(records)*9)
	for i, record := range records {
		base := i * 9
		valueStrings = append(valueStrings, fmt.Sprintf("(@p%d,@p%d,@p%d,@p%d,@p%d,@p%d,@p%d,@p%d,@p%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9))
		valueArgs = append(valueArgs,
			record.Name, record.Email, record.Age, record.Address,
			record.PhoneNumber, record.CreatedAt, record.Data, record.Status, record.Score)
	}
	query := fmt.Sprintf(
		`INSERT INTO %s (name,email,age,address,phone_number,created_at,data,status,score)
 OUTPUT INSERTED.id
 VALUES %s`,
		lg.tableName, strings.Join(valueStrings, ","))
	rows, err := lg.cm.GetDB().QueryContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		lg.metrics.RecordInsertedID(id)
	}
	return rows.Err()
}
func (lg *LoadGenerator) batchInsertWithoutTracking(ctx context.Context, records []TestRecord) error {
	if len(records) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(records))
	valueArgs := make([]interface{}, 0, len(records)*9)
	for i, record := range records {
		base := i * 9
		valueStrings = append(valueStrings, fmt.Sprintf("(@p%d,@p%d,@p%d,@p%d,@p%d,@p%d,@p%d,@p%d,@p%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9))
		valueArgs = append(valueArgs,
			record.Name, record.Email, record.Age, record.Address,
			record.PhoneNumber, record.CreatedAt, record.Data, record.Status, record.Score)
	}
	query := fmt.Sprintf(
		`INSERT INTO %s (name,email,age,address,phone_number,created_at,data,status,score)
 VALUES %s`,
		lg.tableName, strings.Join(valueStrings, ","))
	_, err := lg.cm.GetDB().ExecContext(ctx, query, valueArgs...)
	return err
}
func (lg *LoadGenerator) performUpdate(ctx context.Context, rng *rand.Rand) {
	start := time.Now()
	// Get a random row ID using TOP 1 with NEWID() for SQL Server
	var id int64
	query := fmt.Sprintf("SELECT TOP 1 id FROM %s ORDER BY NEWID()", lg.tableName)
	err := lg.cm.GetDB().QueryRowContext(ctx, query).Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			lg.metrics.RecordError()
		}
		return
	}
	record := lg.generateRecord()
	updateQuery := fmt.Sprintf(`
UPDATE %s
SET name        = @p1,
    age         = @p2,
    address     = @p3,
    updated_at  = @p4,
    score       = @p5,
    data        = @p6
WHERE id = @p7
`, lg.tableName)
	_, err = lg.cm.GetDB().ExecContext(ctx, updateQuery,
		record.Name, record.Age, record.Address, time.Now(), record.Score, record.Data, id)
	latency := time.Since(start)
	if err != nil {
		lg.metrics.RecordError()
		return
	}
	lg.metrics.RecordUpdate(latency, record.TotalBytes())
}
func (lg *LoadGenerator) generateRecord() TestRecord {
	statuses := []string{"active", "inactive", "pending"}
	return TestRecord{
		Name:        generateRandomName(),
		Email:       generateRandomEmail(),
		Age:         rand.Intn(80) + 18,
		Address:     generateRandomAddress(),
		PhoneNumber: generateRandomPhone(),
		CreatedAt:   time.Now(),
		Data:        generateRandomData(1024),
		Status:      statuses[rand.Intn(len(statuses))],
		Score:       rand.Intn(1000),
	}
}

// Stop gracefully stops the load generator
func (lg *LoadGenerator) Stop() {
	lg.stopOnce.Do(func() {
		close(lg.stopChan)
		lg.wg.Wait()
		fmt.Println("All workers stopped")
	})
}

// CheckDataLoss verifies how many inserted records are actually in the database
func (lg *LoadGenerator) CheckDataLoss(ctx context.Context) (int64, int64, error) {
	insertedIDs := lg.metrics.GetInsertedIDs()
	if len(insertedIDs) == 0 {
		return 0, 0, nil
	}
	totalInserted := int64(len(insertedIDs))
	fmt.Printf("Checking data loss for %d inserted records...\n", totalInserted)
	if len(insertedIDs) > 100000 {
		fmt.Println("  Using optimized count-based verification for large dataset...")
		return lg.checkDataLossOptimized(ctx, totalInserted)
	}
	// Batch IN query approach for smaller datasets
	batchSize := 2000 // SQL Server IN clause limit
	found := int64(0)
	totalBatches := (len(insertedIDs) + batchSize - 1) / batchSize
	for i := 0; i < len(insertedIDs); i += batchSize {
		select {
		case <-ctx.Done():
			return 0, 0, fmt.Errorf("data loss check cancelled: %w", ctx.Err())
		default:
		}
		end := i + batchSize
		if end > len(insertedIDs) {
			end = len(insertedIDs)
		}
		batch := insertedIDs[i:end]
		currentBatch := (i / batchSize) + 1
		if currentBatch%5 == 0 || currentBatch == totalBatches {
			fmt.Printf("  Progress: Checked %d/%d batches (%.1f%%)\n",
				currentBatch, totalBatches, float64(currentBatch)*100/float64(totalBatches))
		}
		var idStrs []string
		for _, id := range batch {
			idStrs = append(idStrs, fmt.Sprintf("%d", id))
		}
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id IN (%s)",
			lg.tableName, strings.Join(idStrs, ","))
		var count int64
		if err := lg.cm.GetDB().QueryRowContext(ctx, query).Scan(&count); err != nil {
			return 0, 0, fmt.Errorf("failed to check data loss (batch %d): %w", currentBatch, err)
		}
		found += count
	}
	lost := totalInserted - found
	fmt.Printf("Data loss check complete: %d found, %d lost out of %d inserted\n",
		found, lost, totalInserted)
	return totalInserted, lost, nil
}
func (lg *LoadGenerator) checkDataLossOptimized(ctx context.Context, totalInserted int64) (int64, int64, error) {
	var tot int64
	if err := lg.cm.GetDB().QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", lg.tableName)).Scan(&tot); err != nil {
		return 0, 0, fmt.Errorf("failed to count records: %w", err)
	}
	lost := totalInserted - tot
	if lost < 0 {
		lost = 0
	}
	fmt.Printf("Data loss check (count): %d in DB, ~%d lost out of %d inserted\n", tot, lost, totalInserted)
	return totalInserted, lost, nil
}

// Cleanup removes the test table
func (lg *LoadGenerator) Cleanup(ctx context.Context) error {
	fmt.Println("Cleaning up test table...")
	_, err := lg.cm.GetDB().ExecContext(ctx,
		fmt.Sprintf("IF OBJECT_ID(N'%s', N'U') IS NOT NULL DROP TABLE %s", lg.tableName, lg.tableName))
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}
	fmt.Println("Cleanup completed")
	return nil
}

// Helper functions to generate random data
func generateRandomName() string {
	firstNames := []string{"John", "Jane", "Michael", "Emily", "David", "Sarah", "Robert", "Lisa", "William", "Jennifer"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}
func generateRandomEmail() string {
	domains := []string{"example.com", "test.com", "email.com", "mail.com"}
	return fmt.Sprintf("user%d@%s", rand.Intn(1000000), domains[rand.Intn(len(domains))])
}
func generateRandomAddress() string {
	streets := []string{"Main St", "Oak Ave", "Maple Dr", "Cedar Ln", "Pine Rd"}
	cities := []string{"Springfield", "Riverside", "Madison", "Georgetown", "Franklin"}
	return fmt.Sprintf("%d %s, %s, CA %05d",
		rand.Intn(9999)+1,
		streets[rand.Intn(len(streets))],
		cities[rand.Intn(len(cities))],
		rand.Intn(99999),
	)
}
func generateRandomPhone() string {
	return fmt.Sprintf("+1-%03d-%03d-%04d",
		rand.Intn(900)+100,
		rand.Intn(900)+100,
		rand.Intn(9000)+1000,
	)
}
func generateRandomData(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
