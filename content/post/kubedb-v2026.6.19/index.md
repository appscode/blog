---
title: Announcing KubeDB v2026.6.19
date: "2026-06-19"
weight: 15
authors:
- Saurov Chandra Biswas
tags:
- aerospike
- clickhouse
- cloud-native
- database
- documentdb
- druid
- elasticsearch
- gitops
- hanadb
- ignite
- kafka
- kubedb
- kubernetes
- mariadb
- migration
- milvus
- mongodb
- mssql
- mysql
- neo4j
- oracle
- percona-xtradb
- postgres
- qdrant
- rabbitmq
- redis
- singlestore
- solr
- vector-database
- weaviate
---

KubeDB **v2026.6.19** is a landmark release focused on **live database migration**, **broad StorageClass migration support**, **expanded OpsRequest coverage**, and **GitOps-first database management** across the entire ecosystem. This release introduces **network-level security via Cilium**, adds **live migration tools for MySQL, MariaDB, and MongoDB**, and delivers **DocumentDB high-availability clustering** along with a full suite of OpsRequests and autoscaling.

Notable additions include **full OpsRequest + TLS + Autoscaler support for Weaviate and Milvus**, **TLS and day-2 operations for HanaDB**, **Oracle OpsRequests and autoscaling**, and **Git-Sync initialization** across six databases. GitOps support lands for ClickHouse, Solr, SingleStore, PerconaXtraDB, Neo4j, RabbitMQ, and Druid; the KubeDB Recommendation Engine now covers Solr, SingleStore, PerconaXtraDB, Neo4j, Milvus, and RabbitMQ.

---

## Key Highlights

* **Cilium NetworkPolicy** — generic, database-agnostic network security for all KubeDB-managed databases
* **Live Database Migration** — MySQL, MariaDB, and MongoDB migration tools with CDC streaming and minimal downtime
* **StorageClass Migration** — 15+ databases now support seamless PVC storage class migration
* **DocumentDB HA Clustering** — multi-replica high-availability with Raft-based failover, full OpsRequests, and autoscaling
* **Weaviate** — full OpsRequest suite (Restart, Reconfigure, HScale, VScale, VolumeExpansion, ReconfigureTLS, RotateAuth, StorageMigration) plus TLS and Autoscaler
* **Milvus** — TLS (external mTLS + internal TLS), full OpsRequests, Autoscaler, and Recommendations
* **Neo4j** — Online Backup & Restore via KubeStash, StorageMigration, Git-Sync, GitOps, Autoscaling, and Recommendations
* **Oracle** — full OpsRequest coverage (Restart, Reconfigure, VerticalScaling, RotateAuth, VolumeExpansion) plus Autoscaler
* **HanaDB** — TLS support and day-2 operations (VerticalScaling, Restart, ReconfigureTLS, RotateAuth, VolumeExpansion, StorageMigration)
* **Git-Sync initialization** — ClickHouse, SingleStore, PerconaXtraDB, Neo4j, and Elasticsearch
* **GitOps support** — ClickHouse, Solr, SingleStore, PerconaXtraDB, Neo4j, RabbitMQ, and Druid
* **Recommendation Engine** — Solr, SingleStore, PerconaXtraDB, Neo4j, Milvus, and RabbitMQ

---

## Common Improvements

### Cilium NetworkPolicy Support

KubeDB v2026.6.19 adds support for **Cilium NetworkPolicy** in a generic, database-agnostic way. Two network policy flavors are now supported: `kubernetes` and `cilium`. The feature is disabled by default and can be enabled at install/upgrade time:

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version v2026.6.19 \
  --namespace kubedb --create-namespace \
  --set global.networkPolicy.enabled=true \
  --set global.networkPolicy.flavor=cilium \
  --set-file global.license=license.txt \
  --wait --burst-limit=10000 --debug
```

For full configuration reference, see the [network policy docs](https://kubedb.com/docs/v2026.6.19/setup/install/kubedb/configuration/#network-policy).

---

## MySQL

### MySQL Migration Tool

KubeDB now ships a **MySQL migration tool** that enables live migration from any source environment — **AWS RDS**, **Azure Database for MySQL**, **Google Cloud SQL**, or **self-hosted** — to KubeDB-managed MySQL clusters with minimal downtime.

The migration operates in three phases:

1. **Schema Migration** — transfers the complete database schema (tables, indexes, constraints)
2. **Snapshot Migration** — bulk initial data transfer with configurable parallelism
3. **CDC Streaming** — continuous replication via Change Data Capture until cutover

**Prerequisite:** source MySQL must have binary logging enabled with `binlog_format=ROW` and `binlog_row_image=FULL`. The migration user must have `REPLICATION SLAVE` and `REPLICATION CLIENT` privileges.

First deploy a KubeDB-managed MySQL (see the [quickstart guide](https://kubedb.com/docs/v2026.6.19/guides/mysql/quickstart/quickstart/)). Then create an `AppBinding` for the source and apply the following `Migrator` CR:

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: mysql-migrate
  namespace: demo
spec:
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
  source:
    mysql:
      connectionInfo:
        appBinding:
          name: source-mysql
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
      schema:
        enabled: true
        database: []
        excludeDatabase: []
      snapshot:
        enabled: true
        pipeline:
          workers: 3
          sinkers: 4
          buffer: 12
          write_batch_size: 200
          read_batch_size: 1000
      streaming:
        enabled: true
  target:
    mysql:
      connectionInfo:
        appBinding:
          name: target-mysql
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
```

Monitor the `LAG` field with `kubectl get migrator -n demo`. Once LAG reaches zero, stop writes to the source and delete the `Migrator` CR to complete the cutover.

> See the [MySQL Migration Guide](https://kubedb.com/docs/v2026.6.19/guides/mysql/migration/databaseMigration/) for TLS setup, platform-specific binary log configuration, and AppBinding creation.

### Virtual Secrets

MySQL now supports the **Virtual Secrets** feature, allowing users to keep database auth secrets outside the cluster (e.g., Vault, AWS Secrets Manager, Azure Key Vault) while still using them with KubeDB-managed databases. See the [Virtual Secrets blog post](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/) for details.

---

## MariaDB

### MariaDB Migration Tool

KubeDB now ships a **MariaDB migration tool** that enables live migration from **AWS RDS**, **Azure Database for MariaDB**, or **self-hosted** MariaDB instances to KubeDB-managed clusters with minimal downtime.

The tool operates in the same three-phase model as the MySQL migration tool: Schema Migration → Snapshot Migration → CDC Streaming.

**Prerequisite:** source MariaDB must have `binlog_format=ROW` and `binlog_row_image=FULL`, with the migration user having `REPLICATION SLAVE` and `REPLICATION CLIENT` privileges.

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: mariadb-migrate
  namespace: demo
spec:
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
  source:
    mariadb:
      connectionInfo:
        appBinding:
          name: source-mariadb
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
      schema:
        enabled: true
        database: []
        excludeDatabase: []
      snapshot:
        enabled: true
        pipeline:
          workers: 3
          sinkers: 4
          buffer: 12
          write_batch_size: 200
          read_batch_size: 1000
      streaming:
        enabled: true
  target:
    mariadb:
      connectionInfo:
        appBinding:
          name: target-mariadb
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
```

> See the [MariaDB Migration Guide](https://kubedb.com/docs/v2026.6.19/guides/mariadb/migration/databaseMigration/) for full details.

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: sample
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 30m
```
---

## MongoDB

### MongoDB Migration Tool

KubeDB now ships a **MongoDB migration tool** built on **mongoshake** that enables live migration from **MongoDB Atlas**, **AWS DocumentDB**, **Azure Cosmos DB (MongoDB API)**, or **self-hosted** MongoDB instances with minimal downtime.

The migration operates in two phases:

1. **Full Sync** — initial bulk transfer of all collections and documents (upsert semantics for re-sync safety)
2. **Incremental Sync (Oplog Tailing)** — continuous replication from the source replica set oplog until cutover

**Prerequisite:** source MongoDB must be a **replica set with oplog enabled**; the migration user needs **read privileges** on all databases; the source must be **network-reachable** from within the cluster.

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: mongodb-migrate
  namespace: demo
spec:
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
  source:
    mongodb:
      connectionInfo:
        appBinding:
          name: source-mongodb
          namespace: demo
      mongoshake:
        syncMode: all
        extraConfiguration:
          full_sync.executor.insert_on_dup_update: "true"
  target:
    mongodb:
      connectionInfo:
        appBinding:
          name: target-mongodb
          namespace: demo
```

> See the [MongoDB Migration Guide](https://kubedb.com/docs/v2026.6.19/guides/mongodb/migration/databaseMigration/) for TLS and platform-specific setup.

### Virtual Secrets

MongoDB now supports the **Virtual Secrets** feature, allowing users to keep database auth secrets outside the cluster (e.g., Vault, AWS Secrets Manager, Azure Key Vault) while still using them with KubeDB-managed databases. See the [Virtual Secrets blog post](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/) for details.

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: sample
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 30m
```

---

## DocumentDB

This release introduces three major capabilities for DocumentDB: **high-availability clustering**, a comprehensive **OpsRequest** suite, and **autoscaling**.

### High-Availability Clustering

DocumentDB now supports multi-replica HA clustering built on PostgreSQL streaming replication with a Raft-based coordinator.

**Key Highlights:**
- Multi-replica HA via PostgreSQL streaming replication
- Automatic leader election and failover (Raft-based coordinator)
- Writes on the primary replicated to all standbys; standbys are read-only
- Automatic standby re-sync (`pg_rewind`) and rejoin after failover
- No data loss across pod failure, failover, and restart scenarios
- MongoDB wire-protocol access (`mongosh`) on port `10260`

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: dcdb
  namespace: demo
spec:
  version: 'pg17-0.110.0'
  storageType: Durable
  deletionPolicy: Halt
  replicas: 3
  podTemplate:
    spec:
      containers:
        - name: documentdb
          resources:
            requests:
              cpu: 500m
              memory: 2Gi
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

Connect using any MongoDB client:

```bash
mongosh 'mongodb://<user_name>:<password>@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true'
```

### OpsRequests

KubeDB now supports the following `DocumentDBOpsRequest` types:

- **Restart** — graceful rolling restart with zero configuration changes
- **VerticalScaling** — scale CPU/memory for DocumentDB and coordinator containers; auto-regenerates `pgtune.conf` when tuning is enabled
- **HorizontalScaling** — add/remove replicas, including standalone→HA and HA→standalone transitions
- **Reconfigure** — apply custom PostgreSQL settings inline via `user.conf`
- **RotateAuth** — operator-generated or external-secret credentials rotation
- **VolumeExpansion** — expand PVC storage (Online or Offline modes)
- **StorageMigration** — migrate to a different StorageClass

Example — Horizontal Scale Up:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: dcdb-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: dcdb
  horizontalScaling:
    replicas: 5
```

### Autoscaling

The new `DocumentDBAutoscaler` CRD supports both **compute** and **storage** autoscaling:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DocumentDBAutoscaler
metadata:
  name: dcdb-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: dcdb
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    documentdb:
      trigger: "On"
      podLifeTimeThreshold: 1m
      resourceDiffPercentage: 5
      minAllowed:
        cpu: 600m
        memory: 1.5Gi
      maxAllowed:
        cpu: "2"
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

---

## Neo4j

This release brings six new capabilities to Neo4j: **Online Backup & Restore**, **StorageMigration**, **Git-Sync initialization**, **GitOps support**, **Autoscaling**, and the **Recommendation Engine**.

### Online Backup & Restore (KubeStash)

Neo4j databases can now be backed up and restored online using KubeStash with no downtime, using Neo4j's native online backup tooling.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: neo4j-backup-config
  namespace: demo
spec:
  backends:
  - name: s3-backend
    retentionPolicy:
      name: demo-retention
      namespace: demo
    storageRef:
      name: s3-storage
      namespace: demo
  sessions:
  - addon:
      name: neo4j-addon
      tasks:
      - name: logical-backup
    name: frequent-backup
    repositories:
    - backend: s3-backend
      directory: /backup
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      name: s3-neo4j-repo
    scheduler:
      jobTemplate:
        backoffLimit: 1
      schedule: '*/5 * * * *'
    sessionHistoryLimit: 1
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    name: neo4j-backup
    namespace: demo
```

### StorageMigration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: neo4j-test
  migration:
    storageClassName: custom-longhorn
    oldPVReclaimPolicy: Delete
  timeout: 3000s
```

### Initialization via Git-Sync

Neo4j can now be initialized with Cypher scripts pulled from a public or private Git repository:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-git
  namespace: demo
spec:
  replicas: 3
  deletionPolicy: WipeOut
  version: "2025.11.2"
  init:
    script:
      scriptPath: "current"
      git:
        args:
        - --repo=git@github.com:fr-sarker/neo4j-init-script.git
        - --link=current
        - --root=/root
        - --one-time
        authSecret:
          name: git-creds
        securityContext:
          runAsUser: 65533
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 3Gi
```

### GitOps Support

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Neo4j
metadata:
  name: neo4j-dev
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 4
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2.5Gi
  deletionPolicy: WipeOut
```

### Autoscaling

A new `Neo4jAutoscaler` CRD adds compute and storage autoscaling for Neo4j server nodes:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: Neo4jAutoscaler
metadata:
  name: neo4j-as
  namespace: demo
spec:
  databaseRef:
    name: neo4j
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    neo4j:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 500m
        memory: 2Gi
      maxAllowed:
        cpu: 1
        memory: 4Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
  storage:
    neo4j:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

### Recommendation Engine

Neo4j is now supported by the KubeDB Recommendation Engine. The Ops-manager watches for version updates, TLS certificate expiry, and auth rotation, creating `Recommendation` CRs whenever maintenance is due.

---

## Oracle

### OpsRequests

This release adds full Oracle ops coverage:

- **Restart** — controlled restart of Oracle instances
- **Reconfigure** — apply configuration via secret or inline
- **VerticalScaling** — update CPU/memory for Oracle nodes
- **RotateAuth** — rotate Oracle authentication credentials
- **VolumeExpansion** — expand Oracle storage volumes (Online/Offline)

Example — Reconfigure:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: reconfigure-test
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sa-test
  configuration:
    configSecret:
      name: oracle-custom-config
```

Example — VolumeExpansion:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: oracle-volume-more
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: long-horn-sa-test
  volumeExpansion:
    mode: "Offline"
    node: 12Gi
```

### Oracle Autoscaler

Both compute and storage autoscaling are now supported via `OracleAutoscaler`:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: OracleAutoscaler
metadata:
  name: oracle-as-compute
  namespace: demo
spec:
  databaseRef:
    name: sa-test
  opsRequestOptions:
    timeout: 15m
    apply: IfReady
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 10
      minAllowed:
        cpu: "2500m"
        memory: 9Gi
      maxAllowed:
        cpu: "4"
        memory: 11Gi
      containerControlledValues: "RequestsAndLimits"
      controlledResources: ["cpu", "memory"]
```

---

## Weaviate

### OpsRequest Support

This release introduces a full suite of `WeaviateOpsRequest` types:

- **Restart** — controlled rolling restart
- **Reconfigure** — apply configuration changes via secret or inline
- **HorizontalScaling** — scale Weaviate nodes up or down
- **VerticalScaling** — adjust CPU/memory for Weaviate containers
- **VolumeExpansion** — expand PVCs (Online or Offline mode)
- **ReconfigureTLS** — add, update, rotate, or remove TLS (cert-manager backed)
- **RotateAuth** — rotate API key authentication credentials
- **StorageMigration** — migrate to a different StorageClass

Example — HorizontalScaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: weaviate-sample
  horizontalScaling:
    node: 5
```

### TLS Support

KubeDB now supports TLS for Weaviate, enabling HTTPS for REST client communication. When `clientAuth: true` is set, Weaviate enforces mTLS. Certificates are provisioned and managed by cert-manager:

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
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: weaviate-issuer
    clientAuth: true
  deletionPolicy: WipeOut
```

### Autoscaler

The new `WeaviateAutoscaler` CRD provides storage autoscaling:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: WeaviateAutoscaler
metadata:
  name: weaviate-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: weaviate-sample
  storage:
    weaviate:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 50
      expansionMode: "Online"
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

---

## Milvus

### TLS Support

KubeDB now supports TLS for Milvus with independent configuration for external and internal traffic:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: demo
spec:
  version: "2.6.9"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
  tls:
    issuerRef:
      name: milvus-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    external:
      mode: mTLS
    internal:
      mode: TLS
```

### OpsRequest Support

A full suite of `MilvusOpsRequest` types is now supported: HorizontalScaling, Reconfigure, Restart, RotateAuth, UpdateVersion, VerticalScaling, VolumeExpansion, ReconfigureTLS, and StorageMigration.

Example — Horizontal Scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-hscale-up-querynode
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: milvus-cluster
  horizontalScaling:
    topology:
      proxy: 2
      streamingnode: 2
```

### Autoscaler

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MilvusAutoscaler
metadata:
  name: milvus-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: milvus-cluster
  storage:
    streamingnode:
      trigger: "On"
      usageThreshold: 22
      expansionMode: "Offline"
      scalingRules:
        - appliesUpto: "100Ti"
          threshold: "50%"
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

### Recommendation Engine

Milvus is now supported by the KubeDB Recommendation Engine for version updates, TLS certificate expiry, and auth rotation.

---

## HanaDB

KubeDB now supports TLS and a comprehensive set of day-2 operations for SAP HANA.

### TLS Support

Enable TLS via `ReconfigureTLS` OpsRequest using a cert-manager Issuer:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hana-sr-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hana-sr
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: hdb-ca-issuer
  timeout: 30m
  apply: IfReady
```

### OpsRequests

The following `HanaDBOpsRequest` types are now supported:

- **VerticalScaling** — scale CPU and memory for HanaDB containers
- **Restart** — rolling restart with no configuration changes
- **ReconfigureTLS** — add, rotate, or remove TLS certificates
- **RotateAuth** — rotate SYSTEM password (operator-generated or user-defined)
- **VolumeExpansion** — expand storage (StorageClass must support expansion)
- **StorageMigration** — migrate PVCs to a new StorageClass

Example — VerticalScaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hana-sr-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: hana-sr
  verticalScaling:
    hanadb:
      resources:
        requests:
          cpu: "2100m"
          memory: "8448Mi"
        limits:
          cpu: "4"
          memory: "14Gi"
  timeout: 30m
  apply: IfReady
```

---

## Elasticsearch

### Git-Sync Initialization

Elasticsearch can now be initialized with scripts pulled from a public or private Git repository:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-git
  namespace: demo
spec:
  init:
    script:
      scriptPath: "current"
      git:
        args:
          - --repo=https://github.com/appscode/elasticsearch-init-scripts.git
          - --depth=1
          - --link=current
          - --root=/git
          - --one-time
  version: xpack-8.18.8
  enableSSL: true
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: es-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: es
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
```

For clusters in topology mode (master/data/ingest), the target StorageClass can be set per node type:

```yaml
spec:
  type: StorageMigration
  databaseRef:
    name: es
  migration:
    topology:
      data:
        storageClassName: longhorn-custom
      master:
        storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
```

---

## ClickHouse

### Git-Sync Initialization

ClickHouse can now be initialized from a public or private Git repository.

**From a public repository:**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: ch-standalone-init-script
  namespace: demo
spec:
  version: 25.7.1
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      scriptPath: "current"
      git:
        args:
          - --repo=https://github.com/ShuvoKumarMondal/clickhouse-init-script.git
          - --link=current
          - --root=/root
          - --one-time
  deletionPolicy: WipeOut
```

**From a private repository**, first create a secret with your SSH keys:

```bash
ssh-keyscan $YOUR_GIT_HOST > /tmp/known_hosts
kubectl create secret generic -n demo git-creds \
   --from-file=ssh=$HOME/.ssh/id_rsa \
   --from-file=known_hosts=/tmp/known_hosts
```

Then reference it in the `git.authSecret` field. See the [git-sync docs](https://github.com/kubernetes/git-sync/tree/master/docs) for additional authentication mechanisms.

### GitOps Support

ClickHouse can now be provisioned and managed declaratively through the `gitops.kubedb.com` CRD group:

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: ClickHouse
metadata:
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          storageClassName: "local-path"
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
          storageClassName: "local-path"
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  deletionPolicy: WipeOut
```

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  timeout: 5h
  type: StorageMigration
  databaseRef:
    name: clickhouse-prod
  migration:
    cluster:
      storageClassName: longhorn-custom
      oldPVReclaimPolicy: Delete
    clickHouseKeeper:
      storageClassName: longhorn
      oldPVReclaimPolicy: Delete
```

---

## Solr

### Recommendation Engine

Solr is now supported by the KubeDB Recommendation Engine. Once a Solr database is `Ready`, the Ops-manager watches it for version updates, TLS certificate expiry, and auth secret rotation, creating `Recommendation` CRs whenever maintenance is due.

### GitOps Support

Solr can now be provisioned and managed declaratively through the `gitops.kubedb.com` CRD group.

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  timeout: 5h
  type: StorageMigration
  databaseRef:
    name: solr-combined
  migration:
    node:
      storageClassName: longhorn-custom
      oldPVReclaimPolicy: Delete
```

---

## SingleStore

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  timeout: 60m
  type: StorageMigration
  databaseRef:
    name: sample-sdb
  migration:
    aggregator:
      storageClassName: longhorn-custom
      oldPVReclaimPolicy: Delete
    leaf:
      storageClassName: longhorn-custom
      oldPVReclaimPolicy: Delete
```

### Git-Sync Initialization

SingleStore now supports initializing from a Git repository (public or private) using the `init.script.git` field.

### RotateAuth

A new `RotateAuth` OpsRequest is now available for SingleStore:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-sdb
```

### GitOps Support and Recommendation Engine

SingleStore can now be provisioned and managed via the `gitops.kubedb.com` CRD group. It is also now supported by the KubeDB Recommendation Engine.

---

## PerconaXtraDB

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  timeout: 30m
  type: StorageMigration
  databaseRef:
    name: sample-pc
  migration:
    storageClassName: local-path
    oldPVReclaimPolicy: Delete
```

### Git-Sync Initialization and GitOps

PerconaXtraDB now supports Git-Sync initialization and can be managed through the `gitops.kubedb.com` CRD group. It is also supported by the KubeDB Recommendation Engine.

---

## RabbitMQ

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: rabbitmq-dev
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 10m
```

### GitOps Support and Recommendation Engine

RabbitMQ can now be provisioned and managed declaratively through the `gitops.kubedb.com` CRD group. It is also now supported by the KubeDB Recommendation Engine for version updates, TLS certificate expiry, and auth secret rotation.

---

## Druid

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druid-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: druid-cluster
  migration:
    topology:
      historicals:
        storageClassName: longhorn-custom
      middleManagers:
        storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
```

### GitOps Support

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: "30.0.1"
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    historicals:
      replicas: 2
    middleManagers:
      replicas: 2
    coordinators:
      replicas: 1
    brokers:
      replicas: 1
    routers:
      replicas: 1
  deletionPolicy: WipeOut
```

---

## Qdrant

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: qdrant-sample
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 30m
```

### Autoscaler

KubeDB now supports autoscaling for Qdrant via the `QdrantAutoscaler` CRD. Both compute and storage autoscaling are supported:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: QdrantAutoscaler
metadata:
  name: qdrant-as-storage
  namespace: demo
spec:
  databaseRef:
    name: qdrant-sample
  storage:
    node:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

---

## Kafka

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: kafka-prod
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 30m
```

For topology-mode clusters, storage can be migrated independently for broker and controller nodes:

```yaml
spec:
  type: StorageMigration
  databaseRef:
    name: kafka-prod
  migration:
    broker:
      storageClassName: longhorn-broker
      oldPVReclaimPolicy: Delete
    controller:
      storageClassName: longhorn-controller
      oldPVReclaimPolicy: Delete
  timeout: 30m
```

---

## Redis

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: redis-prod
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 30m
```

Along with StorageMigration, this release adds **ACL merging during Reconfigure OpsRequests** and fixes a bug during ACL deletion.

---

## SQL Server (MSSQL)

### StorageClass Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: mssql-prod
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 30m
```

---

## Ignite

### StorageClass Migration

KubeDB now supports `StorageMigration` for Apache Ignite, enabling PVC migration from one StorageClass to another:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ignite-smigrate-storage-migration
  namespace: demo-ignite-smigrate
spec:
  type: StorageMigration
  databaseRef:
    name: ignite-smigrate
  migration:
    storageClassName: local-path-migrated
  timeout: 3600s
  apply: IfReady
```
---
## ProxySQL

### Rotate Authentication

KubeDB now supports `RotateAuth` for ProxySQL:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: rotate-px-auth
  namespace: demo
spec:
  type: RotateAuth
  proxyRef:
    name: sample
  apply: Always
```

## Support

- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
