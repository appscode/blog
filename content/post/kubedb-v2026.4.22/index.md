---
title: Announcing KubeDB v2026.4.22
date: "2026-04-22"
weight: 15
authors:
- Saurov Chandra Biswas
tags:
- ai
- cloud-native
- configuration
- database
- elasticsearch
- gitops
- hanadb
- kubedb
- kubernetes
- mariadb
- milvus
- mysql
- neo4j
- oracle
- postgres
- qdrant
- redis
- vector-database
- weaviate
---

KubeDB **v2026.4.22** focuses on **horizontal scalability, operational reliability, and database ecosystem expansion**. This release introduces sharding support for the KubeDB Autoscaler operator, new OpsRequest capabilities for Neo4j, Qdrant, and other databases, and enhanced GitOps workflows.

This release brings **SQL Server vertical scaling** for additional components, **Oracle custom configuration and initialization**, **DocumentDB standalone support**, **ClickHouse backup/restore**, and improved monitoring for **Milvus** and **HanaDB**.

---

## Key Highlights

* Autoscaler Operator Sharding for horizontal scalability
* Neo4j OpsRequests (Reconfigure, HorizontalScaling, VerticalScaling, VolumeExpansion, UpdateVersion, RotateAuth)
* Qdrant OpsRequests support
* SQL Server vertical scaling for Arbiter, Coordinator, Exporter
* Oracle custom configuration and initialization
* DocumentDB standalone deployment
* ClickHouse backup and restore with KubeStash
* Enhanced GitOps workflows
* Milvus native monitoring
* HanaDB monitoring support with Prometheus

---

## Common Improvements

### Autoscaler Operator Sharding

In this release, we are introducing **sharding support for the KubeDB Autoscaler operator**, enabling horizontal scalability for autoscaler resources across all supported databases. This enhancement ensures that both the autoscaler controllers and the recommender component can efficiently distribute workload across multiple operator pods.

**Key Features:**

- Supports all autoscaler CRDs (autoscaling.kubedb.com)
- No changes required to existing Autoscaler resources
- Even distribution using consistent hashing
- Autoscaler Recommender is now shard-aware (no duplicate processing)

**How It Works**

The operator-shard-manager:
- Labels each autoscaler resource with a shard index: `shard.operator.k8s.appscode.com/kubedb-autoscaler: "0"`
- Each autoscaler operator pod:
  - Detects its assigned shard index
  - Processes only the resources belonging to that shard

**ShardConfiguration for Autoscaler**

```yaml
apiVersion: operator.k8s.appscode.com/v1alpha1
kind: ShardConfiguration
metadata:
  name: kubedb-autoscaler
spec:
  controllers:
  - apiGroup: apps
    kind: StatefulSet
    name: kubedb-autoscaler
    namespace: kubedb
  resources:
  - apiGroup: autoscaling.kubedb.com
```

To enable autoscaler sharding while installing KubeDB:

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2026.4.27 \
  --namespace kubedb --create-namespace \
  --set operator-shard-manager.enabled=true \
  --set kubedb-autoscaler.replicaCount=3 \
  --set-file global.license=license.txt \
  --wait --burst-limit=10000 --debug
```

---

### GitOps Improvements

This release improves GitOps workflows by making configuration updates, TLS reconfiguration, authentication rotation, and horizontal scaling more reliable across databases.

**Configuration Updates:**
Improved generation and deduplication of Reconfigure OpsRequests by enhancing config comparison for PostgreSQL and Redis/Valkey.

**TLS Reconfiguration Enhancements:**
Extended TLS reconfiguration to detect database-specific field changes:
- Elasticsearch, MariaDB, MySQL: Changes to `RequireSSL` now trigger TLS reconfigure OpsRequests
- PostgreSQL: Changes to `ClientAuthMode` now trigger a Restart OpsRequest

**Horizontal Scaling:**
Fixed horizontal scaling logic for Redis.

**Rotate Auth:**
GitOps-based rotate authentication is now supported for Redis.

This update makes GitOps-driven database management safer, more predictable, and operationally reliable.

---

## SQL Server

### Vertical Scaling for Arbiter, Coordinator, and Exporter

In this release, we have extended SQL Server vertical scaling support for additional components.

Now, you can also configure resources for:
- Arbiter
- Exporter
- Coordinator

You can now define resource requests and limits for these components directly in MSSQLServerOpsRequest:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: ag-cluster
  verticalScaling:
    mssqlserver:
      resources:
        requests:
          memory: "4Gi"
          cpu: 2
        limits:
          memory: "4Gi"
    coordinator:
      resources:
        limits:
          memory: 512Mi
        requests:
          cpu: 500m
          memory: 512Mi
    arbiter:
      resources:
        limits:
          memory: 512Mi
        requests:
          cpu: 500m
          memory: 512Mi
  timeout: 2m
  apply: IfReady
```

This enhancement completes SQL Server vertical scaling support by allowing all components to be scaled, not just the database container.

---

## MariaDB

### New MariaDB bin-Log Retention Feature

We're excited to announce a new bin-log retention management feature in our MariaDB operator, designed to give users more control over their binlog storage.

**Key Highlights:**

- **Automatic bin-log Deletion**: Previous bin-log files can now be automatically deleted based on user-defined retention policies. This process runs through the sidekick pod.

- **Flexible Retention Policies**: In `MariaDBArchiver.spec.logBackup`, you can now define:
  - `logRetentionHistoryLimit`: Number of retention stats to keep (e.g., 5)
  - `retentionPeriod`: Duration after which old bin-log files are deleted (e.g., 10d)
  - `retentionSchedule`: Cron-style schedule for running the deletion process (e.g., "*/15 * * * *")

**Defaults:**
- retentionPeriod: 1 year
- retentionSchedule: every 3 months

- **Incremental Snapshot Support**: The system maintains the last logRetentionHistoryLimit retention stats in your incremental snapshots. binlog files older than the retentionPeriod will be automatically removed.

> If you change any of these retention fields, a sidekick pod restart is required for the new configuration to take effect.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MariaDBArchiver
metadata:
  name: mariadbarchiver-sample
  namespace: aws-demo14
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: aws-demo14
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: mariadb-retention-policy
    namespace: aws-demo14
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "aws-demo14"
  fullBackup:
    driver: "VolumeSnapshotter"
    task:
      params:
        volumeSnapshotClassName: "longhorn-snapshot-vsc"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/3 * * * *"
    sessionHistoryLimit: 2
  logBackup:
    logRetentionHistoryLimit: 5
    retentionPeriod: 10d
    retentionSchedule: "*/2 * * * *"
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/3 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "s3-storage3"
      namespace: "backupstorage"
```

---

## MySQL

In this release, we have significantly improved MySQL Group Replication support. The default value for `group_replication_unreachable_majority_timeout` has been set to 20 seconds.

**Key enhancements include:**

- Fixed coordinator bugs to make it more resilient under various failure scenarios
- Added automatic healing capabilities
- Improved split-brain resolution to ensure cluster stability
- Fixed an issue where the coordinator would incorrectly bootstrap a new group when an existing group was already running
- Ensured no data loss during recovery and group operations
- Fixed behavior related to write availability on nodes after failures

---

## Neo4j

KubeDB now supports day-2 operations for Neo4j via OpsRequest CRDs. All operations perform rolling updates with zero downtime and respect `apply: IfReady` semantics — operations proceed only when the cluster is in a healthy state.

---

### OpsRequest: Reconfigure

Apply custom `neo4j.conf` settings to a running cluster via rolling restart. Supports `applyConfig` inline or via a referenced `ConfigSecret`. Use `removeCustomConfig: true` to revert to defaults.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: neo4j
  configuration:
    configSecret:
      name: new-custom-config
    removeCustomConfig: true
    applyConfig:
      server.metrics.csv.interval: "40s"
  timeout: 5m
  apply: IfReady
```

---

### OpsRequest: HorizontalScaling

Add or remove Neo4j server nodes from a running cluster. After scaling, KubeDB automatically runs the database reallocation command to redistribute hosted databases across the updated node set.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neoops-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: neo4j
  horizontalScaling:
    server: 5
    reallocate:
      strategy: "full"
      batchSize: 1
```

---

### OpsRequest: VerticalScaling

Update CPU and memory requests/limits for Neo4j server containers. Applied without a rolling restart — resource changes take effect in-place without disrupting the running cluster.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: neo4j
  verticalScaling:
    server:
      resources:
        limits:
          cpu: 1500m
          memory: 4Gi
        requests:
          cpu: 700m
          memory: 4Gi
```

---

### OpsRequest: VolumeExpansion

Expand PVC storage size for Neo4j server nodes. Supports both `Online` (no downtime) and `Offline` (rolling restart) modes depending on the StorageClass. KubeDB patches the PVC and waits for the filesystem resize to complete before proceeding.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-volumeexpansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: neo4j
  volumeExpansion:
    mode: "Offline"
    server: 4Gi
---
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-volumeexpansiononline
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: neo4j
  volumeExpansion:
    mode: "Online"
    server: 6Gi
```

---

### OpsRequest: UpdateVersion

Upgrade Neo4j to a newer version via a rolling restart. KubeDB validates the upgrade path, updates the container image per node, and ensures the cluster reaches a healthy state before continuing to the next pod.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: neo4j
  updateVersion:
    targetVersion: 2025.11.2
```

---

### OpsRequest: RotateAuth

Rotate Neo4j credentials by referencing an updated auth secret. Applied without a rolling restart — KubeDB propagates the new credentials to all servers in-place. Supports providing an external secret via `authentication.secretRef`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neoops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: neo4j
  authentication:
    secretRef:
      kind: Secret
      name: external-neo4j-auth
  timeout: 5m
  apply: IfReady
```

---

## Redis

In this release we have fixed the `only ACL apply` issue which causes restart while updating the ACL in Redis.

Fixed cluster join issue after pod restart.

---

## Pgpool

We’ve added Pgpool load balancing support for PostgreSQL read replicas. You can now configure load balancing via dedicated API fields, making it straightforward to route and distribute read traffic.

For better understanding [read](https://appscode.com/blog/post/kubedb-v2026.2.26/#postgresql).

example:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: quick-pgpool
  namespace: pool
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  sslMode: disable
  clientAuthMode: md5
  syncUsers: true
  deletionPolicy: WipeOut
  configuration:
    backends:
groupName: reporting
weight: 2
groupName: user-facing
weight: 1
groupName: analytics
weight: 5
```
If you want to configure backends, run PgpoolOpsRequest. For example:
```yaml
spec:
  configuration:
    backend:
      sync:
        - name: demo
          weight: 2
        - name: demo2
          weight: 10
        - name: STANDBY
          weight: 8
      delete:
        - PRIMARY
    removeCustomConfig: true
    configSecret:
      name: pp-custom-config
    applyConfig:
      pgpool.conf: |-
        max_pool = 72
        num_init_children = 9
```


---

## Qdrant

Various OpsRequests support has been added in this release.

### Horizontal Scaling

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-hor-scaling
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: qdrant-sample
  horizontalScaling:
    node: 3
```

### Reconfigure

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: qdrant-sample
  configuration:
    applyConfig:
      config.yaml: |
        log_level: INFO
  timeout: 5m
  apply: IfReady
```

### Restart

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: qdrant-sample
  timeout: 5m
  apply: Always
```

### Rotate Auth

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: qdrant-sample
  authentication:
    secretRef:
      kind: Secret
      name: qdrant-rotate-auth
  timeout: 5m
  apply: IfReady
```

### Update Version

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: qdrant-sample
  updateVersion:
    targetVersion: 1.17.0
  timeout: 5m
  apply: IfReady
```

### Vertical Scaling

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-vertical-scaling
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: qdrant-sample
  verticalScaling:
    node:
      resources:
        requests:
          memory: "500Mi"
          cpu: "500m"
        limits:
          memory: "500Mi"
          cpu: "500m"
  timeout: 5m
  apply: IfReady
```

### Volume Expansion

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-volume-exp
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: qdrant-sample
  volumeExpansion:
    node: 2Gi
    mode: Offline
```

---

## HanaDB

### Monitoring HanaDB

KubeDB supports Prometheus operator based monitoring for HanaDB deployments.

When monitoring is enabled:
- A stats service is created for HanaDB exporter metrics
- A ServiceMonitor is created for Prometheus Operator
- HanaDB internal metrics are exposed by the HanaDB exporter
- KubeDB summary metrics are exposed through Panopticon

### Enable Monitoring in HanaDB

Add the following monitor section in your HanaDB spec:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hana-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: Durable
  topology:
    mode: SystemReplication
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9668
      serviceMonitor:
        labels:
          release: prometheus
        interval: 30s
```

This enables:
- HanaDB exporter metrics on port 9668
- Prometheus scraping through a generated ServiceMonitor

### Install Monitoring Stack

**Install Prometheus Operator**

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace \
  --set grafana.image.tag=7.5.5
```

**Install Panopticon**

Panopticon is required for KubeDB summary metrics.

```bash
helm upgrade -i panopticon oci://ghcr.io/appscode-charts/panopticon \
  --version v202.9.30 \
  -n monitoring --create-namespace --wait \
  --set-file license=/path/to/license-file.txt \
  --set monitoring.serviceMonitor.labels.release=prometheus
```

**Install KubeDB Metrics Configurations**

```bash
helm upgrade -i kubedb-metrics oci://ghcr.io/appscode-charts/kubedb-metrics \
  --version v2026.4.13 \
  -n kubedb --create-namespace \
  --set featureGates.HanaDB=true
```

### Grafana Dashboards

There are two dashboards for HanaDB:
- KubeDB / HanaDB / Summary
- SAP HANA

### Alerting

KubeDB also provides HanaDB alert rules through the hanadb-alerts chart.

```bash
helm upgrade -i hana-cluster oci://ghcr.io/appscode-charts/hanadb-alerts \
  --version v2026.2.24 \
  -n demo \
  --set form.alert.labels.release=prometheus
```

---

## Milvus

KubeDB now supports native monitoring for Milvus deployments using Prometheus and Grafana. Metrics are exposed directly from Milvus and seamlessly integrated with the KubeDB monitoring stack.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: kubedb
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
    distributed:
      proxy:
        replicas: 2
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9091
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 500m
            memory: 256Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
              - ALL
          runAsGroup: 1000
          runAsNonRoot: true
          runAsUser: 1000
          seccompProfile:
            type: RuntimeDefault
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    storageClassName: local-path
    resources:
      requests:
        storage: 10Gi
```

---

## Oracle

This release adds two Oracle database features for KubeDB:

1. **Custom database configuration** through `configuration`
2. **Database initialization** through `init`

### 1) Custom Oracle configuration

You can now provide Oracle configuration through a **Secret**, **inline configuration**, or **both at the same time**.

Use the file name `oracle.cnf` for both the custom configuration file and the inline configuration entry.

**Create a custom configuration file**

```bash
# Create an oracle.cnf file with the following content:
PROCESSES = 800
```

**Create a Secret from the configuration file**

```bash
kubectl create secret generic -n demo oracle-custom-config --from-file=./oracle.cnf
```

**Create an Oracle database with custom configuration**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: standalone-inline-conf-can
  namespace: demo
spec:
  configuration:
    secretName: oracle-custom-config
    inline:
      oracle.cnf: |
        SGA_TARGET=5G
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

### 2) Oracle initialization with InitConfig

This release also adds support for Oracle database initialization using `InitConfig`.

> **Important:** The file stored in the init ConfigMap or init Secret must be named `setup.sql`.

**Create an init ConfigMap**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oracle-init-script
  namespace: demo
data:
  setup.sql: |
    CREATE USER test IDENTIFIED BY test;
    GRANT CONNECT, RESOURCE TO test;

    CREATE TABLE test.demo (
      id NUMBER,
      name VARCHAR2(50)
    );
```

**Create an Oracle database with init config**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-standalone-init-config
  namespace: demo
spec:
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  init:
    script:
      configMap:
        name: oracle-init-script
  deletionPolicy: WipeOut
```

---

## DocumentDB

KubeDB now supports standalone deployments for DocumentDB with built-in health check capabilities. The health check mechanism continuously verifies the status of the DocumentDB instance, ensuring it is running and responsive.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb
  namespace: demo
spec:
  version: 'pg17-0.109.0'
  storageType: Durable
  deletionPolicy: Delete
  replicas: 1
  podTemplate:
    spec:
      containers:
        - name: documentdb
          resources:
            requests:
              cpu: 500m
              memory: 2Gi
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

---

## ClickHouse

KubeDB now supports backup and restore for ClickHouse databases using KubeStash. This enables reliable data protection with cloud storage backends (e.g., S3, Azure).

**BackupConfiguration:**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-clickhouse-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    namespace: demo
    name: ch
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      backupTimeout: 300s
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-clickhouse-repo
          backend: s3-backend
          directory: /clickhouse
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: clickhouse-addon
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 101
              runAsGroup: 101
              fsGroup: 101
        tasks:
          - name: logical-backup
```

**RestoreSession:**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-clickhouse
  namespace: demo
spec:
  restoreTimeout: 300s
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    namespace: demo
    name: ch2
  dataSource:
    repository: s3-clickhouse-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: clickhouse-addon
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 101
          runAsGroup: 101
          fsGroup: 101
    tasks:
      - name: logical-backup-restore
```

---

## Support

- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).