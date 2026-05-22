# SQL Server High Load Test Client
A high-concurrency load testing client for **Microsoft SQL Server**.

## Features
- **Mixed workload**: Configurable % of Reads, Inserts, and Updates
- **Concurrent workers**: Supports from a few to thousands of goroutines
- **Batch inserts**: Uses SQL Server `OUTPUT INSERTED.id` for efficient bulk inserts with ID tracking
- **Data loss detection**: After the test, verifies rows are actually in the database
- **Connection monitoring**: Tracks active sessions via `sys.dm_exec_sessions`
- **Latency percentiles**: Reports Avg, P95, P99 latency per operation type
- **Kubernetes ready**: Includes Job / ConfigMap / Secret manifests
## SQL Server Specific Notes
| Feature | PostgreSQL | SQL Server |
|---|---|---|
| Auto-increment | `BIGSERIAL` | `BIGINT IDENTITY(1,1)` |
| Get inserted IDs | `RETURNING id` | `OUTPUT INSERTED.id` |
| Param placeholder | `$1, $2, ...` | `@p1, @p2, ...` |
| Random order | `ORDER BY RANDOM()` | `ORDER BY NEWID()` |
| Result limit | `LIMIT n` | `SELECT TOP n` |
| Datetime | `TIMESTAMP` / `NOW()` | `DATETIME2` / `SYSDATETIME()` |
| Text type | `TEXT` | `NVARCHAR(MAX)` |
| Drop table | `DROP TABLE IF EXISTS` | `IF OBJECT_ID(...) IS NOT NULL DROP TABLE` |
| Max connections | `max_connections` (PostgreSQL config) | `sys.configurations WHERE name='max connections'` |
| Active sessions | `pg_stat_activity` | `sys.dm_exec_sessions` |
| Parameter limit | 65535 | 2100 → max batch = 200 rows (9 cols × 200 = 1800 params) |
## Configuration
| Env Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | SQL Server host |
| `DB_PORT` | `1433` | SQL Server port |
| `DB_USER` | `sa` | Database user |
| `DB_PASSWORD` | `` | Database password |
| `DB_NAME` | `testdb` | Database name |
| `DB_ENCRYPT` | `disable` | Encryption: `disable`, `false`, `true` |
| `DB_TRUST_SERVER_CERT` | `true` | Trust server certificate |
| `DB_MAX_OPEN_CONNS` | `50` | Max open connections in pool |
| `DB_MAX_IDLE_CONNS` | `10` | Max idle connections in pool |
| `DB_MIN_FREE_CONNS` | `5` | Minimum free connections required |
| `CONCURRENT_WRITERS` | `10` | Number of concurrent goroutines |
| `TEST_RUN_DURATION` | `300` | Test duration in seconds |
| `BATCH_SIZE` | `100` | Records per batch insert (max 200) |
| `REPORT_INTERVAL` | `10` | Metrics report interval in seconds |
| `READ_PERCENT` | `0` | % of read operations |
| `INSERT_PERCENT` | `70` | % of insert operations |
| `UPDATE_PERCENT` | `30` | % of update operations |
| `TABLE_NAME` | `load_test_data` | Table name for load testing |
| `READ_BATCH_SIZE` | `10` | Records fetched per read |
| `SEED_DATA_ROWS` | `50000` | Seed rows to pre-populate |
> **Note:** `READ_PERCENT + INSERT_PERCENT + UPDATE_PERCENT` must equal 100.
> **Note:** `BATCH_SIZE` must be ≤ 200 due to SQL Server's 2100 parameter limit.
## Quick Start
### Local
```bash
export DB_HOST=localhost
export DB_PORT=1433
export DB_USER=sa
export DB_PASSWORD='YourStrong@Passw0rd'
export DB_NAME=testdb
export CONCURRENT_WRITERS=20
export TEST_RUN_DURATION=60
export READ_PERCENT=20
export INSERT_PERCENT=60
export UPDATE_PERCENT=20
make run
```
### Docker
```bash
make docker-build
docker run --rm \
  -e DB_HOST=host.docker.internal \
  -e DB_PASSWORD='YourStrong@Passw0rd' \
  mssql-load-test:latest
```
### Kubernetes
```bash
# Edit k8s/01-configmap.yaml and k8s/02-secret.yaml with your SQL Server details
make deploy-k8s
kubectl logs -f -n mssql-load-test job/mssql-load-test
```
## Architecture
```
mssql-load-test/
├── main.go                      # Entry point (v2: read + write)
├── config/
│   └── config.go                # Environment-based configuration
├── clients/mssql/
│   ├── client.go                # DB client type
│   ├── connection_manager.go    # Connection pool + SQL Server stats
│   ├── load_generator.go        # Write-only load generator (v1)
│   └── load_generator_v2.go     # Mixed read/write load generator (v2)
├── metrics/
│   ├── metrics.go               # Write-only metrics
│   └── metrics_v2.go            # Read+Write metrics with percentiles
└── k8s/                         # Kubernetes manifests
```
## Data Loss Detection
After the test completes, the client:
1. Compares `totalRows` counter (tracked in memory) against `SELECT COUNT(*) FROM table`
2. Reports the difference as potential data loss
3. This is useful for chaos engineering to detect if SQL Server failover/crash lost data
## Chaos Testing Use Case
This client is designed for use with **chaos experiments** (e.g., chaos-mesh):
- Run this client continuously against a SQL Server HA deployment
- Trigger chaos experiments (pod kill, network partition, etc.)
- After the test, check if any data was lost during the failover
