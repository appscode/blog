package main

import (
	"context"
	"fmt"
	"github.com/Neaj-Morshad-101/mssql-load-test/clients/mssql"
	"github.com/Neaj-Morshad-101/mssql-load-test/config"
	"github.com/Neaj-Morshad-101/mssql-load-test/metrics"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=================================================================")
	fmt.Println("SQL Server High Concurrency Load Testing Client v2")
	fmt.Println("Supports Read + Write Operations for High Concurrent Users")
	fmt.Println("=================================================================")
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\nConfiguration:")
	fmt.Printf("  Database: %s@%s:%d/%s\n", cfg.DB.User, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)
	fmt.Printf("  Concurrent Workers: %d\n", cfg.Load.ConcurrentWriters)
	fmt.Printf("  Test Duration: %v\n", cfg.Load.Duration)
	fmt.Printf("  Batch Size: %d records (inserts), %d records (reads)\n",
		cfg.Load.BatchSize, cfg.Workload.ReadBatchSize)
	fmt.Printf("  Workload: %d%% Reads, %d%% Inserts, %d%% Updates\n",
		cfg.Workload.ReadPercent, cfg.Workload.InsertPercent, cfg.Workload.UpdatePercent)
	fmt.Printf("  Report Interval: %v\n", cfg.Load.ReportInterval)
	fmt.Println()
	if cfg.Load.ConcurrentWriters >= 5000 {
		fmt.Println("WARNING: HIGH CONCURRENCY MODE")
		fmt.Printf("  Running with %d concurrent workers\n", cfg.Load.ConcurrentWriters)
		fmt.Println("  Ensure SQL Server is configured for high concurrency:")
		fmt.Println("  - max connections >= 1000 (or 0 for unlimited)")
		fmt.Println("  - Sufficient memory and CPU resources")
		fmt.Println()
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Load.Duration+30*time.Second)
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	var cm *mssql.ConnectionManager
	for {
		tcm, err := mssql.NewConnectionManager(&cfg.DB)
		if err != nil {
			fmt.Printf("Failed to create connection manager: %v\n", err)
			fmt.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}
		cm = tcm
		break
	}
	defer cm.Close()
	m := metrics.NewV2()
	lg := mssql.NewLoadGeneratorV2(cm, cfg, m)
	if err := lg.Initialize(ctx); err != nil {
		fmt.Printf("Failed to initialize load generator: %v\n", err)
		os.Exit(1)
	}
	monitorCtx, monitorCancel := context.WithCancel(ctx)
	defer monitorCancel()
	go cm.MonitorConnections(monitorCtx, 5*time.Second, func(stats *mssql.ConnectionStats) {
		m.UpdateConnectionMetrics(stats.CurrentConnections, stats.MaxConnections, stats.AvailableConnections)
	})
	go func() {
		ticker := time.NewTicker(cfg.Load.ReportInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				snapshot := m.GetSnapshot()
				snapshot.Print()
			}
		}
	}()
	lg.Start(ctx)
	testTimer := time.NewTimer(cfg.Load.Duration)
	fmt.Printf("\nStarting load test for %v...\n", cfg.Load.Duration)
	if cfg.Load.ConcurrentWriters >= 5000 {
		fmt.Printf("Simulating %d concurrent users\n", cfg.Load.ConcurrentWriters)
	}
	fmt.Println("Press Ctrl+C to stop early")
	fmt.Println()
	select {
	case <-testTimer.C:
		fmt.Println("\nTest duration completed")
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal, stopping...")
	case <-ctx.Done():
		fmt.Println("\nContext cancelled, stopping...")
	}
	lg.Stop()
	fmt.Println("\nFinal Results:")
	finalSnapshot := m.GetSnapshot()
	finalSnapshot.Print()
	fmt.Println("\n=================================================================")
	fmt.Println("Performance Summary:")
	fmt.Printf("  Average Throughput: %.2f operations/sec\n",
		float64(finalSnapshot.TotalOperations)/finalSnapshot.Duration.Seconds())
	if finalSnapshot.TotalReads > 0 {
		fmt.Printf("  Read Operations: %d (%.2f/sec avg)\n",
			finalSnapshot.TotalReads,
			float64(finalSnapshot.TotalReads)/finalSnapshot.Duration.Seconds())
	}
	if finalSnapshot.TotalInserts > 0 {
		fmt.Printf("  Insert Operations: %d (%.2f/sec avg)\n",
			finalSnapshot.TotalInserts,
			float64(finalSnapshot.TotalInserts)/finalSnapshot.Duration.Seconds())
	}
	if finalSnapshot.TotalUpdates > 0 {
		fmt.Printf("  Update Operations: %d (%.2f/sec avg)\n",
			finalSnapshot.TotalUpdates,
			float64(finalSnapshot.TotalUpdates)/finalSnapshot.Duration.Seconds())
	}
	total := float64(finalSnapshot.TotalOperations + finalSnapshot.TotalErrors)
	if total > 0 {
		fmt.Printf("  Error Rate: %.4f%%\n",
			float64(finalSnapshot.TotalErrors)*100/total)
	}
	fmt.Printf("  Total Data Transferred: %.2f GB\n", float64(finalSnapshot.TotalBytes)/(1024*1024*1024))
	fmt.Println("=================================================================")
	fmt.Println("\n=================================================================")
	fmt.Println("Checking for Data Loss...")
	fmt.Println("=================================================================")
	dataLossCtx, dataLossCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer dataLossCancel()
	totalInDB, lostRecords, err := lg.CheckDataLoss(dataLossCtx)
	if err != nil {
		fmt.Printf("Warning: Data loss check failed: %v\n", err)
	} else {
		dataLossPercent := 0.0
		totalRows := lg.GetTotalRows()
		if totalRows > 0 {
			dataLossPercent = float64(lostRecords) * 100.0 / float64(totalRows)
		}
		fmt.Println("\n=================================================================")
		fmt.Println("Data Loss Report:")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Printf("  Total Rows Tracked (inserts + seed): %d\n", totalRows)
		fmt.Printf("  Records Found in DB: %d\n", totalInDB)
		fmt.Printf("  Records Lost: %d\n", lostRecords)
		fmt.Printf("  Data Loss Percentage: %.2f%%\n", dataLossPercent)
		fmt.Println("=================================================================")
		if lostRecords > 0 {
			fmt.Printf("\nWARNING: %d records were inserted but not found in database!\n", lostRecords)
			fmt.Println("This may indicate:")
			fmt.Println("  - SQL Server crash/restart occurred during test")
			fmt.Println("  - Failover to a replica that was behind")
			fmt.Println("  - Transaction rollback due to replication issues")
		} else if totalInDB > 0 {
			fmt.Println("\nNo data loss detected - all inserted records are present in database")
		}
	}
	fmt.Println("\nCleaning up test data...")
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()
	if cfg.Workload.SkipCleanup {
		fmt.Println("SKIP_CLEANUP=true: leaving table in place for manual inspection.")
		fmt.Printf("  Table: %s (in database %s)\n", cfg.Workload.TableName, cfg.DB.DBName)
		fmt.Printf("  To inspect: SELECT COUNT(*) FROM %s;\n", cfg.Workload.TableName)
	} else {
		if err := lg.Cleanup(cleanupCtx); err != nil {
			fmt.Printf("Warning: Cleanup failed: %v\n", err)
		} else {
			fmt.Println("Test data table deleted successfully")
		}
	}
	fmt.Println("\nTest completed successfully!")
}
