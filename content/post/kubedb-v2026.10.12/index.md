---
title: Announcing KubeDB v2026.10.12
date: "2026-10-12"
weight: 15
authors:
- Saurov Chandra Biswas
tags:
- clickhouse
- cloud-native
- database
- documentdb
- elasticsearch
- etcd
- hanadb
- kubedb
- kubernetes
- milvus
- mssql
- mysql
- neo4j
- opensearch
- oracle
- pgbouncer
- postgres
- proxysql
- redis
- solr
- valkey
- virtual-secrets
---

KubeDB **v2026.10.12** introduces **KubeDB Branching**: safe, independent copies of your production database for testing, experimenting, and debugging, on the same cluster or on another one. Branching is available for Postgres, MySQL, MariaDB, MongoDB, Redis, and Elasticsearch.

This release also brings a brand-new **etcd** database, **Transparent Data Encryption (TDE)** and **Cross Data Center Disaster Recovery** for Postgres, **point-in-time recovery for Milvus and Neo4j**, **Oracle RMAN backup to S3**, **Solr backup and restore**, **Amazon Aurora MySQL support in ProxySQL**, **MySQL Standalone to Group Replication promotion**, and **Virtual Secrets** support across most of the KubeDB portfolio.

---

## Key Highlights

* **KubeDB Branching** — create an independent copy of a database from a point-in-time snapshot, in the same cluster or across clusters
* **etcd** — KubeDB now provisions and manages etcd, with backup and restore via KubeStash
* **Postgres TDE** — Transparent Data Encryption with Percona `pg_tde`, plus `RotatePrincipalKey` and `EnableWALEncryption` ops requests
* **Postgres Cross-DC Disaster Recovery** — controlled cross data center failover for distributed Postgres
* **Postgres Git-sync Init Scripts** — pull initialization scripts from a Git repository
* **Milvus PITR** — continuous archiving and point-in-time recovery via `MilvusArchiver`; KubeDB-managed etcd for meta storage; GPU and SR-IOV support
* **Neo4j PITR** — continuous archiving via `Neo4jArchiver`, composite database backup and restore, and Neo4j `2026.08.1`
* **Oracle Backup** — full physical RMAN backups streamed to S3 using KubeStash
* **Solr Backup & Restore** — logical backup and restore via KubeStash
* **ClickHouse Keeper Ops** — HorizontalScaling, VerticalScaling, and VolumeExpansion for internally managed ClickHouse Keeper
* **ProxySQL for Aurora MySQL** — query routing, load balancing, and health checks for Amazon Aurora MySQL
* **MySQL Promotion** — promote a Standalone MySQL to Group Replication
* **DocumentDB** — TLS, `ReconfigureTLS`, and KubeStash backup and restore
* **Redis/Valkey** — primary switchover before restart, Announce ops, and operator-managed ACL secrets
* **Virtual Secrets** — now supported by 23 more databases, and by ops requests and backup/restore plugins
* **New versions** — Elasticsearch `xpack-9.5.3`, OpenSearch `opensearch-3.8.0`, Neo4j `2026.08.1`, Milvus `3.0.0`, etcd `3.5.21` and `3.6.4`, Postgres `17.9-percona`

---

## KubeDB Branching

A **branch** is a separate database created from a point-in-time snapshot of the source database's volumes. Changes made on a branch never reach the source, and changes made on the source never reach the branch, so you can try schema changes, rehearse upgrades, or debug against production-like data without touching production.

Branching is available for **Postgres, MySQL, MariaDB, MongoDB, Redis, and Elasticsearch**, and works in two modes:

- **Same cluster:** the branch is created alongside the source.
- **Cross cluster:** the branch is created on another cluster. This needs Open Cluster Management (OCM) to connect the clusters, and both clusters must use the same storage backend.

The following example uses Postgres. First, deploy a Postgres database:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: prod-pg
  namespace: demo
spec:
  version: "16.1"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

Then create a branch of it on the same cluster, in the `dev` namespace:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: prod-pg-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
  target:
    namespace: dev
    name: prod-pg-dev
```

To create the branch on another cluster instead, set `target.clusterName` to the name of the target cluster in OCM and apply the Branch on the source cluster:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: prod-pg-remote
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
  target:
    clusterName: cluster-b # name of the target cluster in OCM
    namespace: demo
    name: prod-pg-remote
```

---

## etcd

### Introducing KubeDB etcd

KubeDB now provisions and manages etcd, a natively clustered Raft datastore. There is no separate standalone mode: `spec.replicas: 1` is a one-member cluster, and the default is 3. Members are added and removed through etcd's own membership API, one at a time. A new member joins as a learner and is promoted once it has caught up. This release adds the `3.5.21` and `3.6.4` EtcdVersions.

etcd is disabled by default. Enable it when installing the `kubedb` chart with `global.featureGates.Etcd=true`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-cluster
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

The `Etcd` resource also supports custom configuration and tuning (`spec.configuration`), monitoring, an auth secret, and halting (`spec.halted`).

### Backup and Restore

Snapshot backup and restore use KubeStash with an `EtcdArchiver` and the new `etcd-addon` (built on the new `etcd-restic-plugin`). A backup takes a point-in-time snapshot of the live cluster and stores it in a restic repository. An `EtcdArchiver` schedules a full-backup session and a manifest-backup session.

There is no continuous archiving or point-in-time recovery for etcd. `recoveryTimestamp` selects the newest full snapshot taken at or before that time.

To restore into a new database, set `spec.init.archiver`. The restore runs before the first member starts. `recoveryTimestamp` is required, and `fullDBRepository` and `manifestRepository` are optional.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-restored
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  init:
    archiver:
      recoveryTimestamp: "2026-10-10T10:00:00Z"
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: etcd-full
        namespace: demo
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

---

## Postgres

### Transparent Data Encryption (TDE) with Percona `pg_tde`

KubeDB now supports Transparent Data Encryption for Postgres using the Percona `pg_tde` extension. Table data (and optionally WAL) is encrypted at rest, and the principal key is held in an external key provider: HashiCorp Vault, a KMIP server, or a local key file. The HA coordinator uses `pg_tde_rewind` and `pg_tde_basebackup` for TDE databases, and physical backup and point-in-time recovery are supported.

TDE is available on the new `17.9-percona` PostgresVersion.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: tde-pg
  namespace: demo
spec:
  version: "17.9-percona"
  replicas: 3
  tde:
    cipher: aes_256            # aes_128 (default) or aes_256
    encryptWAL: true           # needs a Vault or KMIP provider
    defaultEncryptedTables: true
    keyProvider:
      vault:
        address: https://vault.demo.svc:8200
        mountPath: secret
        tokenSecretRef:
          name: vault-token    # Secret with key "token"
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

Notes:

- The `file` key provider is only valid for a standalone database and cannot back WAL encryption.
- Set `enforceEncryption: true` to block creating unencrypted tables.

Two new `PostgresOpsRequest` types manage TDE on a running database. `RotatePrincipalKey` rotates the principal key without rewriting data. `EnableWALEncryption` turns on WAL encryption with a rolling restart.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: rotate-principal-key
  namespace: demo
spec:
  type: RotatePrincipalKey
  databaseRef:
    name: tde-pg
  rotatePrincipalKey:
    keyName: pg-tde-key-2      # optional, derived if empty
```

### Cross Data Center Disaster Recovery (DC-DR)

A distributed Postgres can now run across data centers (OCM clusters) with a controlled cross-DC failover. Opt in with the `dr.kubedb.com/enabled: "true"` annotation on a distributed Postgres (`spec.distributed: true`).

- `status.disasterRecovery` reports the active data center, the per-DC health, replication lag and writability, and whether the database is currently protected by a healthy standby DC.
- Set `spec.replication.bestEffortCrossDCLagBytesForFailover` to cap how much un-replicated WAL an unplanned failover may lose. A surviving DC that is further behind refuses to promote and the database stays down until you accept the loss by annotating it with `dr.kubedb.com/accept-failover-data-loss=true`. This is a best-effort budget, not a data-loss guarantee.
- Scale each data center independently with a `HorizontalScaling` ops request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: scale-dc
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: dr-pg
  horizontalScaling:
    dataCenters:
      - clusterName: cluster-a
        replicas: 3
      - clusterName: cluster-b
        replicas: 3
```

- The coordinator avoids false re-seeds and write-locks a primary that the failover fence refuses.
- Remote replicas now run a coordinator. A remote replica can self-heal with `pg_rewind` or re-initialize, a multi-replica remote replica can be cascaded, and its DR metrics are served on the raft-metrics port.

### Git-sync for Initialization Scripts

Initialization scripts can now be pulled from a Git repository. A git-sync init container fetches the repository before `initdb`.

```yaml
spec:
  init:
    script:
      scriptPath: scripts
      git:
        args:
          - --repo=https://github.com/<org>/<repo>
          - --ref=main
        authSecret:
          name: git-credentials   # optional
```

### Other Improvements

- **Branching:** Postgres supports the new KubeDB Branch resource (see [KubeDB Branching](#kubedb-branching) above). A branched database waits for its cloned data volumes before provisioning.
- **Custom `pg_hba` rules:** add rules under `spec.configuration.inline["user_hba.conf"]` or as a `user_hba.conf` key in your `configSecret`. They are applied to `pg_hba.conf` in the order you write them.
- **Auth secrets:** generated passwords now include symbols, and `RotateAuth` uses the same generator as provisioning.
- **Archiver:** archiver sidekicks now inherit the database's custom configuration, license and image pull secrets.
- **Backup:** KubeStash backups can run while the database is in the `Critical` phase, because the primary is still serving. Restore still requires `Ready`.

---

## MySQL

### Standalone to Group Replication Promotion

This release adds support for promoting a MySQL Standalone to Group Replication through a `ReplicationModeTransformation` ops request.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: promote-to-gr
  namespace: demo
spec:
  type: ReplicationModeTransformation
  databaseRef:
    name: my-standalone
  replicationModeTransformation:
    targetTopologyMode: GroupReplication
    mode: Single-Primary
  timeout: 15m
  apply: Always
```

---

## ProxySQL

### Amazon Aurora MySQL Support

KubeDB now supports **Amazon Aurora MySQL** with ProxySQL, providing robust high availability, query routing, and load balancing.

- **Aurora Integration:** ProxySQL manages traffic across Aurora MySQL nodes.
- **Automatic Hostgroup Assignment:** nodes are dynamically assigned to read/write hostgroups.
- **Health Checks:** faulty nodes are automatically removed from routing with graceful failover.

Create the following three objects to connect with Aurora MySQL:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aurora-auth
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: "<password>"
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: aurora-appbinding
  namespace: demo
spec:
  type: kubedb.com/aws-aurora
  clientConfig:
    url: tcp://aurora.cluster-xxxx.us-east-2.rds.amazonaws.com:3306
    caBundle: <base64 encoded pem file>
  secret:
    name: aurora-auth
  version: "8.0.mysql_aurora.3.10.3"
```

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: aurora-proxy
  namespace: demo
spec:
  version: "3.0.1-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: aurora-appbinding
  configuration:
    init:
      inline:
        mysqlAWSAuroraHostgroups:
          max_lag_ms: 30000
          check_interval_ms: 10000
  deletionPolicy: WipeOut
```

---

## Neo4j

### Continuous Archiving and Point-in-Time Recovery

KubeDB now supports continuous archiving for Neo4j. With a `Neo4jArchiver` resource, transaction logs can be archived to remote storage so a database can be recovered to a selected point in time. This helps reduce data loss between scheduled backups and supports recovery after an operational mistake or failure.

For setup and recovery steps, see the Neo4j continuous archiving and point-in-time recovery guide in the KubeDB documentation.

### Composite Database Backup and Restore

KubeStash logical backup and restore now supports Neo4j composite databases. A composite database provides a unified query context over its constituent databases, so backing up and restoring the associated databases together helps preserve the data needed by the composite deployment. KubeStash uses Neo4j administration tools for logical backup and restore, and its backup configuration can be customized to select databases.

To learn more, see the Neo4j composite database backup and restore guide in the KubeDB documentation.

### New Neo4j Version

This release adds support for Neo4j `2026.08.1`.

---

## Elasticsearch and OpenSearch

This release adds the following database versions:

* Elasticsearch `xpack-9.5.3`
* OpenSearch `opensearch-3.8.0`

---

## Solr

### Backup and Restore

KubeDB now supports backup and restore for Solr using KubeStash. This enables reliable data protection with cloud storage backends (S3 and GCS).

**BackupConfiguration:**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: solr-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Solr
    namespace: demo
    name: solr-cluster
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: solr-frequent-backup
      sessionHistoryLimit: 10
      scheduler:
        schedule: "*/30 * * * *"
        successfulJobsHistoryLimit: 1
        failedJobsHistoryLimit: 1
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: solr-s3-repo
          backend: s3-backend
          directory: cluster
      addon:
        name: solr-addon
        tasks:
          - name: logical-backup
```

**RestoreSession:**

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: solr-restore
  namespace: demo
spec:
  target:
    name: solr-restored
    namespace: demo
    apiGroup: kubedb.com
    kind: Solr
  dataSource:
    snapshot: latest
    repository: solr-s3-repo
  addon:
    name: solr-addon
    tasks:
      - name: logical-backup-restore
```

---

## ClickHouse

### Ops Requests for Internally Managed ClickHouse Keeper

KubeDB-managed ClickHouse now supports the following ops requests for an internally managed ClickHouse Keeper.

**Horizontal Scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chkeeper-horizontal-scaling
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: clickhouse-prod
  horizontalScaling:
    clickHouseKeeper: 5
```

**Vertical Scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chkeeper-scale-vertical-cluster
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: clickhouse-prod
  verticalScaling:
    clickHouseKeeper:
      resources:
        requests:
          memory: "3Gi"
          cpu: "3"
        limits:
          memory: "3Gi"
          cpu: "3"
```

**Volume Expansion:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chkeeper-offline-volume-expansion
  namespace: demo
spec:
  apply: "IfReady"
  type: VolumeExpansion
  databaseRef:
    name: clickhouse-prod
  volumeExpansion:
    mode: "Offline"
    clickHouseKeeper: 4Gi
```

---

## Oracle

### Backup (RMAN to S3)

This release adds physical backup support for KubeDB-managed Oracle databases using KubeStash and the new `oracle-backup-plugin`.

- Take full physical backups of Oracle with RMAN, streamed directly to S3-compatible object storage through the Oracle Secure Backup Cloud Module (OSBWS / `SBT_TAPE`).
- The OSBWS wallet is configured automatically. The plugin builds the wallet config secret from the `BackupStorage` and triggers a `Reconfigure` `OracleOpsRequest` to set up the wallet on the database. Later backups reuse the wallet and refresh it only when the backup storage changes.
- Each backup covers the database, archived redo logs, the current control file and the SPFILE, all under a single RMAN tag (`KS-...`).
- Recovery metadata (RMAN tag, DBID, until-SCN, incarnation, control file and SPFILE handles) is recorded on the `Snapshot`.
- Works with Standalone and DataGuard Oracle. Use the optional `targetPod` parameter to run RMAN on a specific pod.

#### Backup Storage

Create the S3 credential secret. The keys must be `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3-secret
  namespace: demo
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: <access-key>
  AWS_SECRET_ACCESS_KEY: <secret-key>
```

Then create the `BackupStorage`:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: s3.amazonaws.com
      region: us-east-1
      bucket: oracle-backup
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: WipeOut
```

Edit these fields:

- `spec.storage.s3.endpoint`: S3 endpoint.
- `spec.storage.s3.bucket`: bucket where the RMAN backup pieces are written.
- `spec.storage.s3.secretName`: secret that holds the S3 credentials.

#### Backup Configuration

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: oracle-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Oracle
    name: sa-test
    namespace: demo
  backends:
    - name: s3-backend
      storageRef:
        name: s3-storage
        namespace: demo
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/30 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: oracle-s3-repo
          backend: s3-backend
          directory: /oracle
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: oracle-addon
        tasks:
          - name: physical-backup
```

Edit these fields:

- `spec.target.name`: target Oracle database name.
- `spec.sessions[].scheduler.schedule`: cron schedule for the backup.
- `spec.sessions[].repositories`: **use exactly one repository per session**. Multi-repository backup sessions are not supported.
- `spec.sessions[].addon.tasks[].params.targetPod`: (optional) Oracle pod to run RMAN on.

#### Limitations

- Only S3-compatible `BackupStorage` is supported.
- The S3 `prefix` is not applied to RMAN data. The SBT module writes backup pieces at the root of the bucket, and the prefix applies only to KubeStash metadata. Don't use the prefix to scope bucket lifecycle rules or manual cleanup.
- Only full backups are supported in this release.

---

## Milvus

### KubeDB etcd for Meta Storage

Milvus's meta etcd is now created and managed by KubeDB's own etcd operator (the new `Etcd` resource) instead of the upstream etcd-operator. `spec.metaStorage.authSecret` is passed through to the generated `Etcd`.

### External etcd with Auth and TLS

For an etcd you manage yourself (`metaStorage.externallyManaged: true`), you can now set:

- `authSecret`: a basic-auth Secret with `username` and `password`. Omit it if the etcd has authentication disabled.
- `clientTLSSecret`: a Secret with `tls.crt`, `tls.key` and `ca.crt`.
- `rootPath`: the etcd key prefix. It defaults to `<milvus name>-etcd`. Set it to the Helm default `by-dev` to adopt an existing deployment's metadata.

```yaml
spec:
  metaStorage:
    externallyManaged: true
    rootPath: by-dev
    endpoints:
      - https://etcd.helm-ns.svc:2379
    authSecret:
      name: etcd-auth          # username, password
    clientTLSSecret:
      name: etcd-client        # tls.crt, tls.key, ca.crt
```

### GPU and SR-IOV

Each Milvus component can request GPUs with `gpu` (`resourceName`, `count`, `nodeSelector`, `tolerations`). SR-IOV networking through Multus is set with `network.sriov` (`attachmentRef`, `resourceName`, optional `rdmaResourceName`, `interface`, and `gds` for a second GPU Direct Storage network). A Distributed role can be split into named `groups`, each with its own replicas, pod template, GPU and network settings, for example two GPU classes of QueryNode. `HorizontalScaling` and `VerticalScaling` ops requests accept `groups`.

### Point-in-Time Recovery (PITR)

KubeDB can now continuously archive a Milvus database and recover it to a point in time. A `MilvusArchiver` combines three things:

- **Full backups:** a physical backup on a schedule, which becomes the base of a recovery. The `fullBackup.driver` is `Restic` or `VolumeSnapshotter` (the default).
- **Continuous log backup:** a `milvus-archiver` sidekick that runs next to the database and archives its write-ahead log.
- **Manifest backup:** the KubeDB manifests of the database.

Create the archiver, then label your Milvus database so the archiver's selector matches it. The operator links the two by setting `spec.archiver.ref` on the database. You can also set `spec.archiver.ref` yourself to pin a specific archiver.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MilvusArchiver
metadata:
  name: milvus-archiver
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
        archiver: restic
  retentionPolicy:
    name: milvus-retention
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
    subDir: milvus
  deletionPolicy: WipeOut
  fullBackup:
    driver: Restic
    scheduler:
      schedule: "0 */6 * * *"
      successfulJobsHistoryLimit: 2
      failedJobsHistoryLimit: 2
    sessionHistoryLimit: 3
  logBackup:
    retentionPeriod: 2d
    retentionSchedule: "0 */6 * * *"
    successfulLogHistoryLimit: 5
    failedLogHistoryLimit: 5
  manifestBackup:
    scheduler:
      schedule: "0 */6 * * *"
    sessionHistoryLimit: 2
```

The database opts in with the label from the archiver's selector:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-pitr
  namespace: demo
  labels:
    archiver: restic
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-storage
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

To recover, create a new Milvus with `spec.init.archiver`. KubeDB selects the newest successful full backup taken at or before `recoveryTimestamp`, restores it, and starts the database only after the metadata and objects are restored.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-restored
  namespace: demo
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-restore-storage   # an empty bucket, with the same rootPath as the source
  init:
    archiver:
      recoveryTimestamp: "2026-09-30T16:22:29Z"
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: milvus-pitr-full
        namespace: demo
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

`recoveryTimestamp`, `fullDBRepository` and `encryptionSecret` are all required.

Limitations:

- The archiver is not supported on Milvus 3.x yet. Use a 2.6.x version.
- The meta etcd must be KubeDB-managed and dedicated to this Milvus, so `metaStorage.externallyManaged: true` is rejected.
- A Standalone Milvus created by an older operator, which keeps its write-ahead log in RocksMQ, cannot be archived. Migrate it with a logical backup restored into a new Milvus.

### Other Improvements

- Milvus `3.0.0` is added to the catalog.
- `UpdateVersion` now restarts pods one at a time, in a fixed role order, and waits for each pod to be Ready before restarting the next.
- `spec.restrictedSCC: true` supports the OpenShift `restricted-v2` SCC. Set it when you create the object.
- The operator runs when the KubeStash and Sidekick CRDs are not installed.
- `Restart`, `StorageMigration` and `VolumeExpansion` ops requests now fail once their pod or PetSet wait retries are exhausted.

---

## DocumentDB

### TLS and ReconfigureTLS

DocumentDB now supports TLS through cert-manager.

- `spec.tls.dbTLS` secures the Postgres backend.
- `spec.tls.gatewayTLS` secures the MongoDB-wire gateway, and falls back to `dbTLS` if unset.
- `spec.tls.gatewayMutualTLSEnabled` controls gateway mutual TLS, which is on by default.
- `spec.sslMode` sets the Postgres `sslmode` (`disable`, `allow`, `prefer`, `require`, `verify-ca` or `verify-full`).
- `spec.clientAuthMode` is `scram` (default) or `cert`.

A `ReconfigureTLS` ops request can add, update, rotate or remove TLS.

### Backup and Restore

KubeStash backup and restore now work for DocumentDB with `documentdb-addon`:

- **Backup tasks:** `logical-backup` (`pg_dumpall` by default), `physical-backup`, `volume-snapshot` (new `documentdb-csi-snapshotter-plugin`) and `manifest-backup`.
- **Restore tasks:** `logical-backup-restore`, `physical-backup-restore` and `manifest-restore`.

The operator publishes a `<db>-admin` AppBinding that the backup plugins use to connect to the backend Postgres. It carries the server CA, and the client certificate when `clientAuthMode` is `cert`.

### Other Improvements

- When `spec.configuration.tuning` is set, a `Reconfigure` ops request can change it (`profile`, `maxConnections`, `storageType`, `disableAutoTune`), and a `VerticalScaling` ops request now updates the tuning configuration for the new resources.
- Ops requests such as `RotateAuth` now work on a standalone (single-replica) DocumentDB. The primary is detected correctly when there is no replication.

---

## HanaDB

- **HorizontalScaling** ops for System Replication clusters (replicas ≥ 2). Scale-up adds raft nodes and registers SR secondaries. Scale-down unregisters sites and deletes their PVCs. The arbiter is added or removed as the count flips between odd and even.
- `spec.enforceVolumePermission` is now honored. It adds an init container that fixes volume ownership.
- **Breaking:** the standby Service is renamed from `<db>-secondary` to `<db>-standby`. The `serviceTemplates` alias `secondary` is now `standby`, and server-cert DNS names change.

---

## MSSQL

- **Breaking:** the secondary Service is renamed from `<db>-secondary` to `<db>-standby`.
- **Git-sync init container:** the catalog now has `gitSyncer` (git-sync `v4.7.0`) in all six MSSQLServerVersions.

---

## Redis/Valkey

### Switchover Primary Before Restart

There are no new fields. Existing ops now switch the primary first on Sentinel-mode Redis (≥ 6.2, 2+ replicas), for `Restart`, rolling `VerticalScaling`, `UpdateVersion`, `Reconfigure` and `ReconfigureTLS`. For example:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: rd-demo
  namespace: demo
spec:
  version: 8.2.2
  mode: Sentinel
  replicas: 3
  sentinelRef:
    name: sen-demo
    namespace: demo
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    accessModes:
      - ReadWriteOnce
  deletionPolicy: Halt
---
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-restart
  namespace: demo
spec:
  type: Restart        # the old primary is demoted before its pod is evicted
  databaseRef:
    name: rd-demo
```

### Announce Ops: ip → ip (Cluster Mode)

The database must already use `announce.type: ip`. A `hostname → ip` change now fails with "not supported". Endpoints use `host:port@busport`, and each shard needs one endpoint per replica.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-announce-ip
  namespace: demo
spec:
  type: Announce
  databaseRef:
    name: redis-announce
  announce:
    type: ip
    shards:
      - endpoints: ["10.10.0.11:6379@16379", "10.10.0.12:6379@16379"]
      - endpoints: ["10.10.0.21:6379@16379", "10.10.0.22:6379@16379"]
      - endpoints: ["10.10.0.31:6379@16379", "10.10.0.32:6379@16379"]
```

### Operator-Managed ACL Secret

The operator seeds `<db>-acl-<last 6 of UID>` from `configuration.acl.secretRef`. ACL ops merge into that secret instead of editing yours.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-instance
  namespace: demo
spec:
  version: 8.2.2
  mode: Cluster
  cluster: {shards: 3, replicas: 2}
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    accessModes: [ReadWriteOnce]
  configuration:
    acl:
      secretRef:
        name: old-acl-secret          # holds k1, k2 ...
      rules:
        - userName1 ${k1} allkeys +@string +@set -SADD
        - userName2 ${k2} allkeys +@string +@set -SADD
  deletionPolicy: WipeOut
---
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-acl-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: redis-instance
  configuration:
    auth:
      syncACL:
        - userName1 ${k1} +get ~mykeys:*
        - userName10 ${k10} +get ~mykeys:*
      deleteUsers:
        - userName2
      secretRef:
        name: new-acl-secret
```

---

## PgBouncer

### RotateAuth with Virtual Secrets

`RotateAuth` works for rotating to or from a virtual secret, and also for reload-config and the TLS common-name lookup. The operator decides based on `apiGroup == virtual-secrets.dev`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pb-rotate-to-virtual
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: pgbouncer-server
  authentication:
    secretRef:
      apiGroup: virtual-secrets.dev
      kind: secret
      name: new-virtual-secret
  timeout: 5m
  apply: IfReady
```

---

## Virtual Secrets

Virtual Secrets store database auth secret data in an external secret manager while behaving like native Kubernetes secrets for users. Support started with PostgreSQL and was later extended to PgBouncer, Pgpool, Redis, MongoDB and MySQL, and now covers most of the KubeDB database portfolio.

In this release, Virtual Secrets support is added for **Cassandra, ClickHouse, DB2, Druid, Elasticsearch, HanaDB, Hazelcast, Ignite, Kafka, MariaDB, Memcached, Milvus, MSSQLServer, Neo4j, Oracle, Percona XtraDB, ProxySQL, Qdrant, RabbitMQ, Singlestore, Solr, Weaviate and ZooKeeper**.

- **Ops requests:** Virtual Secrets now work with ops requests for PostgreSQL, PgBouncer, Pgpool, Redis and MongoDB.
- **Backup and restore:** the backup and restore plugins can now read Virtual Secrets.
- **CRD detection:** operators now watch Virtual Secrets only when the Virtual Secrets CRD is installed in the cluster.

Learn more about Virtual Secrets [here](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/), and follow the step-by-step guide to use them with KubeDB [here](https://appscode.com/blog/post/virtual-secrets-v2025.3.14/#use-virtual-secrets-with-kubedb).

---

## Support

- **Contact Us**: Reach out via [our website](https://appscode.com/contact/).
- **Release Updates**: Join our [google group](https://groups.google.com/a/appscode.com/g/releases) for release updates.
- **Stay Updated**: Follow us on [Twitter/X](https://x.com/KubeDB) for product announcements.
- **Tutorials**: Subscribe to our [YouTube channel](https://youtube.com/@appscode) for tutorials on production-grade Kubernetes tools.
- **Learn More**: Explore [Production-Grade Databases in Kubernetes](https://kubedb.com/).
- **Report Issues**: File bugs or feature requests on [GitHub](https://github.com/kubedb/project/issues/new).
