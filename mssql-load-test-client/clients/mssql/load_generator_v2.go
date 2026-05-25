package mssql

import (
	"context"
	"fmt"
	"github.com/Neaj-Morshad-101/mssql-load-test/config"
	"github.com/Neaj-Morshad-101/mssql-load-test/metrics"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LoadGeneratorV2 generates mixed read/write load on SQL Server
type LoadGeneratorV2 struct {
	cm        *ConnectionManager
	config    *config.Config
	metrics   *metrics.MetricsV2
	wg        sync.WaitGroup
	stopChan  chan struct{}
	stopOnce  sync.Once
	tableName string
	totalRows atomic.Int64
}

// NewLoadGeneratorV2 creates a new enhanced load generator with read support
func NewLoadGeneratorV2(cm *ConnectionManager, cfg *config.Config, m *metrics.MetricsV2) *LoadGeneratorV2 {
	return &LoadGeneratorV2{
		cm:        cm,
		config:    cfg,
		metrics:   m,
		stopChan:  make(chan struct{}),
		tableName: cfg.Workload.TableName,
	}
}

// Initialize sets up the test table and prepares the database
func (lg *LoadGeneratorV2) Initialize(ctx context.Context) error {
	fmt.Println("Initializing enhanced SQL Server load generator with read support...")
	createTableSQL := fmt.Sprintf(`
IF OBJECT_ID(N'%s', N'U') IS NULL
CREATE TABLE %s (
id           BIGINT IDENTITY(1,1) PRIMARY KEY,
name         NVARCHAR(255) NOT NULL,
email        NVARCHAR(255) NOT NULL,
age          INT NOT NULL,
address      NVARCHAR(MAX),
phone_number NVARCHAR(20),
created_at   DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
updated_at   DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
data         NVARCHAR(MAX),
status       NVARCHAR(50) DEFAULT 'active',
score        INT DEFAULT 0
)
`, lg.tableName, lg.tableName)
	if _, err := lg.cm.GetDB().ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	// Create indexes
	indexes := []struct{ name, def string }{
		{"idx_" + lg.tableName + "_email",
			fmt.Sprintf("CREATE INDEX idx_%s_email ON %s(email)", lg.tableName, lg.tableName)},
		{"idx_" + lg.tableName + "_created_at",
			fmt.Sprintf("CREATE INDEX idx_%s_created_at ON %s(created_at)", lg.tableName, lg.tableName)},
		{"idx_" + lg.tableName + "_status_score",
			fmt.Sprintf("CREATE INDEX idx_%s_status_score ON %s(status,score)", lg.tableName, lg.tableName)},
		{"idx_" + lg.tableName + "_name",
			fmt.Sprintf("CREATE INDEX idx_%s_name ON %s(name)", lg.tableName, lg.tableName)},
	}
	for _, idx := range indexes {
		sql := fmt.Sprintf(
			"IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name='%s' AND object_id=OBJECT_ID('%s')) %s",
			idx.name, lg.tableName, idx.def)
		if _, err := lg.cm.GetDB().ExecContext(ctx, sql); err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}
	var count int64
	if err := lg.cm.GetDB().QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", lg.tableName)).Scan(&count); err != nil {
		return fmt.Errorf("failed to count records: %w", err)
	}
	if count == 0 {
		fmt.Println("Seeding table with initial data...")
		if err := lg.seedInitialData(ctx, lg.config.Workload.SeedDataRows); err != nil {
			return fmt.Errorf("failed to seed initial data: %w", err)
		}
		count = int64(lg.config.Workload.SeedDataRows)
		fmt.Printf("Seeded %d initial records\n", count)
	} else {
		fmt.Printf("Table already contains %d records\n", count)
	}
	lg.totalRows.Store(count)
	fmt.Println("Enhanced load generator initialized successfully")
	return nil
}
func (lg *LoadGeneratorV2) seedInitialData(ctx context.Context, count int) error {
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
func (lg *LoadGeneratorV2) Start(ctx context.Context) {
	fmt.Printf("Starting %d concurrent workers with mixed read/write workload...\n", lg.config.Load.ConcurrentWriters)
	fmt.Printf("  Workload: %d%% Reads, %d%% Inserts, %d%% Updates\n",
		lg.config.Workload.ReadPercent,
		lg.config.Workload.InsertPercent,
		lg.config.Workload.UpdatePercent)
	for i := 0; i < lg.config.Load.ConcurrentWriters; i++ {
		lg.wg.Add(1)
		go lg.worker(ctx, i)
	}
	fmt.Println("All workers started successfully")
}
func (lg *LoadGeneratorV2) worker(ctx context.Context, workerID int) {
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
			if roll < lg.config.Workload.ReadPercent {
				lg.performRead(ctx, rng)
			} else if roll < lg.config.Workload.ReadPercent+lg.config.Workload.InsertPercent {
				lg.performInsert(ctx, rng)
			} else {
				lg.performUpdate(ctx, rng)
			}
		}
	}
}

// performRead executes a read/SELECT operation
func (lg *LoadGeneratorV2) performRead(ctx context.Context, rng *rand.Rand) {
	start := time.Now()
	var err error
	var bytesRead int64
	switch rng.Intn(4) {
	case 0:
		bytesRead, err = lg.readByIDRange(ctx, rng)
	case 1:
		bytesRead, err = lg.readByStatus(ctx, rng)
	case 2:
		bytesRead, err = lg.readRecentRecords(ctx)
	case 3:
		bytesRead, err = lg.readByNamePattern(ctx, rng)
	}
	latency := time.Since(start)
	if err != nil {
		lg.metrics.RecordError()
		return
	}
	lg.metrics.RecordRead(latency, bytesRead)
}
func (lg *LoadGeneratorV2) readByIDRange(ctx context.Context, rng *rand.Rand) (int64, error) {
	totalRows := lg.totalRows.Load()
	if totalRows == 0 {
		return 0, fmt.Errorf("no rows to read")
	}
	startID := rng.Int63n(totalRows)
	limit := lg.config.Workload.ReadBatchSize
	query := fmt.Sprintf(`
SELECT TOP (@p1) id,name,email,age,address,phone_number,status,score,data
FROM %s
WHERE id >= @p2
ORDER BY id
`, lg.tableName)
	rows, err := lg.cm.GetDB().QueryContext(ctx, query, limit, startID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var bytesRead int64
	for rows.Next() {
		var r TestRecord
		if err := rows.Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.Address, &r.PhoneNumber, &r.Status, &r.Score, &r.Data); err != nil {
			return bytesRead, err
		}
		bytesRead += r.TotalBytes()
	}
	return bytesRead, rows.Err()
}
func (lg *LoadGeneratorV2) readByStatus(ctx context.Context, rng *rand.Rand) (int64, error) {
	statuses := []string{"active", "inactive", "pending"}
	status := statuses[rng.Intn(len(statuses))]
	query := fmt.Sprintf(`
SELECT TOP (@p1) id,name,email,score,data
FROM %s
WHERE status = @p2
ORDER BY score DESC
`, lg.tableName)
	rows, err := lg.cm.GetDB().QueryContext(ctx, query, lg.config.Workload.ReadBatchSize, status)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var bytesRead int64
	for rows.Next() {
		var id int64
		var name, email, data string
		var score int
		if err := rows.Scan(&id, &name, &email, &score, &data); err != nil {
			return bytesRead, err
		}
		bytesRead += int64(len(name) + len(email) + len(data))
	}
	return bytesRead, rows.Err()
}
func (lg *LoadGeneratorV2) readRecentRecords(ctx context.Context) (int64, error) {
	query := fmt.Sprintf(`
SELECT TOP (@p1) id,name,email,created_at,data
FROM %s
ORDER BY created_at DESC
`, lg.tableName)
	rows, err := lg.cm.GetDB().QueryContext(ctx, query, lg.config.Workload.ReadBatchSize)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var bytesRead int64
	for rows.Next() {
		var id int64
		var name, email, data string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &email, &createdAt, &data); err != nil {
			return bytesRead, err
		}
		bytesRead += int64(len(name) + len(email) + len(data))
	}
	return bytesRead, rows.Err()
}
func (lg *LoadGeneratorV2) readByNamePattern(ctx context.Context, rng *rand.Rand) (int64, error) {
	firstNames := []string{"John", "Jane", "Michael", "Emily", "David", "Sarah", "Robert", "Lisa", "William", "Jennifer"}
	pattern := firstNames[rng.Intn(len(firstNames))] + "%"
	query := fmt.Sprintf(`
SELECT TOP (@p1) id,name,email,data
FROM %s
WHERE name LIKE @p2
`, lg.tableName)
	rows, err := lg.cm.GetDB().QueryContext(ctx, query, lg.config.Workload.ReadBatchSize, pattern)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var bytesRead int64
	for rows.Next() {
		var id int64
		var name, email, data string
		if err := rows.Scan(&id, &name, &email, &data); err != nil {
			return bytesRead, err
		}
		bytesRead += int64(len(name) + len(email) + len(data))
	}
	return bytesRead, rows.Err()
}
func (lg *LoadGeneratorV2) performInsert(ctx context.Context, rng *rand.Rand) {
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
	lg.totalRows.Add(int64(len(records)))
	lg.metrics.RecordInsert(latency, bytesWritten)
}
func (lg *LoadGeneratorV2) batchInsert(ctx context.Context, records []TestRecord) error {
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
func (lg *LoadGeneratorV2) batchInsertWithoutTracking(ctx context.Context, records []TestRecord) error {
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
		`INSERT INTO %s (name,email,age,address,phone_number,created_at,data,status,score) VALUES %s`,
		lg.tableName, strings.Join(valueStrings, ","))
	_, err := lg.cm.GetDB().ExecContext(ctx, query, valueArgs...)
	return err
}
func (lg *LoadGeneratorV2) performUpdate(ctx context.Context, rng *rand.Rand) {
	start := time.Now()
	totalRows := lg.totalRows.Load()
	if totalRows == 0 {
		return
	}
	randomID := rng.Int63n(totalRows) + 1
	record := lg.generateRecord()
	updateQuery := fmt.Sprintf(`
UPDATE %s
SET name       = @p1,
    age        = @p2,
    address    = @p3,
    updated_at = @p4,
    score      = @p5,
    data       = @p6
WHERE id = @p7
`, lg.tableName)
	_, err := lg.cm.GetDB().ExecContext(ctx, updateQuery,
		record.Name, record.Age, record.Address, time.Now(), record.Score, record.Data, randomID)
	latency := time.Since(start)
	if err != nil {
		lg.metrics.RecordError()
		return
	}
	lg.metrics.RecordUpdate(latency, record.TotalBytes())
}
func (lg *LoadGeneratorV2) generateRecord() TestRecord {
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
func (lg *LoadGeneratorV2) Stop() {
	lg.stopOnce.Do(func() {
		close(lg.stopChan)
		lg.wg.Wait()
		fmt.Println("All workers stopped")
	})
}

// CheckDataLoss verifies how many inserted records are actually in the database
func (lg *LoadGeneratorV2) CheckDataLoss(ctx context.Context) (int64, int64, error) {
	var tot int64
	if err := lg.cm.GetDB().QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", lg.tableName)).Scan(&tot); err != nil {
		return 0, 0, fmt.Errorf("failed to count total records for data loss check: %w", err)
	}
	lost := lg.totalRows.Load() - tot
	if lost < 0 {
		lost = 0
	}
	return tot, lost, nil
}

// Cleanup removes the test table
func (lg *LoadGeneratorV2) Cleanup(ctx context.Context) error {
	fmt.Println("Cleaning up test table...")
	_, err := lg.cm.GetDB().ExecContext(ctx,
		fmt.Sprintf("IF OBJECT_ID(N'%s', N'U') IS NOT NULL DROP TABLE %s", lg.tableName, lg.tableName))
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}
	fmt.Println("Cleanup completed")
	return nil
}

// GetTotalRows returns the current row count (including seeded rows)
func (lg *LoadGeneratorV2) GetTotalRows() int64 {
	return lg.totalRows.Load()
}
