package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsV2 tracks all performance metrics including read operations
type MetricsV2 struct {
	totalReads       atomic.Int64
	totalInserts     atomic.Int64
	totalUpdates     atomic.Int64
	totalErrors      atomic.Int64
	totalBytes       atomic.Int64
	insertedIDs      sync.Map
	totalInsertedIDs atomic.Int64
	readLatencies    []time.Duration
	insertLatencies  []time.Duration
	updateLatencies  []time.Duration
	latencyMutex     sync.RWMutex
	activeConns      atomic.Int32
	maxConns         atomic.Int32
	availableConns   atomic.Int32
	startTime        time.Time
	lastReportTime   time.Time
	lastReadCount    int64
	lastInsertCount  int64
	lastUpdateCount  int64
	lastErrorCount   int64
	lastBytesCount   int64
}

// MetricsSnapshotV2 represents metrics at a point in time
type MetricsSnapshotV2 struct {
	Duration         time.Duration
	TotalReads       int64
	TotalInserts     int64
	TotalUpdates     int64
	TotalOperations  int64
	TotalErrors      int64
	TotalBytes       int64
	TotalInsertedIDs int64
	LostRecords      int64
	DataLossPercent  float64
	ReadsPerSec      float64
	InsertsPerSec    float64
	UpdatesPerSec    float64
	OpsPerSec        float64
	ErrorsPerSec     float64
	BytesPerSec      float64
	AvgReadLatency   time.Duration
	P95ReadLatency   time.Duration
	P99ReadLatency   time.Duration
	AvgInsertLatency time.Duration
	P95InsertLatency time.Duration
	P99InsertLatency time.Duration
	AvgUpdateLatency time.Duration
	P95UpdateLatency time.Duration
	P99UpdateLatency time.Duration
	ActiveConns      int32
	MaxConns         int32
	AvailableConns   int32
}

// NewV2 creates a new MetricsV2 instance
func NewV2() *MetricsV2 {
	return &MetricsV2{
		startTime:       time.Now(),
		lastReportTime:  time.Now(),
		readLatencies:   make([]time.Duration, 0, LimitArraySize),
		insertLatencies: make([]time.Duration, 0, LimitArraySize),
		updateLatencies: make([]time.Duration, 0, LimitArraySize),
	}
}
func (m *MetricsV2) RecordRead(latency time.Duration, bytesRead int64) {
	m.totalReads.Add(1)
	m.totalBytes.Add(bytesRead)
	m.latencyMutex.Lock()
	m.readLatencies = append(m.readLatencies, latency)
	if len(m.readLatencies) > LimitArraySize {
		m.readLatencies = m.readLatencies[len(m.readLatencies)-LimitArraySize:]
	}
	m.latencyMutex.Unlock()
}
func (m *MetricsV2) RecordInsert(latency time.Duration, bytesWritten int64) {
	m.totalInserts.Add(1)
	m.totalBytes.Add(bytesWritten)
	m.latencyMutex.Lock()
	m.insertLatencies = append(m.insertLatencies, latency)
	if len(m.insertLatencies) > LimitArraySize {
		m.insertLatencies = m.insertLatencies[len(m.insertLatencies)-LimitArraySize:]
	}
	m.latencyMutex.Unlock()
}
func (m *MetricsV2) RecordInsertedID(id int64) {
	m.insertedIDs.Store(id, true)
	m.totalInsertedIDs.Add(1)
}
func (m *MetricsV2) GetInsertedIDs() []int64 {
	ids := make([]int64, 0)
	m.insertedIDs.Range(func(key, value interface{}) bool {
		if id, ok := key.(int64); ok {
			ids = append(ids, id)
		}
		return true
	})
	return ids
}
func (m *MetricsV2) RecordUpdate(latency time.Duration, bytesWritten int64) {
	m.totalUpdates.Add(1)
	m.totalBytes.Add(bytesWritten)
	m.latencyMutex.Lock()
	m.updateLatencies = append(m.updateLatencies, latency)
	if len(m.updateLatencies) > LimitArraySize {
		m.updateLatencies = m.updateLatencies[len(m.updateLatencies)-LimitArraySize:]
	}
	m.latencyMutex.Unlock()
}
func (m *MetricsV2) RecordError() {
	m.totalErrors.Add(1)
}
func (m *MetricsV2) UpdateConnectionMetrics(active, max, available int32) {
	m.activeConns.Store(active)
	m.maxConns.Store(max)
	m.availableConns.Store(available)
}
func (m *MetricsV2) GetSnapshot() MetricsSnapshotV2 {
	now := time.Now()
	duration := now.Sub(m.startTime)
	intervalDuration := now.Sub(m.lastReportTime)
	snapshot := MetricsSnapshotV2{
		Duration:       duration,
		TotalReads:     m.totalReads.Load(),
		TotalInserts:   m.totalInserts.Load(),
		TotalUpdates:   m.totalUpdates.Load(),
		TotalErrors:    m.totalErrors.Load(),
		TotalBytes:     m.totalBytes.Load(),
		ActiveConns:    m.activeConns.Load(),
		MaxConns:       m.maxConns.Load(),
		AvailableConns: m.availableConns.Load(),
	}
	snapshot.TotalOperations = snapshot.TotalReads + snapshot.TotalInserts + snapshot.TotalUpdates
	if intervalDuration.Seconds() > 0 {
		readsDiff := snapshot.TotalReads - m.lastReadCount
		insertsDiff := snapshot.TotalInserts - m.lastInsertCount
		updatesDiff := snapshot.TotalUpdates - m.lastUpdateCount
		errorsDiff := snapshot.TotalErrors - m.lastErrorCount
		bytesDiff := snapshot.TotalBytes - m.lastBytesCount
		snapshot.ReadsPerSec = float64(readsDiff) / intervalDuration.Seconds()
		snapshot.InsertsPerSec = float64(insertsDiff) / intervalDuration.Seconds()
		snapshot.UpdatesPerSec = float64(updatesDiff) / intervalDuration.Seconds()
		snapshot.OpsPerSec = float64(readsDiff+insertsDiff+updatesDiff) / intervalDuration.Seconds()
		snapshot.ErrorsPerSec = float64(errorsDiff) / intervalDuration.Seconds()
		snapshot.BytesPerSec = float64(bytesDiff) / intervalDuration.Seconds()
	}
	m.latencyMutex.RLock()
	if len(m.readLatencies) > 0 {
		snapshot.AvgReadLatency = calculateAvg(m.readLatencies)
		snapshot.P95ReadLatency = calculatePercentile(m.readLatencies, 95)
		snapshot.P99ReadLatency = calculatePercentile(m.readLatencies, 99)
	}
	if len(m.insertLatencies) > 0 {
		snapshot.AvgInsertLatency = calculateAvg(m.insertLatencies)
		snapshot.P95InsertLatency = calculatePercentile(m.insertLatencies, 95)
		snapshot.P99InsertLatency = calculatePercentile(m.insertLatencies, 99)
	}
	if len(m.updateLatencies) > 0 {
		snapshot.AvgUpdateLatency = calculateAvg(m.updateLatencies)
		snapshot.P95UpdateLatency = calculatePercentile(m.updateLatencies, 95)
		snapshot.P99UpdateLatency = calculatePercentile(m.updateLatencies, 99)
	}
	m.latencyMutex.RUnlock()
	m.lastReadCount = snapshot.TotalReads
	m.lastInsertCount = snapshot.TotalInserts
	m.lastUpdateCount = snapshot.TotalUpdates
	m.lastErrorCount = snapshot.TotalErrors
	m.lastBytesCount = snapshot.TotalBytes
	m.lastReportTime = now
	return snapshot
}

// Print prints the metrics snapshot in a readable format
func (s *MetricsSnapshotV2) Print() {
	fmt.Println("=================================================================")
	fmt.Printf("Test Duration: %v\n", s.Duration.Round(time.Second))
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("Cumulative Statistics:")
	fmt.Printf("  Total Operations: %d (Reads: %d, Inserts: %d, Updates: %d)\n",
		s.TotalOperations, s.TotalReads, s.TotalInserts, s.TotalUpdates)
	fmt.Printf("  Total Errors: %d\n", s.TotalErrors)
	fmt.Printf("  Total Data Transferred: %.2f MB\n", float64(s.TotalBytes)/(1024*1024))
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("Current Throughput (interval):")
	fmt.Printf("  Operations/sec: %.2f (Reads: %.2f/s, Inserts: %.2f/s, Updates: %.2f/s)\n",
		s.OpsPerSec, s.ReadsPerSec, s.InsertsPerSec, s.UpdatesPerSec)
	fmt.Printf("  Throughput: %.2f MB/s\n", s.BytesPerSec/(1024*1024))
	fmt.Printf("  Errors/sec: %.2f\n", s.ErrorsPerSec)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("Latency Statistics:")
	if s.AvgReadLatency > 0 {
		fmt.Printf("  Reads   - Avg: %v, P95: %v, P99: %v\n",
			s.AvgReadLatency.Round(time.Microsecond),
			s.P95ReadLatency.Round(time.Microsecond),
			s.P99ReadLatency.Round(time.Microsecond))
	}
	if s.AvgInsertLatency > 0 {
		fmt.Printf("  Inserts - Avg: %v, P95: %v, P99: %v\n",
			s.AvgInsertLatency.Round(time.Microsecond),
			s.P95InsertLatency.Round(time.Microsecond),
			s.P99InsertLatency.Round(time.Microsecond))
	}
	if s.AvgUpdateLatency > 0 {
		fmt.Printf("  Updates - Avg: %v, P95: %v, P99: %v\n",
			s.AvgUpdateLatency.Round(time.Microsecond),
			s.P95UpdateLatency.Round(time.Microsecond),
			s.P99UpdateLatency.Round(time.Microsecond))
	}
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("Connection Pool:")
	fmt.Printf("  Active: %d, Max: %d, Available: %d\n",
		s.ActiveConns, s.MaxConns, s.AvailableConns)
	fmt.Println("=================================================================")
}
