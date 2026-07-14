---
title: Announcing KubeDB v2026.7.10
date: "2026-07-10"
weight: 15
authors:
- Saurov Chandra Biswas
tags:
- clickhouse
- cloud-native
- database
- gitops
- kubedb
- kubernetes
- milvus
- neo4j
- oracle
- postgres
- proxysql
- qdrant
- vector-database
- weaviate
---

KubeDB **v2026.7.10** delivers a broad set of quality-of-life improvements and new capabilities across the ecosystem. The headline feature is **In-Place Vertical Scaling** — databases can now resize CPU and memory without restarting pods, leveraging the Kubernetes In-Place Pod Resize API (GA in Kubernetes 1.35). This release also ships **extension-enabled PostgreSQL images** (pgvector, PostGIS, pg_repack, pg_cron, pgaudit, and pg_stat_statements out of the box) and a new **`synchronousReplicationConfig` API** for fine-grained control over PostgreSQL synchronous replication.

Other highlights include **ClickHouse PITR (Point-in-Time Recovery)** via the new `ClickHouseArchiver` CRD, **ClickHouse shard scaling**, **Oracle TLS reconfiguration**, **Weaviate native monitoring**, **Qdrant logical backup & restore**, **Milvus GitOps**, and a major expansion of GitOps support to 20+ additional databases.

---

## Key Highlights

* **In-Place Vertical Scaling** — resize container CPU/memory on live pods with zero restarts (Kubernetes 1.35+)
* **PostgreSQL Extensions** — new `-ext` PostgresVersions bundle pgvector, PostGIS, pg_repack, pg_cron, pgaudit, and pg_stat_statements
* **PostgreSQL Synchronous Replication Config** — new `spec.synchronousReplicationConfig` API with quorum (`Any`), priority (`First`), and wildcard modes
* **ClickHouse Archiver** — continuous WAL-level archiving and point-in-time recovery for ClickHouse
* **ClickHouse Shard Scaling** — add shards to a running ClickHouse cluster via HorizontalScaling OpsRequest
* **ClickHouse Alerting** — Alertmanager integration for ClickHouse metrics
* **Oracle ReconfigureTLS** — add, remove, and rotate TLS certificates for Oracle databases
* **Neo4j 2026.05.0** — new version support with backup-channel TLS
* **Weaviate Monitoring** — native Prometheus/Grafana monitoring with three dashboards and alerting
* **Qdrant Backup & Restore** — logical backup and restore via KubeStash
* **Milvus GitOps** — declarative management via the `gitops.kubedb.com` CRD group
* **ProxySQL Recommendation Engine** — automated maintenance recommendations
* **Expanded GitOps** — 20+ databases now supported including Memcached, PgBouncer, PgPool, Cassandra, Hazelcast, Zookeeper, Ignite, and more; GitOps StorageClass Migration added to all GitOps-enabled databases

---

## Common: In-Place Vertical Scaling

Vertical scaling (`VerticalScaling` OpsRequest) now supports an **`InPlace`** mode. Previously, any CPU or memory change evicted and restarted the target pods. With `mode: InPlace`, the operator resizes live containers through the Kubernetes `pods/resize` subresource — no eviction, no downtime.

A new `mode` field controls the behavior:

- `InPlace` — resizes running containers without restarting them.
- `Restart` — the previous behavior (evict and recreate pods). This remains the default when `mode` is omitted, so all existing OpsRequests continue to work unchanged.

If the kubelet rejects an in-place resize for a specific pod (for example, shrinking memory below current usage is `Infeasible`), the operator automatically falls back to a restart for that pod so the scaling always completes.

> **Requirement:** In-Place Pod Resize is GA in Kubernetes 1.35. On older clusters the field is accepted but silently falls back to a restart.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-scale-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: sample-postgres
  verticalScaling:
    mode: InPlace
    postgres:
      resources:
        requests:
          memory: "2Gi"
          cpu: "4"
        limits:
          memory: "2Gi"
          cpu: "4"
```

The same `mode` field is available on all database-specific `VerticalScaling` OpsRequests across the KubeDB ecosystem.

---

## PostgreSQL

### Extension-Enabled PostgresVersions

KubeDB now ships a set of **extension-enabled** PostgreSQL images — the `-ext` versions. These are the same official PostgreSQL images with a curated set of popular extensions compiled in, so you can enable them with a plain `CREATE EXTENSION` — no custom image build required.

Every `-ext` image bundles:

| Extension | `CREATE EXTENSION` name | Needs `shared_preload_libraries`? | What it does |
|---|---|---|---|
| pgvector | `vector` | No | Vector similarity search for embeddings |
| PostGIS | `postgis` | No | Geospatial types and functions |
| pg_repack | `pg_repack` | No | Rebuild tables/indexes without long locks |
| pg_cron | `pg_cron` | **Yes** | In-database cron job scheduler |
| pgaudit | `pgaudit` | **Yes** (recommended) | Session/object audit logging |
| pg_stat_statements | `pg_stat_statements` | **Yes** (already preloaded by KubeDB) | SQL execution statistics |

Available versions (Alpine and Debian Bookworm flavors):

```bash
$ kubectl get postgresversions | grep -E 'NAME|ext'
NAME                      VERSION   DISTRIBUTION   DB_IMAGE
16.13-bookworm-ext        16.13     KubeDB         ghcr.io/appscode-images/postgres:16.13-bookworm-ext
16.13-ext                 16.13     KubeDB         ghcr.io/appscode-images/postgres:16.13-alpine-ext
17.9-bookworm-ext         17.9      KubeDB         ghcr.io/appscode-images/postgres:17.9-bookworm-ext
17.9-ext                  17.9      KubeDB         ghcr.io/appscode-images/postgres:17.9-alpine-ext
18.3-bookworm-ext         18.3      KubeDB         ghcr.io/appscode-images/postgres:18.3-bookworm-ext
18.3-ext                  18.3      KubeDB         ghcr.io/appscode-images/postgres:18.3-alpine-ext
```

**Only `pg_cron` and `pgaudit` require `shared_preload_libraries`.** For pgvector, PostGIS, and pg_repack you can skip the config Secret entirely and just run `CREATE EXTENSION`.

#### Deploy with extensions

Create a config Secret to preload `pg_cron` and `pgaudit` (keep `pg_stat_statements`, which KubeDB preloads by default):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pg-extensions-config
  namespace: demo
stringData:
  user.conf: |-
    shared_preload_libraries='pg_stat_statements,pg_cron,pgaudit'
    cron.database_name='postgres'
```

Then deploy a Postgres using an `-ext` version:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-extensions
  namespace: demo
spec:
  version: "18.3-ext"
  replicas: 1
  configSecret:
    name: pg-extensions-config   # omit if you only need pgvector / PostGIS / pg_repack
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Enable extensions in `psql`:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_repack;
CREATE EXTENSION IF NOT EXISTS pg_cron;
CREATE EXTENSION IF NOT EXISTS pgaudit;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

> The `-ext` PostgresVersions ship by default from **KubeDB v2026.7.10**. On earlier releases you can create the `PostgresVersion` objects by hand — copy an existing same-version object and change `spec.db.image` to the `-ext` image.

---

### Synchronous Replication Config

A new `spec.synchronousReplicationConfig` API gives fine-grained control over PostgreSQL's `synchronous_standby_names` and `synchronous_commit` settings.

> **Availability:** `spec.synchronousReplicationConfig` requires **KubeDB v2026.7.10** and `postgres-init >= 0.20.0`. Verify your version: `kubectl get postgresversion <ver> -o jsonpath='{.spec.initContainer.image}'`

#### API reference

| Field | Type | Default | Description |
|---|---|---|---|
| `mode` | `Any` \| `First` | `Any` | `Any` = quorum (`ANY N`); `First` = priority order (`FIRST N`) |
| `numSyncReplicas` | integer | `1` | The `N` in `ANY N` / `FIRST N`; must be `>= 1` and `< spec.replicas` |
| `commitLevel` | `On` \| `RemoteApply` \| `RemoteWrite` \| `Local` \| `Off` | `RemoteWrite` | Maps to `synchronous_commit` |
| `standbyNames` | `[]string` | auto (all pods) | Explicit ordered standby `application_name`s; mutually exclusive with `useWildcard` |
| `useWildcard` | bool | `false` | Use `*` to match any connected standby; mutually exclusive with `standbyNames` |

> **YAML gotcha:** `On` and `Off` are parsed as booleans. Always quote them: `commitLevel: "On"`.

#### Example 1 — Quorum (Any)

Wait for **any 2** standbys to acknowledge each commit:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sync-postgres
  namespace: demo
spec:
  version: "17.4"
  replicas: 3
  standbyMode: Hot
  streamingMode: Synchronous
  synchronousReplicationConfig:
    mode: Any
    numSyncReplicas: 2
    commitLevel: RemoteWrite
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

After the cluster is `Ready`:

```bash
$ kubectl exec -n demo sync-postgres-0 -c postgres -- \
    psql -U postgres -tAc "SHOW synchronous_standby_names;"
ANY 2 ("sync-postgres-1","sync-postgres-2")

$ kubectl exec -n demo sync-postgres-0 -c postgres -- \
    psql -U postgres -c "select application_name, state, sync_state from pg_stat_replication order by 1;"
 application_name |   state   | sync_state
------------------+-----------+------------
 sync-postgres-1  | streaming | quorum
 sync-postgres-2  | streaming | quorum
```

#### Example 2 — Priority order (First + standbyNames)

Make one standby the preferred synchronous replica:

```yaml
spec:
  streamingMode: Synchronous
  synchronousReplicationConfig:
    mode: First
    numSyncReplicas: 1
    commitLevel: "On"          # quoted — see YAML gotcha above
    standbyNames:
    - sync-postgres-2          # highest priority
    - sync-postgres-1
```

```bash
$ kubectl exec -n demo sync-postgres-0 -c postgres -- \
    psql -U postgres -tAc "SHOW synchronous_standby_names;"
FIRST 1 ("sync-postgres-2","sync-postgres-1")

# sync-postgres-2 is active sync; sync-postgres-1 is potential (promoted if -2 fails)
```

#### Example 3 — Wildcard

Accept any connected standby without naming them:

```yaml
spec:
  streamingMode: Synchronous
  synchronousReplicationConfig:
    mode: Any
    numSyncReplicas: 1
    useWildcard: true
```

```bash
$ kubectl exec -n demo sync-postgres-0 -c postgres -- \
    psql -U postgres -tAc "SHOW synchronous_standby_names;"
ANY 1 (*)
```

#### `commitLevel` reference

| `commitLevel` | PostgreSQL `synchronous_commit` | Meaning |
|---|---|---|
| `Off` | `off` | Returns without waiting for local WAL flush. Fastest, least durable. |
| `Local` | `local` | Waits for local WAL flush only. |
| `RemoteWrite` | `remote_write` | **(default)** Waits until a standby has written WAL to its OS buffer. |
| `On` | `on` | Waits until a standby has flushed WAL to disk. |
| `RemoteApply` | `remote_apply` | Waits until a standby has applied WAL. Highest latency. |

---

## ClickHouse

### ClickHouse Archiver (PITR)

KubeDB now supports **continuous archiving and point-in-time recovery (PITR)** for ClickHouse via the new `ClickHouseArchiver` CRD. You need KubeStash installed in your cluster.

Create a `ClickHouseArchiver` with a reference to a `BackupStorage`, `RetentionPolicy`, and an encryption secret:

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: ClickHouseArchiver
metadata:
  name: clickhouse-archiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: demo-retention
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "ClickHouseBackup"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/30 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "s3-storage"
      namespace: "demo"
```

Reference the archiver from a ClickHouse cluster via `spec.archiver`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: ch
  namespace: demo
spec:
  version: 25.7.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      name: appscode-cluster
      shards: 2
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  deletionPolicy: WipeOut
  archiver:
    ref:
      name: clickhouse-archiver-sample
      namespace: demo
```

To restore to a specific point in time, create a new ClickHouse with `spec.init.archiver`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-restored
  namespace: demo
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: ch-full
        namespace: demo
      recoveryTimestamp: "2026-07-02T06:38:42Z"
  version: 25.7.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      name: appscode-cluster
      shards: 2
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  deletionPolicy: WipeOut
```

### Shard Scaling

ClickHouse clusters can now be scaled horizontally by adding shards via a `HorizontalScaling` OpsRequest:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-shard-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: clickhouse-prod
  horizontalScaling:
    shards: 3
```

### Alerting

KubeDB now supports Alertmanager-based alerting for ClickHouse. Users can configure alert rules to receive notifications when a ClickHouse metric exceeds a configured threshold. This integrates with the standard KubeDB monitoring stack.

---

## Oracle

### ReconfigureTLS

This release adds TLS reconfiguration support for Oracle via `OracleOpsRequest`.

**Add or update TLS** (using a cert-manager Issuer):

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: reconfigure-tls-dg-sample
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: dg-sample
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: oracle-ca-issuer
    certificates:
      - alias: server
      - alias: client
      - alias: metrics-exporter
  timeout: 30m
```

**Remove TLS:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: remove-tls-dg-sample
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: dg-sample
  tls:
    remove: true
```

**Rotate certificates:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: rotate-tls-dg-sample
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: dg-sample
  tls:
    rotateCertificates: true
```

---

## Neo4j

### Version 2026.05.0 and Backup TLS

KubeDB now supports Neo4j version **2026.05.0**. This release also introduces **TLS-encrypted backup channels** via the `spec.tls.backup.mode` field, allowing you to secure backup traffic independently of bolt/HTTP TLS:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  version: 2026.05.0
  replicas: 3
  storageType: Durable
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-ca-issuer
    backup:
      mode: TLS
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

---

## Weaviate

### Native Monitoring

KubeDB now supports Prometheus Operator-based monitoring for Weaviate. Weaviate exposes native Prometheus metrics from its built-in metrics endpoint, and KubeDB wires that endpoint into the standard monitoring stack.

When monitoring is enabled:
- Weaviate Prometheus metrics are exposed inside the database container
- A stats service is created for scraping
- A `ServiceMonitor` is created for Prometheus Operator
- KubeDB summary metrics are exposed via Panopticon
- Three Grafana dashboards are available: **KubeDB / Weaviate / Summary**, **KubeDB / Weaviate / Pod**, and **KubeDB / Weaviate / Database**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 2112
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

Install the monitoring stack:

```bash
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace \
  --set grafana.image.tag=7.5.5

helm upgrade -i kubedb-metrics oci://ghcr.io/appscode-charts/kubedb-metrics \
  --version v2026.7.10 \
  -n kubedb --create-namespace \
  --set featureGates.Weaviate=true
```

Alerting is also available through the `weaviate-alerts` chart. The default alert groups cover scrape failures, pod restarts, CPU/memory pressure, persistent volume usage, HTTP/gRPC error rate, p95 latency, and replication signals.

---

## Qdrant

### Logical Backup & Restore

KubeDB now supports logical backup and restore for Qdrant via KubeStash, with support for selective collection backup.

**BackupConfiguration:**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: qdrant-sample-backup
spec:
  target:
    kind: Qdrant
    name: qdrant-sample
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
      addon:
        name: qdrant-addon
        tasks:
          - name: logical-backup
            params:
              collections: "demo_collection"
```

**RestoreSession:**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-qdrant-sample
spec:
  target:
    kind: Qdrant
    name: qdrant-sample-restore
  dataSource:
    repository: minio-qdrant-repo
    snapshot: latest
  addon:
    name: qdrant-addon
    tasks:
      - name: logical-backup-restore
```

---

## Milvus

### GitOps Support

Milvus can now be provisioned and managed declaratively through the `gitops.kubedb.com` CRD group. Commit changes to your Milvus spec and the GitOps operator automatically translates them into the correct `MilvusOpsRequest`:

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
    distributed:
      streamingnode:
        storageType: Durable
        storage:
          accessModes:
            - ReadWriteOnce
          storageClassName: longhorn-custom
          resources:
            requests:
              storage: 1Gi
```

---

## ProxySQL

### Recommendation Engine

ProxySQL is now supported by the KubeDB **Recommendation Engine**. Once a ProxySQL instance is `Ready`, the Ops-manager watches it for version updates, TLS certificate expiry, and auth secret rotation, and creates a `Recommendation` CR whenever maintenance is due. The Supervisor then executes it — immediately, on approval, or inside a configured maintenance window.

---

## Expanded GitOps Support

This release greatly expands the set of databases that support declarative GitOps management via the `gitops.kubedb.com` CRD group. GitOps is now available for:

**Newly added in v2026.7.10:** Memcached, PgBouncer, PgPool, ProxySQL, Cassandra, HanaDB, Hazelcast, Ignite, Zookeeper — in addition to the databases that already had GitOps support.

**GitOps StorageClass Migration** is also now available for all GitOps-enabled databases, allowing storage class changes to be driven declaratively through Git commits.

---

## Support

- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
