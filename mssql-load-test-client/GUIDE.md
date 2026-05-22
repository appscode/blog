# MSSQL Load Test — Developer Guide

A deep-dive reference for understanding, configuring, and running the **mssql-load-test** tool.  
It covers project structure, every configuration knob, the internal execution model, metrics, and step-by-step test scenarios.

---

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Architecture](#architecture)
   - [Execution Flow](#execution-flow)
   - [Package Dependency Graph](#package-dependency-graph)
4. [Configuration Reference](#configuration-reference)
   - [Database Connection](#database-connection)
   - [Connection Pool](#connection-pool)
   - [Load Parameters](#load-parameters)
   - [Workload Distribution](#workload-distribution)
5. [Database Schema](#database-schema)
6. [Load Generator Deep-Dive](#load-generator-deep-dive)
   - [V1 — Write-Only Generator](#v1--write-only-generator)
   - [V2 — Mixed Read/Write Generator (active)](#v2--mixed-readwrite-generator-active)
   - [Seeding](#seeding)
   - [Batch INSERT and the OUTPUT Clause](#batch-insert-and-the-output-clause)
   - [Read Operations](#read-operations)
   - [Update Strategies](#update-strategies)
   - [Data Loss Check](#data-loss-check)
7. [Metrics System](#metrics-system)
   - [Metrics (V1)](#metrics-v1)
   - [MetricsV2 (active)](#metricsv2-active)
   - [Latency Percentiles](#latency-percentiles)
   - [Connection Metrics](#connection-metrics)
8. [Connection Manager](#connection-manager)
9. [KubeDB Client Builder](#kubedb-client-builder)
10. [Building and Running](#building-and-running)
    - [Local Run (plain SQL Server)](#local-run-plain-sql-server)
    - [Build Binary](#build-binary)
    - [Docker Image](#docker-image)
    - [Kubernetes Job](#kubernetes-job)
11. [Testing Scenarios](#testing-scenarios)
    - [Quick Smoke Test](#quick-smoke-test)
    - [Write-Heavy Benchmark](#write-heavy-benchmark)
    - [Read-Heavy Benchmark](#read-heavy-benchmark)
    - [High-Concurrency Stress Test](#high-concurrency-stress-test)
    - [Chaos / Failover Test](#chaos--failover-test)
12. [Understanding the Output](#understanding-the-output)
13. [Tuning Guidance](#tuning-guidance)
14. [SQL Server–Specific Notes](#sql-server-specific-notes)

---

## Overview

`mssql-load-test` is a Go program that hammers a Microsoft SQL Server instance with configurable concurrent workers performing **batch INSERTs**, **random UPDATEs**, and **SELECT** reads.  
After the test it verifies **data durability** (no silent data loss) and drops the test table.

Primary use-case: **chaos / failover testing** of KubeDB-managed SQL Server instances on Kubernetes.

---

## Project Structure

```
mssql-load-test/
├── main.go                         # Entry point – wires everything together
├── go.mod / go.sum                 # Module: github.com/Neaj-Morshad-101/mssql-load-test
├── Makefile                        # build / run / docker / k8s helpers
├── Dockerfile                      # Multi-stage distroless image
│
├── config/
│   └── config.go                   # All env-var config parsing + validation
│
├── clients/mssql/
│   ├── client.go                   # Thin wrapper: type Client struct { *sql.DB }
│   ├── connection_manager.go       # sql.DB lifecycle + connection stats from sys views
│   ├── load_generator.go           # V1 – write-only (INSERT + UPDATE), used as reference
│   ├── load_generator_v2.go        # V2 – mixed R/W, the one used by main.go
│   └── kubedb_client_builder.go    # KubeDB-native builder (reads creds from K8s Secrets)
│
├── metrics/
│   ├── metrics.go                  # V1 metrics (insert/update only)
│   └── metrics_v2.go              # V2 metrics (read + insert + update), used by main.go
│
└── k8s/
    ├── 00-namespace.yaml
    ├── 01-configmap.yaml           # All non-secret env vars
    ├── 02-secret.yaml              # DB_PASSWORD
    └── 03-job.yaml                 # Kubernetes Job definition
```

---

## Architecture

### Execution Flow

```
main()
  │
  ├─ config.LoadFromEnv()           ← reads + validates all env vars
  │
  ├─ retry loop: mssql.NewConnectionManager()
  │     └─ sql.Open("sqlserver", dsn)
  │     └─ db.Ping()
  │     └─ query sys.configurations  (max connections)
  │     └─ query sys.dm_exec_sessions (active sessions)
  │     └─ db.SetMaxOpenConns / SetMaxIdleConns
  │
  ├─ metrics.NewV2()
  │
  ├─ mssql.NewLoadGeneratorV2(cm, cfg, m)
  │
  ├─ lg.Initialize(ctx)
  │     └─ CREATE TABLE IF NOT EXISTS load_test_data (...)
  │     └─ CREATE INDEX (x4, idempotent)
  │     └─ COUNT(*) — if 0 → seedInitialData(SeedDataRows)
  │
  ├─ goroutine: cm.MonitorConnections()  ← every 5 s, updates metrics
  │
  ├─ goroutine: ticker → m.GetSnapshot().Print()  ← every ReportInterval
  │
  ├─ lg.Start(ctx)
  │     └─ spawn ConcurrentWriters goroutines, each running worker()
  │           └─ roll = rand(100)
  │                 roll < ReadPercent        → performRead()
  │                 roll < Read+InsertPercent → performInsert()
  │                 else                      → performUpdate()
  │
  ├─ wait for: timer | SIGINT | ctx cancel
  │
  ├─ lg.Stop()           ← close(stopChan), wg.Wait()
  ├─ m.GetSnapshot().Print()
  ├─ lg.CheckDataLoss()  ← COUNT(*) vs tracked totalRows
  └─ lg.Cleanup()        ← DROP TABLE IF EXISTS
```

### Package Dependency Graph

```
main
 ├── config          (no internal deps)
 ├── metrics         (no internal deps)
 └── clients/mssql
      ├── config
      ├── metrics
      └── (external) github.com/microsoft/go-mssqldb
                     kubedb.dev/apimachinery  (kubedb_client_builder only)
                     sigs.k8s.io/controller-runtime
                     k8s.io/api
```

---

## Configuration Reference

All configuration is done via **environment variables**. There are no config files.

### Database Connection

| Env var | Default | Description |
|---------|---------|-------------|
| `DB_HOST` | `localhost` | SQL Server hostname or IP |
| `DB_PORT` | `1433` | TCP port |
| `DB_USER` | `sa` | Login name |
| `DB_PASSWORD` | _(empty)_ | Password — set via Secret in K8s |
| `DB_NAME` | `testdb` | Database to connect to (must exist) |
| `DB_ENCRYPT` | `disable` | TLS encryption: `disable`, `false`, `true` |
| `DB_TRUST_SERVER_CERT` | `true` | Skip TLS certificate verification |
| `DB_CERTIFICATE` | _(empty)_ | Path to server certificate file (optional) |

**Connection string format** (built by `config.DBConfig.GetConnectionString()`):
```
sqlserver://sa:password@host:1433?database=testdb&encrypt=disable&TrustServerCertificate=true&connection+timeout=30&dial+timeout=30
```

### Connection Pool

| Env var | Default | Description |
|---------|---------|-------------|
| `DB_MAX_OPEN_CONNS` | `50` | `sql.DB.SetMaxOpenConns` — max live connections |
| `DB_MAX_IDLE_CONNS` | `10` | `sql.DB.SetMaxIdleConns` — kept open when idle |
| `DB_MIN_FREE_CONNS` | `5` | Minimum free server-side slots required at startup |

Connection lifetime is hardcoded to **1 hour** max, **15 minutes** max idle.

### Load Parameters

| Env var | Default | Description |
|---------|---------|-------------|
| `CONCURRENT_WRITERS` | `10` | Number of goroutine workers |
| `TEST_RUN_DURATION` | `300` | Test duration in **seconds** |
| `BATCH_SIZE` | `100` | Rows per INSERT batch (max **200**, enforced by validation) |
| `REPORT_INTERVAL` | `10` | Metrics print interval in **seconds** |

> **Why max 200?** SQL Server caps query parameters at **2100**. Each row uses 9 parameters → floor(2100 / 9) = 233. The tool caps it at 200 to stay safely under the limit.

### Workload Distribution

| Env var | Default | Description |
|---------|---------|-------------|
| `READ_PERCENT` | `0` | % of operations that are SELECTs |
| `INSERT_PERCENT` | `70` | % of operations that are batch INSERTs |
| `UPDATE_PERCENT` | `30` | % of operations that are UPDATE by ID |
| `TABLE_NAME` | `load_test_data` | Name of the test table |
| `READ_BATCH_SIZE` | `10` | Rows fetched per SELECT query |
| `SEED_DATA_ROWS` | `50000` | Rows pre-loaded before the test starts |

> **Constraint**: `READ_PERCENT + INSERT_PERCENT + UPDATE_PERCENT` **must equal 100**, or the program exits with an error.

---

## Database Schema

The tool creates this table (idempotent via `IF OBJECT_ID(...) IS NULL`):

```sql
CREATE TABLE load_test_data (
    id           BIGINT IDENTITY(1,1) PRIMARY KEY,
    name         NVARCHAR(255) NOT NULL,
    email        NVARCHAR(255) NOT NULL,
    age          INT NOT NULL,
    address      NVARCHAR(MAX),
    phone_number NVARCHAR(20),
    created_at   DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
    updated_at   DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
    data         NVARCHAR(MAX),          -- ~1 KB random payload
    status       NVARCHAR(50) DEFAULT 'active',
    score        INT DEFAULT 0
);
```

Four **non-clustered indexes** (idempotent via `sys.indexes` check):

| Index name | Columns | Purpose |
|------------|---------|---------|
| `idx_<table>_email` | `email` | Email-based lookups |
| `idx_<table>_created_at` | `created_at` | Recent-records ordered scan |
| `idx_<table>_status_score` | `status, score` | Status-filtered score sort |
| `idx_<table>_name` | `name` | LIKE prefix scans |

---

## Load Generator Deep-Dive

### V1 — Write-Only Generator

File: `clients/mssql/load_generator.go`  
Type: `LoadGenerator`  
**Not used by `main.go`** — kept for reference / alternative wiring.

Differences from V2:
- No read operations (only INSERT + UPDATE).
- `performUpdate` picks a random row via `SELECT TOP 1 id FROM table ORDER BY NEWID()` — a true random pick, but can be slow on large tables.
- Data loss check tracks exact inserted IDs: for ≤100 000 rows it does batched `IN (...)` queries; for larger sets it falls back to `COUNT(*)`.
- Seeds 10 000 rows (hardcoded) vs V2's configurable `SEED_DATA_ROWS`.

### V2 — Mixed Read/Write Generator (active)

File: `clients/mssql/load_generator_v2.go`  
Type: `LoadGeneratorV2`  
**This is what `main.go` uses.**

Each worker goroutine independently rolls a random number every iteration:

```
roll = rand(100)
```

| roll range | action |
|------------|--------|
| `[0, ReadPercent)` | `performRead` |
| `[ReadPercent, ReadPercent+InsertPercent)` | `performInsert` |
| `[ReadPercent+InsertPercent, 100)` | `performUpdate` |

Workers use their own seeded `rand.Rand` (seeded with `time.Now().UnixNano() + workerID`) to avoid lock contention on the global source.

### Seeding

Before workers start, `Initialize` checks `COUNT(*)`. If the table is empty it calls `seedInitialData(SeedDataRows)`:

- Rows are inserted in batches of **100** (hardcoded) using `batchInsertWithoutTracking` (no `OUTPUT INSERTED.id` — seed rows are not tracked for data-loss).
- Default: **50 000** seed rows (configurable via `SEED_DATA_ROWS`).
- The seed row count is stored in `totalRows` atomic so UPDATE workers immediately have valid IDs to target.

### Batch INSERT and the OUTPUT Clause

`batchInsert` builds a single multi-row `INSERT ... VALUES` statement using numbered parameters:

```sql
INSERT INTO load_test_data (name, email, age, address, phone_number, created_at, data, status, score)
OUTPUT INSERTED.id
VALUES (@p1,@p2,@p3,@p4,@p5,@p6,@p7,@p8,@p9),
       (@p10,@p11,...),
       ...
```

Parameter naming: `@p1`, `@p2`, ... (SQL Server / `go-mssqldb` convention, unlike Postgres `$1`, `$2`).

`OUTPUT INSERTED.id` returns the auto-generated `BIGINT IDENTITY` values which are recorded in `MetricsV2.insertedIDs` (`sync.Map`) for later data-loss checking.

`batchInsertWithoutTracking` is identical but omits `OUTPUT INSERTED.id` and uses `ExecContext` — used for seeding where ID tracking is unnecessary.

### Read Operations

`performRead` randomly picks one of four query patterns:

| Pattern | Query | Use case |
|---------|-------|----------|
| `readByIDRange` | `SELECT TOP N ... WHERE id >= @p2 ORDER BY id` | Sequential scan from a random offset |
| `readByStatus` | `SELECT TOP N ... WHERE status = @p2 ORDER BY score DESC` | Filtered + sorted scan |
| `readRecentRecords` | `SELECT TOP N ... ORDER BY created_at DESC` | Hot-path "latest rows" query |
| `readByNamePattern` | `SELECT TOP N ... WHERE name LIKE @p2` | Index prefix scan |

The `TOP (@p1)` clause is parameterised (correct usage with `go-mssqldb`).  
`READ_BATCH_SIZE` controls `N` (default 10).

### Update Strategies

**V2** (`performUpdate`):
```sql
UPDATE load_test_data
SET name=@p1, age=@p2, address=@p3, updated_at=@p4, score=@p5, data=@p6
WHERE id = @p7
```
The target `id` is computed as `rand(totalRows) + 1`. This is an in-process random pick — very fast but may occasionally miss a gap if a row was never inserted at that ID. Missing rows are a no-op (`0 rows affected`), not an error.

**V1** (`performUpdate`):
```sql
SELECT TOP 1 id FROM load_test_data ORDER BY NEWID()
```
Then separate UPDATE by `id`. Slower (two round-trips) but guaranteed to hit an existing row. Uses `sql.ErrNoRows` guard for the empty-table edge case.

### Data Loss Check

After `lg.Stop()`, `CheckDataLoss` is called with a 5-minute timeout:

```go
SELECT COUNT(*) FROM load_test_data
```

```
lost = totalRows.Load() - COUNT(*)
if lost < 0 { lost = 0 }   // more rows than expected (e.g., table pre-existed)
```

`totalRows` includes both seed rows and all successfully committed inserts tracked via `totalRows.Add(int64(len(records)))` in `performInsert`.

> **Why `lost < 0` is possible**: if the table already existed with data before the test ran, or if `SEED_DATA_ROWS` was used and the actual DB already had rows.

---

## Metrics System

### Metrics (V1)

`metrics/metrics.go` — tracks inserts + updates only.  
Used by the V1 `LoadGenerator` (not wired in `main.go`).

### MetricsV2 (active)

`metrics/metrics_v2.go` — tracks reads + inserts + updates.

All counters use `sync/atomic` primitives for lock-free increments:

| Atomic field | Type | Tracks |
|---|---|---|
| `totalReads` | `atomic.Int64` | read operations |
| `totalInserts` | `atomic.Int64` | insert batches |
| `totalUpdates` | `atomic.Int64` | update operations |
| `totalErrors` | `atomic.Int64` | any DB error |
| `totalBytes` | `atomic.Int64` | data volume (reads + writes) |
| `insertedIDs` | `sync.Map` | set of every inserted `id` |

Latency slices (`readLatencies`, `insertLatencies`, `updateLatencies`) are protected by a single `sync.RWMutex`. They are capped at **1 000 000** entries (constant `LimitArraySize`) to bound memory.

`GetSnapshot()` returns a point-in-time `MetricsSnapshotV2` that also computes **interval throughput** (ops/sec since last snapshot) using the `last*` counters.

### Latency Percentiles

`calculatePercentile(durations, p)`:
1. Copy the slice.
2. Sort ascending.
3. Return `sorted[ floor(len * p / 100) ]`.

Reported: **Avg**, **P95**, **P99** for each operation type.

### Connection Metrics

`UpdateConnectionMetrics(active, max, available int32)` is called every 5 seconds from the `MonitorConnections` goroutine, which queries:

```sql
-- Server max connections (0 = unlimited → treated as 32767)
SELECT CAST(value_in_use AS INT)
FROM sys.configurations
WHERE name = 'max connections'

-- Active user sessions
SELECT COUNT(*)
FROM sys.dm_exec_sessions
WHERE is_user_process = 1
```

---

## Connection Manager

`clients/mssql/connection_manager.go`

Startup sequence:
1. `sql.Open("sqlserver", connStr)` — registers the driver, no network call yet.
2. `db.PingContext(30s)` — actual TCP connect + login.
3. `GetConnectionStats()` — query `sys.configurations` and `sys.dm_exec_sessions`.
4. Check `AvailableConnections >= MinFreeConns` — fail fast if server is already saturated.
5. Configure pool: `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime(1h)`, `SetConnMaxIdleTime(15m)`.

The tool retries the entire `NewConnectionManager` call in a loop (5-second backoff) until success. This is useful when the load test pod starts before SQL Server is ready.

---

## KubeDB Client Builder

`clients/mssql/kubedb_client_builder.go`

This provides an alternative way to build a connection when running inside a Kubernetes cluster alongside a **KubeDB-managed MSSQLServer** CR. Instead of reading env vars, it:

1. Takes a `controller-runtime` `client.Client` and a `*dbapi.MSSQLServer` object.
2. Reads username/password from the K8s `Secret` referenced by `db.Spec.AuthSecret`.
3. Builds the DSN based on `db.Spec.TLS`:
   - `Spec.TLS == nil` → `encrypt=disable`
   - `Spec.TLS != nil` → `encrypt=true&TrustServerCertificate=true`
4. Opens `sql.DB`, pings, returns `*Client`.

Builder methods (fluent API):

```go
client, err := mssql.NewKubeDBClientBuilder(k8sClient, mssqlCR).
    WithPod("mssql-0").           // connect to a specific pod
    WithDatabase("mydb").         // default: "master"
    WithContext(ctx).
    GetMSSQLClient()
```

> `WithURL` can override the host entirely (useful for port-forwarding or external access).

---

## Building and Running

### Local Run (plain SQL Server)

**Start SQL Server with Docker:**
```bash
docker run -e "ACCEPT_EULA=Y" \
           -e "SA_PASSWORD=YourStrong@Passw0rd" \
           -p 1433:1433 \
           --name sqlserver \
           -d mcr.microsoft.com/mssql/server:2022-latest
```

**Create the test database:**
```bash
docker exec -it sqlserver /opt/mssql-tools/bin/sqlcmd \
  -S localhost -U sa -P 'YourStrong@Passw0rd' \
  -Q "CREATE DATABASE testdb"
```

**Run the load test:**
```bash
cd /path/to/mssql-load-test

export DB_HOST=localhost
export DB_PORT=1433
export DB_USER=sa
export DB_PASSWORD='YourStrong@Passw0rd'
export DB_NAME=testdb
export DB_ENCRYPT=disable

# Workload: 5 workers, 60 s, 70% insert / 30% update
export CONCURRENT_WRITERS=5
export TEST_RUN_DURATION=60
export INSERT_PERCENT=70
export UPDATE_PERCENT=30
export READ_PERCENT=0
export BATCH_SIZE=50
export SEED_DATA_ROWS=1000

go run main.go
```

### Build Binary

```bash
make build
# produces: bin/mssql-load-test
./bin/mssql-load-test
```

Or manually:
```bash
go build -o mssql-load-test .
```

### Docker Image

```bash
# Build
make docker-build
# or with custom tag:
make docker-build IMAGE_TAG=v1.0.0

# Push
make docker-push

# Run locally
docker run --rm \
  -e DB_HOST=host.docker.internal \
  -e DB_PASSWORD='YourStrong@Passw0rd' \
  -e DB_NAME=testdb \
  -e CONCURRENT_WRITERS=10 \
  -e TEST_RUN_DURATION=60 \
  -e INSERT_PERCENT=70 \
  -e UPDATE_PERCENT=30 \
  -e READ_PERCENT=0 \
  ghcr.io/neaj-morshad-101/mssql-load-test:latest
```

### Kubernetes Job

**1. Edit `k8s/01-configmap.yaml`** — set `DB_HOST` to your SQL Server service FQDN:
```yaml
DB_HOST: "mssql.demo.svc.cluster.local"
DB_PORT: "1433"
DB_NAME: "testdb"
```

**2. Edit `k8s/02-secret.yaml`** — set the actual SA password:
```yaml
stringData:
  DB_PASSWORD: "YourStrong@Passw0rd"
```

**3. Deploy:**
```bash
make deploy-k8s
# equivalent to:
kubectl apply -f k8s/
```

**4. Watch logs:**
```bash
kubectl logs -n mssql-load-test -l job-name=mssql-load-test -f
```

**5. Clean up:**
```bash
make cleanup-k8s
```

---

## Testing Scenarios

### Quick Smoke Test

Verify connectivity and basic functionality in under 30 seconds.

```bash
export DB_HOST=localhost
export DB_PASSWORD='YourStrong@Passw0rd'
export DB_NAME=testdb
export DB_ENCRYPT=disable
export CONCURRENT_WRITERS=2
export TEST_RUN_DURATION=20
export BATCH_SIZE=10
export SEED_DATA_ROWS=100
export INSERT_PERCENT=70
export UPDATE_PERCENT=30
export READ_PERCENT=0
go run main.go
```

Expected: no errors, data loss = 0, table dropped at end.

---

### Write-Heavy Benchmark

Measure max INSERT throughput.

```bash
export CONCURRENT_WRITERS=20
export TEST_RUN_DURATION=120
export BATCH_SIZE=200          # maximum allowed
export SEED_DATA_ROWS=10000
export INSERT_PERCENT=90
export UPDATE_PERCENT=10
export READ_PERCENT=0
export DB_MAX_OPEN_CONNS=50
go run main.go
```

Watch for:
- `Operations/sec` (Inserts/s line) during intervals.
- Error rate — should be 0.00% in steady state.
- P99 insert latency.

---

### Read-Heavy Benchmark

Simulate an OLTP read workload.

```bash
export CONCURRENT_WRITERS=30
export TEST_RUN_DURATION=120
export BATCH_SIZE=50
export SEED_DATA_ROWS=100000   # large dataset for realistic reads
export READ_PERCENT=70
export INSERT_PERCENT=20
export UPDATE_PERCENT=10
export READ_BATCH_SIZE=50
go run main.go
```

Watch for:
- `Reads/s` throughput.
- Avg / P95 / P99 read latency.

---

### High-Concurrency Stress Test

Test SQL Server connection handling under many clients.

```bash
export CONCURRENT_WRITERS=500
export TEST_RUN_DURATION=300
export BATCH_SIZE=100
export SEED_DATA_ROWS=50000
export INSERT_PERCENT=60
export UPDATE_PERCENT=20
export READ_PERCENT=20
export DB_MAX_OPEN_CONNS=200
export DB_MAX_IDLE_CONNS=50
export DB_MIN_FREE_CONNS=10
go run main.go
```

> **Note**: The tool prints a `WARNING: HIGH CONCURRENCY MODE` banner when `CONCURRENT_WRITERS >= 5000`. Ensure the SQL Server `max connections` setting can accommodate this.

Watch for:
- Connection pool exhaustion errors.
- Rising P99 latency.
- `Active` connection count in pool metrics.

---

### Chaos / Failover Test

This is the primary intended use-case — run the load test against a KubeDB HA SQL Server (AG or standalone) and inject failures mid-test.

**1. Start the load test as a Kubernetes Job** (see [Kubernetes Job](#kubernetes-job)).

**2. While it runs, inject chaos:**

```bash
# Kill the primary pod
kubectl delete pod mssql-0 -n demo

# Or use Chaos Mesh to simulate a network partition
kubectl apply -f chaos-network-partition.yaml
```

**3. After the test completes, check the output:**

```
Data Loss Report:
-----------------------------------------------------------------
  Total Rows Tracked (inserts + seed): 387450
  Records Found in DB:                 387450
  Records Lost:                        0
  Data Loss Percentage:                0.00%
```

A non-zero `Records Lost` value means SQL Server lost committed transactions during failover — this is what the tool is designed to detect.

---

## Understanding the Output

### Startup Output

```
=================================================================
SQL Server High Concurrency Load Testing Client v2
=================================================================

Configuration:
  Database: sa@localhost:1433/testdb
  Concurrent Workers: 10
  Test Duration: 5m0s
  Batch Size: 100 records (inserts), 10 records (reads)
  Workload: 20% Reads, 60% Inserts, 20% Updates
  Report Interval: 10s
```

### Periodic Metrics Report (every `REPORT_INTERVAL` seconds)

```
=================================================================
Test Duration: 30s
-----------------------------------------------------------------
Cumulative Statistics:
  Total Operations: 1240 (Reads: 248, Inserts: 744, Updates: 248)
  Total Errors: 0
  Total Data Transferred: 74.40 MB
-----------------------------------------------------------------
Current Throughput (interval):
  Operations/sec: 124.00 (Reads: 24.80/s, Inserts: 74.40/s, Updates: 24.80/s)
  Throughput: 7.44 MB/s
  Errors/sec: 0.00
-----------------------------------------------------------------
Latency Statistics:
  Reads   - Avg: 2.1ms,  P95: 8.3ms,  P99: 15.2ms
  Inserts - Avg: 18.4ms, P95: 45.6ms, P99: 78.9ms
  Updates - Avg: 3.2ms,  P95: 9.1ms,  P99: 16.4ms
-----------------------------------------------------------------
Connection Pool:
  Active: 847, Max: 32767, Available: 31920
=================================================================
```

### Final Summary

```
Performance Summary:
  Average Throughput: 118.33 operations/sec
  Read Operations: 5916 (23.66/sec avg)
  Insert Operations: 35496 (141.98/sec avg)
  Update Operations: 11832 (47.33/sec avg)
  Error Rate: 0.0000%
  Total Data Transferred: 2.13 GB
```

### Data Loss Report

```
Data Loss Report:
-----------------------------------------------------------------
  Total Rows Tracked (inserts + seed):  85496
  Records Found in DB:                  85496
  Records Lost:                         0
  Data Loss Percentage:                 0.00%
```

---

## Tuning Guidance

| Goal | Knob | Direction |
|------|------|-----------|
| Higher INSERT throughput | `BATCH_SIZE` ↑ (max 200), `CONCURRENT_WRITERS` ↑ | Increase |
| Reduce connection pool errors | `DB_MAX_OPEN_CONNS` ↑ | Increase |
| Reduce SQL Server connection exhaustion | `DB_MAX_OPEN_CONNS` ↓ | Decrease |
| Realistic OLTP read mix | `READ_PERCENT` = 60-80 | Adjust |
| Test AG failover detect speed | `TEST_RUN_DURATION` ↑, kill primary mid-test | Increase duration |
| Reduce memory in the load-test pod | `SEED_DATA_ROWS` ↓, `CONCURRENT_WRITERS` ↓ | Decrease |
| Avoid `max connections` exhaustion | `DB_MIN_FREE_CONNS` ↑ | Increase |

> **Rule of thumb**: Keep `CONCURRENT_WRITERS × BATCH_SIZE` proportional to `DB_MAX_OPEN_CONNS`. If workers block waiting for a free connection, throughput flatlines and latency spikes.

---

## SQL Server–Specific Notes

### Parameter Placeholders
`go-mssqldb` uses **`@p1`, `@p2`, ...`@pN`** (not `$1/$2` like Postgres or `?` like MySQL). The batch INSERT builds these dynamically.

### Parameter Limit
SQL Server enforces a **2100 parameter** hard limit per query. With 9 columns per row, maximum batch = **floor(2100/9) = 233**. The tool caps at **200** and validates on startup.

### `TOP` vs `LIMIT`
SQL Server uses `SELECT TOP N ...` or `SELECT TOP (@param) ...` — not `LIMIT`. All queries in this tool use the parameterized `TOP (@p1)` form.

### IDENTITY Columns
`BIGINT IDENTITY(1,1)` is the SQL Server equivalent of `SERIAL` / `BIGSERIAL` in Postgres. The `OUTPUT INSERTED.id` clause is the canonical way to retrieve auto-generated IDs from a multi-row INSERT.

### Idempotent DDL
SQL Server lacks `CREATE TABLE IF NOT EXISTS`. The tool uses:
```sql
IF OBJECT_ID(N'table_name', N'U') IS NULL CREATE TABLE ...
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name=... AND object_id=...) CREATE INDEX ...
```

### TLS
- `encrypt=disable` — plain text, suitable for in-cluster or dev.
- `encrypt=true` with `TrustServerCertificate=true` — encrypted but no cert chain validation (suitable for KubeDB-issued self-signed certs).
- `encrypt=true` with `TrustServerCertificate=false` — full chain validation (requires a trusted CA).

### `NEWID()` for Random Ordering
`ORDER BY NEWID()` (V1 `performUpdate`) is functionally equivalent to Postgres's `ORDER BY RANDOM()`. It forces a full table scan + sort, making it expensive on large tables — hence V2 uses an in-process `rand.Int63n(totalRows)` instead.

